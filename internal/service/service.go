package service

import (
	"context"
	"fmt"
	"go-photo/internal/model"
	"mime/multipart"
)

type UserService interface {
	Get(ctx context.Context, uuid string) (model.User, error)
	UploadFile(ctx context.Context, uuid string, photo multipart.File, fileName string) (int64, error)
}

var FolderConflictNameErr = fmt.Errorf("asdfasd asdfas dasdf: %s")
