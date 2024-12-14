package service

import (
	"context"
	"go-photo/internal/model"
	serviceModel "go-photo/internal/service/photo/model"
	"mime/multipart"
)

//go:generate mockgen -destination=mocks/mock.go -source=service.go

type UserService interface {
	// Login выполняет аутентификацию пользователя по логину и паролю. Возвращает JWT token
	Login(ctx context.Context, login string, password string) (string, error)
	Get(ctx context.Context, uuid string) (model.User, error)
	GetAll(ctx context.Context) ([]model.User, error)
}

type PhotoService interface {
	// UploadPhoto загружает одну фотографию
	UploadPhoto(ctx context.Context, userUUID string, photoFile *multipart.FileHeader) (int, error)
	// UploadBatchPhotos загружает несколько фотографий конкурентно. Возвращает список информации о загруженных фотографиях.
	// Если возникла ошибка во время загрузки фотографии, то прикрепляет информацию об ошибке.
	UploadBatchPhotos(ctx context.Context, userUUID string, photoFiles []*multipart.FileHeader) (*serviceModel.UploadInfoList, error)
	GetPhotoVersions(ctx context.Context, photoID int) ([]model.PhotoVersion, error)
}
