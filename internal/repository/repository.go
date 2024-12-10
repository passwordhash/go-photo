package repository

import (
	"context"
	repoModel "go-photo/internal/repository/photo/model"
)

type PhotoRepository interface {
	// CreateOriginalPhoto создает новую запись фото в БД, создается только оригинальная версия
	CreateOriginalPhoto(ctx context.Context, photo *repoModel.CreateOriginalPhotoParams) (int, error)
	GetPhotoVersions(ctx context.Context, photoID int) ([]repoModel.PhotoVersion, error)
}
