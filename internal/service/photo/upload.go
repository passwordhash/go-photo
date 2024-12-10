package photo

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"go-photo/internal/config"
	repoModel "go-photo/internal/repository/photo/model"
	"go-photo/internal/utils"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

// UploadBatchPhotos загружает несколько фотографий
func (s *service) UploadBatchPhotos(ctx context.Context, userUUID string, photoFiles []*multipart.FileHeader) ([]string, error) {
	var uploaded []string

	userFolder, err := ensureUserFolder(userUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure user's photos folder exists: %w", err)
	}

	for _, file := range photoFiles {
		_, err := s.processFile(ctx, userUUID, file, userFolder)
		if err != nil {
			return nil, err // Прерываем выполнение, если ошибка
		}
		uploaded = append(uploaded, file.Filename)
	}

	return uploaded, nil
}

// UploadPhoto загружает одну фотографию
func (s *service) UploadPhoto(ctx context.Context, userUUID string, photoFile *multipart.FileHeader) (int, error) {
	userFolder, err := ensureUserFolder(userUUID)
	if err != nil {
		return 0, fmt.Errorf("failed to ensure user's photos folder exists: %w", err)
	}

	id, err := s.processFile(ctx, userUUID, photoFile, userFolder)
	if err != nil {
		return 0, err
	}

	return id, nil
}

// processFile обрабатывает загрузку одного файла: проверяет существование, сохраняет, записывает в базу
func (s *service) processFile(ctx context.Context, userUUID string, file *multipart.FileHeader, userFolder string) (int, error) {
	destPath := filepath.Join(userFolder, file.Filename)

	exist, err := utils.Exist(destPath)
	if err != nil {
		return 0, fmt.Errorf("failed to check if file exists: %w", err)
	}
	if exist {
		return 0, &FileAlreadyExistsError{Filename: file.Filename}
	}

	err = saveFile(file, destPath)
	if err != nil {
		return 0, fmt.Errorf("failed to save photo with name '%s': %w", file.Filename, err)
	}

	id, err := s.photoRepository.CreateOriginalPhoto(ctx, &repoModel.CreateOriginalPhotoParams{
		UserUUID: userUUID,
		Filename: file.Filename,
		Filepath: destPath,
		Size:     file.Size,
	})
	if err != nil {
		// Если не удалось сохранить фото в базу, удаляем его с диска
		if rollbackErr := removePhotoFromDisk(destPath); rollbackErr != nil {
			log.Errorf("failed to rollback local save: %v", rollbackErr)
		}
		return 0, fmt.Errorf("failed to save photo in database: %w", err)
	}

	return id, nil
}

// ensureUserFolder проверяет или создает директорию для пользователя
func ensureUserFolder(userUUID string) (string, error) {
	userFolder := filepath.Join(config.PhotosDir, userUUID)
	return userFolder, utils.EnsureDirectoryExists(userFolder)
}

// saveFile сохраняет загружаемый файл на диск
func saveFile(file *multipart.FileHeader, destPath string) error {
	src, err := file.Open()
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	if err != nil {
		return fmt.Errorf("failed to write file to disk: %w", err)
	}

	return nil
}

// removePhotoFromDisk удаляет файл с диска
func removePhotoFromDisk(photoPath string) error {
	if err := os.Remove(photoPath); err != nil {
		return fmt.Errorf("failed to remove photo file %s: %w", photoPath, err)
	}
	return nil
}
