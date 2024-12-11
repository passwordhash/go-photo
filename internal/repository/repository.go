package repository

import (
	"context"
	repoModel "go-photo/internal/repository/photo/model"
)

type PhotoRepository interface {
	// CreateOriginalPhoto создает новую запись фото в БД, создается только оригинальная версия
	CreateOriginalPhoto(ctx context.Context, photo *repoModel.CreateOriginalPhotoParams) (int, error)

	// GetPhotoByID возвращает фото по его ID.
	// Если фото не найдено, возвращает ошибку PhotoNotFound
	GetPhotoByID(ctx context.Context, photoID int) (*repoModel.Photo, error)

	// GetPhotoVersions возвращает все версии фото по его ID отсортированные по размеру по возрастанию.
	//
	GetPhotoVersions(ctx context.Context, photoID int) ([]repoModel.PhotoVersion, error)
}
