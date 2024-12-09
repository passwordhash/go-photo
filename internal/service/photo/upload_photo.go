package photo

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"go-photo/internal/config"
	repoModel "go-photo/internal/repository/photo/model"
	_ "go-photo/internal/service"
	"go-photo/internal/utils"
	"mime/multipart"
	"os"
	"path/filepath"
)

func (s *service) UploadPhoto(ctx context.Context, userUUID string, photoFile multipart.File, photoName string) (int, error) {
	photoPath, photoSize, err := savePhotoToDisk(userUUID, photoName, photoFile)
	if err != nil {
		return 0, fmt.Errorf("failed to save photo locally: %w", err)
	}
	fmt.Println("photoPath", photoPath)

	// TEMP
	foldername := config.DefaultUsersFoldername

	folderID, err := s.photoRepository.MustGetFolder(ctx, foldername, userUUID)
	if err != nil {
		return 0, fmt.Errorf("failed to get user folders: %w", err)
	}

	id, err := s.photoRepository.CreateOriginalPhoto(ctx, &repoModel.CreateOriginalPhotoParams{
		UserUUID: userUUID,
		Filename: photoName,
		FolderID: folderID,
		Filepath: photoPath,
		Size:     photoSize,
	})
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
