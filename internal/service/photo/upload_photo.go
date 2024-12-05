package photo

import (
	"context"
	"errors"
	"fmt"
	"go-photo/internal/config"
	"go-photo/internal/utils"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

var FileAlreadyExistsError = errors.New("photo with the same name already exists")

func (s *Service) UploadPhoto(_ context.Context, uuid string, photoFile multipart.File, photoName string) (int64, error) {
	userFolder := filepath.Join(config.PhotosDir, uuid)
	if _, err := os.Stat(userFolder); os.IsNotExist(err) {
		err := os.Mkdir(userFolder, os.ModePerm)
		if err != nil {
			return 0, fmt.Errorf("user subfolder creation error")
		}
	}

	photoPath := filepath.Join(userFolder, photoName)

	exists, err := utils.Exist(photoPath)
	if err != nil {
		return 0, fmt.Errorf("cannot check photo path")
	}
	if exists {
		return 0, FileAlreadyExistsError
	}

	file, err := os.Create(photoPath)
	if err != nil {
		return 0, fmt.Errorf("cannot create photo file")
	}
	defer file.Close()

	fileSize, err := io.Copy(file, photoFile)
	if err != nil {
		return 0, fmt.Errorf("failed to save file")
	}

	return fileSize, nil
}
