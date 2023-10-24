package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
)

var DefaultConfigFilename = "config.json"

type Config struct {
	HTTP       string `json:"http,omitempty"`
	StartBlock uint64 `json:"startBlock,omitempty"`
	EndBlock   uint64 `json:"endBlock,omitempty"`

	PostgresDBName   string `json:"postgresDBName,omitempty"`
	PostgresHost     string `json:"postgresHost,omitempty"`
	PostgresPort     uint32 `json:"postgresPort,omitempty"`
	PostgresUser     string `json:"postgresUser,omitempty"`
	PostgresPassword string `json:"postgresPassword,omitempty"`

	PullStartHeight uint64 `json:"pullStartHeight,omitempty"`
	PullEndHeight   uint64 `json:"pullEndHeight,omitempty"`

	StatisticsDateBegin string `json:"statisticsDateBegin,omitempty"`
	StatisticsDateEnd   string `json:"statisticsDateEnd,omitempty"`

	CoinHistoryPrice string `json:"coinPriceHistory,omitempty"`
}

func DefaultConfig() Config {
	return Config{
		HTTP:             "https://solitary-responsive-putty.celo-mainnet.quiknode.pro/40a3938f2f03f6ae973996eccf6106a9ab27c418",
		StartBlock:       0,
		EndBlock:         0,
		PostgresDBName:   "",
		PostgresHost:     "",
		PostgresPort:     5432,
		PostgresUser:     "",
		PostgresPassword: "",

		PullStartHeight: 0,
		PullEndHeight:   0,

		StatisticsDateBegin: "",
		StatisticsDateEnd:   "",
	}
}

func LoadConfig() (*Config, error) {
	preCfg := DefaultConfig()

	file, err := ioutil.ReadFile(DefaultConfigFilename)
	if err != nil {
		return nil, err
	}
	file = bytes.TrimPrefix(file, []byte("\xef\xbb\xbf"))
	err = json.Unmarshal(file, &preCfg)
	if err != nil {
		return nil, err
	}

	if err := preCfg.ValidateConfig(); err != nil {
		return nil, err
	}
	return &preCfg, nil
}

func (cfg *Config) ValidateConfig() error {
	if cfg.PullStartHeight > cfg.PullEndHeight {
		return errors.New("pull start height is smaller to pull end height")
	}
	if cfg.PullEndHeight == 0 {
		return errors.New("PullEndHeight is empty")
	}
	if cfg.PostgresDBName == "" {
		return errors.New("PostgresDBName is empty")
	}
	if cfg.PostgresHost == "" {
		return errors.New("PostgresHost is empty")
	}
	if cfg.PostgresPort == 0 {
		return errors.New("PostgresPort is 0")
	}

	if cfg.CoinHistoryPrice == "" {
		return errors.New("CoinHistoryPrice is empty")
	}
	return nil
}
