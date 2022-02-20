package main

import (
	"encoding/xml"
	"fmt"
	"os"
)

func main() {
	f, err := os.Open("LEB.xml")
	defer f.Close()
	if err != nil {
		panic(err)
	}

	var data Bible
	err = xml.NewDecoder(f).Decode(&data)

	fmt.Printf("%d books in the Bible\n", len(data.Book))
	fmt.Printf("First is %s\n", data.Book[0].Num)
	fmt.Printf("Last is %s\n", data.Book[len(data.Book)-1].Num)
	fmt.Printf("%s\n", data.Book[0].Chapter[0].Verse[0].Text)

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
