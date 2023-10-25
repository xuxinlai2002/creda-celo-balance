package main

import (
	"fmt"
	"os"
	"path/filepath"
	godebug "runtime/debug"
	"sync"

	"github.com/xuxinlai2002/creda-celo-balance/build"
	"github.com/xuxinlai2002/creda-celo-balance/config"
	"github.com/xuxinlai2002/creda-celo-balance/db"
	"github.com/xuxinlai2002/creda-celo-balance/log"
	"github.com/xuxinlai2002/creda-celo-balance/signal"
	"github.com/xuxinlai2002/creda-celo-balance/tokens"
	"github.com/xuxinlai2002/creda-celo-balance/transactions"
)

const (
	defaultLogFilename = "creda-celo.log"
)

func main() {
	var wg sync.WaitGroup
	// LogWriter is the root logger that all of the daemon's subloggers are
	// hooked up to.
	logWriter := build.NewRotatingLogWriter()

	// Hook interceptor for os signals.
	shutdownInterceptor, err := signal.Intercept()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Println("tokens start failed", "error", err)
		panic(any(err.Error()))
	}

	log.SetupLoggers(logWriter, shutdownInterceptor)
	err = logWriter.InitLogRotator(
		filepath.Join(cfg.LogDir, defaultLogFilename),
		cfg.MaxLogFileSize, cfg.MaxLogFiles,
	)
	if err != nil {
		fmt.Printf("log rotation setup failed: %v", err)
		panic(any(err.Error()))
	}

	// Parse, validate, and set debug log level(s).
	err = build.ParseAndSetDebugLevels(cfg.DebugLevel, logWriter)
	if err != nil {
		fmt.Printf("error parsing debug level: %v", err)
		panic(any(err.Error()))
	}

	// Show version at startup.
	log.MainLog.Infof("Version: %s commit=%s, build=%s, logging=%s, "+
		"debuglevel=%s", build.Version(), build.Commit,
		build.Deployment, build.LoggingType, cfg.DebugLevel)

	log.MainLog.Infof("Sanitizing Go's GC trigger %d%%", 80)
	godebug.SetGCPercent(int(80))

	err = db.CreateDataBase(cfg.PostgresDBName, cfg.PostgresUser, cfg.PostgresPassword, cfg.PostgresHost, cfg.PostgresPort)
	if err != nil {
		log.MainLog.Errorf("Create DataBase failed: %v", err)
		//panic(any(err.Error()))
	}

	tokensService, err := tokens.NewService(cfg, &wg)
	if err != nil {
		log.MainLog.Errorf("new tokens services err: %v", err)
		panic(any(err.Error()))
	}
	tokensService.Start(shutdownInterceptor)

	pullBlock, err := transactions.New(cfg, &wg)
	if err != nil {
		log.MainLog.Errorf("pullBlock initialized failed: %v", err)
		panic(any(err.Error()))
	}
	pullBlock.Start(shutdownInterceptor)

	wg.Wait()
}
