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
	"net/url"
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

type Parser func(string) (interface{}, error)

type param struct {
	Type   string
	Column string // for filters only
	Parser Parser
}

func IntParser(v string) (interface{}, error) {
	return strconv.Atoi(v)
}

func NoopParser(v string) (interface{}, error) {
	return v, nil
}
func SearchParser(v string) (interface{}, error) {
	terms := strings.Split(v, " ")
	phrase := false
	t1 := strings.Index(terms[0], "\"")
	last := terms[len(terms)-1]
	tfinal := strings.Index(last, "\"")
	if t1 == 0 && tfinal == len(last)-1 {
		phrase = true
	}

	var queryTerms string
	if phrase {
		queryTerms = strings.Join(terms, "<->")
	} else {
		queryTerms = strings.Join(terms, "&")
	}
	return queryTerms, nil
}

func PaginationParser(v string) (interface{}, error) {
	offset, err := strconv.Atoi(v)
	if err != nil || offset <= 0 {
		offset = 0
	} else {
		offset = (offset - 1) * 50
	}
	return offset, nil
}

func NewIndexHandler(db *sql.DB, indexTemp *template.Template) *indexHandler {
	return &indexHandler{db: db, indexTemp: indexTemp}
}

func (h indexHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "404 not found! This is the index hanlder!", http.StatusNotFound)
		return
	}

	query, args, err := genQuery(r.URL.Query())
	if err != nil {
		log.Println("error parsing params or generating query", err)
		http.Error(w, "search error", http.StatusInternalServerError)
		return
	}

	ftsQuery, results, resultCount, phrase, err := searchBible(query, args, h.db)
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

	if err := h.indexTemp.ExecuteTemplate(w, "layout", data); err != nil {
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

func genQuery(values url.Values) (string, []interface{}, error) {

	var (
		validParams = map[string]param{
			"search": {
				Type:   "search",
				Parser: SearchParser,
			},
			"p": {
				Type:   "offset",
				Parser: PaginationParser,
			},
			"testamentFilter": {
				Type:   "filter",
				Column: "k.testament",
			},
			"sectionFilter": {
				Type:   "filter",
				Column: "k.section",
			},
			"bookFilter": {
				Type:   "filter",
				Column: "k.short_name",
			},
		}
		searchArg string
		offsetArg string
		filters   []string
		args      []interface{}
		count     = 1
	)

	for key := range values {
		f, found := validParams[key]
		if !found {
			// fail silently and skip to next param
			continue
		}
		// leave out params with no values
		if str := strings.TrimSpace(values.Get(key)); str == "" {
			continue
		}

		parser := f.Parser
		if parser == nil {
			// if a parser wasn't supplied, use the noop
			parser = NoopParser
		}

		val, err := parser(values.Get(key))
		if err != nil {
			return "", nil, fmt.Errorf("query generator: error parsing key %s: %w", key, err)
		}

		args = append(args, val)
		switch f.Type {
		case "filter":
			col := f.Column
			filters = append(filters, fmt.Sprintf("%s = $%d", col, count))
		case "search":
			searchArg = fmt.Sprintf("$%d", count)
		case "offset":
			offsetArg = fmt.Sprintf("$%d", count)
		}
		count++
	}

	// no pagination param provided set default to zero
	if offsetArg == "" {
		offsetArg = fmt.Sprintf("$%d", count)
		args = append(args, 0)
	}

	where := strings.Join(filters, " AND ")
	if where != "" {
		where = "WHERE " + where
	}
	query := fmt.Sprintf(`SELECT b.content,
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
      to_tsquery('english', %s) query
    WHERE query @@ fts
  ) as b
  LEFT JOIN books k on k.short_name = b.book
	%s
ORDER BY k.book_num LIMIT 50 OFFSET %s;`, searchArg, where, offsetArg)
	return query, args, nil
}

func searchBible(search string, args []interface{}, DB *sql.DB) ([]string, []*VerseResult, int, bool, error) {
	isPhrase := false
	if strings.Contains(search, "<->") {
		isPhrase = true
	}

	rows, err := DB.Query(search, args...)
	if err != nil {
		// handle this error
		return nil, nil, 0, isPhrase, err
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
			return nil, nil, 0, isPhrase, err
		}

		err = xml.Unmarshal([]byte(content), &result.Content)
		if err != nil {
			return nil, nil, 0, isPhrase, err
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
		return nil, nil, 0, isPhrase, err
	}

	return ftsQuery, results, resultCount, isPhrase, err

}
