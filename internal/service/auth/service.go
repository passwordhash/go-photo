package auth

import (
	ssogrpc "go-photo/internal/client/sso/grpc"
	def "go-photo/internal/service"
	"log/slog"
)

var _ def.AuthService = (*service)(nil)

type service struct {
	log *slog.Logger

	authClient *ssogrpc.Client

	appName   string
	appSecret string
}

func New(authClient *ssogrpc.Client, appName string, appSecret string) *service {
	return &service{
		log:        authClient.Log,
		authClient: authClient,
		appName:    appName,
		appSecret:  appSecret,
	}
}
