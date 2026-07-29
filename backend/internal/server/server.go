package server

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/NutcrackerCom/todo_server/backend/internal/db"
	"github.com/NutcrackerCom/todo_server/backend/internal/handler"
	"github.com/go-chi/chi/v5"
)

type Server struct {
	Logger *log.Logger
	Server http.Server
}

func NewServer(logger *log.Logger, port int, db *db.Db) Server {
	r := chi.NewRouter()
	//r.Get("/api/nextdate", handler.GetNextDate)
	r.Post("/api/task", handler.AddTask(db))
	r.Get("/api/task", handler.GetTask(db))
	r.Get("/api/tasks", handler.GetTasks(db))
	r.Put("/api/task", handler.UpdateTask(db))
	r.Delete("/api/task", handler.DeleteTask(db))
	r.Post("/api/task/done", handler.DoneTask(db))
	r.Handle("/*", http.FileServer(http.Dir("./web")))

	return Server{
		Logger: logger,
		Server: http.Server{
			Addr:         fmt.Sprintf(":%d", port),
			Handler:      r,
			ErrorLog:     logger,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  15 * time.Second,
		},
	}
}
