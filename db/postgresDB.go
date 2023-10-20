package db

import (
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	_ "github.com/lib/pq"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/xuxinlai2002/creda-celo-balance/types"
)

type PostgresDB struct {
	db   *sql.DB
	lock sync.Mutex
}

func NewDB(dbName, user, password, host string, port uint32) (*PostgresDB, error) {
	db, err := sql.Open("postgres", fmt.Sprintf("user=%s dbname=%s sslmode=disable password=%s host=%s port=%d", user, dbName, password, host, port))
	if err != nil {
		fmt.Println("failed open databases", err)
		return nil, err
	}

	self := &PostgresDB{
		db: db,
	}
	return self, nil
}

func (p *PostgresDB) Close() error {
	return p.db.Close()
}

func (p *PostgresDB) CreateRecordTable(tableName string) error {
	p.lock.Lock()
	defer p.lock.Unlock()

	createTableSQL := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s ("+
		"id SERIAL PRIMARY KEY,"+
		"coinID INT,"+
		"blocknumber INT,"+
		"timestamp INT,"+
		"txhash VARCHAR(66),"+
		"fromAddress VARCHAR(42),"+
		"toAddress VARCHAR(42),"+
		"value TEXT"+
		");", tableName)
	_, err := p.db.Exec(createTableSQL)
	if err != nil {
		return errors.New(fmt.Sprintf("create sql table %s err: %v", tableName, err))
	}

	return nil
}

func (p *PostgresDB) InsertRecords(tableName string, records []*types.TokenRecord) error {
	p.lock.Lock()
	defer p.lock.Unlock()

	tx, err := p.db.Begin()
	if err != nil {
		return errors.New(fmt.Sprintf("db begin err: %v", err))
	}
	defer tx.Rollback()

	for _, record := range records {
		sqlInsert := fmt.Sprintf("INSERT INTO %s ("+
			"coinID,"+
			"blocknumber,"+
			"timestamp,"+
			"txhash,"+
			"fromAddress,"+
			"toAddress,"+
			"value"+
			") VALUES ($1,$2,$3,$4,$5,$6,$7)", tableName)
		stmt, err := tx.Prepare(sqlInsert)
		if err != nil {
			return errors.New(fmt.Sprintf("db prepare err: %v", err))
		}
		defer stmt.Close()

		_, err = stmt.Exec(record.CoinID, record.BlockNumber, record.Timestamp, record.TxHash.String(), record.From.String(), record.To.String(), record.Value.String())
		if err != nil {
			return errors.New(fmt.Sprintf("db stmt exec err: %v", err))
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return errors.New(fmt.Sprintf("db tx commit err: %v", err))
	}

	return nil
}

func (p *PostgresDB) ReadPullTxHistory(timestamp uint64) ([]*types.TokenRecord, error) {
	tableName := p.getPullTxTableNameByDate(timestamp)
	return p.queryTable(tableName)
}

func (p *PostgresDB) ReadTokenTransferHistory(timestamp uint64) ([]*types.TokenRecord, error) {
	tableName := p.getTokensTableNameByDate(timestamp)
	return p.queryTable(tableName)
}

func (p *PostgresDB) queryTable(tableName string) ([]*types.TokenRecord, error) {
	p.lock.Lock()
	defer p.lock.Unlock()
	query := fmt.Sprintf("SELECT coinid, blocknumber, timestamp, txhash, fromaddress, toaddress, value FROM %s", tableName)
	rows, err := p.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]*types.TokenRecord, 0)
	for rows.Next() {
		var coinID, blocknumber, timestamp, txhash, fromAddress, toAddress, value string
		if err := rows.Scan(&coinID, &blocknumber, &timestamp, &txhash, &fromAddress, &toAddress, &value); err != nil {
			return nil, err
		}
		coinid, ok := big.NewInt(0).SetString(coinID, 10)
		if !ok {
			return nil, errors.New(fmt.Sprintf("coindID is error%s", coinID))
		}
		number, ok := big.NewInt(0).SetString(blocknumber, 10)
		if !ok {
			return nil, errors.New(fmt.Sprintf("blocknumber is error%s", blocknumber))
		}
		time, ok := big.NewInt(0).SetString(timestamp, 10)
		if !ok {
			return nil, errors.New(fmt.Sprintf("timestamp is error%s", timestamp))
		}
		txID := common.HexToHash(txhash)
		from := common.HexToAddress(fromAddress)
		to := common.HexToAddress(toAddress)
		amount, ok := big.NewInt(0).SetString(value, 10)
		if !ok {
			return nil, errors.New(fmt.Sprintf("value is error%s", value))
		}

		record := &types.TokenRecord{
			CoinID:      coinid.Uint64(),
			BlockNumber: number.Uint64(),
			Timestamp:   time.Uint64(),
			TxHash:      txID,
			From:        from,
			To:          to,
			Value:       amount,
		}
		records = append(records, record)
	}
	return records, err
}

func (p *PostgresDB) getPullTxTableNameByDate(timestamp uint64) string {
	t := time.Unix(int64(timestamp), 0)
	tableName := fmt.Sprintf("tx_%04d%02d%02d", t.Year(), int(t.Month()), t.Day())
	return tableName
}

func (p *PostgresDB) getTokensTableNameByDate(timestamp uint64) string {
	t := time.Unix(int64(timestamp), 0)
	date := fmt.Sprintf("%04d%02d%02d", t.Year(), int(t.Month()), t.Day())
	tableName := fmt.Sprintf("event%v.txt", date)
	return tableName
}

func (p *PostgresDB) tableExists(tableName string) (bool, error) {
	str := fmt.Sprintf("SELECT * FROM information_schema.tables WHERE table_schema ='%s' AND table_name='%s';", "public", tableName)

	fmt.Printf("str is %s", str)

	var exists bool
	err := p.db.QueryRow(str).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
