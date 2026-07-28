package config

import (
	"fmt"
	"os"
	"strconv"
)

const (
	defaultPort = 7540
)

type Config struct {
	Port int
}

func Load() (Config, error) {
	port := defaultPort

	value := os.Getenv("TODO_PORT")
	if value != "" {
		parse, err := strconv.Atoi(value)
		if err != nil {
			return Config{}, fmt.Errorf("Error TODO_PORT %q: %w", value, err)
		}
		port = parse
	}
	return Config{
		Port: port,
	}, nil
}
