package repository

import (
	"context"
	repoModel "go-photo/internal/repository/photo/model"
)

//go:generate mockgen -source=repository.go -destination=mock/mocks.go

type PhotoRepository interface {
	// CreateOriginalPhoto создает новую запись repoModel.Photo в БД и к ней repoModel.PhotoVersion.
	// Гарантируется, что у фото будет original версия.
	CreateOriginalPhoto(ctx context.Context, photo *repoModel.CreateOriginalPhotoParams) (int, error)

	// CreatePhotoPublishedInfo создает новую запись repoModel.PublishedPhotoInfo в БД.
	// Возвращает уникальный токен для доступа к фото.
	// Если запись уже существует, возвращает ошибку.
	CreatePhotoPublishedInfo(ctx context.Context, photoID int) (string, error)

	// GetPhotoByID возвращает фото по его ID.
	// Если фото не найдено, возвращает ошибку PhotoNotFound.
	GetPhotoByID(ctx context.Context, photoID int) (*repoModel.Photo, error)

	// GetPhotoVersions возвращает все версии фото по его ID.
	GetPhotoVersions(ctx context.Context, photoID int) ([]repoModel.PhotoVersion, error)

	// GetPhotoVersionByToken возвращает версию фото по токену и версии.
	GetPhotoVersionByToken(ctx context.Context, token string, filterParams *repoModel.FilterParams) (*repoModel.PhotoVersion, error)

	// GetPublicPhotosByTokenPrefix возвращает все публичные фото, которые начинаются с заданного токена.
	// TODO: tests
	GetPublicPhotosByTokenPrefix(ctx context.Context, tokenPrefix string, filterParams *repoModel.FilterParams) ([]repoModel.PhotoWithPhotoVersion, error)

	// DeletePhotoPublishedInfo удаляет запись repoModel.PublishedPhotoInfo из БД.
	// Если запись не найдена, возвращает ошибку.
	DeletePhotoPublishedInfo(ctx context.Context, photoID int) error
}
