package app

import (
	ssogrpc "go-photo/internal/client/sso/grpc"
	"go-photo/internal/config"
	"go-photo/internal/repository"
	photoRepository "go-photo/internal/repository/photo"
	"go-photo/internal/service"
	"go-photo/internal/service/auth"
	photoService "go-photo/internal/service/photo"
	"go-photo/internal/service/token"
	pkgRepo "go-photo/pkg/repository"

	"github.com/jmoiron/sqlx"
)

type serviceProvider struct {
	bc       config.Config
	pgConfig *pkgRepo.PSQLConfig

	photoRepository repository.PhotoRepository

	authService  service.AuthService
	tokenService service.TokenService
	photoService service.PhotoService
}

func newServiceProvider() *serviceProvider {
	return &serviceProvider{}
}

func (s *serviceProvider) PhotoRepository(db *sqlx.DB) repository.PhotoRepository {
	if s.photoRepository == nil {
		s.photoRepository = photoRepository.NewRepository(db)
	}

	return s.photoRepository
}

func (s *serviceProvider) AuthService(client *ssogrpc.Client, appSecret string) service.AuthService {
	if s.authService == nil {
		s.authService = auth.New(client, APP_NAME, appSecret)
	}

	return s.authService
}

// func (s *serviceProvider) UserService(accountClient desc.AuthClient) service.UserService {
// 	if s.userSevice == nil {
// 		s.userSevice = userService.NewService(accountClient, nil)
// 	}

// 	return s.userSevice
// }

func (s *serviceProvider) TokenService(appSecret string) service.TokenService {
	if s.tokenService == nil {
		s.tokenService = token.New(appSecret)
	}

	return s.tokenService
}

func (s *serviceProvider) PhotoService(db *sqlx.DB, storageFolder string) service.PhotoService {
	if s.photoService == nil {
		deps := photoService.Deps{
			StorageFolderPath: storageFolder,
		}
		s.photoService = photoService.NewService(deps, s.PhotoRepository(db), nil)
	}

	return s.photoService
}
