package main

import (
	"fmt"
	godebug "runtime/debug"

	"github.com/xuxinlai2002/creda-celo-balance/config"
	"github.com/xuxinlai2002/creda-celo-balance/tokens"
	"github.com/xuxinlai2002/creda-celo-balance/transactions"
)

func main() {
	fmt.Println("Sanitizing Go's GC trigger", "percent", int(80))
	godebug.SetGCPercent(int(80))
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Println("tokens start failed", "error", err)
		panic(any(err.Error()))
	}

	go func() {
		err = tokens.Start(cfg)
		if err != nil {
			fmt.Println("tokens start failed", "error", err)
		}
	}()

	resultCh := make(chan error, 1)
	pullBlock, err := transactions.New(cfg)
	if err != nil {
		fmt.Println("pullBlock new failed", "error", err)
		panic(any(err.Error()))
	}
	pullBlock.Start(resultCh)

	for {
		select {
		case failedError := <-resultCh:
			if failedError != nil {
				fmt.Println("transactions pull failed", "error", failedError)
			} else {
				fmt.Println("block pull completed")
			}
			return
		}
	}
}
