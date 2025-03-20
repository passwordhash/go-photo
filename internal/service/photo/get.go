package photo

import (
	"context"
	"go-photo/internal/model"
	"go-photo/internal/repository/photo/converter"
	repoModel "go-photo/internal/repository/photo/model"
	serviceErr "go-photo/internal/service/error"
)

func (s *service) GetPhotoVersions(ctx context.Context, userUUID string, photoID int) ([]model.PhotoVersion, error) {
	photo, err := s.photoRepository.GetPhotoByID(ctx, photoID)
	if err := s.HandleRepoErr(err); err != nil {
		return nil, err
	}

	if photo.UserUUID != userUUID {
		return nil, serviceErr.AccessDeniedError
	}

	repoVersions, err := s.photoRepository.GetPhotoVersions(ctx, photoID)
	if err := s.HandleRepoErr(err); err != nil {
		return nil, err
	}

	versions := converter.ToPhotoVersionsFromRepo(repoVersions)

	return versions, nil
}

// getUserPhoto возвращает фотографию пользователя по ее ID.
// Если фотография не найдена, возвращает ошибку PhotoNotFoundError.
// Если фотография найдена, но принадлежит другому пользователю, возвращает ошибку AccessDeniedError.
func (s *service) getUserPhoto(ctx context.Context, userUUID string, photoID int) (*repoModel.Photo, error) {
	photo, err := s.photoRepository.GetPhotoByID(ctx, photoID)
	if err := s.HandleRepoErr(err); err != nil {
		return &repoModel.Photo{}, err
	}

	if photo.UserUUID != userUUID {
		return &repoModel.Photo{}, serviceErr.AccessDeniedError
	}

	return photo, nil
}
