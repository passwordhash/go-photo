package photo

// FileAlreadyExistsError возвращается, если файл в папке пользователя уже существует (не в базе данных)
type FileAlreadyExistsError struct {
	Filename string
}

func (e *FileAlreadyExistsError) Error() string {
	return "photo with the same name already exists: " + e.Filename
}
