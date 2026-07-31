package config

import (
	"fmt"
	"os"
	"strconv"
)

const (
	defaultPort = 7540
	dbPath      = "scheduler.db"
)

type Config struct {
	Port     int
	Db       string
	Password string
}

func Load() (Config, error) {
	port := defaultPort
	db := dbPath
	value := os.Getenv("TODO_PORT")
	if value != "" {
		parse, err := strconv.Atoi(value)
		if err != nil {
			return Config{}, fmt.Errorf("некорректное значение TODO_PORT %q: %w", value, err)
		}
		port = parse
	}

	dbPathVal := os.Getenv("TODO_DBFILE")
	if dbPathVal != "" {
		db = dbPathVal
	}

	return Config{
		Port:     port,
		Db:       db,
		Password: os.Getenv("TODO_PASSWORD"),
	}, nil
}
