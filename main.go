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

	err = tokens.Start(cfg)
	if err != nil {
		fmt.Println("tokens start failed", "error", err)
	}

	err = transactions.Start()
	if err != nil {
		fmt.Println("tokens start failed", "error", err)
	}

}
