package handlers

import (
	"database/sql"
	"encoding/xml"
	"html/template"
	"log"
	"strconv"

	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type chapterResult struct {
	Reference string
	Results   []*VerseResult
}

type ChapterData struct {
	Reference Reference
	Results   []*VerseResult
}

func NewChapterHandler(db *sql.DB, chapterTemp *template.Template) *chapterHandler {
	return &chapterHandler{db: db, chapterTemp: chapterTemp}
}

type chapterHandler struct {
	db          *sql.DB
	chapterTemp *template.Template
}

type Reference struct {
	Book    string
	Chapter int
}

func (h chapterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	book := vars["book"]
	chapter, err := strconv.Atoi(vars["chapter"])

	results, err := getChapters(Reference{book, chapter}, h.db)
	if err != nil {
		log.Println("error searching for chapter", err)
		http.Error(w, "search error", http.StatusInternalServerError)
		return
	}

	var ref = Reference{book, chapter}

	data := ChapterData{
		Reference: ref,
		Results:   results,
	}

	if err := h.chapterTemp.Execute(w, data); err != nil {
		log.Println("error executing tempalte", err)
		http.Error(w, "tempalte error", http.StatusInternalServerError)
		return
	}
}

func getChapters(ref Reference, DB *sql.DB) ([]*VerseResult, error) {

	rows, err := DB.Query(`SELECT content, book, chapter, verse
	FROM bible WHERE book = $1 AND chapter = $2
	ORDER BY verse;`, ref.Book, ref.Chapter)
	if err != nil {
		// handle this error
		return nil, err
	}
	defer rows.Close()

	var (
		results []*VerseResult
	)
	for rows.Next() {

		var result VerseResult
		var content string
		err = rows.Scan(&content, &result.Book, &result.Chapter, &result.Verse)
		if err != nil {
			// handle this error
			return nil, err
		}

		err = xml.Unmarshal([]byte(content), &result.Content)
		if err != nil {
			return nil, err
		}

		results = append(results, &result)
	}
	// get any error encountered during iteration
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return results, err

}
