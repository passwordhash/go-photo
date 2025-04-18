package photo

import (
	"go-photo/internal/repository"
	def "go-photo/internal/service"
	"go-photo/internal/utils"
)

// Проверка на соответствие интерфейсу UserService (для статической проверки)
var _ def.PhotoService = (*service)(nil)

type Deps struct {
	// абсолютный путь к папке с фотографиями
	StorageFolderPath string
}

type service struct {
	d               Deps
	utils           utils.Interface
	photoRepository repository.PhotoRepository
}

func NewService(d Deps, photoRepository repository.PhotoRepository, u utils.Interface) *service {
	if u == nil {
		u = utils.New()
	}
	return &service{d: d, utils: u, photoRepository: photoRepository}
}
