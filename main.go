package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/xuxinlai2002/creda-celo-balance/config"
	"github.com/xuxinlai2002/creda-celo-balance/db"
	"github.com/xuxinlai2002/creda-celo-balance/signal"
	"github.com/xuxinlai2002/creda-celo-balance/tokens"
	"github.com/xuxinlai2002/creda-celo-balance/transactions"
	godebug "runtime/debug"
)

func main() {
	var wg sync.WaitGroup
	// Hook interceptor for os signals.
	shutdownInterceptor, err := signal.Intercept()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("Sanitizing Go's GC trigger", "percent", int(80))
	godebug.SetGCPercent(int(80))
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Println("tokens start failed", "error", err)
		panic(any(err.Error()))
	}

	err = db.CreateDataBase(cfg.PostgresDBName, cfg.PostgresUser, cfg.PostgresPassword, cfg.PostgresHost, cfg.PostgresPort)
	if err != nil {
		fmt.Println("Create DataBase failed", "error", err)
		//panic(any(err.Error()))
	}

	tokensService, err := tokens.NewService(cfg, &wg)
	if err != nil {
		fmt.Println("new tokens services err: ", err)
		panic(any(err.Error()))
	}
	tokensService.Start(shutdownInterceptor)

	pullBlock, err := transactions.New(cfg, &wg)
	if err != nil {
		fmt.Println("pullBlock initialized failed", "error", err)
		panic(any(err.Error()))
	}
	pullBlock.Start(shutdownInterceptor)

	wg.Wait()
}
