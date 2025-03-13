package main

import (
	"context"
	_ "go-photo/docs"
	"go-photo/internal/app"
	"log"
)

// @title Go-Photo API
// @version 1.0
// @description API Server for Go-Photo app

// @contact.name   Yaroslav Molodcov
// @contact.email  iam@it-yaroslav.ru

// @basePath /

// @securityDefinitions.apikey JWTAuth
// @in header
// @name Authorization
func main() {
	ctx := context.Background()

	a, err := app.NewApp(ctx)
	if err != nil {
		log.Fatalf("failed to init app: %s", err.Error())
	}

	err = a.Run()
	if err != nil {
		log.Fatalf("failed to run app: %s", err.Error())
	}
}
