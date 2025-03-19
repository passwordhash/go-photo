package photo

import (
	"context"
	"errors"
	repoErr "go-photo/internal/repository/error"
	serviceErr "go-photo/internal/service/error"
)

func (s *service) PublicatePhoto(ctx context.Context, userUUID string, photoID int) (string, error) {
	photo, err := s.getUserPhoto(ctx, userUUID, photoID)
	if err != nil {
		return "", err
	}

	publicToken, err := s.photoRepository.CreatePhotoPublishedInfo(ctx, photo.ID)
	if errors.Is(err, repoErr.ConflictError) {
		return "", serviceErr.AlreadyExists
	}
	if err != nil {
		return "", err
	}

	return publicToken, nil
}
