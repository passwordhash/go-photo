package utils

//go:generate mockgen -destination=mock/mocks.go -source=utils.go

type Inteface interface {
	EncryptPassword(publicKey *string, password string) (string, error)
}

type Utils struct {
}

func New() *Utils {
	return &Utils{}
}
