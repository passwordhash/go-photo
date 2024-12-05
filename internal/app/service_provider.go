package app

import (
	"go-photo/internal/config"
	"go-photo/internal/service"
	photoService "go-photo/internal/service/photo"
	userService "go-photo/internal/service/user"
	desc "go-photo/pkg/account_v1"
	"log"
)

type serviceProvider struct {
	bc config.Config

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

func (s *serviceProvider) UserService(accountClient desc.AccountServiceClient) service.UserService {
	if s.userSevice == nil {
		s.userSevice = userService.NewUserService(accountClient)
	}

	return s.userSevice
}

func (s *serviceProvider) PhotoService() service.PhotoService {
	if s.photoService == nil {
		s.photoService = photoService.NewPhotoService()
	}

	return s.photoService
}
