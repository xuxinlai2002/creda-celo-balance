package main

import (
	"fmt"
	"sync"

	"github.com/xuxinlai2002/creda-celo-balance/config"
	"github.com/xuxinlai2002/creda-celo-balance/statistics/account"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(1)
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Println("tokens start failed", "error", err)
		panic(any(err.Error()))
	}

	bal, err := account.New(cfg, &wg)
	if err != nil {
		fmt.Println(fmt.Sprintf("token start height: %v", err))
		panic(any(err))
	}
	bal.Start()

	wg.Wait()
}
