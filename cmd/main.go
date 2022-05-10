//Starts CMS server and DB
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"

	dbH "backend/database"

	_ "github.com/lib/pq"
)

//serveView serves view pages
func serveView(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hasPageID := regexp.MustCompile(`/(view)/(\d+)`)
		m := hasPageID.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		pageNum, _ := strconv.Atoi(m[2])
		P, err := dbH.GetPage(pageNum, db)
		if err != nil {
			panic(err)
		}
		t, _ := template.ParseFiles(dbH.Conf.Host.Path + "view.html")
		t.Execute(w, P)
	}
}

//serveEdit serves edit pages
func serveEdit(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hasPageID := regexp.MustCompile(`/(edit)/(\d+)`)
		m := hasPageID.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		pageNum, _ := strconv.Atoi(m[2])

		P, err := dbH.GetPage(pageNum, db)
		if err != nil {
			panic(err)
		}
		t, _ := template.ParseFiles(dbH.Conf.Host.Path + "edit.html")
		t.Execute(w, P)
	}
}

//deletePage parses /delete/<PID> calls b.DeletePage with the parsed PID
//redirects to home page
func deletePage(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		exp := regexp.MustCompile(`/delete/(\d+)`)
		m := exp.FindStringSubmatch(r.URL.Path)
		log.Println(m)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		num, err := strconv.Atoi(m[1])
		if err != nil {
			http.NotFound(w, r)
			return
		}
		err = dbH.DeletePage(num, db)
		if err != nil {
			http.NotFound(w, r)
			log.Println(err)
		}
		http.Redirect(w, r, "/", 300)
	}
}

//updatePage takes form data from user and writes it to the row specified in the request url
func updatePage(db *sql.DB, w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" {
		http.Error(w, "expected POST request type", 400)
		return errors.New("updatePage: wrong http method")
	}

	hasPageID := regexp.MustCompile(`/(update)/(\d+)`)
	m := hasPageID.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		return errors.New("updatePage: invalid page id")
	}
	pid, err := strconv.Atoi(m[2])
	if err != nil {
		return errors.New("updatePage: invalid page id")
	}
	exists, err := dbH.PageExists(pid, db)
	if err != nil {
		return errors.New("updatePage: " + err.Error())
	}

	if !exists {
		http.Error(w, "Could not save page", 400)
		return errors.New("updatePage: could not save page, page already exists")
	}

	err = r.ParseForm()
	if err != nil {
		return errors.New("updatePage: could not parse form")
	}

	// fmt.Println(r.Form["body"][0])
	err = dbH.EditPage(dbH.Page{Title: r.Form["title"][0], Body: r.Form["body"][0], ID: pid}, db)
	if err != nil {
		return errors.New("updatePage:" + err.Error())
	}
	http.Redirect(w, r, "/", 300)
	return nil
}

//addPage inserts row into database and serves edit page to edit that row
func addPage(db *sql.DB, w http.ResponseWriter, r *http.Request) error {
	P, err := dbH.InsertPage(dbH.Page{}, db)
	if err != nil {
		return errors.New("InsertPage: " + err.Error())
	}

	t, err := template.ParseFiles(dbH.Conf.Host.Path + "edit.html")
	if err != nil {
		return err
	}
	err = t.Execute(w, P)
	if err != nil {
		return err
	}
	return nil
}

func handleErr(f func(*sql.DB, http.ResponseWriter, *http.Request) error, db *sql.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		err := f(db, w, r)
		if err != nil {
			switch err.(type) {
			case dbH.DBFatal:
				log.Println("db fatal: " + err.Error())
			default:
				log.Println(err)
			}
		}
	}
}

//Starts CMS server and DB
func main() {

	var server http.Server
	server.Addr = ":" + dbH.Conf.Host.Port

	//opens connection with DB
	db, err := dbH.ConnectDB()
	defer db.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	// to restart db connection, must have closed all querys, otherwise using the rows will result in a seg error
	// maybe use a pointer to db connection with a mutexe or something of the sort
	// when connection is restarting, make db unavailable too all but utility function that does restart
	// use shutdown method on

	//Defines handlers for /edit/ /view/ /delete/ and / endpoints
	http.HandleFunc("/edit/", serveEdit(db))
	http.HandleFunc("/view/", serveView(db))
	http.HandleFunc("/delete/", deletePage(db))
	http.HandleFunc("/update/", handleErr(updatePage, db))
	http.HandleFunc("/add/", handleErr(addPage, db))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		file, err := os.ReadFile(dbH.Conf.Host.Path + "homepage.html")
		if err != nil {
			log.Println(err)
			http.NotFound(w, r)
			return
		}
		Pages, err := dbH.GetTitles(db)
		t := template.Must(template.New("").Parse(string(file)))
		t.Execute(w, Pages)
	})
	http.HandleFunc("/css.css", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, dbH.Conf.Host.Path+"css.css")
	})
	http.HandleFunc("/bundle.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, dbH.Conf.Host.Path+"bundle.js")
	})

	// returns current list of titles of pages from database
	// client side fetch request initiates this call
	http.HandleFunc("/titles.json", func(w http.ResponseWriter, r *http.Request) {

		Pages, err := dbH.GetTitles(db)

		b, err := json.Marshal(Pages)
		if err != nil {
			log.Println(err)
		}
		_, err = w.Write(b)
		if err != nil {
			log.Println(err)
		}
	})
	// starts localhost server on port 8081
	// will stop program after logging error
	// log.Fatal(http.ListenAndServe(":"+dbH.Conf.Host.Port, nil))
	go func() {
		server.Shutdown(context.TODO())
	}()
	log.Println(server.ListenAndServe())
}

/*
PANIC HANDLING:
if there is a serious db error:
try reseting connection:
if that works, continue
if it doesn't work
exit the application
*/
