package auth

import (
	"context"
	"fmt"
	"log/slog"

	serviceAuthModel "go-photo/internal/service/auth/model"
	serviceErr "go-photo/internal/service/error"

	def "github.com/passwordhash/protos/gen/go/go-sso"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TEMP:
const tmpAppID = 101

func (s *service) Register(
	ctx context.Context,
	params serviceAuthModel.RegisterParams,
) (string, error) {
	const op = "service.auth.Register"

	log := s.log.With(
		slog.String("op", op),
		slog.String("email", params.Email),
	)

	log.Info("registration user")

	resp, err := s.authAPI.Register(ctx, &def.RegisterRequest{
		Email:    params.Email,
		Password: params.Password,
	})
	if err != nil {
		return "", s.handleGRPCErr(ctx, log, err)
	}

	log.Info("user registered", "userID", resp.UserUuid)

	return resp.UserUuid, nil
}

func (s *service) Login(ctx context.Context, email string, password string) (string, error) {
	const op = "service.auth.Login"

	log := s.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	log.Info("logging in user")

	user, err := s.authAPI.Login(ctx, &def.LoginRequest{
		Email:    email,
		Password: password,
		AppName:  s.appName,
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

	log.WarnContext(ctx, "grpc error",
		"code", st.Code().String(),
		"message", st.Message(),
	)

	switch st.Code() {
	case codes.NotFound:
		return fmt.Errorf("%w: %v", serviceErr.UserNotFoundError, err)
	case codes.Unauthenticated:
		return fmt.Errorf("%w: %v", serviceErr.UserUnauthtenticatedError, err)
	case codes.AlreadyExists:
		return fmt.Errorf("%w: %v", serviceErr.UserAlreadyExistsError, err)
	default:
		log.ErrorContext(ctx, "unexpected grcp error", "error", err)
		return err
	}
}
