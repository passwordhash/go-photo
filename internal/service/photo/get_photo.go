package photo

import (
	"context"
	"go-photo/internal/model"
	"go-photo/internal/repository/photo/converter"
)

func (s *service) GetPhoto(ctx context.Context, photoID int) (*model.Photo, error) {
	var photo *model.Photo
	photoInfo, err := s.photoRepository.GetPhotoByID(ctx, photoID)
	if err != nil {
		return nil, err
	}
	photoVersionsInfo, err := s.photoRepository.GetPhotoVersions(ctx, photoID)
	if err != nil {
		return nil, err
	}

	photo = converter.ToPhotoFromRepo(photoInfo, photoVersionsInfo)
	return photo, nil
}
