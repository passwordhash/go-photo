package user

import (
	"context"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	serviceErr "go-photo/internal/service/error"
	serviceUserModel "go-photo/internal/service/user/model"
	"go-photo/internal/utils"
	mock_utils "go-photo/internal/utils/mock"
	def "go-photo/pkg/account_v1"
	mock_account_v1 "go-photo/pkg/account_v1/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"testing"
	"time"
)

var PublicKey = "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAqv7uJc+0TqvV9uFPOQeV\nQ/VoDYuYVztUlB5whWxtwRbX4YczgHf77V04QIC5LEuE5+Vo+3eDXgUO43gUb6i7\ntggx3x8n8F6bZgApsrF+uPnNlTHx8p4/uxQWXfrB4IaRG4Xrr9G/KFfjt3+RpQlX\nFlLQKZmHRR5PpOkBKGPvl5ew7NfBGNR4Peexz84WY2Im+DN/zVvENPLSMY4BqGjQ\n8EzlgF5XFFJX6bQ0BXIbMR7+iAed5y9ahLciJbWNVPaOjyHOf1Rv3TOktU91ZnDX\nx0gZIHgDQCQHclURIVSYFZSvx5W8keQ/XsWr5jP/Y44gpzPiGJQchRtYT4/GPj4t\nXQIDAQAB\n-----END PUBLIC KEY-----"

func TestService_Login(t *testing.T) {
	type mockBehavior func(*mock_utils.MockInteface, *mock_account_v1.MockAccountServiceClient, string, string)

	tests := []struct {
		name          string
		email         string
		password      string
		mockBehavior  mockBehavior
		expectedToken string
		expectedError error
	}{
		{
			name:     "Valid",
			email:    "john@doe.ru",
			password: "password",
			mockBehavior: func(u *mock_utils.MockInteface, a *mock_account_v1.MockAccountServiceClient, email, password string) {
				u.EXPECT().
					EncryptPassword(&PublicKey, password).
					Return("encrypted-password", nil).
					Times(1)
				a.EXPECT().
					Login(gomock.Any(), &def.LoginRequest{Email: email, EncryptedPassword: "encrypted-password"}).
					Return(&def.LoginResponse{JwtToken: "jwt-token"}, nil).
					Times(1)

			},
			expectedToken: "jwt-token",
			expectedError: nil,
		},
		{
			name:     "Bad credentials",
			email:    "john@doe.ru",
			password: "password",
			mockBehavior: func(u *mock_utils.MockInteface, a *mock_account_v1.MockAccountServiceClient, email, password string) {
				u.EXPECT().
					EncryptPassword(&PublicKey, password).
					Return("encrypted-password", nil).
					Times(1)
				a.EXPECT().
					Login(gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.NotFound, "user not found")).
					Times(1)
			},
			expectedToken: "",
			expectedError: serviceErr.UserNotFoundError,
		},
		{
			name:     "Invalid public key",
			email:    "john@doe.ru",
			password: "password",
			mockBehavior: func(u *mock_utils.MockInteface, a *mock_account_v1.MockAccountServiceClient, email, password string) {
				u.EXPECT().
					EncryptPassword(&PublicKey, password).
					Return("", utils.InvalidPublicKeyError).
					Times(1)
			},
			expectedToken: "",
			expectedError: serviceErr.UnexpectedError,
		},
		{
			name:     "Enrypt password error",
			email:    "john@doe.ru",
			password: "password",
			mockBehavior: func(u *mock_utils.MockInteface, a *mock_account_v1.MockAccountServiceClient, email, password string) {
				u.EXPECT().
					EncryptPassword(&PublicKey, password).
					Return("", errors.New("some encrypt password func error")).
					Times(1)
			},
			expectedToken: "",
			expectedError: serviceErr.UnexpectedError,
		},
		{
			name:     "GRPC internal error",
			email:    "john@doe.ru",
			password: "password",
			mockBehavior: func(u *mock_utils.MockInteface, a *mock_account_v1.MockAccountServiceClient, email, password string) {
				u.EXPECT().
					EncryptPassword(&PublicKey, password).
					Return("encrypted-password", nil).
					Times(1)
				a.EXPECT().
					Login(gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.Internal, "some grpc internal error")).
					Times(1)
			},
			expectedToken: "",
			expectedError: serviceErr.UnexpectedError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUtils := mock_utils.NewMockInteface(ctrl)
			mockAccountClient := mock_account_v1.NewMockAccountServiceClient(ctrl)
			s := NewService(mockAccountClient, mockUtils)
			s.publicKeyCache = publicKeyCache{
				key: PublicKey,
				ttl: time.Now().Add(time.Minute),
			}

			tt.mockBehavior(mockUtils, mockAccountClient, tt.email, tt.password)

			token, err := s.Login(nil, tt.email, tt.password)
			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedToken, token)
			}
		})
	}
}

