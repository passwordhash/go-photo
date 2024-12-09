package repository

import (
	"context"
	"go-photo/internal/model"
	repoModel "go-photo/internal/repository/photo/model"
)

type PhotoRepository interface {
	// GetFolders возвращает список папок пользователя. Если папок нет, возвращает пустой список
	GetFolders(ctx context.Context, userUUID string) ([]repoModel.Folder, error)
	CreateFolder(ctx context.Context, folderpath, userUUID string) error

	CreatePhoto(ctx context.Context, photo *model.Photo) (int, error)
	GetPhotoVersions(ctx context.Context, photoID int) ([]repoModel.PhotoVersion, error)
}
