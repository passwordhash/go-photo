package error

import (
	"errors"
	"fmt"
)

// FileAlreadyExistsError возвращается, если файл в папке пользователя уже существует (не в базе данных)
type FileAlreadyExistsError struct {
	Filename string
}

func (e *FileAlreadyExistsError) Error() string {
	return fmt.Sprintf("file %s already exists", e.Filename)
}

var (
	ServiceError = errors.New("service error")
	DbError      = errors.New("db error")

	ParticalSuccessError = errors.New("partical success")
	AllFailedError       = errors.New("all failed")

	UserNotFoundError      = errors.New("user not found")
	UserAlreadyExistsError = errors.New("user already exists")
)
