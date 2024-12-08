package photo

import (
	"context"
	"fmt"
	"go-photo/internal/config"
	"go-photo/internal/model"
	_ "go-photo/internal/service"
	"go-photo/internal/utils"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

func (s *service) UploadPhoto(ctx context.Context, userUUID string, photoFile multipart.File, photoName string) (int, error) {
	userFolder := filepath.Join(config.PhotosDir, userUUID)
	if _, err := os.Stat(userFolder); os.IsNotExist(err) {
		if err := os.MkdirAll(userFolder, os.ModePerm); err != nil {
			return 0, fmt.Errorf("failed to create user folder %s: %w", userFolder, err)
		}
	}

	photoPath := filepath.Join(userFolder, photoName)

	exists, err := utils.Exist(photoPath)
	if err != nil {
		return 0, fmt.Errorf("failed to check photo path existence: %w", err)
	}
	if exists {
		return 0, FileAlreadyExistsError
	}

	file, err := os.Create(photoPath)
	if err != nil {
		return 0, fmt.Errorf("failed to create photo file %s: %w", photoPath, err)
	}
	defer file.Close()

	fileSize, err := io.Copy(file, photoFile)
	if err != nil {
		return 0, fmt.Errorf("failed to save file to disk: %w", err)
	}

	photo := model.Photo{
		Filename: photoName,
		UserUUID: userUUID,
		Folder: model.Folder{
			Folderpath: config.DefaultUsersFoldername,
		},
		Versions: []model.PhotoVersion{
			{
				VersionType: model.Original,
				Filepath:    photoPath,
				Size:        fileSize,
			},
		},
	}

	id, err := s.photoRepository.CreatePhoto(ctx, &photo)
	if err != nil {
		return 0, fmt.Errorf("failed to save photo in database: %w", err)
	}

	return id, nil
}
