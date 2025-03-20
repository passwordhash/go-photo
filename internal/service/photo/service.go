package photo

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"go-photo/internal/repository"
	repoErr "go-photo/internal/repository/error"
	def "go-photo/internal/service"
	serviceErr "go-photo/internal/service/error"
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

func (s *service) HandleError(err error) error {
	if errors.Is(err, repoErr.NotFoundError) {
		return fmt.Errorf("%w: %v", serviceErr.PhotoNotFoundError, err)
	}
	if errors.Is(err, repoErr.ConflictError) {
		return fmt.Errorf("%w: %v", serviceErr.AlreadyExists, err)
	}
	if err != nil {
		log.Errorf("%v: %v", serviceErr.UnexpectedError, err)
		return fmt.Errorf("%w: %v", serviceErr.UnexpectedError, err)
	}

	return nil
}
