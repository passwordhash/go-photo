package config

import (
	"errors"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

const (
	httpPortEnvName    = "HTTP_PORT"
	logLevelEnvName    = "LOG_LEVEL"
	grpcAddrEnvName    = "GRPC_ADDR"
	grpcTimeoutEnvName = "GRPC_TIMEOUT"
	storageFolderPath  = "STORAGE_FOLDER"
)

type Config interface {
	AppSecret() string
	SetAppSecret(secret string)

	HTTPAddr() string
	GRPCAddr() string
	GRPCTimeout() time.Duration

	LogLevel() string

	StorageFolder() string
}

type baseConfig struct {
	httpPort          string
	grpcAddr          string
	grpcTimeout       time.Duration
	logLevel          string
	storageFolderPath string
	clients           clientsConfig
	appSecret         string
}

type clientsConfig struct {
	sso grpcClient
}

type grpcClient struct {
	address      string
	timeout      time.Duration
	retriesCount int
	Insecure     bool
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

	grpcTimeout := os.Getenv(grpcTimeoutEnvName)
	if len(grpcTimeout) == 0 {
		return nil, errors.New("grpc timeout not found")
	}
	grpcTimeoutI, err := strconv.Atoi(grpcTimeout)
	if err != nil {
		return nil, errors.New("grpc timeout invalid")
	}

	storageFolder := os.Getenv(storageFolderPath)
	if len(storageFolder) == 0 {
		storageFolder = DefaultStorageFolderPath
	}

	return &baseConfig{
		httpPort:          port,
		grpcAddr:          grpcAddr,
		grpcTimeout:       time.Duration(grpcTimeoutI) * time.Second,
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

func (c *baseConfig) AppSecret() string {
	return c.appSecret
}

func (c *baseConfig) SetAppSecret(secret string) {
	c.appSecret = secret
}

func (c *baseConfig) HTTPAddr() string {
	return net.JoinHostPort("0.0.0.0", c.httpPort)
}

func (c *baseConfig) GRPCAddr() string {
	return c.grpcAddr
}

func (c *baseConfig) GRPCTimeout() time.Duration {
	return c.grpcTimeout
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
