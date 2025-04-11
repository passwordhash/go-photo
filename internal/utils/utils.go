package utils

import (
	"github.com/google/uuid"
	"path/filepath"
)

func (u *Utils) UUIDFilename(filename string) string {
	ext := filepath.Ext(filename)
	uuidName := uuid.NewString()
	return uuidName + ext
}
