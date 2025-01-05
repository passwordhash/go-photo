package app

import (
	"github.com/jmoiron/sqlx"
	"go-photo/internal/config"
	"go-photo/internal/repository"
	photoRepository "go-photo/internal/repository/photo"
	"go-photo/internal/service"
	photoService "go-photo/internal/service/photo"
	userService "go-photo/internal/service/user"
	desc "go-photo/pkg/account_v1"
	pkgRepo "go-photo/pkg/repository"
	"log"
)

type serviceProvider struct {
	bc       config.Config
	pgConfig *pkgRepo.PSQLConfig

	photoRepository repository.PhotoRepository

	userSevice   service.UserService
	photoService service.PhotoService
}

func newServiceProvider() *serviceProvider {
	return &serviceProvider{}
}

func (s *serviceProvider) BaseConfig() config.Config {
	if s.bc == nil {
		cfg, err := config.NewConfig()
		if err != nil {
			log.Fatalf("failed to get base config: %s", err.Error())
		}

		s.bc = cfg
	}

	return s.bc
}

func (s *serviceProvider) PSQLConfig() pkgRepo.PSQLConfig {
	if s.pgConfig == nil {
		cfg, err := config.NewPSQLConfig()
		if err != nil {
			log.Fatalf("failed to get psql config: %s", err.Error())
		}

		s.pgConfig = &cfg
	}

	return *s.pgConfig
}

func (s *serviceProvider) PhotoRepository(db *sqlx.DB) repository.PhotoRepository {
	if s.photoRepository == nil {
		s.photoRepository = photoRepository.NewRepository(db)
	}

	return s.photoRepository
}

func (s *serviceProvider) UserService(accountClient desc.AccountServiceClient) service.UserService {
	if s.userSevice == nil {
		s.userSevice = userService.NewService(accountClient, nil)
	}

	return s.userSevice
}

func (s *serviceProvider) TokenService(accountClient desc.AccountServiceClient) service.TokenService {
	return userService.NewService(accountClient, nil)
}

func (s *serviceProvider) PhotoService(db *sqlx.DB) service.PhotoService {
	if s.photoService == nil {
		deps := photoService.Deps{
			StorageFolderPath: s.BaseConfig().StorageFolder(),
		}
		s.photoService = photoService.NewService(deps, s.PhotoRepository(db))
	}

	return s.photoService
}
