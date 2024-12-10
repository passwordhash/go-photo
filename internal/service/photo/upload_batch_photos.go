package photo

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"go-photo/internal/config"
	repoModel "go-photo/internal/repository/photo/model"
	"go-photo/internal/utils"
	"mime/multipart"
	"path/filepath"
)

// UploadBatchPhotos загружает несколько фотографий в папку пользователя.
// Возвращает список названий загруженных фотографий.

// Если хотя бы одна фотография уже существует, не загружает ни одной фотографии,
// возвращает ошибку FileAlreadyExistsError и список названий уже существующих фотографий.
func (s *service) UploadBatchPhotos(ctx context.Context, userUUID string, photoFiles []*multipart.FileHeader) ([]string, error) {
	var uploaded []string

	userFolder, err := ensureUserFolder(userUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure user's photos folder exists: %w", err)
	}

	folderID, err := s.photoRepository.MustGetFolder(ctx, config.DefaultUsersFoldername, userUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user folders: %w", err)
	}

	for _, file := range photoFiles {
		destPath := filepath.Join(userFolder, file.Filename)

		exist, _ := utils.Exist(destPath)
		if exist {
			continue
		}

		err = saveFile(file, destPath)
		if err != nil {
			return nil, fmt.Errorf("failed to save photo with name '%s': %w", file.Filename, err)
		}

		_, err := s.photoRepository.CreateOriginalPhoto(ctx, &repoModel.CreateOriginalPhotoParams{
			UserUUID: userUUID,
			Filename: file.Filename,
			FolderID: folderID,
			Filepath: destPath,
			Size:     file.Size,
		})
		if err != nil {
			// Если не удалось сохранить фото в базу, удаляем его с диска
			if rollbackErr := removePhotoFromDisk(destPath); rollbackErr != nil {
				log.Errorf("failed to rollback local save: %v", rollbackErr)
			}
			return nil, fmt.Errorf("failed to save photo in database: %w", err)
		}

		uploaded = append(uploaded, file.Filename)
	}

	return uploaded, nil
}
