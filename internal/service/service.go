package service

import (
	"context"
	"go-photo/internal/model"
	serviceModel "go-photo/internal/service/photo/model"
	"mime/multipart"
)

//go:generate mockgen -destination=mocks/mock.go -source=service.go

type UserService interface {
	Get(ctx context.Context, uuid string) (model.User, error)
	GetAll(ctx context.Context) ([]model.User, error)
}

type PhotoService interface {
	UploadPhoto(ctx context.Context, userUUID string, photoFile *multipart.FileHeader) (int, error)
	UploadBatchPhotos(ctx context.Context, userUUID string, photoFiles []*multipart.FileHeader) (*serviceModel.UploadInfoList, error)
	GetPhotoVersions(ctx context.Context, photoID int) ([]model.PhotoVersion, error)
}
