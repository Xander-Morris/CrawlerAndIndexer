package database

import (
	"database/sql"
	"log"
)

func SearchForKeyword(keyword string) {
	db, err := sql.Open("sqlite", databaseFileName)

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()
}