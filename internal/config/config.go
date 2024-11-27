package config

import (
	"errors"
	"github.com/joho/godotenv"
	"net"
	"os"
)

const (
	httpPortEnvName = "HTTP_PORT"
)

type Config interface {
	HTTPAddr() string
}

type baseConfig struct {
	httpPort string
}

func NewConfig() (Config, error) {
	port := os.Getenv(httpPortEnvName)
	if len(port) == 0 {
		return nil, errors.New("http port not found")
	}

	return &baseConfig{
		httpPort: port,
	}, nil
}

func Load(path string) error {
	err := godotenv.Load(path)
	if err != nil {
		return err
	}

	return nil
}

func (c *baseConfig) HTTPAddr() string {
	return net.JoinHostPort("localhost", c.httpPort)
}
