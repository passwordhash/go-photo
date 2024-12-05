package user

import (
	def "go-photo/internal/service"
	desc "go-photo/pkg/account_v1"
)

// Проверка на соответствие интерфейсу UserService (для статической проверки)
var _ def.UserService = (*Service)(nil)

type Service struct {
	accountClient desc.AccountServiceClient

	//userRepo repository.UserRepository
}

func NewUserService(accountClient desc.AccountServiceClient) *Service {
	return &Service{accountClient: accountClient}
}
