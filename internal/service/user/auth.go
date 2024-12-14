package user

import (
	"context"
	"fmt"
	serviceErr "go-photo/internal/service/error"
	def "go-photo/pkg/account_v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *service) Login(ctx context.Context, login string, password string) (string, error) {
	// TODO encytped password
	resp, err := s.accountClient.Login(ctx, &def.LoginRequest{Email: login, EncryptedPassword: password})
	if err != nil {
		return "", s.handleGRPCErr(err)
	}

	return resp.JwtToken, nil
}

func (s *service) handleGRPCErr(err error) error {
	st, ok := status.FromError(err)
	if !ok {
		return serviceErr.ServiceError
	}

	switch st.Code() {
	case codes.NotFound:
		return serviceErr.UserNotFoundError
	case codes.Internal:
		return serviceErr.ServiceError
	}

	return fmt.Errorf("unhandled error: %v", err)
}
