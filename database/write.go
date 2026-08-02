package database

import (
	"database/sql"
	"fmt"
	"log"
	"main/scraper"
	"strings"

	_ "github.com/glebarez/go-sqlite"
)

var tables = map[string][]map[string]string{
	"words": {
		{"id": "INTEGER PRIMARY KEY AUTOINCREMENT"},
		{"word": "TEXT NOT NULL"},
	},
	"urls": {
		{"id": "INTEGER PRIMARY KEY AUTOINCREMENT"},
		{"url": "TEXT NOT NULL"},
	},
	"words_to_urls": {
		{"word_id": "INTEGER REFERENCES words(id)"},
		{"url_id": "INTEGER REFERENCES urls(id)"},
	},
}

func WriteToDatabase(wordsToUrls *scraper.WordsToUrls) {
	db, err := sql.Open("sqlite", "words_to_urls.db")

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	db.Exec("PRAGMA journal_mode=WAL;")
	db.Exec("PRAGMA synchronous=NORMAL;")
	tx, err := db.Begin()

	for tableName, columns := range tables {
		var colDefs []string

		for _, colMap := range columns {
			for colName, colType := range colMap {
				colDefs = append(colDefs, fmt.Sprintf("%s %s", colName, colType))
			}
		}

		query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s);", tableName, strings.Join(colDefs, ", "))
		_, err := db.Exec(query)

		if err != nil {
			log.Fatal(err)
		}
	}

	insertWordSQL := `INSERT INTO words (word) VALUES (?);`
	insertUrlSQL := `INSERT INTO urls (url) VALUES (?);`
	insertWordToUrlSQL := `INSERT INTO words_to_urls (word_id, url_id) VALUES (?, ?);`
	wordStmt, _ := tx.Prepare(insertWordSQL)
	urlStmt, _ := tx.Prepare(insertUrlSQL)
	wordToUrlStmt, _ := tx.Prepare(insertWordToUrlSQL)
	defer wordStmt.Close()
	defer urlStmt.Close()
	defer wordToUrlStmt.Close()

	for word, urls := range wordsToUrls.GetItems() {
		result, _ := wordStmt.Exec(word)
		word_id, _ := result.LastInsertId()

		for url := range urls {
			result, _ = urlStmt.Exec(url, word_id)
			url_id, _ := result.LastInsertId()
			wordToUrlStmt.Exec(word_id, url_id)
		}
	}

	tx.Commit()
}
