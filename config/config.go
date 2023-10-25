package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

var DefaultConfigFilename = "config.json"

type Config struct {
	DebugLevel     string `json:"debugLevel,omitempty"` // Logging level for all subsystems {trace, debug, info, warn, error, critical}
	LogDir         string `json:"logDir,omitempty"`
	MaxLogFiles    int    `json:"maxLogFiles,omitempty"`    // Maximum logfiles to keep (0 for no rotation)
	MaxLogFileSize int    `json:"maxLogFileSize,omitempty"` // Maximum logfile size in MB

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
		DebugLevel:       "Info",
		LogDir:           "",
		MaxLogFiles:      1,
		MaxLogFileSize:   100,
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
	cfg.LogDir = CleanAndExpandPath(cfg.LogDir)

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

// CleanAndExpandPath expands environment variables and leading ~ in the
// passed path, cleans the result, and returns it.
// This function is taken from https://github.com/btcsuite/btcd
func CleanAndExpandPath(path string) string {
	if path == "" {
		return ""
	}

	// Expand initial ~ to OS specific home directory.
	if strings.HasPrefix(path, "~") {
		var homeDir string
		u, err := user.Current()
		if err == nil {
			homeDir = u.HomeDir
		} else {
			homeDir = os.Getenv("HOME")
		}

		path = strings.Replace(path, "~", homeDir, 1)
	}

	// NOTE: The os.ExpandEnv doesn't work with Windows-style %VARIABLE%,
	// but the variables can still be expanded via POSIX-style $VARIABLE.
	return filepath.Clean(os.ExpandEnv(path))
}
