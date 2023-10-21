package account

import (
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/xuxinlai2002/creda-celo-balance/config"
	"github.com/xuxinlai2002/creda-celo-balance/db"
)

type Account struct {
	Address common.Address
	Balance *big.Int
	cfg     *config.Config
	db      *db.PostgresDB
}

func New(cfg *config.Config) (*Account, error) {
	database, err := db.NewDB(cfg.PostgresDBName, cfg.PostgresUser, cfg.PostgresPassword, cfg.PostgresHost, cfg.PostgresPort)
	if err != nil {
		return nil, err
	}
	acc := &Account{
		cfg: cfg,
		db:  database,
	}
	return acc, nil
}

func (a *Account) Start() {
	go func() {
		err := a.statisticsBalance()
		if err != nil {
			fmt.Println("statisticsBalance failed", "error", err)
		}
	}()
}

func (a *Account) statisticsBalance() error {
	layout := "2006-01-02"
	startDate, err := time.Parse(layout, a.cfg.StatisticsDateBegin)
	if err != nil {
		return errors.New("StatisticsDateBegin time format error " + err.Error())
	}
	endDate, err := time.Parse(layout, a.cfg.StatisticsDateEnd)
	if err != nil {
		return errors.New("StatisticsDateEnd time format error " + err.Error())
	}
	for i := startDate; i.Before(endDate); i = i.AddDate(0, 0, 1) {
		fmt.Println("i==", i.String(), " ", i.AddDate(0, 0, 1))
		pullTxRecords, err := a.db.ReadPullTxHistory(i)
		if err != nil {
			fmt.Println("ReadPullTxHistory failed", "error", err)
		}
		tokenRecords, err := a.db.ReadTokenTransferHistory(i)
		for _, r := range pullTxRecords {
			fmt.Println("pullTx ==r.from", r.From.String(), "to", r.To.String(), "va ", r.Value.String(), "tx", r.TxHash.String())
		}
		for _, r := range tokenRecords {
			fmt.Println("token === .from", r.From.String(), "to", r.To.String(), "va ", r.Value.String(), "tx", r.TxHash.String())
		}
	}
	fmt.Println("333333333")
	return nil
}
