package photo

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"go-photo/internal/config"
	"go-photo/internal/model"
	_ "go-photo/internal/service"
	"go-photo/internal/utils"
	"mime/multipart"
	"os"
	"path/filepath"
)

func (s *service) UploadPhoto(ctx context.Context, userUUID string, photoFile multipart.File, photoName string) (int, error) {
	photoPath, fileSize, err := savePhotoToDisk(userUUID, photoName, photoFile)
	if err != nil {
		return 0, fmt.Errorf("failed to save photo locally: %w", err)
	}

	photo := model.Photo{
		Filename: photoName,
		UserUUID: userUUID,
		Folder: model.Folder{
			Folderpath: config.DefaultUsersFoldername,
			UserUUID:   userUUID,
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
		// Если не удалось сохранить фото в базу, удаляем его с диска
		if rollbackErr := removePhotoFromDisk(photoPath); rollbackErr != nil {
			log.Errorf("failed to rollback local save: %v", rollbackErr)
		}
		return 0, fmt.Errorf("failed to save photo in database: %w", err)
	}

	return id, nil
}

func savePhotoToDisk(userUUID, photoName string, photoFile multipart.File) (string, int64, error) {
	userFolder := filepath.Join(config.PhotosDir, userUUID)

	err := utils.EnsureDirectoryExists(userFolder)
	if err != nil {
		return "", 0, fmt.Errorf("failed to ensure user folder exists: %w", err)
	}

	photoPath := filepath.Join(userFolder, photoName)

	if exists, err := utils.Exist(photoPath); err != nil {
		return "", 0, fmt.Errorf("failed to check if photo exists: %w", err)
	} else if exists {
		return "", 0, FileAlreadyExistsError
	}

	fileSize, err := utils.WriteFile(photoPath, photoFile)
	if err != nil {
		return "", 0, fmt.Errorf("failed to write file: %w", err)
	}

	return photoPath, fileSize, nil
}

func removePhotoFromDisk(photoPath string) error {
	if err := os.Remove(photoPath); err != nil {
		return fmt.Errorf("failed to remove photo file %s: %w", photoPath, err)
	}
	return nil
}
