package utils

import (
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
)

var imageExt = map[string]struct{}{
	".jpg":  {},
	".jpeg": {},
	".png":  {},
	".svg":  {},
	".webp": {},
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
