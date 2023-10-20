package tokens

import (
	"bufio"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/celo-org/celo-blockchain/common"
	_ "github.com/lib/pq"
	"github.com/xuxinlai2002/creda-celo-balance/client"
	"github.com/xuxinlai2002/creda-celo-balance/config"
)

var rpcClient *client.Client

type LogTransfer struct {
	From   common.Address
	To     common.Address
	Tokens *big.Int
}

type TokenRecord struct {
	CoinID      uint64
	BlockNumber uint64
	Timestamp   uint64
	TxHash      common.Hash
	From        common.Address
	To          common.Address
	Value       *big.Int
}

func Start(cfg *config.Config) error {
	cli, err := client.Dial(cfg.HTTP)
	if err != nil {
		return err
	}
	rpcClient = cli

	processERC20Tokens(cfg)

	return nil
}

func saveToFile(date string, data []TokenRecord, cfg *config.Config) error {
	filename := fmt.Sprintf("event%v.txt", date)
	filePath := cfg.OutputDir + filename
	fmt.Printf("save %v record to %v\n", len(data), filePath)
	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	title := "coinid,blocknumber,time,txhash,from,to,value\n"

	defer f.Close()

	writer := bufio.NewWriter(f)
	writer.WriteString(title)
	for _, d := range data {
		line := fmt.Sprintf("%d,%d,%d,%s,%s,%s,%d\n", d.CoinID, d.BlockNumber, d.Timestamp, d.TxHash, d.From, d.To, d.Value)
		writer.WriteString(line)
	}

	writer.Flush()

	return nil
}

func saveToPostgreSql(date string, data []TokenRecord) error {
	// 连接数据库
	dbName := "postgres"
	userName := "postgres"
	passwd := "12345678"
	db, err := sql.Open("postgres", fmt.Sprintf("user=%s dbname=%s sslmode=disable password=%s", userName, dbName, passwd))
	if err != nil {
		return errors.New(fmt.Sprintf("open sql err: %v", err))
	}

	defer db.Close()

	// 创建表
	tableName := fmt.Sprintf("event%s", date)
	sqlCreateTable := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s ("+
		"id SERIAL PRIMARY KEY,"+
		"coinid INT,"+
		"blocknumber INT,"+
		"timestamp INT,"+
		"txhash VARCHAR(66),"+
		"fromaddress VARCHAR(42),"+
		"toaddress VARCHAR(42),"+
		"value TEXT"+
		");", tableName)
	_, err = db.Exec(sqlCreateTable)
	if err != nil {
		return errors.New(fmt.Sprintf("create sql table %s err: %v", tableName, err))
	}
	fmt.Printf("create table: %s\n", tableName)

	// 插入数据
	tx, err := db.Begin()
	if err != nil {
		return errors.New(fmt.Sprintf("db begin err: %v", err))
	}

	for _, d := range data {
		sqlInsert := fmt.Sprintf("INSERT INTO %s ("+
			"coinid,"+
			"blocknumber,"+
			"timestamp,"+
			"txhash,"+
			"fromaddress,"+
			"toaddress,"+
			"value"+
			") VALUES ($1,$2,$3,$4,$5,$6,$7)", tableName)
		stmt, err := tx.Prepare(sqlInsert)
		if err != nil {
			return errors.New(fmt.Sprintf("db tx prepare err: %v", err))
		}
		defer stmt.Close()

		if _, err := stmt.Exec(d.CoinID, d.BlockNumber, d.Timestamp, d.TxHash.String(), d.From.String(), d.To.String(), d.Value.String()); err != nil {
			return errors.New(fmt.Sprintf("db stmt exec err: %v", err))
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return errors.New(fmt.Sprintf("db tx commit err: %v", err))
	}

	fmt.Printf("saved into table %s\n", tableName)

	return nil
}

func processERC20Tokens(cfg *config.Config) {
	tokens := ERC20Tokens
	distance := uint64(10000)
	toBlock := uint64(0)
	logTransferSig := []byte("Transfer(address,address,uint256)")

	tokenDaysData := make(map[string][]TokenRecord)

	for i := cfg.StartBlock; i < cfg.EndBlock; i = toBlock + 1 {
		if i+distance < cfg.EndBlock {
			toBlock = i + distance
		} else {
			toBlock = cfg.EndBlock
		}
		fmt.Printf("pull block from %v to %v\n", i, toBlock)
		for address, tokenInfo := range tokens {
			query := rpcClient.BuildQuery(address, logTransferSig, big.NewInt(0).SetUint64(i), big.NewInt(0).SetUint64(toBlock))
			logs, err := rpcClient.FilterLogs(context.Background(), query)
			if err != nil {
				fmt.Printf("filter logs failed, error: %v\n", err)
			} else if len(logs) > 0 {
				fmt.Printf("addr: %v, len(logs): %v\n", address, len(logs))
				//fmt.Println("Date,CoinID,BlockNumber,Time,TxHash,From,To,Value")
				for _, vlog := range logs {
					bn := big.NewInt(0)
					bn.SetUint64(vlog.BlockNumber)
					b, err := rpcClient.BlockByNumber(context.Background(), bn)
					if err != nil {
						fmt.Printf("rpc.BlockByNumber err: %v\n", err)
					} else {
						var transferEvent LogTransfer
						transferEvent.Tokens = big.NewInt(0).SetBytes(vlog.Data)
						transferEvent.From = common.HexToAddress(vlog.Topics[1].Hex())
						transferEvent.To = common.HexToAddress(vlog.Topics[2].Hex())
						if transferEvent.Tokens.Cmp(big.NewInt(0)) <= 0 {
							continue
						}
						tr := TokenRecord{
							CoinID:      tokenInfo.CoinID,
							BlockNumber: vlog.BlockNumber,
							Timestamp:   b.Header().Time,
							TxHash:      vlog.TxHash,
							From:        transferEvent.From,
							To:          transferEvent.To,
							Value:       transferEvent.Tokens,
						}
						t := time.Unix(int64(tr.Timestamp), 0)
						date := fmt.Sprintf("%04d%02d%02d", t.Year(), int(t.Month()), t.Day())

						//fmt.Printf("-> %v,%v(%v),%v,%v,%s,%s,%s,%d\n",
						//	date, tr.CoinID, tokenInfo.Name, tr.BlockNumber, tr.Timestamp, tr.TxHash, tr.From, tr.To, tr.Value)

						_, ok := tokenDaysData[date]
						if ok {
							tokenDaysData[date] = append(tokenDaysData[date], tr)
						} else {
							if len(tokenDaysData) > 0 {
								for k, v := range tokenDaysData {
									if err := saveToPostgreSql(k, v); err != nil {
										fmt.Printf("save token event to db err: %v\n", err)
									}
									//if err := saveToFile(k, v); err != nil {
									//	fmt.Printf("save token event to file err: %v\n", err)
									//}
									delete(tokenDaysData, k)
								}
							}
							tokenDaysData[date] = []TokenRecord{tr}
						}
					}
				}
			}
		}
	}

	if len(tokenDaysData) > 0 {
		for k, v := range tokenDaysData {
			if err := saveToFile(k, v, cfg); err != nil {
				fmt.Printf("save token event to file err: %v\n", err)
			}
			delete(tokenDaysData, k)
		}
	}
}
