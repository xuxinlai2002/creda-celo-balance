package tokens

import (
	"context"
	"errors"
	"fmt"
	"github.com/xuxinlai2002/creda-celo-balance/signal"
	"github.com/xuxinlai2002/creda-celo-balance/utils"
	"math/big"
	"sync"
	"time"

	"github.com/celo-org/celo-blockchain/common"
	_ "github.com/lib/pq"
	"github.com/xuxinlai2002/creda-celo-balance/client"
	"github.com/xuxinlai2002/creda-celo-balance/config"
	"github.com/xuxinlai2002/creda-celo-balance/db"
	ctypes "github.com/xuxinlai2002/creda-celo-balance/types"
)

type TokenService struct {
	cli      *client.Client
	cfg      *config.Config
	records  map[string][]*ctypes.TokenRecord
	database *db.PostgresDB
	wg       *sync.WaitGroup
}

func NewService(cfg *config.Config, wg *sync.WaitGroup) (*TokenService, error) {
	cli, err := client.Dial(cfg.HTTP)
	if err != nil {
		return nil, err
	}

	database, err := db.NewDB(cfg.PostgresDBName, cfg.PostgresUser, cfg.PostgresPassword, cfg.PostgresHost, cfg.PostgresPort)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("new db err: %v", err))
	}

	return &TokenService{
		cli:      cli,
		cfg:      cfg,
		records:  make(map[string][]*ctypes.TokenRecord),
		database: database,
		wg:       wg,
	}, nil
}

func (s *TokenService) Start(interceptor signal.Interceptor) error {
	s.wg.Add(1)

	go func() {
		defer s.wg.Done()
		s.processERC20Tokens(interceptor)
	}()

	return nil
}

func (s *TokenService) persistToDB(date string, records []*ctypes.TokenRecord) error {
	tableName := fmt.Sprintf("event%s", date)
	if err := s.database.CreateRecordTable(tableName); err != nil {
		return errors.New(fmt.Sprintf("token service persist db err: %v", err))
	}

	if err := s.database.InsertRecords(tableName, records); err != nil {
		return errors.New(fmt.Sprintf("token service persist db err: %v", err))
	}

	fmt.Printf("saved into table %s\n", tableName)

	return nil
}

func (s *TokenService) processERC20Tokens(interceptor signal.Interceptor) {
	tokens := ERC20Tokens
	distance := uint64(10000)
	toBlock := uint64(0)
	logTransferSig := []byte("Transfer(address,address,uint256)")

	startHeight := s.cfg.StartBlock

	progress, err := utils.GetTokenCurrentHeight()
	fmt.Println("token start height: ", progress)
	if err == nil && progress > startHeight {
		startHeight = progress + 1
	}

	for i := startHeight; i < s.cfg.EndBlock; i = toBlock + 1 {
		if i+distance < s.cfg.EndBlock {
			toBlock = i + distance
		} else {
			toBlock = s.cfg.EndBlock
		}
		fmt.Printf("pull block from %v to %v\n", i, toBlock)
		for address, tokenInfo := range tokens {
			select {
			default:
				query := s.cli.BuildQuery(address, logTransferSig, big.NewInt(0).SetUint64(i), big.NewInt(0).SetUint64(toBlock))
				logs, err := s.cli.FilterLogs(context.Background(), query)
				if err != nil {
					fmt.Printf("filter logs failed, error: %v\n", err)
				} else if len(logs) > 0 {
					fmt.Printf("addr: %v, len(logs): %v\n", address, len(logs))
					//fmt.Println("Date,CoinID,BlockNumber,Time,TxHash,From,To,Value")
					for _, vlog := range logs {
						bn := big.NewInt(0)
						bn.SetUint64(vlog.BlockNumber)
						b, err := s.cli.BlockByNumber(context.Background(), bn)
						if err != nil {
							fmt.Printf("rpc.BlockByNumber err: %v\n", err)
						} else {
							tr := &ctypes.TokenRecord{
								CoinID:      tokenInfo.CoinID,
								BlockNumber: vlog.BlockNumber,
								Timestamp:   b.Header().Time,
								TxHash:      vlog.TxHash,
								From:        common.HexToAddress(vlog.Topics[1].Hex()),
								To:          common.HexToAddress(vlog.Topics[2].Hex()),
								Value:       big.NewInt(0).SetBytes(vlog.Data),
							}
							if tr.Value.Cmp(big.NewInt(0)) <= 0 {
								continue
							}

							t := time.Unix(int64(tr.Timestamp), 0)
							date := fmt.Sprintf("%04d%02d%02d", t.Year(), int(t.Month()), t.Day())

							//fmt.Printf("-> %v,%v(%v),%v,%v,%s,%s,%s,%d\n",
							//	date, tr.CoinID, tokenInfo.Name, tr.BlockNumber, tr.Timestamp, tr.TxHash, tr.From, tr.To, tr.Value)

							if _, ok := s.records[date]; ok {
								s.records[date] = append(s.records[date], tr)
							} else {
								if len(s.records) > 0 {
									for k, v := range s.records {
										records := v
										delete(s.records, k)

										s.wg.Add(1)
										go func() {
											s.wg.Done()
											if err := s.persistToDB(k, records); err != nil {
												fmt.Printf("persist token event to db err: %v\n", err)
											}
										}()
									}
								}
								s.records[date] = []*ctypes.TokenRecord{tr}
							}
						}
					}
					utils.WriteTokenCurrentHeight(toBlock)
				}

			case <-interceptor.ShutdownChannel():
				fmt.Println("token service shutting down...")
				goto shutdown
			}
		}
	}

shutdown:
	if len(s.records) > 0 {
		for k, v := range s.records {
			if err := s.persistToDB(k, v); err != nil {
				fmt.Printf("persist token event to db err: %v\n", err)
			}

			delete(s.records, k)
		}
	}

	s.database.Close()
	fmt.Println("token service finished")
}
