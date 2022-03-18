package main

import (
	"database/sql"
	"html/template"

	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/zapponejosh/go-bible/handlers"
)

func main() {

	// create DB connection
	DB, err := connectDB()
	if err != nil {
		panic(err)
	}
	err = DB.Ping()
	if err != nil {
		panic(err)
	}
	defer DB.Close()

	// templates
	// var baseTmpl = [...]string{"static/templates/form.html", "static/templates/layout.html"} TODO need to refactor to use this
	tmpl := template.Must(template.ParseFiles("static/templates/index.html", "static/templates/form.html", "static/templates/layout.html"))
	chapterTmpl := template.Must(template.ParseFiles("static/templates/chapter.html", "static/templates/form.html", "static/templates/layout.html"))

	r := mux.NewRouter()

	// serve static file (ccs, images, etc)
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	// API routes
	api := r.PathPrefix("/api").Subrouter()
	booksHand := handlers.NewBooksHandler(DB)
	api.Handle("/books", booksHand)

	// chapter handler
	chapterHand := handlers.NewChapterHandler(DB, chapterTmpl)
	r.Handle("/{book}/{chapter}", chapterHand)

	// search handler
	hand := handlers.NewIndexHandler(DB, tmpl)
	r.Handle("/", hand)

	// Start server
	fmt.Println("Starting the server on http://localhost:3333")
	http.ListenAndServe(":3333", r)

}

func connectDB() (*sql.DB, error) {
	const (
		host     = "localhost"
		port     = 5432
		user     = "bibleapp"
		password = ""
		dbname   = "bible"
	)

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"dbname=%s sslmode=disable",
		host, port, user, dbname)

	db, err := sql.Open("postgres", psqlInfo)

	return db, err
}
