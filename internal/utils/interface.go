package utils

//go:generate mockgen -destination=mock/mocks.go -source=interface.go

type Interface interface {
	EncryptPassword(publicKey *string, password string) (string, error)

	// UUIDFilename возвращает имя файла, сгенерированное с помощью UUID,
	// сохраняя расширение оригинального файла.
	UUIDFilename(filename string) string
}

type Utils struct {
}

func New() *Utils {
	return &Utils{}
}
