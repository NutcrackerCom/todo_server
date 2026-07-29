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
		return fmt.Errorf("не удалось создать структуру базы: %w", err)
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
		return nil, false, fmt.Errorf("не удалось проверить файл базы %w", err)
	}

	database, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return nil, false, fmt.Errorf("не удалось открыть базу %w", err)
	}
	if err := database.Ping(); err != nil {
		return nil, false, fmt.Errorf("не удалось подключиться к базе %w", err)
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
		return 0, fmt.Errorf("не удалось добавить задачу: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("не удалось получить идентификатор задачи: %w", err)
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
		return nil, fmt.Errorf("задача не найдена")
	}

	if err != nil {
		return nil, fmt.Errorf(
			"не удалось получить задачу: %w",
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
		return nil, fmt.Errorf("не удалось получить список задач: %w", err)
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
			return nil, fmt.Errorf("не удалось прочитать задач: %w", err)
		}
		tasks = append(tasks, &task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка чтения списка задач: %w", err)
	}

	return tasks, nil
}

func (s *Db) UpdateTask(task *Task) error {
	const query = `
		UPDATE scheduler
		SET
			date = ?,
			title = ?,
			comment = ?,
			repeat = ?
		WHERE id = ?
	`

	result, err := s.DB.Exec(
		query,
		task.Date,
		task.Title,
		task.Comment,
		task.Repeat,
		task.ID,
	)
	if err != nil {
		return fmt.Errorf(
			"не удалось изменить задачу: %w",
			err,
		)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("не удалось определить количество изменённых задач: %w", err)
	}

	if count == 0 {
		return fmt.Errorf("задача не найдена")
	}

	return nil
}

func (s *Db) DeleteTask(id string) error {
	const query = `
		DELETE FROM scheduler
		WHERE id = ?
	`

	result, err := s.DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("не удалось удалить задачу: %w", err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("не удалось определить количество удалённых задач: %w", err)
	}

	if count == 0 {
		return fmt.Errorf("задача не найдена")
	}

	return nil
}

func (s *Db) UpdateDate(nextDate, id string) error {
	const query = `
		UPDATE scheduler
		SET date = ?
		WHERE id = ?
	`

	result, err := s.DB.Exec(
		query,
		nextDate,
		id,
	)
	if err != nil {
		return fmt.Errorf("не удалось изменить дату задачи: %w", err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("не удалось определить количество изменённых задач: %w", err)
	}

	if count == 0 {
		return fmt.Errorf("задача не найдена")
	}

	return nil
}
