package user

import (
	"context"
	"errors"
	"go-photo/internal/config"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

var (
	ErrFailedToCreateFolder  = errors.New("user")
	ErrFileAlreadyExists     = errors.New("photo with the same name already exists")
	ErrInvalidFilePermission = errors.New("invalid file permission")
	ErrFailedToSavePhoto     = errors.New("failed to save file")
)

func (s *Service) UploadFile(_ context.Context, uuid string, photoFile multipart.File, photoName string) (int64, error) {
	userFolder := filepath.Join(config.PhotosDir, uuid)
	if _, err := os.Stat(userFolder); os.IsNotExist(err) {
		err := os.Mkdir(userFolder, os.ModePerm)
		if err != nil {
			return 0, ErrFailedToCreateFolder
		}
	}

	photoPath := filepath.Join(userFolder, photoName)
	if _, err := os.Stat(photoPath); err == nil {
		return 0, ErrFileAlreadyExists
	}

	file, err := os.Create(photoPath)
	if err != nil {
		return 0, ErrInvalidFilePermission
	}
	defer file.Close()

	fileSize, err := io.Copy(file, photoFile)
	if err != nil {
		return 0, ErrFailedToSavePhoto
	}

	return fileSize, nil
}
