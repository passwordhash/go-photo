package photo

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	repoModel "go-photo/internal/repository/photo/model"
	serviceErr "go-photo/internal/service/error"
	"go-photo/internal/utils"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

// UploadBatchPhotos загружает несколько фотографий
// Если файл с таким именем уже существует, возвращается ошибка FileAlreadyExistsError
func (s *service) UploadBatchPhotos(ctx context.Context, userUUID string, photoFiles []*multipart.FileHeader) ([]string, error) {
	var uploaded []string

	userFolder, err := ensureUserFolder(s.d.StorageFolderPath, userUUID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", serviceErr.ServiceError, err)
	}

	for _, file := range photoFiles {
		_, err := s.processFile(ctx, userUUID, file, userFolder)
		if err != nil {
			return nil, fmt.Errorf("processing file %s err: %w", file.Filename, err)
		}
		uploaded = append(uploaded, file.Filename)
	}

	return uploaded, nil
}

// UploadPhoto загружает одну фотографию
func (s *service) UploadPhoto(ctx context.Context, userUUID string, photoFile *multipart.FileHeader) (int, error) {
	userFolder, err := ensureUserFolder(s.d.StorageFolderPath, userUUID)
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
func (s *service) processFile(ctx context.Context, userUUID string, photoFile *multipart.FileHeader, userFolder string) (int, error) {
	photoPath := filepath.Join(userFolder, photoFile.Filename)

	exist, err := utils.Exist(photoPath)
	if err != nil {
		return 0, fmt.Errorf("failed to check if photoFile exists: %w", err)
	}
	if exist {
		return 0, &FileAlreadyExistsError{Filename: photoFile.Filename}
	}

	err = saveFile(photoFile, photoPath)
	if err != nil {
		return 0, fmt.Errorf("failed to save photo with name '%s' in '%s': %w", photoFile.Filename, photoPath, err)
	}

	id, err := s.photoRepository.CreateOriginalPhoto(ctx, &repoModel.CreateOriginalPhotoParams{
		UserUUID: userUUID,
		Filename: photoFile.Filename,
		Filepath: photoPath,
		Size:     photoFile.Size,
	})
	if err != nil {
		// Если не удалось сохранить фото в базу, удаляем его с диска
		if rollbackErr := removePhotoFromDisk(photoPath); rollbackErr != nil {
			log.Errorf("failed to rollback local save: %v", rollbackErr)
		}
		return 0, fmt.Errorf("save photo %w: %v", serviceErr.DbError, err)
	}

	return id, nil
}

// ensureUserFolder проверяет или создает директорию для пользователя
func ensureUserFolder(storageFolderPath, userUUID string) (string, error) {
	userFolder := filepath.Join(storageFolderPath, userUUID)
	err := utils.EnsureDirectoryExists(userFolder)
	if err != nil {
		return "", fmt.Errorf("failed to ensure user's photos folder exists: %w", err)
	}
	return userFolder, nil
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
