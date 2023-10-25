package tokens

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/celo-org/celo-blockchain/common"
	_ "github.com/lib/pq"
	"github.com/xuxinlai2002/creda-celo-balance/client"
	"github.com/xuxinlai2002/creda-celo-balance/config"
	"github.com/xuxinlai2002/creda-celo-balance/db"
	"github.com/xuxinlai2002/creda-celo-balance/signal"
	ctypes "github.com/xuxinlai2002/creda-celo-balance/types"
	"github.com/xuxinlai2002/creda-celo-balance/utils"
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

	log.Infof("saved into table %s", tableName)

	return nil
}

func (s *TokenService) processERC20Tokens(interceptor signal.Interceptor) {
	tokens := ERC20Tokens
	distance := uint64(10000)
	toBlock := uint64(0)
	logTransferSig := []byte("Transfer(address,address,uint256)")

	startHeight := s.cfg.StartBlock

	progress, err := utils.GetTokenCurrentHeight()
	log.Infof("token start height: %v", progress)
	if err == nil && progress > startHeight {
		startHeight = progress + 1
	}

	for i := startHeight; i < s.cfg.EndBlock; i = toBlock + 1 {
		if i+distance < s.cfg.EndBlock {
			toBlock = i + distance
		} else {
			toBlock = s.cfg.EndBlock
		}
		log.Infof("pull block from %v to %v", i, toBlock)
		for address, tokenInfo := range tokens {
			select {
			default:
				query := s.cli.BuildQuery(address, logTransferSig, big.NewInt(0).SetUint64(i), big.NewInt(0).SetUint64(toBlock))
				logs, err := s.cli.FilterLogs(context.Background(), query)
				if err != nil {
					log.Errorf("filter logs failed, error: %v", err)
				} else if len(logs) > 0 {
					log.Infof("addr: %v, len(logs): %v", address, len(logs))
					for _, vlog := range logs {
						bn := big.NewInt(0)
						bn.SetUint64(vlog.BlockNumber)
						b, err := s.cli.BlockByNumber(context.Background(), bn)
						if err != nil {
							log.Errorf("rpc.BlockByNumber err: %v", err)
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
												log.Errorf("persist token event to db err: %v", err)
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
				log.Infof("token service shutting down...")
				goto shutdown
			}
		}
	}

shutdown:
	if len(s.records) > 0 {
		for k, v := range s.records {
			if err := s.persistToDB(k, v); err != nil {
				log.Errorf("persist token event to db err: %v", err)
			}

			delete(s.records, k)
		}
	}

	s.database.Close()
	log.Infof("token service finished")
}
