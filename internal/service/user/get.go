package user

import (
	"context"
	"go-photo/internal/model"
	"google.golang.org/protobuf/types/known/emptypb"
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

func (s Service) GetAll(ctx context.Context) ([]model.User, error) {
	resp, err := s.accountClient.GetAll(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}

	var users []model.User
	for _, account := range resp.Accounts {
		// TODO: create converter
		users = append(users, model.User{
			UUID:         account.Uuid,
			Email:        account.Email,
			PasswordHash: account.HashedPassword,
			IsVerified:   account.IsVerified,
			CreatedAt:    time.Unix(account.CreatedAt.Seconds, int64(account.CreatedAt.Nanos)),
		})
	}

	return users, nil
}
