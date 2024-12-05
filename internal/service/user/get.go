package user

import (
	"context"
	"go-photo/internal/model"
	"time"
)

func (s *Service) Get(_ context.Context, uuid string) (model.User, error) {
	mockUser := model.User{
		UUID:         uuid,
		Email:        "mock-email",
		PasswordHash: "mock-password-hash",
		IsVerified:   true,
		CreatedAt:    time.Now(),
	}

	return mockUser, nil
}
