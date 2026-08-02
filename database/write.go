package database

import (
	"database/sql"
	"fmt"
	"log"
	"main/data_structures"
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
			"idx_word": "CREATE UNIQUE INDEX IF NOT EXISTS idx_word ON words(word);",
		},
		InsertStatement: `INSERT INTO words (word) VALUES (?) ON CONFLICT(word) DO UPDATE SET word=excluded.word RETURNING id;`,
	},
	"urls": TableDefinition{
		Columns: []map[string]string{
			{"id": "INTEGER PRIMARY KEY AUTOINCREMENT"},
			{"url": "TEXT NOT NULL"},
		},
		Indexes: map[string]string{
			"idx_url": "CREATE UNIQUE INDEX IF NOT EXISTS idx_url ON urls(url);",
		},
		InsertStatement: `INSERT INTO urls (url) VALUES (?) ON CONFLICT(url) DO UPDATE SET url=excluded.url RETURNING id;`,
	},
	"words_to_urls": TableDefinition{
		Columns: []map[string]string{
			{"word_id": "INTEGER REFERENCES words(id)"},
			{"url_id": "INTEGER REFERENCES urls(id)"},
		},
		Indexes: map[string]string{
			"word_url_pair": "CREATE UNIQUE INDEX IF NOT EXISTS word_url_pair ON words_to_urls(word_id, url_id);",
		},
		InsertStatement: `INSERT OR IGNORE INTO words_to_urls (word_id, url_id) VALUES (?, ?);`,
	},
}

const databaseFileName = "words_to_urls.db"

func createTables(db *sql.DB) error {
	for tableName, tableInfo := range tables {
		var colDefs []string

		for _, colMap := range tableInfo.Columns {
			for colName, colType := range colMap {
				colDefs = append(colDefs, fmt.Sprintf("%s %s", colName, colType))
			}
		}

		query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s);", tableName, strings.Join(colDefs, ", "))

		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("create table %s: %w", tableName, err)
		}

		for indexName, indexCmd := range tableInfo.Indexes {
			if _, err := db.Exec(indexCmd); err != nil {
				return fmt.Errorf("create index %s: %w", indexName, err)
			}
		}
	}

	return nil
}

func prepareInsertStatements(tx *sql.Tx) (map[string]*sql.Stmt, error) {
	tableNameToInsertStmt := make(map[string]*sql.Stmt)

	for tableName, tableInfo := range tables {
		stmt, err := tx.Prepare(tableInfo.InsertStatement)

		if err != nil {
			return nil, fmt.Errorf("prepare insert for %s: %w", tableName, err)
		}

		tableNameToInsertStmt[tableName] = stmt
	}

	return tableNameToInsertStmt, nil
}

func writeWordsToUrls(wordsToUrls *data_structures.WordsToUrls, tableNameToInsertStmt map[string]*sql.Stmt) error {
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

func WriteToDatabase(wordsToUrls *data_structures.WordsToUrls) error {
	db, err := sql.Open("sqlite", databaseFileName)

	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	defer db.Close()

	if _, err := db.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		log.Printf("Failed to set journal_mode: %v", err)
	}

	if _, err := db.Exec("PRAGMA synchronous=NORMAL;"); err != nil {
		log.Printf("Failed to set synchronous: %v", err)
	}

	if err := createTables(db); err != nil {
		return err
	}

	tx, err := db.Begin()

	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	defer tx.Rollback()

	tableNameToInsertStmt, err := prepareInsertStatements(tx)

	if err != nil {
		return err
	}

	for _, stmt := range tableNameToInsertStmt {
		defer stmt.Close()
	}

	if err := writeWordsToUrls(wordsToUrls, tableNameToInsertStmt); err != nil {
		return fmt.Errorf("failed to write words to urls: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
