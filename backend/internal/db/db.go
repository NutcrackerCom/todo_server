package db

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

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

func (db *Db) AddTask(task *Task) (int64, error) {
	const query = `
		INSERT INTO scheduler (
			date,
			title,
			comment,
			repeat
		)
		VALUES (?, ?, ?, ?)
	`
	result, err := db.DB.Exec(
		query,
		task.Date,
		task.Title,
		task.Comment,
		task.Repeat,
	)
	if err != nil {
		return 0, fmt.Errorf("couldn't add task: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("couldn't get the issue ID: %w", err)
	}

	return id, nil
}

func (s *Db) GetTask(id string) (*Task, error) {
	const query = `
		SELECT id, date, title, comment, repeat
		FROM scheduler
		WHERE id = ?
	`

	var task Task

	err := s.DB.QueryRow(query, id).Scan(
		&task.ID,
		&task.Date,
		&task.Title,
		&task.Comment,
		&task.Repeat,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("issue not found")
	}

	if err != nil {
		return nil, fmt.Errorf(
			"couldn't get the issue: %w",
			err,
		)
	}

	return &task, nil
}

func (s *Db) Tasks(search string, limit int) ([]*Task, error) {
	if limit <= 0 {
		limit = 50
	}

	tasks := make([]*Task, 0)

	var (
		query string
		args  []any
	)

	switch {
	case search == "":
		query = `
			SELECT id, date, title, comment, repeat
			FROM scheduler
			ORDER BY date, id
			LIMIT ?
		`
		args = []any{limit}

	default:
		searchDate, err := time.Parse("02.01.2006", search)
		if err == nil {
			query = `
				SELECT id, date, title, comment, repeat
				FROM scheduler
				WHERE date = ?
				ORDER BY date, id
				LIMIT ?
			`
			args = []any{
				searchDate.Format("20060102"),
				limit,
			}
		} else {
			query = `
				SELECT id, date, title, comment, repeat
				FROM scheduler
				WHERE title LIKE ? OR comment LIKE ?
				ORDER BY date, id
				LIMIT ?
			`

			searchPattern := "%" + search + "%"

			args = []any{
				searchPattern,
				searchPattern,
				limit,
			}
		}
	}

	rows, err := s.DB.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("couldn't get a list of issues: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var task Task

		if err := rows.Scan(
			&task.ID,
			&task.Date,
			&task.Title,
			&task.Comment,
			&task.Repeat,
		); err != nil {
			return nil, fmt.Errorf("couldn't read the issue: %w", err)
		}
		tasks = append(tasks, &task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error reading the task list: %w", err)
	}

	return tasks, nil
}
