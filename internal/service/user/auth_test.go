package user

import (
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	serviceErr "go-photo/internal/service/error"
	serviceUserModel "go-photo/internal/service/user/model"
	def "go-photo/pkg/account_v1"
	mock_account_v1 "go-photo/pkg/account_v1/mocks"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"testing"
	"time"
)

const PublicKey = "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAqv7uJc+0TqvV9uFPOQeV\nQ/VoDYuYVztUlB5whWxtwRbX4YczgHf77V04QIC5LEuE5+Vo+3eDXgUO43gUb6i7\ntggx3x8n8F6bZgApsrF+uPnNlTHx8p4/uxQWXfrB4IaRG4Xrr9G/KFfjt3+RpQlX\nFlLQKZmHRR5PpOkBKGPvl5ew7NfBGNR4Peexz84WY2Im+DN/zVvENPLSMY4BqGjQ\n8EzlgF5XFFJX6bQ0BXIbMR7+iAed5y9ahLciJbWNVPaOjyHOf1Rv3TOktU91ZnDX\nx0gZIHgDQCQHclURIVSYFZSvx5W8keQ/XsWr5jP/Y44gpzPiGJQchRtYT4/GPj4t\nXQIDAQAB\n-----END PUBLIC KEY-----"

func TestService_Login(t *testing.T) {
	type encryptBehavior func(string, string) (string, error)
	type mockBehavior func(*mock_account_v1.MockAccountServiceClient, string, string)

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
			mockBehavior: func(m *mock_account_v1.MockAccountServiceClient, email, password string) {
				m.EXPECT().Login(gomock.Any(), gomock.Any()).Return(&def.LoginResponse{
					JwtToken: "jwt-token",
				}, nil)
			},
			expectedToken: "jwt-token",
			expectedError: nil,
		},
		{
			name:     "Bad credentials",
			email:    "john@doe.ru",
			password: "password",
			mockBehavior: func(m *mock_account_v1.MockAccountServiceClient, email, password string) {
				m.EXPECT().Login(gomock.Any(), gomock.Any()).Return(nil, status.Error(codes.NotFound, "user not found"))
			},
			expectedToken: "",
			expectedError: serviceErr.UserNotFoundError,
		},
		{
			name:     "Internal error",
			email:    "john@doe.ru",
			password: "password",
			mockBehavior: func(m *mock_account_v1.MockAccountServiceClient, email, password string) {
				m.EXPECT().Login(gomock.Any(), gomock.Any()).Return(nil, status.Error(codes.Internal, "internal error"))
			},
			expectedToken: "",
			expectedError: serviceErr.InternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAccountClient := mock_account_v1.NewMockAccountServiceClient(ctrl)
			s := NewService(mockAccountClient)
			s.publicKeyCache = publicKeyCache{
				key: PublicKey,
				ttl: time.Now().Add(time.Minute),
			}

			tt.mockBehavior(mockAccountClient, tt.email, tt.password)

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
	type mockBehavior func(*mock_account_v1.MockAccountServiceClient, serviceUserModel.RegisterParams)

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
			mockBehavior: func(m *mock_account_v1.MockAccountServiceClient, params serviceUserModel.RegisterParams) {
				m.EXPECT().Signup(gomock.Any(), gomock.Any()).Return(&def.CreateResponse{
					Uuid:     "uuid",
					JwtToken: "jwt-token",
				}, nil)
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
			mockBehavior: func(m *mock_account_v1.MockAccountServiceClient, params serviceUserModel.RegisterParams) {
				m.EXPECT().Signup(gomock.Any(), gomock.Any()).Return(nil, status.Error(codes.AlreadyExists, "user already exists"))
			},
			expectedInfo:  serviceUserModel.RegisterInfo{},
			expectedError: serviceErr.UserAlreadyExistsError,
		},
		{
			name: "Internal error",
			params: serviceUserModel.RegisterParams{
				Email:    "john@doe.ru",
				Password: "password",
			},
			mockBehavior: func(m *mock_account_v1.MockAccountServiceClient, params serviceUserModel.RegisterParams) {
				m.EXPECT().Signup(gomock.Any(), gomock.Any()).Return(nil, status.Error(codes.Internal, "internal error"))
			},
			expectedInfo:  serviceUserModel.RegisterInfo{},
			expectedError: serviceErr.InternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAccountClient := mock_account_v1.NewMockAccountServiceClient(ctrl)
			tt.mockBehavior(mockAccountClient, tt.params)

			s := NewService(mockAccountClient)
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
			expectedError: serviceErr.InternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAccountClient := mock_account_v1.NewMockAccountServiceClient(ctrl)
			tt.mockBehavior(mockAccountClient)

			s := NewService(mockAccountClient)
			key, err := s.getPublicKey(nil)
			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, PublicKey, key)
			}
		})
	}
}
