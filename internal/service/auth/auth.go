package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	serviceAuthModel "go-photo/internal/service/auth/model"

	def "github.com/passwordhash/protos/gen/go/go-sso"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	UserNotFoundError         = errors.New("user not found")
	UserAlreadyExistsError    = errors.New("user already exists")
	UserUnauthtenticatedError = errors.New("user unauthenticated")
)

// TEMP:
const tmpAppID = 101

func (s *service) Register(
	ctx context.Context,
	params serviceAuthModel.RegisterParams,
) (int64, error) {
	const op = "service.auth.Register"

	log := s.log.With(
		slog.String("op", op),
		slog.String("email", params.Email),
	)

	log.Info("registration user")

	resp, err := s.authClient.Api.Register(ctx, &def.RegisterRequest{
		Email:    params.Email,
		Password: params.Password,
	})
	if err != nil {
		return 0, s.handleGRPCErr(ctx, log, err)
	}

	log.Info("user registered", "userID", resp.UserId)

	return resp.UserId, nil
}

func (s *service) Login(ctx context.Context, email string, password string) (string, error) {
	const op = "service.auth.Login"

	log := s.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	log.Info("logging in user")

	user, err := s.authClient.Api.Login(ctx, &def.LoginRequest{
		Email:    email,
		Password: password,
		AppId:    tmpAppID,
	})
	if err != nil {
		return "", s.handleGRPCErr(ctx, log, err)
	}

	return user.Token, nil
}

func (s *service) handleGRPCErr(ctx context.Context, log *slog.Logger, err error) error {
	st, ok := status.FromError(err)
	if !ok {
		log.ErrorContext(ctx, "not a grpc err when grpc err expected")
		return err
	}

	switch st.Code() {
	case codes.NotFound:
		log.Warn("user not found", "error", err)
		return fmt.Errorf("%w: %v", UserNotFoundError, err)
	case codes.AlreadyExists:
		log.Warn("user already exists", "error", err)
		return UserAlreadyExistsError
	case codes.Unauthenticated:
		log.Warn("user unauthenticated", "error", err)
		return fmt.Errorf("%w: %v", UserUnauthtenticatedError, err)
	}

	// return fmt.Errorf("%w: %v", serviceErr.UnexpectedError, err)
	log.ErrorContext(ctx, "unexpected grcp error", "error", err)
	return err
}
