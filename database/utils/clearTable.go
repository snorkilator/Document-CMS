package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	connStr := "user=cmsdb dbname=cmsdb port=5433 password=admin"

	deleteAllRows := "delete from pages where 0=0"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)

	}
	_, err = db.Query(deleteAllRows)
	if err != nil {
		log.Fatal(err)
	}
}
