package app

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"go-photo/internal/config"
	"go-photo/internal/handler"
	"go-photo/internal/handler/v1/photos"
	"go-photo/internal/handler/v1/user"
	"os"
	"time"
)

type App struct {
	httpServer *gin.Engine
	sp         *serviceProvider
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
	folders := []string{config.PhotosDir, config.LogsDir}

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

func (a *App) initHTTPServer(_ context.Context) error {
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

	usersHandler := user.NewUserHandler(a.sp.UserService())
	photosHandler := photos.NewPhotosHandler(a.sp.PhotoService())

	usersHandler.RegisterRoutes(v1)
	photosHandler.RegisterRoutes(v1)

	a.httpServer = router

	return nil
}

func (a *App) runHTTPServer() error {
	return a.httpServer.Run(a.sp.BaseConfig().HTTPAddr())
}
