package main

import (
	"database/sql"
	"encoding/xml"
	"html/template"
	"log"

	"fmt"
	"net/http"
	"strings"

	_ "github.com/lib/pq"
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
	tmpl := template.Must(template.ParseFiles("static/index.html"))

	// serve static file (ccs, images, etc)
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	hand := NewIndexHandler(DB, tmpl)
	fmt.Println("Starting the server on http://localhost:3333")
	http.Handle("/", hand)
	http.ListenAndServe(":3333", nil)

}

type ResultData struct {
	Terms []string
	// Results map[string](template.HTML)
	Results []*VerseResult
}

func NewIndexHandler(db *sql.DB, indexTemp *template.Template) *indexHandler {
	return &indexHandler{db: db, indexTemp: indexTemp}
}

type indexHandler struct {
	db        *sql.DB
	indexTemp *template.Template
}

func (h indexHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "404 not found!", http.StatusNotFound)
		return
	}

	q := r.URL.Query()
	terms := strings.Split(strings.Join(q["search"], " "), " ")

	ftsQuery, results, err := searchBible(terms, h.db)
	if err != nil {
		log.Println("error searching terms", err)
		http.Error(w, "search error", http.StatusInternalServerError)
		return
	}

	highlightTerm(ftsQuery, results)

	data := ResultData{
		Terms:   ftsQuery,
		Results: results,
	}

	if err := h.indexTemp.Execute(w, data); err != nil {
		log.Println("error executing tempalte", err)
		http.Error(w, "tempalte error", http.StatusInternalServerError)
		return
	}
}

func highlightTerm(fts []string, results []*VerseResult) {

	parseResults := make(map[string]template.HTML)

	for _, v := range results {

		ref := fmt.Sprintf("%s %d:%d", v.Book, v.Chapter, v.Verse)

		//highlight term
		verse := v.Content.Text
		for _, t := range fts {
			// clean terms
			t = strings.Trim(strings.ReplaceAll(t, "'", ""), " ")

			verse = strings.Replace(verse, t, fmt.Sprintf("<i class='term'>%s</i>", t), -1)
			verse = strings.Replace(verse, strings.Title(t), fmt.Sprintf("<i class='term'>%s</i>", strings.Title(t)), -1)

		}
		textParsed := template.HTML(verse)

		parseResults[ref] = textParsed

		v.VerseHtml = template.HTML(verse)
	}

}

func searchBible(terms []string, DB *sql.DB) ([]string, []*VerseResult, error) {
	queryTerms := strings.Join(terms, "&")

	rows, err := DB.Query(`SELECT content, book, chapter, verse, query
	FROM bible,
		to_tsquery('english', $1) query
	WHERE query @@ fts LIMIT 50;`, queryTerms)
	if err != nil {
		// handle this error
		return nil, nil, err
	}
	defer rows.Close()

	var (
		results  []*VerseResult
		ftsQuery []string
	)
	for rows.Next() {

		var result VerseResult
		var content string
		var query string
		err = rows.Scan(&content, &result.Book, &result.Chapter, &result.Verse, &query)
		if err != nil {
			// handle this error
			return nil, nil, err
		}

		err = xml.Unmarshal([]byte(content), &result.Content)
		if err != nil {
			return nil, nil, err
		}

		ftsQuery = strings.Split(query, "&")

		results = append(results, &result)
	}
	// get any error encountered during iteration
	err = rows.Err()
	if err != nil {
		return nil, nil, err
	}

	return ftsQuery, results, err

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
