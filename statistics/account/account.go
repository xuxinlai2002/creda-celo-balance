package account

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"math/big"
	"os"
	"sync"
	"time"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/xuxinlai2002/creda-celo-balance/client"
	"github.com/xuxinlai2002/creda-celo-balance/config"
	"github.com/xuxinlai2002/creda-celo-balance/db"
	"github.com/xuxinlai2002/creda-celo-balance/types"
)

var zeroAddress = common.ZeroAddress.String()

type Account struct {
	cfg    *config.Config
	db     *db.PostgresDB
	client *client.Client

	accounts         map[types.ADDRESS]map[types.COINID]*big.Int
	coinPriceHistory map[types.COINID]map[types.DATE]*big.Float

	wg *sync.WaitGroup
}

func New(cfg *config.Config, wg *sync.WaitGroup) (*Account, error) {
	database, err := db.NewDB(cfg.PostgresDBName, cfg.PostgresUser, cfg.PostgresPassword, cfg.PostgresHost, cfg.PostgresPort)
	if err != nil {
		return nil, err
	}
	acc := &Account{
		cfg: cfg,
		db:  database,
		wg:  wg,
	}
	err = acc.loadCoinPrice(cfg.CoinHistoryPrice)
	if err != nil {
		return nil, err
	}

	cli, err := client.Dial(cfg.HTTP)
	if err != nil {
		return nil, err
	}
	acc.client = cli
	return acc, nil
}

func (a *Account) Start() {
	go func() {
		err := a.statisticsBalance()
		if err != nil {
			fmt.Println("statisticsBalance failed", "error", err)
		}
		a.wg.Done()
	}()
}

func (a *Account) loadCoinPrice(path string) error {
	coinHistoryFile, err := os.Open(path)
	if err != nil {
		return err
	}
	defer coinHistoryFile.Close()
	a.coinPriceHistory = make(map[types.COINID]map[types.DATE]*big.Float)
	scanner := bufio.NewScanner(coinHistoryFile)
	for scanner.Scan() {
		line := scanner.Text()
		var coinid, dateStr, priceStr string
		_, err := fmt.Sscanf(line, "%s %s %s", &coinid, &dateStr, &priceStr)
		if err != nil {
			return err
		}

		cointype, ok := big.NewInt(0).SetString(coinid, 10)
		if !ok {
			return errors.New("coinID is not number " + coinid)
		}
		coinID := types.COINID(cointype.Uint64())
		_, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			return err
		}

		price, ok := new(big.Float).SetString(priceStr)
		if !ok {
			return err
		}
		if _, exists := a.coinPriceHistory[coinID]; !exists {
			a.coinPriceHistory[coinID] = make(map[types.DATE]*big.Float)
		}
		a.coinPriceHistory[coinID][types.DATE(dateStr)] = price
	}
	return nil
}

func (a *Account) statisticsBalance() error {
	layout := "2006-01-02"
	startDate, err := time.Parse(layout, a.cfg.StatisticsDateBegin)
	if err != nil {
		return errors.New("StatisticsDateBegin time format error " + err.Error())
	}
	endDate, err := time.Parse(layout, a.cfg.StatisticsDateEnd)
	endDate = endDate.AddDate(0, 0, 1)
	if err != nil {
		return errors.New("StatisticsDateEnd time format error " + err.Error())
	}
	a.accounts = make(map[types.ADDRESS]map[types.COINID]*big.Int)
	for i := startDate; i.Before(endDate); i = i.AddDate(0, 0, 1) {
		fmt.Println("read date", i.String())
		pullTxRecords, _ := a.db.ReadPullTxHistory(i)
		if err != nil {
			pullTxRecords = make([]*types.TokenRecord, 0)
		}
		tokenRecords, err := a.db.ReadTokenTransferHistory(i)
		if err != nil {
			tokenRecords = make([]*types.TokenRecord, 0)
		}
		for _, r := range pullTxRecords {
			a.calcAccountBalance(r)
		}
		for _, r := range tokenRecords {
			a.calcAccountBalance(r)
		}
		err = a.calcUSDValue(i)
		if err != nil {
			fmt.Println("calaUSD Value error", "error", err)
		}
	}
	return nil
}

func (a *Account) calcUSDValue(date time.Time) error {
	dateStr := date.Format("2006-01-02")
	tableName := "ods_balance_" + date.Format("20060102")
	err := a.db.CreateBalanceTable(tableName)
	if err != nil {
		return err
	}
	err = a.db.InsertAccountHistoryBalance(tableName, types.DATE(dateStr), a.accounts, a.coinPriceHistory)
	return err
}

func (a *Account) calcAccountBalance(record *types.TokenRecord) {
	from := types.ADDRESS(record.From.String())
	to := types.ADDRESS(record.To.String())
	coinID := types.COINID(record.CoinID)
	intValue := record.Value
	if string(from) != zeroAddress {
		if _, exists := a.accounts[from]; !exists {
			a.accounts[from] = make(map[types.COINID]*big.Int)
		}
		balance := a.accounts[from][coinID]
		if balance == nil {
			balance = big.NewInt(0)
			a.accounts[from][coinID] = balance
		}
		a.accounts[from][coinID] = balance.Sub(balance, intValue)
		if a.accounts[from][coinID].Sign() < 0 && coinID == types.CELO_COINID {
			b, err := a.client.BalanceAt(context.Background(), record.From, big.NewInt(0).SetUint64(record.BlockNumber))
			if err != nil {
				panic("balance at error " + err.Error())
			}
			a.accounts[from][coinID] = b
		}
	}

	if string(to) != zeroAddress {
		if _, exists := a.accounts[to]; !exists {
			a.accounts[to] = make(map[types.COINID]*big.Int)
		}
		balance := a.accounts[to][coinID]
		if balance == nil {
			balance = big.NewInt(0)
			a.accounts[to][coinID] = balance
		}
		a.accounts[to][coinID] = balance.Add(balance, intValue)
	}
}
