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

func NewServer(logger *log.Logger, port int, db *db.Db, password string) Server {
	r := chi.NewRouter()
	r.Post("/api/signin", handler.SignIn(password))
	r.Group(func(protected chi.Router) {
		protected.Use(handler.RequireAuth(password))
		protected.Get("/api/nextdate", handler.GetNextDate)
		protected.Post("/api/task", handler.AddTask(db))
		protected.Get("/api/task", handler.GetTask(db))
		protected.Get("/api/tasks", handler.GetTasks(db))
		protected.Put("/api/task", handler.UpdateTask(db))
		protected.Delete("/api/task", handler.DeleteTask(db))
		protected.Post("/api/task/done", handler.DoneTask(db))
	})

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
