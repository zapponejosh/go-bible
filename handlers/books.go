package handlers

import (
	"database/sql"
	"encoding/json"
	"log"

	"net/http"

	_ "github.com/lib/pq"
)

type booksHandler struct {
	db *sql.DB
}

func NewBooksHandler(db *sql.DB) *booksHandler {
	return &booksHandler{db: db}
}

func (h booksHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var books []*Book

	books, err := getBooks(h.db)
	if err != nil {
		log.Println("error fetching books", err)
		http.Error(w, "search error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(books); err != nil {
		http.Error(w, "encoding error", http.StatusInternalServerError)
	}
}

type Book struct {
	ShortName string `json:"short_name,omitempty"`
	LongName  string `json:"long_name,omitempty"`
	Testament string `json:"testament,omitempty"`
	BookNum   int    `json:"book_num,omitempty"`
	Chapters  int    `json:"chapters,omitempty"`
	Section   string `json:"section,omitempty"`
}

func getBooks(DB *sql.DB) ([]*Book, error) {

	rows, err := DB.Query(`SELECT * FROM books ORDER BY book_num;`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []*Book

	for rows.Next() {
		var book Book
		err = rows.Scan(&book.ShortName, &book.LongName, &book.Testament, &book.BookNum, &book.Chapters, &book.Section)
		if err != nil {
			return nil, err
		}
		books = append(books, &book)
	}
	return books, nil
}
