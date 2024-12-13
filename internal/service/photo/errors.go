package photo

import "fmt"

// FileAlreadyExistsError возвращается, если файл в папке пользователя уже существует (не в базе данных)
type FileAlreadyExistsError struct {
	Filename string
}

func (e *FileAlreadyExistsError) Error() string {
	return fmt.Sprintf("file %s already exists", e.Filename)
}
