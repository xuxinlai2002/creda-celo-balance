package main

import (
	"fmt"
	"github.com/xuxinlai2002/creda-celo-balance/config"
	"github.com/xuxinlai2002/creda-celo-balance/tokens"
	"github.com/xuxinlai2002/creda-celo-balance/transactions"
	godebug "runtime/debug"
)

func main() {
	fmt.Println("Sanitizing Go's GC trigger", "percent", int(80))
	godebug.SetGCPercent(int(80))
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Println("tokens start failed", "error", err)
		panic(any(err.Error()))
	}
	_ = cfg
	err = tokens.Start(cfg)
	if err != nil {
		fmt.Println("tokens start failed", "error", err)
	}

	resultCh := make(chan error, 1)
	transactions.Start(cfg, resultCh)
	if err != nil {
		fmt.Println("tokens start failed", "error", err)
	}

	for {
		select {
		case failedError := <-resultCh:
			fmt.Println("transactions pull failed", "error", failedError)
			return
		}
	}
}
