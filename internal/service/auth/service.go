package auth

import (
	def "go-photo/internal/service"
	"log/slog"

	gossov1 "github.com/passwordhash/protos/gen/go/go-sso"
)

var _ def.AuthService = (*service)(nil)

type service struct {
	log *slog.Logger

	// authClient *ssogrpc.Client
	authAPI gossov1.AuthClient

	appName   string
	appSecret string
}

// func New(authClient *ssogrpc.Client, appName string, appSecret string) *service {
func New(log *slog.Logger, authAPI gossov1.AuthClient, appName string, appSecret string) *service {
	return &service{
		log:       log,
		authAPI:   authAPI,
		appName:   appName,
		appSecret: appSecret,
	}
}
