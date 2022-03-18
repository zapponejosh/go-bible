package handlers

import (
	"database/sql"
	"encoding/xml"
	"fmt"
	"html/template"
	"log"
	"regexp"
	"strconv"

	"net/http"
	"strings"
)

type VerseResult struct {
	Content   Verse
	Book      string
	Chapter   int
	Verse     int
	VerseHtml template.HTML
}

type indexHandler struct {
	db        *sql.DB
	indexTemp *template.Template
}

type TemplateData struct {
	Terms   []string
	Results []*VerseResult
	Count   int
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
	// format query in slice
	terms := strings.Split(strings.Join(q["search"], " "), " ")
	offset, err := strconv.Atoi(strings.Join(q["p"], ""))
	if err != nil || offset <= 0 {
		offset = 0
	} else {
		offset = (offset - 1) * 50
	}
	// allow for phrases
	phrase := false
	t1 := strings.Index(terms[0], "\"")
	last := terms[len(terms)-1]
	tfinal := strings.Index(last, "\"")
	if t1 == 0 && tfinal == len(last)-1 {
		phrase = true
	}

	ftsQuery, results, resultCount, err := searchBible(terms, h.db, phrase, offset)
	if err != nil {
		log.Println("error searching terms", err)
		http.Error(w, "search error", http.StatusInternalServerError)
		return
	}

	highlightTerm(ftsQuery, results, phrase)

	data := TemplateData{
		Terms:   ftsQuery,
		Results: results,
		Count:   resultCount,
	}

	if err := h.indexTemp.Execute(w, data); err != nil {
		log.Println("error executing tempalte", err)
		http.Error(w, "tempalte error", http.StatusInternalServerError)
		return
	}
}

func highlightTerm(fts []string, results []*VerseResult, isPhrase bool) {
	for _, v := range results {

		//highlight term
		verse := v.Content.Text
		for _, t := range fts {
			// clean terms
			t = strings.Trim(strings.ReplaceAll(t, "'", ""), " ")

			// convert verse string to []string and for each item do strings.contains
			vSlice := strings.Split(verse, " ")

			for i, wrd := range vSlice {
				preSpan := "<span class='highlight'>"
				postSpan := "</span>"
				if strings.Contains(wrd, t) || strings.Contains(wrd, strings.Title(t)) {

					// if last character is a symbol pull it out and add it after postSpan
					lastCh := wrd[len(wrd)-1:]
					matched, _ := regexp.MatchString(`\W`, lastCh)
					if matched {
						vSlice[i] = preSpan + wrd[:len(wrd)-1] + postSpan + lastCh
					} else {
						vSlice[i] = preSpan + wrd + postSpan
					}

				}
			}
			verse = strings.Join(vSlice, " ")

		}

		v.VerseHtml = template.HTML(verse)
	}

}

func searchBible(terms []string, DB *sql.DB, isPhrase bool, offset int) ([]string, []*VerseResult, int, error) {
	var queryTerms string
	if isPhrase {
		queryTerms = strings.Join(terms, "<->")
	} else {
		queryTerms = strings.Join(terms, "&")
	}

	rows, err := DB.Query(`SELECT b.content,
  b.book,
  b.chapter,
  b.verse,
  b.query,
	COUNT(*) OVER() AS result_count
FROM (
    SELECT content,
      book,
      chapter,
      verse,
      query
    FROM bible,
      to_tsquery('english', $1) query
    WHERE query @@ fts
  ) as b
  LEFT JOIN books k on k.short_name = b.book
ORDER BY k.book_num LIMIT 50 OFFSET $2;`, queryTerms, offset)
	if err != nil {
		// handle this error
		return nil, nil, 0, err
	}
	defer rows.Close()

	var (
		results     []*VerseResult
		ftsQuery    []string
		resultCount int
	)
	for rows.Next() {

		var result VerseResult
		var content string
		var query string
		err = rows.Scan(&content, &result.Book, &result.Chapter, &result.Verse, &query, &resultCount)
		if err != nil {
			// handle this error
			return nil, nil, 0, err
		}

		err = xml.Unmarshal([]byte(content), &result.Content)
		if err != nil {
			return nil, nil, 0, err
		}

		if isPhrase {
			phraseFts := query
			for i := 1; i < 5; i++ {
				phraseFts = strings.ReplaceAll(phraseFts, fmt.Sprintf("<%d>", i), "&")
			}
			ftsQuery = strings.Split(phraseFts, "&")
		} else {
			ftsQuery = strings.Split(query, "&")
		}

		results = append(results, &result)
	}
	// get any error encountered during iteration
	err = rows.Err()
	if err != nil {
		return nil, nil, 0, err
	}

	return ftsQuery, results, resultCount, err

}
