package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/NutcrackerCom/todo_server/backend/internal/db"
	nextdate "github.com/NutcrackerCom/todo_server/backend/pkg/api"
)

type taskResponse struct {
	ID    string `json:"id,omitempty"`
	Error string `json:"error,omitempty"`
}

type tasksResponse struct {
	Tasks []*db.Task `json:"tasks"`
}

func GetNextDate(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	nowValue := query.Get("now")
	dateValue := query.Get("date")
	repeatValue := query.Get("repeat")

	now := time.Now()
	if nowValue != "" {
		parsedNow, err := time.Parse(nextdate.DateLayout, nowValue)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		now = parsedNow
	}
	nextDate, err := nextdate.NextDate(now, dateValue, repeatValue)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(nextDate))
}

func writeJSON(w http.ResponseWriter, data any) {
	w.Header().Set(
		"Content-Type",
		"application/json; charset=UTF-8",
	)

	_ = json.NewEncoder(w).Encode(data)
}

func writeTaskError(w http.ResponseWriter, err error) {
	writeJSON(w, taskResponse{
		Error: err.Error(),
	})
}

func checkTaskDate(task *db.Task) error {
	now := time.Now()
	todayString := now.Format(nextdate.DateLayout)
	if task.Date == "" {
		task.Date = todayString
	}
	taskDate, err := time.Parse(nextdate.DateLayout, task.Date)
	if err != nil {
		return fmt.Errorf("некорректная дата %q: %w", task.Date, err)
	}
	today, err := time.Parse(nextdate.DateLayout, todayString)
	if err != nil {
		return fmt.Errorf("не удалось определить текущую дату: %w", err)
	}

	var nextDate string
	if task.Repeat != "" {
		nextDate, err = nextdate.NextDate(now, task.Date, task.Repeat)
		if err != nil {
			return fmt.Errorf("некорректное правило повторения: %w", err)
		}
	}

	if taskDate.Before(today) {
		if task.Repeat == "" {
			task.Date = todayString
		} else {
			task.Date = nextDate
		}
	}

	return nil
}

func AddTask(storage *db.Db) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var task db.Task
		if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
			writeTaskError(w, fmt.Errorf("ошибка чтения JSON: %w", err))
			return
		}
		if strings.TrimSpace(task.Title) == "" {
			writeTaskError(w, fmt.Errorf("не указан заголовок задачи"))
			return
		}
		if err := checkTaskDate(&task); err != nil {
			writeTaskError(w, err)
			return
		}
		id, err := storage.AddTask(&task)
		if err != nil {
			writeTaskError(w, err)
			return
		}

		writeJSON(w, taskResponse{
			ID: strconv.FormatInt(id, 10),
		})
	}
}

func GetTask(storage *db.Db) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimSpace(r.URL.Query().Get("id"))

		if id == "" {
			writeTaskError(w, fmt.Errorf("не указан идентификатор"))
			return
		}

		task, err := storage.GetTask(id)
		if err != nil {
			writeTaskError(w, err)
			return
		}
		writeJSON(w, task)
	}
}

func GetTasks(storage *db.Db) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		search := r.URL.Query().Get("search")

		tasks, err := storage.Tasks(search, 50)
		if err != nil {
			writeTaskError(w, err)
			return
		}

		writeJSON(w, tasksResponse{
			Tasks: tasks,
		})
	}
}

func UpdateTask(storage *db.Db) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var task db.Task

		if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
			writeTaskError(
				w,
				fmt.Errorf("ошибка чтения JSON: %w", err),
			)
			return
		}

		task.ID = strings.TrimSpace(task.ID)
		task.Title = strings.TrimSpace(task.Title)
		task.Repeat = strings.TrimSpace(task.Repeat)

		if task.ID == "" {
			writeTaskError(
				w,
				fmt.Errorf("не указан идентификатор"),
			)
			return
		}

		if task.Title == "" {
			writeTaskError(
				w,
				fmt.Errorf("не указан заголовок задачи"),
			)
			return
		}

		if err := checkTaskDate(&task); err != nil {
			writeTaskError(w, err)
			return
		}

		if err := storage.UpdateTask(&task); err != nil {
			writeTaskError(w, err)
			return
		}

		writeJSON(w, struct{}{})
	}
}

func DeleteTask(storage *db.Db) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimSpace(
			r.URL.Query().Get("id"),
		)

		if id == "" {
			writeTaskError(w, fmt.Errorf("не указан идентификатор"))
			return
		}

		if err := storage.DeleteTask(id); err != nil {
			writeTaskError(w, err)
			return
		}

		writeJSON(w, struct{}{})
	}
}

func DoneTask(storage *db.Db) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimSpace(
			r.URL.Query().Get("id"),
		)

		if id == "" {
			writeTaskError(w, fmt.Errorf("не указан идентификатор"))
			return
		}

		task, err := storage.GetTask(id)
		if err != nil {
			writeTaskError(w, err)
			return
		}

		if task.Repeat == "" {
			if err := storage.DeleteTask(id); err != nil {
				writeTaskError(w, err)
				return
			}

			writeJSON(w, struct{}{})
			return
		}

		nextDate, err := nextdate.NextDate(
			time.Now(),
			task.Date,
			task.Repeat,
		)
		if err != nil {
			writeTaskError(w, fmt.Errorf("не удалось вычислить следующую дату: %w", err))
			return
		}

		if err := storage.UpdateDate(nextDate, id); err != nil {
			writeTaskError(w, err)
			return
		}

		writeJSON(w, struct{}{})
	}
}
