package photo

import def "go-photo/internal/service"

// Проверка на соответствие интерфейсу UserService (для статической проверки)
var _ def.PhotoService = (*Service)(nil)

type Service struct {
	//userRepo repository.UserRepository
}

func NewPhotoService() *Service {
	return &Service{}
}
