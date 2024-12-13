package app

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"go-photo/internal/config"
	"go-photo/internal/handler"
	"go-photo/internal/handler/v1/photos"
	"go-photo/internal/handler/v1/user"
	desc "go-photo/pkg/account_v1"
	"go-photo/pkg/repository"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"os"
	"time"
)

type App struct {
	grpcClient desc.AccountServiceClient
	httpServer *gin.Engine

	db *sqlx.DB

	sp *serviceProvider
}

func NewApp(ctx context.Context) (*App, error) {
	a := &App{}

	err := a.initDeps(ctx)
	if err != nil {
		return nil, err
	}

	return a, nil
}

func (a *App) Run() error {
	return a.runHTTPServer()
}

func (a *App) initDeps(ctx context.Context) error {
	inits := []func(context.Context) error{
		a.initConfig,
		a.initServiceProvider,
		a.initFolders,
		a.initLogging,
		a.initPGConnection,
		a.initGRPCClient,
		a.initHTTPServer,
	}

	for _, f := range inits {
		err := f(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *App) initConfig(_ context.Context) error {
	err := config.Load(".env")
	if err != nil {
		return err
	}

	return nil
}

func (a *App) initServiceProvider(_ context.Context) error {
	a.sp = newServiceProvider()
	return nil
}

func (a *App) initFolders(_ context.Context) error {
	folders := []string{a.sp.BaseConfig().StorageFolder(), config.LogsDir}

	// TODO: move to utils
	for _, folder := range folders {
		if _, err := os.Stat(folder); os.IsNotExist(err) {
			err := os.MkdirAll(folder, os.ModePerm)
			if err != nil {
				return fmt.Errorf("unable to create folder %s: %w", folder, err)
			}
		}
	}
	return nil
}

func (a *App) initLogging(_ context.Context) error {
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.TextFormatter{})

	logLevel, err := log.ParseLevel(a.sp.BaseConfig().LogLevel())
	if err != nil {
		log.Printf("failed to parse log level: %v", err)
		logLevel = log.DebugLevel
	}

	log.SetLevel(logLevel)
	return nil
}

func (a *App) initPGConnection(_ context.Context) error {
	pgConfig := a.sp.PSQLConfig()
	db, err := repository.NewPostgresDB(pgConfig)
	if err != nil {
		return fmt.Errorf("failed to create postgres connection: %w with config: %v", err, pgConfig)
	}

	a.db = db

	return nil
}

func (a *App) initGRPCClient(_ context.Context) error {
	conn, err := grpc.NewClient(a.sp.BaseConfig().GRPCAddr(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to create grpc client: %w", err)
	}

	a.grpcClient = desc.NewAccountServiceClient(conn)

	return nil
}

func (a *App) initHTTPServer(_ context.Context) error {
	if a.grpcClient == nil {
		return fmt.Errorf("grpc client is not initialized")
	}

	router := gin.New()

	router.Use(gin.Recovery())
	router.Use(handler.Logger())

	router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))

	api := router.Group("/api")
	v1 := api.Group("/v1")

	usersHandler := user.NewUserHandler(a.sp.UserService(a.grpcClient))
	photosHandler := photos.NewPhotosHandler(a.sp.PhotoService(a.db))

	usersHandler.RegisterRoutes(v1)
	photosHandler.RegisterRoutes(v1)

	a.httpServer = router

	return nil
}

func (a *App) runHTTPServer() error {
	return a.httpServer.Run(a.sp.BaseConfig().HTTPAddr())
}
