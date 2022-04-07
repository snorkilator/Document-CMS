package database

import (
	"database/sql"

	_ "github.com/lib/pq"
)

//CreatePagesTable creates pages table without creating rows
func CreatePagesTable(db *sql.DB) error {
	_, err := db.Query("CREATE TABLE pages (ID  SERIAL PRIMARY KEY, title TEXT, body TEXT)")
	if err != nil {
		return err
	}
	return nil
}
