package photo

import (
	"context"
	"go-photo/internal/model"
	"go-photo/internal/repository/photo/converter"
)

func (s *service) GetPhotoVersions(ctx context.Context, id int) ([]model.PhotoVersion, error) {
	repoVersions, err := s.photoRepository.GetPhotoVersions(ctx, id)
	// TODO: err handling
	if err != nil {
		return nil, err
	}

	versions := converter.ToPhotoVersionsFromRepo(repoVersions)

	return versions, nil
}
