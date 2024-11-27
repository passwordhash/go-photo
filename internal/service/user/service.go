package user

import (
	def "go-photo/internal/service"
)

// Проверка на соответствие интерфейсу UserService (для статической проверки)
var _ def.UserService = (*Service)(nil)

type Service struct {
	//userRepo repository.UserRepository
}

func NewService() *Service {
	return &Service{}
}
