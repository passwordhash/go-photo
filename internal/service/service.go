package service

import (
	"context"
	"go-photo/internal/model"
	"mime/multipart"
)

type UserService interface {
	Get(ctx context.Context, uuid string) (model.User, error)
	GetAll(ctx context.Context) ([]model.User, error)
}

type PhotoService interface {
	UploadPhoto(ctx context.Context, userUUID string, file multipart.File, photoName string) (int, error)
	GetPhotoVersions(ctx context.Context, photoID int) ([]model.PhotoVersion, error)
}
