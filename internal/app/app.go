package app

import (
	"context"
	"fmt"
	ssogrpc "go-photo/internal/client/sso/grpc"
	"go-photo/internal/config"
	"go-photo/internal/handler/middleware"
	"go-photo/internal/handler/v1/auth"
	"go-photo/internal/handler/v1/docs"
	"go-photo/internal/handler/v1/photos"
	"go-photo/internal/handler/v1/public"
	"go-photo/pkg/repository"
	"log/slog"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/lmittmann/tint"
	gossov1 "github.com/passwordhash/protos/gen/go/go-sso"
)

const APP_NAME = "go-photo"

type App struct {
	log *slog.Logger
	cfg config.Config

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
		// a.initFolders,
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
		slog.Warn(fmt.Sprintf("failed to load config: %v", err))
		slog.Info("loading without .env")
	}

	cfg, err := config.NewConfig()
	if err != nil {
		return fmt.Errorf("failed to create config: %w", err)
	}

	a.cfg = cfg

	return nil
}
func (a *App) initLogging(_ context.Context) error {
	w := os.Stdout

	// dev
	a.log = slog.New(tint.NewHandler(w, &tint.Options{
		Level:      slog.LevelInfo,
		TimeFormat: time.TimeOnly,
	}))

	return nil
}

func (a *App) initGRPCClient(ctx context.Context) error {
	// TODO: timeout from config
	client, err := ssogrpc.New(ctx, a.log,
		a.cfg.GRPCAddr(),
		time.Duration(a.cfg.GRPCTimeout()),
		3)
	if err != nil {
		return fmt.Errorf("failed to create grpc client: %w", err)
	}

	resp, err := client.Api.SigningKey(ctx, &gossov1.SigningKeyRequest{
		AppName: APP_NAME,
	})
	if err != nil {
		return fmt.Errorf("failed to get signing key: %w", err)
	}

	a.ssoClient = client
	a.cfg.SetAppSecret(resp.SigningKey)

	// TODO: health check grpc client

	return nil
}

func (a *App) initServiceProvider(_ context.Context) error {
	a.sp = newServiceProvider()
	return nil
}

// TODO: решить нужно ли это
// func (a *App) initFolders(_ context.Context) error {
// 	folders := []string{a.sp.BaseConfig().StorageFolder(), config.LogsDir}

// 	// TODO: move to utils
// 	for _, folder := range folders {
// 		if _, err := os.Stat(folder); os.IsNotExist(err) {
// 			err := os.MkdirAll(folder, os.ModePerm)
// 			if err != nil {
// 				return fmt.Errorf("unable to create folder %s: %w", folder, err)
// 			}
// 		}
// 	}
// 	return nil
// }

func (a *App) initPGConnection(_ context.Context) error {
	pgConfig, err := config.NewPSQLConfig()
	if err != nil {
		return fmt.Errorf("failed to get psql config: %s", err.Error())
	}

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
	router.Use(middleware.Logger(a.log))

	base := router.Group("/")

	publicHandler := public.NewHandler(a.sp.PhotoService(a.db, a.cfg.StorageFolder()))
	publicHandler.RegisterRoutes(base)

	api := router.Group("/api")
	v1 := api.Group("/v1")

	docsHandler := docs.NewHandler()
	authHandler := auth.NewHandler(a.sp.AuthService(a.ssoClient, a.cfg.AppSecret()))
	// usersHandler := user.NewHandler(a.sp.UserService(a.grpcClient))
	photosHandler := photos.NewHandler(
		a.sp.PhotoService(a.db, a.cfg.StorageFolder()),
		a.sp.TokenService(a.cfg.AppSecret()),
	)

	docsHandler.RegisterRoutes(v1)
	authHandler.RegisterRoutes(v1)
	// usersHandler.RegisterRoutes(v1)
	photosHandler.RegisterRoutes(v1)

	a.httpServer = router

	return nil
}

func (a *App) runHTTPServer() error {
	return a.httpServer.Run(a.cfg.HTTPAddr())
}
