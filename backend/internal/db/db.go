package db

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

const schema = `
CREATE TABLE IF NOT EXISTS scheduler (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date CHAR(8) NOT NULL DEFAULT "",
	title TEXT NOT NULL,
	comment VARCHAR(128) NOT NULL DEFAULT '',
	repeat VARCHAR(128) NOT NULL DEFAULT ''
);
	`

type Db struct {
	DB *sql.DB
}

func (db *Db) createSchema() error {
	if _, err := db.DB.Exec(schema); err != nil {
		return fmt.Errorf("Error: %w", err)
	}

	return nil
}

func (db *Db) Close() error {
	return db.DB.Close()
}

func Init(dbFile string) (*Db, bool, error) {
	var created bool
	_, err := os.Stat(dbFile)

	switch {
	case err == nil:
		created = false
	case errors.Is(err, os.ErrNotExist):
		created = true
	default:
		return nil, false, fmt.Errorf("Error in Init Db %w", err)
	}

	database, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return nil, false, fmt.Errorf("Error in opening Db %w", err)
	}
	if err := database.Ping(); err != nil {
		return nil, false, fmt.Errorf("Error in connecting to Db %w", err)
	}
	storage := &Db{
		DB: database,
	}
	if err := storage.createSchema(); err != nil {
		database.Close()
		return nil, false, err
	}

	return storage, created, nil
}