func TestService_Register(t *testing.T) {
	type mockBehavior func(*mock_utils.MockInteface, *mock_account_v1.MockAccountServiceClient, serviceUserModel.RegisterParams)

	tests := []struct {
		name          string
		params        serviceUserModel.RegisterParams
		mockBehavior  mockBehavior
		expectedInfo  serviceUserModel.RegisterInfo
		expectedError error
	}{
		{
			name: "Valid",
			params: serviceUserModel.RegisterParams{
				Email:    "john@doe.ru",
				Password: "password",
			},
			mockBehavior: func(u *mock_utils.MockInteface, a *mock_account_v1.MockAccountServiceClient, params serviceUserModel.RegisterParams) {
				u.EXPECT().
					EncryptPassword(&PublicKey, params.Password).
					Return("encrypted-password", nil).
					Times(1)
				a.EXPECT().Signup(gomock.Any(), gomock.Any()).
					Return(&def.CreateResponse{
						Uuid:     "uuid",
						JwtToken: "jwt-token"}, nil).
					Times(1)
			},
			expectedInfo: serviceUserModel.RegisterInfo{
				UserUUID: "uuid",
				Token:    "jwt-token",
			},
			expectedError: nil,
		},
		{
			name: "User already exists",
			params: serviceUserModel.RegisterParams{
				Email:    "john@doe.ru",
				Password: "password",
			},
			mockBehavior: func(u *mock_utils.MockInteface, a *mock_account_v1.MockAccountServiceClient, params serviceUserModel.RegisterParams) {
				u.EXPECT().
					EncryptPassword(&PublicKey, params.Password).
					Return("encrypted-password", nil).
					Times(1)
				a.EXPECT().
					Signup(gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.AlreadyExists, "user already exists")).
					Times(1)
			},
			expectedInfo:  serviceUserModel.RegisterInfo{},
			expectedError: serviceErr.UserAlreadyExistsError,
		},
		{
			name: "Enrypt password error",
			params: serviceUserModel.RegisterParams{
				Email:    "john@doe.ru",
				Password: "password",
			},
			mockBehavior: func(u *mock_utils.MockInteface, a *mock_account_v1.MockAccountServiceClient, params serviceUserModel.RegisterParams) {
				u.EXPECT().
					EncryptPassword(&PublicKey, params.Password).
					Return("", errors.New("some encrypt password func error")).
					Times(1)
			},
			expectedInfo:  serviceUserModel.RegisterInfo{},
			expectedError: serviceErr.UnexpectedError,
		},
		{
			name: "Invalid public key",
			params: serviceUserModel.RegisterParams{
				Email:    "john@doe.ru",
				Password: "password",
			},
			mockBehavior: func(u *mock_utils.MockInteface, a *mock_account_v1.MockAccountServiceClient, params serviceUserModel.RegisterParams) {
				u.EXPECT().
					EncryptPassword(&PublicKey, params.Password).
					Return("", utils.InvalidPublicKeyError).
					Times(1)
			},
			expectedInfo:  serviceUserModel.RegisterInfo{},
			expectedError: serviceErr.UnexpectedError,
		},
		{
			name: "GRPC internal error",
			params: serviceUserModel.RegisterParams{
				Email:    "john@doe.ru",
				Password: "password",
			},
			mockBehavior: func(u *mock_utils.MockInteface, a *mock_account_v1.MockAccountServiceClient, params serviceUserModel.RegisterParams) {
				u.EXPECT().
					EncryptPassword(&PublicKey, params.Password).
					Return("encrypted-password", nil).
					Times(1)
				a.EXPECT().
					Signup(gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.Internal, "some grpc internal error")).
					Times(1)
			},
			expectedInfo:  serviceUserModel.RegisterInfo{},
			expectedError: serviceErr.UnexpectedError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUtils := mock_utils.NewMockInteface(ctrl)
			mockAccountClient := mock_account_v1.NewMockAccountServiceClient(ctrl)
			tt.mockBehavior(mockUtils, mockAccountClient, tt.params)

			s := NewService(mockAccountClient, mockUtils)
			s.publicKeyCache = publicKeyCache{
				key: PublicKey,
				ttl: time.Now().Add(time.Minute),
			}

			info, err := s.Register(nil, tt.params)
			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedInfo, info)
			}
		})
	}
}

