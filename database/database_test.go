package database

import (
	"fmt"
	"log"
	"net/http"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

//Only works if id=34 row exists
func TestPageExist(t *testing.T) {

	db, err := ConnectDB()
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 1000; i++ {
		b, err := PageExists(25, db)
		if err != nil {
			t.Fatal(err)
		}
		if !b {
			t.Fatalf("got: %v want: %v", b, true)
		}
		log.Println(i)
		time.Sleep(52 * time.Microsecond)
	}
}

func TestManyDBConnections(t *testing.T) {
	// request homepage over a hundred times
	for i := 0; i < 10000; i++ {
		resp, err := http.Get("http://localhost:8080/view/25")
		if err != nil {
			t.Fatal(err, resp.StatusCode)
		}
		log.Println(i, resp.StatusCode)

	}
}

func TestDBConfig(t *testing.T) {
	insertStr := "insert into pages (title, body) values ('firsttitle', 'firstbody')"
	deleteAllRows := "delete from pages where title='firsttitle'"
	readEntireStr := "select * from pages where title='firsttitle'"
	db, err := ConnectDB()
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Query(insertStr)
	if err != nil {
		t.Fatal(err)
	}
	rows, err := db.Query(readEntireStr)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Query(deleteAllRows)
	if err != nil {
		t.Fatal(err)
	}

	var title string
	var body string
	var id string
	rows.Next()
	err = rows.Scan(&title, &body, &id)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(title, body)
	if body != "firsttitle" {
		t.Fatalf("got title:%s body:%s", title, body)
	}
}
