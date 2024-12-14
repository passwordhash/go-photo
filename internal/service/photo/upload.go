package photo

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	repoModel "go-photo/internal/repository/photo/model"
	serviceErr "go-photo/internal/service/error"
	serviceModel "go-photo/internal/service/photo/model"
	"go-photo/internal/utils"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"sync"
)

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
		return 0, &serviceErr.FileAlreadyExistsError{Filename: photoFile.Filename}
	}

	err = saveFileToDisk(photoFile, photoPath)
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

// UploadBatchPhotos загружает несколько фотографий конкурентно
func (s *service) UploadBatchPhotos(ctx context.Context, userUUID string, photoFiles []*multipart.FileHeader) (*serviceModel.UploadInfoList, error) {
	destFolder, err := ensureUserFolder(s.d.StorageFolderPath, userUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure user's photos folder exists: %w", err)
	}

	workerCount := len(photoFiles)

	saver := make(chan *multipart.FileHeader)
	creater := make(chan *serviceModel.UploadInfo)

	uploaded := &serviceModel.UploadInfoList{}

	go func() {
		for _, file := range photoFiles {
			saver <- file
		}
		close(saver)
	}()

	wgSaver := &sync.WaitGroup{}
	wgSaver.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go func(i int) {
			defer wgSaver.Done()
			for file := range saver {
				info := serviceModel.UploadInfo{
					Filename: file.Filename,
					Size:     file.Size,
				}

				err := saveFileToDisk(file, destFolder)
				if err != nil {
					saveErr := fmt.Errorf("failed to save photo with name '%s' in '%s': %w", file.Filename, destFolder, err)
					log.Errorf(saveErr.Error())
					info.Error = saveErr
				}

				creater <- &info
			}
		}(i)
	}

	wgCreater := &sync.WaitGroup{}
	wgCreater.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go func(i int) {
			defer wgCreater.Done()

			info := <-creater
			if info.Error != nil {
				uploaded.Add(*info)
				return
				// TODO: почему не continue?
			}

			id, err := s.photoRepository.CreateOriginalPhoto(ctx, &repoModel.CreateOriginalPhotoParams{
				UserUUID: userUUID,
				Filename: info.Filename,
				Filepath: filepath.Join(destFolder, info.Filename),
				Size:     info.Size,
			})
			if err != nil {
				dbErr := fmt.Errorf("save photo %w: %v", serviceErr.DbError, err)
				log.Errorf(dbErr.Error())
				info.Error = dbErr
			} else {
				info.PhotoID = id
			}

			uploaded.Add(*info)
		}(i)
	}

	wgCreater.Wait()

	if uploaded.IsAllError() {
		return uploaded, serviceErr.AllFailedError
	}
	if uploaded.IsSomeError() {
		return uploaded, serviceErr.ParticalSuccessError
	}

	return uploaded, nil
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

// saveFileToDisk сохраняет загружаемый файл на диск
func saveFileToDisk(file *multipart.FileHeader, destFolder string) error {
	src, err := file.Open()
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	filePath := filepath.Join(destFolder, file.Filename)
	out, err := os.Create(filePath)
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
