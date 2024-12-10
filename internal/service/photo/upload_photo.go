package photo

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"go-photo/internal/config"
	repoModel "go-photo/internal/repository/photo/model"
	_ "go-photo/internal/service"
	"go-photo/internal/utils"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

func (s *service) UploadPhoto(ctx context.Context, userUUID string, photoFile *multipart.FileHeader) (int, error) {
	userFolder, err := ensureUserFolder(userUUID)
	if err != nil {
		return 0, fmt.Errorf("failed to ensure user's photos folder exists: %w", err)
	}

	photoPath := filepath.Join(userFolder, photoFile.Filename)

	exist, err := utils.Exist(photoPath)
	if err != nil {
		return 0, fmt.Errorf("failed to check if photo exists: %w", err)
	} else if exist {
		return 0, FileAlreadyExistsError
	}

	err = saveFile(photoFile, photoPath)
	if err != nil {
		return 0, fmt.Errorf("failed to save photo with name '%s': %w", photoFile.Filename, err)
	}

	folderID, err := s.photoRepository.MustGetFolder(ctx, config.DefaultUsersFoldername, userUUID)
	if err != nil {
		return 0, fmt.Errorf("failed to get user folders: %w", err)
	}

	id, err := s.photoRepository.CreateOriginalPhoto(ctx, &repoModel.CreateOriginalPhotoParams{
		UserUUID: userUUID,
		Filename: photoFile.Filename,
		FolderID: folderID,
		Filepath: photoPath,
		Size:     photoFile.Size,
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

func ensureUserFolder(userUUID string) (string, error) {
	userFolder := filepath.Join(config.PhotosDir, userUUID)
	return userFolder, utils.EnsureDirectoryExists(userFolder)
}

func saveFile(file *multipart.FileHeader, destPath string) error {
	src, err := file.Open()
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	return err
}

func removePhotoFromDisk(photoPath string) error {
	if err := os.Remove(photoPath); err != nil {
		return fmt.Errorf("failed to remove photo file %s: %w", photoPath, err)
	}
	return nil
}
