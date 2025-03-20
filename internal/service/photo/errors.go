package photo

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	repoErr "go-photo/internal/repository/error"
	serviceErr "go-photo/internal/service/error"
)

// HandleRepoErr обрабатывает ошибки, возвращаемые репозиторием.
// Обрабатывает ошибки:
// - NotFoundError
// - ConflictError
func (s *service) HandleRepoErr(err error) error {
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

