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

	terms := flag.String("terms", "bald", "Comma separated search terms.")
	flag.Parse()
	termsSlice := strings.Split(*terms, ",")

	connectDB(termsSlice)

}

func connectDB(terms []string) *sql.DB {
	const (
		host     = "localhost"
		port     = 5432
		user     = "bibleapp"
		password = ""
		dbname   = "bible"
	)

	queryTerms := strings.Join(terms, "&")

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"dbname=%s sslmode=disable",
		host, port, user, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		panic(err)
	}

	rows, err := db.Query(`SELECT content, book, chapter, verse
	FROM bible,
		to_tsquery('english', $1) query
	WHERE query @@ fts
	LIMIT 20;`, queryTerms)
	if err != nil {
		// handle this error
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		var content string
		var book string
		var chapter int
		var verse int
		err = rows.Scan(&content, &book, &chapter, &verse)
		if err != nil {
			// handle this error
			panic(err)
		}

		fmt.Printf("%s %d:%d\n", book, chapter, verse)
		var verseContent Verse
		err = xml.Unmarshal([]byte(content), &verseContent)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s\n\n", verseContent.Text)
	}
	// get any error encountered during iteration
	err = rows.Err()
	if err != nil {
		panic(err)
	}

	return db
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
