package handlers

import (
	"database/sql"
	"encoding/xml"
	"fmt"
	"html/template"
	"log"

	"net/http"
	"strings"

	_ "github.com/lib/pq"
)

type VerseResult struct {
	Content   Verse
	Book      string
	Chapter   int
	Verse     int
	VerseHtml template.HTML
}
type Verse struct {
	Text        string        `xml:",chardata"`
	Num         string        `xml:"num,attr"`
	Note        []Note        `xml:"note"`
	TransChange []TransChange `xml:"transChange"`
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
type indexHandler struct {
	db        *sql.DB
	indexTemp *template.Template
}

type ResultData struct {
	Terms   []string
	Results []*VerseResult
}

func NewIndexHandler(db *sql.DB, indexTemp *template.Template) *indexHandler {
	return &indexHandler{db: db, indexTemp: indexTemp}
}

func (h indexHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "404 not found! This is the index hanlder!", http.StatusNotFound)
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

			verse = strings.Replace(verse, t, fmt.Sprintf("<i class='highlight'>%s</i>", t), -1)
			verse = strings.Replace(verse, strings.Title(t), fmt.Sprintf("<i class='highlight'>%s</i>", strings.Title(t)), -1)

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
