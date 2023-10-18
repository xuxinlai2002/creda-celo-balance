package tokens

import (
	"bufio"
	"context"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/celo-org/celo-blockchain/common"
	ecommon "github.com/ethereum/go-ethereum/common"
	"github.com/xuxinlai2002/creda-celo-balance/client"
	"github.com/xuxinlai2002/creda-celo-balance/config"
)

var rpcClient *client.Client

type LogTransfer struct {
	From   ecommon.Address
	To     ecommon.Address
	Tokens *big.Int
}

type TokenRecord struct {
	CoinID      uint64
	BlockNumber uint64
	Timestamp   uint64
	TxHash      common.Hash
	From        ecommon.Address
	To          ecommon.Address
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

func saveToFile(date string, data []TokenRecord) error {
	filename := fmt.Sprintf("event%v_%v.txt", date, data[0].BlockNumber)
	fmt.Printf("save %v record to %v\n", len(data), filename)
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0666)
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
		fmt.Printf("pull block %v from to %v\n", i, toBlock)
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

						transferEvent.From = ecommon.HexToAddress(vlog.Topics[1].Hex())
						transferEvent.To = ecommon.HexToAddress(vlog.Topics[2].Hex())

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
									if err := saveToFile(k, v); err != nil {
										fmt.Printf("save token event to file err: %v\n", err)
									}
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
			if err := saveToFile(k, v); err != nil {
				fmt.Printf("save token event to file err: %v\n", err)
			}
			delete(tokenDaysData, k)
		}
	}
}
