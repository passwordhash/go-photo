package repository

import (
	"context"
	repoModel "go-photo/internal/repository/photo/model"
)

type PhotoRepository interface {
	GetPhotoVersions(ctx context.Context, photoID int) ([]repoModel.PhotoVersion, error)
}
