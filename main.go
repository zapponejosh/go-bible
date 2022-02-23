package main

import (
	"database/sql"
	"encoding/xml"
	"flag"
	"fmt"
	"strings"

	_ "github.com/lib/pq"
)

func main() {

	terms := flag.String("terms", "", "Comma separated search terms.")
	flag.Parse()
	termsSlice := strings.Split(*terms, ",")

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

	ftsQuery, results, err := searchBible(termsSlice, DB)
	showResults(ftsQuery, results, terms)

}

func showResults(fts []string, results *[]VerseResult, terms *string) {

	for _, v := range *results {
		fmt.Printf("%s %d:%d\n", v.book, v.chapter, v.verse)

		//highlight term
		verse := v.content.Text
		for _, t := range fts {
			t = strings.Trim(t, "'")
			verse = strings.Replace(verse, t, fmt.Sprintf("{TERM}%s{/TERM}", t), -1)
			verse = strings.Replace(verse, strings.Title(t), fmt.Sprintf("{TERM}%s{/TERM}", strings.Title(t)), -1)

		}
		fmt.Printf("%s\n\n", verse)
	}
	fmt.Printf("%d Results for %s\n", len(*results), *terms)
	fmt.Printf("%s", fts)
}

func searchBible(terms []string, DB *sql.DB) ([]string, *[]VerseResult, error) {

	queryTerms := strings.Join(terms, "&")

	var ftsQuery []string

	rows, err := DB.Query(`SELECT content, book, chapter, verse, query
	FROM bible,
		to_tsquery('english', $1) query
	WHERE query @@ fts;`, queryTerms)
	if err != nil {
		// handle this error
		return nil, nil, err
	}
	defer rows.Close()

	var results []VerseResult

	for rows.Next() {

		var result VerseResult
		var content string
		var query string
		err = rows.Scan(&content, &result.book, &result.chapter, &result.verse, &query)
		if err != nil {
			// handle this error
			return nil, nil, err
		}

		err = xml.Unmarshal([]byte(content), &result.content)
		if err != nil {
			return nil, nil, err
		}

		ftsQuery = strings.Split(query, "&")

		results = append(results, result)
	}
	// get any error encountered during iteration
	err = rows.Err()
	if err != nil {
		return nil, nil, err
	}

	return ftsQuery, &results, err

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
	content Verse
	book    string
	chapter int
	verse   int
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
