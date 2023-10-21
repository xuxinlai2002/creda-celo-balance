package main

import (
	"fmt"
	"time"

	"github.com/xuxinlai2002/creda-celo-balance/config"
	"github.com/xuxinlai2002/creda-celo-balance/statistics/account"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Println("tokens start failed", "error", err)
		panic(any(err.Error()))
	}

	bal, err := account.New(cfg)
	bal.Start()
	time.Sleep(10 * time.Second)
}
