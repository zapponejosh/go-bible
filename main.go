package main

import (
	"database/sql"
	"encoding/xml"
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
	tmpl := template.Must(template.ParseFiles("static/index.html"))
	chapterTmpl := template.Must(template.ParseFiles("static/chapter.html"))

	r := mux.NewRouter()

	// serve static file (ccs, images, etc)
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// chapter handler
	chapterHand := handlers.NewChapterHandler(DB, chapterTmpl)
	r.Handle("{book}/{chapter}", chapterHand)

	// search handler
	hand := handlers.NewIndexHandler(DB, tmpl)
	r.Handle("/", hand)

	// Start server
	fmt.Println("Starting the server on http://localhost:3333")
	http.ListenAndServe(":3333", nil)

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

type VerseResult struct {
	Content   Verse
	Book      string
	Chapter   int
	Verse     int
	VerseHtml template.HTML
}
type Note struct {
	Text        string        `xml:",chardata"`
	TransChange []TransChange `xml:"transChange"`
}

type TransChange struct {
	Text      string `xml:",chardata"`
	Type      string `xml:"type,attr"`
	TransNote string `xml:"note"` // TransNote only provided within Verse (not in Note)
}

type Verse struct {
	Text        string        `xml:",chardata"`
	Num         string        `xml:"num,attr"`
	Note        []Note        `xml:"note"`
	TransChange []TransChange `xml:"transChange"`
}

type Chapter struct {
	Text  string  `xml:",chardata"`
	Num   string  `xml:"num,attr"`
	Verse []Verse `xml:"verse"`
}

type Book struct {
	Text    string    `xml:",chardata"`
	Num     string    `xml:"num,attr"`
	Chapter []Chapter `xml:"chapter"`
}

type Bible struct {
	XMLName xml.Name `xml:"bible"`
	Text    string   `xml:",chardata"`
	Name    string   `xml:"name,attr"`
	Abbrev  string   `xml:"abbrev,attr"`
	Book    []Book   `xml:"book"`
}
