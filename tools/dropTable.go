package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"time"
)

func main() { //tx_20200424 tx_20200814
	startDate := time.Date(2020, 4, 24, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2020, 8, 14, 0, 0, 0, 0, time.UTC)

	dbName := "creadadb"
	user := "creda"
	password := "20231011"
	ndb, err := sql.Open("postgres", fmt.Sprintf("user=%s dbname=%s sslmode=disable password=%s", user, dbName, password))
	if err != nil {
		fmt.Println(err)
	}

	defer ndb.Close()

	for dt := startDate; dt.Before(endDate); dt = dt.AddDate(0, 0, 1) {
		name := fmt.Sprintf("tx_%04d%02d%02d", dt.Year(), int(dt.Month()), dt.Day())
		createTableSQL := `DROP TABLE ` + name
		_, err = ndb.Exec(createTableSQL)
		if err != nil {
			fmt.Println(err)
		}
	}
}
