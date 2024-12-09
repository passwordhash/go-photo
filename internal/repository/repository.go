package repository

import (
	"context"
	repoModel "go-photo/internal/repository/photo/model"
)

type PhotoRepository interface {
	// GetFolderID возвращает ID папки по пути и UUID пользователя.
	// Если папка не найдена, возвращает ошибку EmptyResultError
	GetFolderID(ctx context.Context, folderpath, userUUID string) (int, error)
	// CreateFolder создает новую папку в БД
	CreateFolder(ctx context.Context, folderpath, userUUID string) (int, error)
	// MustGetFolder возвращает ID папки по пути и UUID пользователя, если папка не найдена, то создает новую
	MustGetFolder(ctx context.Context, folderpath, userUUID string) (int, error)

	// CreateOriginalPhoto создает новую запись фото в БД, создается только оригинальная версия
	CreateOriginalPhoto(ctx context.Context, photo *repoModel.CreateOriginalPhotoParams) (int, error)
	GetPhotoVersions(ctx context.Context, photoID int) ([]repoModel.PhotoVersion, error)
}
