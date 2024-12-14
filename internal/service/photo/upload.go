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
	"runtime"
	"sync"
)

// UploadPhoto загружает одну фотографию
func (s *service) UploadPhoto(ctx context.Context, userUUID string, photoFile *multipart.FileHeader) (int, error) {
	userFolder, err := ensureUserFolder(s.d.StorageFolderPath, userUUID)
	if err != nil {
		return 0, fmt.Errorf("failed to ensure user's photos folder exists: %w", err)
	}

	info := s.saveFile(ctx, photoFile, userFolder)
	if info.Error != nil {
		log.Errorf("Failed to save file %s: %v", photoFile.Filename, info.Error)
		return 0, info.Error
	}

	info = s.saveToDatabase(ctx, userUUID, info)
	if info.Error != nil {
		log.Errorf("Failed to save file %s to database: %v", photoFile.Filename, info.Error)
		return 0, info.Error
	}

	return info.PhotoID, nil
}

// UploadBatchPhotos загружает несколько фотографий конкурентно
func (s *service) UploadBatchPhotos(ctx context.Context, userUUID string, photoFiles []*multipart.FileHeader) (*serviceModel.UploadInfoList, error) {
	destFolder, err := ensureUserFolder(s.d.StorageFolderPath, userUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure user's photos folder exists: %w", err)
	}

	uploaded := &serviceModel.UploadInfoList{}
	fileTaskChan := make(chan *multipart.FileHeader)
	dbTaskChan := make(chan serviceModel.UploadInfo)
	resultChan := make(chan serviceModel.UploadInfo)

	fileWorkerCount := runtime.NumCPU() / 3
	dbWorkerCount := runtime.NumCPU() / 3

	fileWg := sync.WaitGroup{}
	for i := 0; i < fileWorkerCount; i++ {
		fileWg.Add(1)
		go func(workerID int) {
			defer fileWg.Done()
			for file := range fileTaskChan {
				select {
				case <-ctx.Done():
					log.Warnf("File worker %d stopped due to context cancellation. Context: %v", workerID, ctx.Err())
				default:
				}

				info := s.saveFile(ctx, file, destFolder)
				dbTaskChan <- info
			}
		}(i)
	}

	dbWg := sync.WaitGroup{}
	for i := 0; i < dbWorkerCount; i++ {
		dbWg.Add(1)
		go func(workerID int) {
			defer dbWg.Done()
			for info := range dbTaskChan {
				select {
				case <-ctx.Done():
					log.Warnf("DB worker %d stopped due to context cancellation. Context: %v", workerID, ctx.Err())
				default:
				}

				info = s.saveToDatabase(ctx, userUUID, info)
				if info.Error != nil {
					log.Errorf("Failed to save file %s: %v", info.Filename, info.Error)
					log.Warnf("Skipping DB save for file %s due to disk save error: %v", info.Filename, info.Error)
					resultChan <- info
					continue
				}

				resultChan <- info
			}
		}(i)
	}

	// Ожидание завершения всех задач
	go func() {
		fileWg.Wait()
		close(dbTaskChan)
		dbWg.Wait()
		close(resultChan)
	}()

	// Отправка задач на сохранение файлов
	go func() {
		for _, file := range photoFiles {
			select {
			case <-ctx.Done():
				log.Warn("File task sending stopped due to context cancellation")
				return
			case fileTaskChan <- file:
			}
		}
		close(fileTaskChan)
	}()

	for res := range resultChan {
		uploaded.Add(res)
	}

	if uploaded.IsAllError() {
		return uploaded, serviceErr.AllFailedError
	}
	if uploaded.IsSomeError() {
		return uploaded, serviceErr.ParticalSuccessError
	}

	return uploaded, nil
}

// saveFile - сохраняет файл на диск
func (s *service) saveFile(_ context.Context, file *multipart.FileHeader, destFolder string) serviceModel.UploadInfo {
	info := serviceModel.UploadInfo{
		Filename: file.Filename,
		Size:     file.Size,
	}

	if err := saveFileToDisk(file, destFolder); err != nil {
		log.Errorf("Failed to save file %s: %v", file.Filename, err)
		info.Error = fmt.Errorf("disk save error: %w", err)
	}

	return info
}

// saveToDatabase сохраняет информацию о файле в базе данных. Если произошла ошибка, файл удаляется с диска
func (s *service) saveToDatabase(ctx context.Context, userUUID string, info serviceModel.UploadInfo) serviceModel.UploadInfo {
	if info.Error != nil {
		return info
	}

	id, err := s.photoRepository.CreateOriginalPhoto(ctx, &repoModel.CreateOriginalPhotoParams{
		UserUUID: userUUID,
		Filename: info.Filename,
		Filepath: filepath.Join(s.d.StorageFolderPath, userUUID, info.Filename),
		Size:     info.Size,
	})
	if err != nil {
		log.Errorf("DB save error for file %s: %v", info.Filename, err)
		info.Error = fmt.Errorf("db save error: %w", err)

		filePath := filepath.Join(s.d.StorageFolderPath, userUUID, info.Filename)
		if rmErr := os.Remove(filePath); rmErr != nil {
			log.Errorf("Failed to remove file %s after DB save error: %v", filePath, rmErr)
			info.Error = fmt.Errorf("%w; additionally, rollback failed: %v", info.Error, rmErr)
		} else {
			log.Infof("File %s removed due to failed DB save", filePath)
		}
	} else {
		info.PhotoID = id
	}

	return info
}

func ensureUserFolder(storageFolderPath, userUUID string) (string, error) {
	userFolder := filepath.Join(storageFolderPath, userUUID)
	err := utils.EnsureDirectoryExists(userFolder)
	if err != nil {
		return "", fmt.Errorf("failed to ensure user's photos folder exists: %w", err)
	}
	return userFolder, nil
}

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
