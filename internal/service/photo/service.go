package photo

import (
	"go-photo/internal/repository"
	def "go-photo/internal/service"
)

// Проверка на соответствие интерфейсу UserService (для статической проверки)
var _ def.PhotoService = (*service)(nil)

type Deps struct {
	// абсолютный путь к папке с фотографиями
	StorageFolderPath string
}

type service struct {
	d               Deps
	photoRepository repository.PhotoRepository
}

func NewService(d Deps, photoRepository repository.PhotoRepository) *service {
	return &service{d: d, photoRepository: photoRepository}
}
