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

func NewDB(dbName, user, password string) (*PostgresDB, error) {
	db, err := sql.Open("postgres", fmt.Sprintf("user=%s dbname=%s sslmode=disable password=%s", user, dbName, password))
	if err != nil {
		fmt.Println("failed open databases", err)
		return nil, err
	}

	//var dbExists bool
	//err = db.QueryRow("SELECT COUNT(*) FROM pg_database WHERE datname = $1", dbName).Scan(&dbExists)
	//if err != nil {
	//	dbExists = false
	//}
	//if !dbExists {
	//	_, err = db.Exec("CREATE DATABASE" + dbName)
	//	if err != nil {
	//		fmt.Println("CREATE dataasee！", err)
	//		return nil, err
	//	}
	//	fmt.Println("database create suc！")
	//} else {
	//	fmt.Println("database existeds！")
	//}

	self := &PostgresDB{
		db: db,
	}
	return self, nil
}

func (p *PostgresDB) Close() error {
	return p.db.Close()
}

func (p *PostgresDB) CreatePullTxTable(tableName string) error {
	exists, err := p.tableExists(tableName)
	if err != nil {
		fmt.Println("CreatePullTxTable get table information failed", "error", err)
		return err
	} else {
		if exists {
			return nil
		} else { //coinid, blocknumber, timestamp, txhash, fromaddress, toaddress, value
			createTableSQL := fmt.Sprintf(`CREATE TABLE %s (coinID INTEGER, blocknumber BIGINT, timestamp TIMESTAMP, txhash VARCHAR(255), fromAddress VARCHAR(255), toAddress VARCHAR(255), value BIGINT)`, tableName)
			fmt.Println("createTableSQL", createTableSQL)
			_, err = p.db.Exec(createTableSQL)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *PostgresDB) CreateTokensTransferTable(timestamp uint64) error {
	tableName := p.getTokensTableNameByDate(timestamp)
	exists, err := p.tableExists(tableName)
	if err != nil {
		fmt.Println("CreateTokensTransferTable get table information failed", "error", err)
		return err
	} else {
		if exists {
			return nil
		} else { //coinid, blocknumber, timestamp, txhash, fromaddress, toaddress, value
			createTableSQL := `CREATE TABLE ` + tableName + ` (coinID INTEGER,blocknumber BIGINT,timestamp TIMESTAMP，txhash VARCHAR(255), fromAddress VARCHAR(255),  toAddress VARCHAR(255), value BIGINT);`
			_, err = p.db.Exec(createTableSQL)
			if err != nil {
				fmt.Println("CreateTokensTransferTable failed", "error", err)
			}
		}
	}
	return nil
}

func (p *PostgresDB) InsertTokenRecords(tableName string, records []*types.TokenRecord) error {
	p.lock.Lock()
	defer p.lock.Unlock()
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, record := range records {
		stmt, err := tx.Prepare("INSERT INTO " + tableName + "(coinID, blocknumber, timestamp, txhash, fromAddress, toAddress, value ) VALUES($1, $2, $3, $4, $5, $6, $7)")
		if err != nil {
			fmt.Println("InsertTokenRecords Prepare failed", "error", err)
			return err
		}
		defer stmt.Close()

		_, err = stmt.Exec(record.CoinID, record.BlockNumber, record.Timestamp, record.TxHash.String(), record.From.String(), record.To.String(), record.Value.Uint64())
		if err != nil {
			fmt.Println("InsertTokenRecords Exec failed", "error", err)
			return err
		}
	}
	// 提交事务
	if err := tx.Commit(); err != nil {
		fmt.Println("InsertTokenRecords Commit failed", "error", err)
		return err
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
