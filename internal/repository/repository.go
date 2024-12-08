package repository

import (
	"context"
	"go-photo/internal/model"
	repoModel "go-photo/internal/repository/photo/model"
)

type PhotoRepository interface {
	GetPhotoVersions(ctx context.Context, photoID int) ([]repoModel.PhotoVersion, error)
	CreatePhoto(ctx context.Context, photo *model.Photo) (int, error)
}
