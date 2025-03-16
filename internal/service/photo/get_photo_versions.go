package photo

import (
	"context"
	"errors"
	"go-photo/internal/model"
	repoErr "go-photo/internal/repository/error"
	"go-photo/internal/repository/photo/converter"
	serviceErr "go-photo/internal/service/error"
)

func (s *service) GetPhotoVersions(ctx context.Context, id int) ([]model.PhotoVersion, error) {
	repoVersions, err := s.photoRepository.GetPhotoVersions(ctx, id)
	if errors.Is(err, repoErr.PhotoNotFound) {
		return nil, serviceErr.PhotoNotFoundError
	}
	if err != nil {
		return nil, err
	}

	versions := converter.ToPhotoVersionsFromRepo(repoVersions)

	return versions, nil
}
