package user

import (
	def "go-photo/internal/service"
	desc "go-photo/pkg/account_v1"
)

// Проверка на соответствие интерфейсу UserService (для статической проверки)
var _ def.UserService = (*service)(nil)

type service struct {
	accountClient desc.AccountServiceClient

	//userRepo repository.UserRepository
}

func NewService(accountClient desc.AccountServiceClient) *service {
	return &service{accountClient: accountClient}
}
