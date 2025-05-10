package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

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
		return "", s.handleGRPCErr(err)
	}

	return user.Token, nil
}

func (s *service) handleGRPCErr(err error) error {
	st, ok := status.FromError(err)
	if !ok {
		// return UnexpectedError
		return err
	}

	switch st.Code() {
	case codes.NotFound:
		return fmt.Errorf("%w: %v", UserNotFoundError, err)
	case codes.AlreadyExists:
		return UserAlreadyExistsError
	case codes.Unauthenticated:
		return fmt.Errorf("%w: %v", UserUnauthtenticatedError, err)
	}

	// return fmt.Errorf("%w: %v", serviceErr.UnexpectedError, err)
	return err
}
