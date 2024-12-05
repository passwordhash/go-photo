package config

import (
	"errors"
	"go-photo/pkg/repository"
	"os"
)

func NewPSQLConfig() (repository.PSQLConfig, error) {
	host := os.Getenv("POSTGRES_HOST")
	if len(host) == 0 {
		host = PostgresDefaultHost
	}

	port := os.Getenv("POSTGRES_PORT")
	if len(port) == 0 {
		port = PostgresDefaultPort
	}

	username := os.Getenv("POSTGRES_USER")
	if len(username) == 0 {
		return repository.PSQLConfig{}, errors.New("POSTGRES_USER is not set")
	}

	password := os.Getenv("POSTGRES_PASSWORD")
	if len(password) == 0 {
		return repository.PSQLConfig{}, errors.New("POSTGRES_PASSWORD is not set")
	}

	dbname := os.Getenv("POSTGRES_DB")
	if len(dbname) == 0 {
		return repository.PSQLConfig{}, errors.New("POSTGRES_DB is not set")
	}

	sslmode := os.Getenv("POSTGRES_SSL_MODE")
	if len(sslmode) == 0 {
		sslmode = PostgresDefaultSSLMode
	}

	return repository.PSQLConfig{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		DBName:   dbname,
		SSLMode:  sslmode,
	}, nil
}
