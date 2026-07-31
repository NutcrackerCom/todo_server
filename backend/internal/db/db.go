package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
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

var (
	ErrTaskNotFound = errors.New("задача не найдена")
	ErrGetTask      = errors.New("не удалось получить задачу")
	ErrDeleteTask   = errors.New("не удалось удалить задачу")
	ErrUpdateTask   = errors.New("не удалось изменить задачу")
)

type Db struct {
	DB *sql.DB
}

func (db *Db) createSchema() error {
	if _, err := db.DB.Exec(schema); err != nil {
		return fmt.Errorf("не удалось создать структуру базы")
	}

	return nil
}

func (db *Db) Close() error {
	return db.DB.Close()
}

func Init(dbFile string, logger *log.Logger) (*Db, bool, error) {
	var created bool
	_, err := os.Stat(dbFile)

	switch {
	case err == nil:
		created = false
	case errors.Is(err, os.ErrNotExist):
		created = true
	default:
		logger.Printf("не удалось открыть базу %v", err)
		return nil, false, fmt.Errorf("не удалось проверить файл базы")
	}

	database, err := sql.Open("sqlite", dbFile)
	if err != nil {
		logger.Printf("не удалось открыть базу: %v", err)
		return nil, false, fmt.Errorf("не удалось открыть базу")
	}
	if err := database.Ping(); err != nil {
		logger.Printf("не удалось подключиться к базе: %v", err)
		return nil, false, fmt.Errorf("не удалось подключиться к базе")
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
		return 0, fmt.Errorf("не удалось добавить задачу")
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("не удалось получить идентификатор задачи")
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
		return nil, ErrTaskNotFound
	}

	if err != nil {
		return nil, ErrGetTask
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
		return nil, fmt.Errorf("не удалось получить список задач")
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
			return nil, fmt.Errorf("не удалось прочитать задач")
		}
		tasks = append(tasks, &task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка чтения списка задач")
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
		return ErrUpdateTask
	}

	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("не удалось определить количество изменённых задач")
	}

	if count == 0 {
		return ErrTaskNotFound
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
		return ErrDeleteTask
	}

	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("не удалось определить количество удалённых задач")
	}

	if count == 0 {
		return ErrTaskNotFound
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
		return fmt.Errorf("не удалось изменить дату задачи")
	}

	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("не удалось определить количество изменённых задач")
	}

	if count == 0 {
		return ErrTaskNotFound
	}

	return nil
}
