package photo

import (
	"go-photo/internal/repository"
	def "go-photo/internal/service"
)

// Проверка на соответствие интерфейсу UserService (для статической проверки)
var _ def.PhotoService = (*service)(nil)

type service struct {
	photoRepository repository.PhotoRepository
}

func NewService(photoRepository repository.PhotoRepository) *service {
	return &service{photoRepository: photoRepository}
}
