package app

import (
	"go-photo/internal/config"
	"go-photo/internal/service"
	userService "go-photo/internal/service/user"
	"log"
)

type serviceProvider struct {
	bc config.Config
	// services, repositories, etc.

	userSevice service.UserService
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

func (s *serviceProvider) UserService() service.UserService {
	if s.userSevice == nil {
		s.userSevice = userService.NewService()
	}

	return s.userSevice
}

// getters ...
