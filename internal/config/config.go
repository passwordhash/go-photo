package config

import (
	"errors"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"net"
	"os"
	"path/filepath"
)

const (
	httpPortEnvName   = "HTTP_PORT"
	logLevelEnvName   = "LOG_LEVEL"
	grpcAddrEnvName   = "GRPC_ADDR"
	storageFolderPath = "STORAGE_FOLDER"
)

type Config interface {
	HTTPAddr() string
	GRPCAddr() string

	LogLevel() string

	StorageFolder() string
}

type baseConfig struct {
	httpPort          string
	grpcAddr          string
	logLevel          string
	storageFolderPath string
}

func NewConfig() (Config, error) {
	port := os.Getenv(httpPortEnvName)
	if len(port) == 0 {
		return nil, errors.New("http port not found")
	}

	logLever := os.Getenv(logLevelEnvName)

	grpcAddr := os.Getenv(grpcAddrEnvName)
	if len(grpcAddr) == 0 {
		return nil, errors.New("grpc addr not found")
	}

	storageFolder := os.Getenv(storageFolderPath)
	if len(storageFolder) == 0 {
		storageFolder = DefaultStorageFolderPath
	}

	return &baseConfig{
		httpPort:          port,
		grpcAddr:          grpcAddr,
		logLevel:          logLever,
		storageFolderPath: storageFolder,
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
	return net.JoinHostPort("0.0.0.0", c.httpPort)
}

func (c *baseConfig) GRPCAddr() string {
	return c.grpcAddr
}

func (c *baseConfig) LogLevel() string {
	return c.logLevel
}

func (c *baseConfig) StorageFolder() string {
	wd, err := os.Getwd()
	if err != nil {
		log.Panicf("failed to get working directory: %v", err)
	}

	// TODO: решить как правильно хранить путь к папке
	return filepath.Join(wd, c.storageFolderPath)
}
