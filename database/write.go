package database

import (
	"database/sql"
	"fmt"
	"log"
	"main/scraper"
	"os"
	"strings"

	_ "github.com/glebarez/go-sqlite"
)

type Schema map[string]TableDefinition

type TableDefinition struct {
	Columns         []map[string]string
	Indexes         map[string]string
	InsertStatement string
}

var tables = Schema{
	"words": TableDefinition{
		Columns: []map[string]string{
			{"id": "INTEGER PRIMARY KEY AUTOINCREMENT"},
			{"word": "TEXT NOT NULL"},
		},
		Indexes: map[string]string{
			"idx_word": "CREATE UNIQUE INDEX idx_word ON words(word);",
		},
		InsertStatement: `INSERT INTO words (word) VALUES (?) ON CONFLICT(word) DO UPDATE SET word=excluded.word RETURNING id;`,
	},
	"urls": TableDefinition{
		Columns: []map[string]string{
			{"id": "INTEGER PRIMARY KEY AUTOINCREMENT"},
			{"url": "TEXT NOT NULL"},
		},
		Indexes: map[string]string{
			"idx_url": "CREATE UNIQUE INDEX idx_url ON urls(url);",
		},
		InsertStatement: `INSERT INTO urls (url) VALUES (?) ON CONFLICT(url) DO UPDATE SET url=excluded.url RETURNING id;`,
	},
	"words_to_urls": TableDefinition{
		Columns: []map[string]string{
			{"word_id": "INTEGER REFERENCES words(id)"},
			{"url_id": "INTEGER REFERENCES urls(id)"},
		},
		Indexes: map[string]string{
			"word_url_pair": "CREATE UNIQUE INDEX word_url_pair ON words_to_urls(word_id, url_id);",
		},
		InsertStatement: `INSERT INTO words_to_urls (word_id, url_id) VALUES (?, ?);`,
	},
}

const databaseFileName = "words_to_urls.db"

func createTables(db *sql.DB) {
	for tableName, tableInfo := range tables {
		var colDefs []string

		for _, colMap := range tableInfo.Columns {
			for colName, colType := range colMap {
				colDefs = append(colDefs, fmt.Sprintf("%s %s", colName, colType))
			}
		}

		query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s);", tableName, strings.Join(colDefs, ", "))
		_, err := db.Exec(query)

		if err != nil {
			log.Fatal(err)
		}

		for _, indexCmd := range tableInfo.Indexes {
			db.Exec(indexCmd)
		}
	}
}

func prepareInsertStatements(tx *sql.Tx) map[string]*sql.Stmt {
	tableNameToInsertStmt := make(map[string]*sql.Stmt)

	for tableName, tableInfo := range tables {
		stmt, err := tx.Prepare(tableInfo.InsertStatement)

		if err != nil {
			log.Fatal(err)
		}

		tableNameToInsertStmt[tableName] = stmt
	}

	return tableNameToInsertStmt
}

func writeWordsToUrls(wordsToUrls *scraper.WordsToUrls, tableNameToInsertStmt map[string]*sql.Stmt) error {
	for word, urls := range wordsToUrls.GetItems() {
		var wordID int64

		if err := tableNameToInsertStmt["words"].QueryRow(word).Scan(&wordID); err != nil {
			return err
		}

		for url := range urls {
			var urlID int64

			if err := tableNameToInsertStmt["urls"].QueryRow(url).Scan(&urlID); err != nil {
				return err
			}

			if _, err := tableNameToInsertStmt["words_to_urls"].Exec(wordID, urlID); err != nil {
				return err
			}
		}
	}

	return nil
}

func WriteToDatabase(wordsToUrls *scraper.WordsToUrls) {
	_ = os.Remove(databaseFileName)
	db, err := sql.Open("sqlite", databaseFileName)

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	db.Exec("PRAGMA journal_mode=WAL;")
	db.Exec("PRAGMA synchronous=NORMAL;")
	createTables(db)
	tx, err := db.Begin()

	if err != nil {
		log.Fatal(err)
		tx.Rollback()
	}

	tableNameToInsertStmt := prepareInsertStatements(tx)

	for _, stmt := range tableNameToInsertStmt {
		defer stmt.Close()
	}

	if err := writeWordsToUrls(wordsToUrls, tableNameToInsertStmt); err != nil {
		tx.Rollback()
		log.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
}