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
	UnexpectedError = errors.New("unexpected error")
	DbError         = errors.New("db error")

	AccessDeniedError = errors.New("access denied")
	AlreadyExists     = errors.New("already exists")

	ParticalSuccessError = errors.New("partical success")
	AllFailedError       = errors.New("all failed")

	UserNotFoundError         = errors.New("user not found")
	UserAlreadyExistsError    = errors.New("user already exists")
	UserUnauthtenticatedError = errors.New("user unauthenticated")

	PhotoNotFoundError      = errors.New("photo not found")
	InvalidVersionTypeError = errors.New("invalid version type")
)
