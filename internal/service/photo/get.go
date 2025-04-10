package photo

import (
	"context"
	"fmt"
	"go-photo/internal/model"
	"go-photo/internal/repository/photo/converter"
	repoModel "go-photo/internal/repository/photo/model"
	serviceErr "go-photo/internal/service/error"
	"os"
	"path/filepath"
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

func (s *service) GetPhotoFileByVersionAndToken(ctx context.Context, token string, version string) ([]byte, error) {
	versionType, err := model.ParseVersionType(version)
	if err != nil {
		return nil, serviceErr.InvalidVersionTypeError
	}

	photoVersion, err := s.photoRepository.GetPhotoVersionByToken(ctx, token, &repoModel.FilterParams{
		VersionType: versionType,
	})
	if err := s.HandleRepoErr(err); err != nil {
		return nil, err
	}

	// TODO: получить uuid пользователя, чтобы найти файл

	photoFilepath := filepath.Join(s.d.StorageFolderPath, photoVersion.UUIDFilename)
	file, err := os.Open(photoFilepath)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to open file: %v", serviceErr.UnexpectedError, err)
	}
	defer file.Close()

	buff := make([]byte, photoVersion.Size)
	_, err = file.Read(buff)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to read file: %v", serviceErr.UnexpectedError, err)
	}

	return buff, nil
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
