package service

import (
	"context"
	"go-photo/internal/model"
)

type UserService interface {
	Get(ctx context.Context, uuid string) (model.User, error)
}
