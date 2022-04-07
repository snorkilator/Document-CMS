//Starts CMS server and DB
package main

import (
	"database/sql"
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
			log.Println(err)
		}
		http.Redirect(w, r, "/", 300)
	}
}

//updatePage takes form data from user and writes it to the row specified in the request url
func updatePage(db *sql.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "expected POST request type", 400)
			return
		}

		hasPageID := regexp.MustCompile(`/(update)/(\d+)`)
		m := hasPageID.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		pid, _ := strconv.Atoi(m[2])
		exists, _ := dbH.PageExists(pid, db) //add error handling
		if exists {
			fmt.Println("exists page")
			_ = r.ParseForm() //add error handling
			fmt.Println(r.Form["body"][0])
			dbH.EditPage(dbH.Page{Title: r.Form["title"][0], Body: r.Form["body"][0], ID: pid}, db)
			http.Redirect(w, r, "/", 300)
			return
		}
		http.Error(w, "Could not save page", 400)
	}
}

//addPage inserts row into database and serves edit page to edit that row
func addPage(db *sql.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		P, err := dbH.InsertPage(dbH.Page{}, db) //handle error!!
		if err != nil {
			panic(err)
		}

		t, _ := template.ParseFiles(dbH.Conf.Host.Path + "edit.html")
		err = t.Execute(w, P)
		if err != nil {
			panic(err)
		}

	}
}

//Starts CMS server and DB
func main() {

	//opens connection with DB
	db, err := dbH.ConnectDB()
	if err != nil {
		panic(err)
	}
	//Defines handlers for /edit/ /view/ /delete/ and / endpoints
	http.HandleFunc("/edit/", serveEdit(db))
	http.HandleFunc("/view/", serveView(db))
	http.HandleFunc("/delete/", deletePage(db))
	http.HandleFunc("/update/", updatePage(db))
	http.HandleFunc("/add/", addPage(db))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		file, err := os.ReadFile(dbH.Conf.Host.Path + "homepage.html")
		if err != nil {
			log.Println(err)
			http.NotFound(w, r)
			return
		}
		Pages := dbH.GetTitles(db)
		t := template.Must(template.New("").Parse(string(file)))
		t.Execute(w, Pages)
	})
	http.HandleFunc("/css.css", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, dbH.Conf.Host.Path+"css.css")
	})

	//starts localhost server on port 8081
	//will stop program after logging error
	log.Fatal(http.ListenAndServe(":"+dbH.Conf.Host.Port, nil))

}
