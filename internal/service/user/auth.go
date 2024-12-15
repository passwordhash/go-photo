package user

import (
	"context"
	"fmt"
	serviceErr "go-photo/internal/service/error"
	serviceUserModel "go-photo/internal/service/user/model"
	def "go-photo/pkg/account_v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *service) Login(ctx context.Context, email string, password string) (string, error) {
	// TODO encrypt password
	resp, err := s.accountClient.Login(ctx, &def.LoginRequest{Email: email, EncryptedPassword: password})
	if err != nil {
		return "", s.handleGRPCErr(err)
	}

	return resp.JwtToken, nil
}

func (s *service) Register(ctx context.Context, input serviceUserModel.RegisterParams) (serviceUserModel.RegisterInfo, error) {
	// TODO encrypt password
	resp, err := s.accountClient.Signup(ctx, &def.CreateRequest{
		Email:             input.Email,
		EncryptedPassword: input.Password,
	})
	if err != nil {
		return serviceUserModel.RegisterInfo{}, s.handleGRPCErr(err)
	}

	info := serviceUserModel.RegisterInfo{
		UserUUID: resp.Uuid,
		Token:    resp.JwtToken,
	}

	return info, nil
}

func (s *service) handleGRPCErr(err error) error {
	st, ok := status.FromError(err)
	if !ok {
		return serviceErr.InternalError
	}

	switch st.Code() {
	case codes.NotFound:
		return serviceErr.UserNotFoundError
	case codes.AlreadyExists:
		return serviceErr.UserAlreadyExistsError
	case codes.Internal:
		return fmt.Errorf("%w: %v", serviceErr.InternalError, st.Message())
	}

	return fmt.Errorf("unhandled error: %v", err)
}
