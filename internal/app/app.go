package app

import (
	"context"
	"fmt"
	ssogrpc "go-photo/internal/client/sso/grpc"
	"go-photo/internal/config"
	"go-photo/internal/handler/middleware"
	"go-photo/internal/handler/v1/auth"
	"go-photo/internal/handler/v1/docs"
	"go-photo/internal/handler/v1/public"
	"go-photo/pkg/repository"
	"log/slog"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

type App struct {
	log *slog.Logger

	ssoClient  *ssogrpc.Client
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
		a.initLogging,
		a.initGRPCClient,
		a.initServiceProvider,
		// TODO: см. ниже
		a.initFolders,
		a.initPGConnection,
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
		log.Warnf("failed to load config: %v", err)
		log.Info("loading without .env")
	}

	cfg, err := config.NewConfig()
	if err != nil {
		return fmt.Errorf("failed to create config: %w", err)
	}

	a.cfg = cfg

	return nil
}
func (a *App) initLogging(_ context.Context) error {
	// TODO: подчистить
	// log.SetOutput(os.Stdout)
	// //log.SetFormatter(&log.TextFormatter{
	// //	ForceColors: true,
	// //})
	// log.SetFormatter(&config.CustomFormatter{
	// 	TimestampFormat: time.DateTime,
	// })

	// logLevel, err := log.ParseLevel(a.sp.BaseConfig().LogLevel())
	// if err != nil {
	// 	log.Printf("failed to parse log level: %v", err)
	// 	log.Printf("use default log level: %s", log.DebugLevel)
	// 	logLevel = log.DebugLevel
	// }

	// log.SetLevel(logLevel)

	jsonHandler := slog.NewJSONHandler(os.Stdout, nil)

	a.log = slog.New(jsonHandler)

	return nil
}

func (a *App) initGRPCClient(ctx context.Context) error {
	fmt.Println("adsfasdfasdfas")
	// TODO: timeout from config
	client, err := ssogrpc.New(ctx, a.log, a.sp.BaseConfig().GRPCAddr(), time.Second, 3)
	if err != nil {
		return fmt.Errorf("failed to create grpc client: %w", err)
	}

	a.ssoClient = client

	// TODO: health check grpc client

	return nil
}

func (a *App) initServiceProvider(_ context.Context) error {
	a.sp = newServiceProvider()
	return nil
}

// TODO: решить нужно ли это
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

func (a *App) initPGConnection(_ context.Context) error {
	pgConfig := a.sp.PSQLConfig()
	db, err := repository.NewPostgresDB(pgConfig)
	if err != nil {
		return fmt.Errorf("failed to create postgres connection: %w with config: %v", err, pgConfig)
	}

	a.db = db

	return nil
}

func (a *App) initHTTPServer(_ context.Context) error {
	if a.ssoClient == nil {
		return fmt.Errorf("grpc client is not initialized")
	}

	router := gin.New()

	router.Use(gin.Recovery())
	router.Use(middleware.Logger())

	base := router.Group("/")

	publicHandler := public.NewHandler(a.sp.PhotoService(a.db))
	publicHandler.RegisterRoutes(base)

	api := router.Group("/api")
	v1 := api.Group("/v1")

	docsHandler := docs.NewHandler()
	authHandler := auth.NewHandler(a.sp.AuthService(a.ssoClient))
	// usersHandler := user.NewHandler(a.sp.UserService(a.grpcClient))
	// photosHandler := photos.NewHandler(a.sp.PhotoService(a.db), a.sp.TokenService(a.grpcClient))

	docsHandler.RegisterRoutes(v1)
	authHandler.RegisterRoutes(v1)
	// usersHandler.RegisterRoutes(v1)
	// photosHandler.RegisterRoutes(v1)

	a.httpServer = router

	return nil
}

func (a *App) runHTTPServer() error {
	return a.httpServer.Run(a.sp.BaseConfig().HTTPAddr())
}
