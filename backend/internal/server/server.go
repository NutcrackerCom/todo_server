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
	r.Post("/api/signin", handler.SignIn(password, logger))
	r.Group(func(protected chi.Router) {
		protected.Use(handler.RequireAuth(password))
		protected.Get("/api/nextdate", handler.GetNextDate)
		protected.Post("/api/task", handler.AddTask(db, logger))
		protected.Get("/api/task", handler.GetTask(db, logger))
		protected.Get("/api/tasks", handler.GetTasks(db, logger))
		protected.Put("/api/task", handler.UpdateTask(db, logger))
		protected.Delete("/api/task", handler.DeleteTask(db, logger))
		protected.Post("/api/task/done", handler.DoneTask(db, logger))
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