func TestService_getPublicKey(t *testing.T) {
	tests := []struct {
		name          string
		mockBehavior  func(*mock_account_v1.MockAccountServiceClient)
		expectedError error
	}{
		{
			name: "Valid",
			mockBehavior: func(m *mock_account_v1.MockAccountServiceClient) {
				m.EXPECT().GetPublicKey(gomock.Any(), gomock.Any()).Return(&def.GetPublicKeyResponse{
					PublicKey: PublicKey,
				}, nil)
			},
			expectedError: nil,
		},
		{
			name: "Internal error",
			mockBehavior: func(m *mock_account_v1.MockAccountServiceClient) {
				m.EXPECT().GetPublicKey(gomock.Any(), gomock.Any()).Return(nil, status.Error(codes.Internal, "internal error"))
			},
			expectedError: serviceErr.UnexpectedError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAccountClient := mock_account_v1.NewMockAccountServiceClient(ctrl)
			mockUtils := mock_utils.NewMockInteface(ctrl)
			s := NewService(mockAccountClient, mockUtils)

			tt.mockBehavior(mockAccountClient)

			key, err := s.getPublicKey(nil)
			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, PublicKey, *key)
			}
		})
	}
}

func TestService_VerifyToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAccountClient := mock_account_v1.NewMockAccountServiceClient(ctrl)

	svc := NewService(mockAccountClient, nil)

	ctx := context.Background()

	t.Run("Valid token", func(t *testing.T) {
		validToken := "valid-token"
		expectedRequest := &def.VerifyTokenRequest{JwtToken: validToken}

		fakeResponse := &def.VerifyTokenResponse{
			Uuid: "12345",
		}

		mockAccountClient.EXPECT().
			VerifyToken(gomock.Any(), expectedRequest).
			Return(fakeResponse, nil).
			Times(1)

		payload, err := svc.VerifyToken(ctx, validToken)
		require.NoError(t, err)
		assert.Equal(t, "12345", payload.UserUUID)
	})

	t.Run("Invalid token", func(t *testing.T) {
		invalidToken := "invalid-token"
		expectedRequest := &def.VerifyTokenRequest{JwtToken: invalidToken}

		grpcErr := status.Error(codes.NotFound, "user not found")
		mockAccountClient.EXPECT().
			VerifyToken(gomock.Any(), expectedRequest).
			Return(nil, grpcErr).
			Times(1)

		payload, err := svc.VerifyToken(ctx, invalidToken)
		require.Error(t, err)
		assert.True(t, errors.Is(err, serviceErr.UserNotFoundError))
		assert.Equal(t, serviceUserModel.TokenPayload{}, payload)
	})
}
