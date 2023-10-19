package main

import (
	"database/sql"
	"fmt"
	"github.com/xuxinlai2002/creda-celo-balance/config"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Println("tokens start failed", "error", err)
		panic(any(err.Error()))
	}

	db, err := sql.Open("postgres", fmt.Sprintf("user=%s dbname=%s sslmode=disable password=%s", user, dbName, password))
	if err != nil {
		g.Log().Error(ctx, err)
	}
	defer db.Close()
}
