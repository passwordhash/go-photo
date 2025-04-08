package utils

//go:generate mockgen -destination=mock/mocks.go -source=interface.go

type Interface interface {
	EncryptPassword(publicKey *string, password string) (string, error)

	UUID() string
}

type Utils struct {
}

func New() *Utils {
	return &Utils{}
}
