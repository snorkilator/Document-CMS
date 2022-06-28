//database defines basic functions to setup and interact with the DB
package database

import (
	"backend/readConfig"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strconv"

	_ "github.com/lib/pq"
)

type DBFatal string

func (err DBFatal) Error() string {
	return string(err)
}

func NewdbFatal(wrap string, err error) error {
	s := wrap + ":" + err.Error()
	return DBFatal(s)
}

var (
	Conf, err = readConfig.GetConfig()
)

//Page is used throughout the module for passing a page's content and ID between functions
type Page struct {
	Title string
	Body  string
	ID    int
}

//ConnectDB connects to database and returns *sql.DB for further communication with DB
func ConnectDB() (*sql.DB, error) {
	connectionStr := fmt.Sprintf("user=%s dbname=%s port=%s password=%s", Conf.Db.User, Conf.Db.Name, Conf.Db.Port, Conf.Db.Password)
	return sql.Open("postgres", connectionStr)
}

//GetTitles takes db connection and returns all titles in pages table as []page
func GetTitles(db *sql.DB) ([]Page, error) {
	data, err := db.Query("SELECT (id, title) FROM pages where 0=0")
	if err != nil {
		return []Page{}, NewdbFatal("GetTitles", err)
	}
	results := []Page{}

	for data.Next() {
		var row string
		data.Scan(&row)
		format := regexp.MustCompile(`\((\d+)\,([^$]*)\)$`)
		m := format.FindStringSubmatch(row)
		num, err := strconv.Atoi(m[1])
		if err != nil {
			return []Page{}, errors.New("GetTitles" + err.Error())
		}

		results = append(results, Page{Title: m[2], ID: num})
	}
	return results, nil
}

// Takes PID, queries database for matching PID and returns
func GetPage(pid int, db *sql.DB) (Page, error) {
	tquery := fmt.Sprintf("SELECT title FROM pages Where id = '%d'", pid)
	title, err := db.Query(tquery)
	if err != nil {
		return Page{}, fmt.Errorf("GetPage: %v", err)
	}
	defer title.Close()
	bquery := fmt.Sprintf("SELECT body FROM pages Where id = '%d'", pid)
	body, err := db.Query(bquery)
	defer body.Close()
	if err != nil {
		return Page{}, fmt.Errorf("GetPage: %v", err)
	}

	t := ""
	b := ""
	title.Next()
	body.Next()
	title.Scan(&t)
	body.Scan(&b)
	return Page{Title: t, Body: b, ID: pid}, nil
}

//InsertPage adds the given page to the pages table
func InsertPage(p Page, db *sql.DB) (Page, error) {
	query := fmt.Sprintf("INSERT INTO pages (title,body) VALUES('%s','%s') RETURNING ID", p.Title, p.Body)
	row := db.QueryRow(query)
	pid := ""
	err := row.Scan(&pid)
	if err != nil {
		return Page{}, err
	}
	i, err := strconv.Atoi(pid)
	if err != nil {
		return Page{}, err
	}
	return Page{ID: i}, nil
}

//DeletePage deletes page from database
func DeletePage(pid int, db *sql.DB) error {
	query := fmt.Sprintf(`DELETE FROM pages where id=%d`, pid)
	_, err := db.Query(query)
	return err
}

//EditPage applies change to row specified in p.ID
func EditPage(p Page, db *sql.DB) error {
	query := fmt.Sprintf("UPDATE pages SET title = '%s', body = '%s' WHERE id='%d'", p.Title, p.Body, p.ID)
	_, err := db.Query(query)
	if err != nil {
		return err
	}
	return nil
}

//PageExists takes PID and returns true and no error if it's in the database. If query fails returns false and error.
func PageExists(pid int, db *sql.DB) (bool, error) {
	qstr := fmt.Sprintf("select exists(select * from pages where id='%d')", pid)
	row := db.QueryRow(qstr)

	exists := ""
	if err := row.Scan(&exists); err != nil {
		return false, errors.New("PageExists: " + err.Error())
	}

	if string(exists) == `true` {
		return true, nil
	}
	return false, nil
}
