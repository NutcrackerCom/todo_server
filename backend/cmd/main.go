package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/NutcrackerCom/todo_server/backend/internal/config"
	"github.com/NutcrackerCom/todo_server/backend/internal/db"
	"github.com/NutcrackerCom/todo_server/backend/internal/server"
)

func createLogger() (*log.Logger, func()) {
	logsDir := "logs"
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		log.Fatalf("не удалось создать папку logs: %v", err)
	}

	logPath := filepath.Join(logsDir, "server.log")

	flog, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	logger := log.New(flog, `serv `, log.LstdFlags|log.Lshortfile)

	closeLogger := func() {
		if err := flog.Close(); err != nil {
			log.Printf("не удалось закрыть файл логов: %v", err)
		}
	}

	return logger, closeLogger
}

func main() {

	logger, close := createLogger()
	defer close()
	cfg, err := config.Load()
	if err != nil {
		logger.Fatalf("ошибка загрузки настроек %v", err)
	}

	dataBase, created, err := db.Init(cfg.Db)
	if err != nil {
		logger.Fatalf("ошибка инициализации базы данных %v", err)
	}

	if created {
		logger.Printf("создана база данных")
	} else {
		logger.Printf("открыта база данных ")
	}

	appServer := server.NewServer(logger, cfg.Port, dataBase)

	if err := appServer.Server.ListenAndServe(); err != nil {
		logger.Fatalf("Error %v", err)
	}
}
