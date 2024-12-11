package utils

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

var imageExt = map[string]struct{}{
	".jpg":  {},
	".jpeg": {},
	".png":  {},
	".svg":  {},
}

// IsAllPhotos проверяет, что все файлы являются фотографиями по расширению.
// При первом несоответствии возвращает false и имя файла.
// Если все файлы являются фотографиями, возвращает пустую строку и true.
func IsAllPhotos(fileHeaders []*multipart.FileHeader) (bool, string) {
	for _, header := range fileHeaders {
		ext := filepath.Ext(header.Filename)
		if !IsPhoto(ext) {
			return false, header.Filename
		}
	}
	return true, ""
}

func IsPhoto(extension string) bool {
	_, ok := imageExt[extension]
	return ok
}

func Exist(path string) (bool, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func EnsureDirectoryExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", path, err)
		}
	}
	return nil
}

func WriteFile(filePath string, photoFile multipart.File) (int64, error) {
	file, err := os.Create(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to create file %s: %w", filePath, err)
	}
	defer file.Close()

	fileSize, err := io.Copy(file, photoFile)
	if err != nil {
		return 0, fmt.Errorf("failed to copy data to file %s: %w", filePath, err)
	}

	return fileSize, nil
}
