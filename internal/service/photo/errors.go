package photo

import "errors"

// FileAlreadyExistsError возвращается, если файл в папке пользователя уже существует (не в базе данных)
var FileAlreadyExistsError = errors.New("photo with the same name already exists")
