package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
)

var DefaultConfigFilename = "tokens_config.json"

type Config struct {
	HTTP       string `json:"http,omitempty"`
	StartBlock uint64 `json:"startBlock,omitempty"`
	EndBlock   uint64 `json:"endBlock,omitempty"`

	PostgresDBName   string `json:"postgresDBName,omitempty"`
	PostgresTable    string `json:"postgresTable,omitempty"`
	PostgresHost     string `json:"postgresHost,omitempty"`
	PostgresPort     uint32 `json:"postgresPort,omitempty"`
	PostgresUser     string `json:"postgresUser,omitempty"`
	PostgresPassword string `json:"postgresPassword,omitempty"`

	StartDate string `json:"StartDate,omitempty"`
	EndDate   string `json:"endDate,omitempty"`
}

func DefaultConfig() Config {
	return Config{
		HTTP:             "https://solitary-responsive-putty.celo-mainnet.quiknode.pro/40a3938f2f03f6ae973996eccf6106a9ab27c418",
		StartBlock:       0,
		EndBlock:         0,
		PostgresDBName:   "",
		PostgresTable:    "",
		PostgresHost:     "",
		PostgresPort:     5432,
		PostgresUser:     "",
		PostgresPassword: "",

		StartDate: "2021-04-20",
		EndDate:   "",
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
	if cfg.StartDate == "" {
		return errors.New("StartDate is empty")
	}
	if cfg.EndDate == "" {
		return errors.New("EndDate is empty")
	}
	if cfg.PostgresDBName == "" {
		return errors.New("PostgresDBName is empty")
	}
	if cfg.PostgresTable == "" {
		return errors.New("PostgresTable is empty")
	}
	if cfg.PostgresHost == "" {
		return errors.New("PostgresHost is empty")
	}
	if cfg.PostgresPort == 0 {
		return errors.New("PostgresPort is 0")
	}
	return nil
}
