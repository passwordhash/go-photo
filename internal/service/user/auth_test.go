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
)

func TestService_Login(t *testing.T) {
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
				m.EXPECT().Login(gomock.Any(), &def.LoginRequest{
					Email:             email,
					EncryptedPassword: password,
				}).Return(&def.LoginResponse{
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
				m.EXPECT().Login(gomock.Any(), &def.LoginRequest{
					Email:             email,
					EncryptedPassword: password,
				}).Return(nil, status.Error(codes.NotFound, "user not found"))
			},
			expectedToken: "",
			expectedError: serviceErr.UserNotFoundError,
		},
		{
			name:     "Internal error",
			email:    "john@doe.ru",
			password: "password",
			mockBehavior: func(m *mock_account_v1.MockAccountServiceClient, email, password string) {
				m.EXPECT().Login(gomock.Any(), &def.LoginRequest{
					Email:             email,
					EncryptedPassword: password,
				}).Return(nil, status.Error(codes.Internal, "internal error"))
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
			tt.mockBehavior(mockAccountClient, tt.email, tt.password)

			s := NewService(mockAccountClient)

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
				m.EXPECT().Signup(gomock.Any(), &def.CreateRequest{
					Email:             params.Email,
					EncryptedPassword: params.Password,
				}).Return(&def.CreateResponse{
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
				m.EXPECT().Signup(gomock.Any(), &def.CreateRequest{
					Email:             params.Email,
					EncryptedPassword: params.Password,
				}).Return(nil, status.Error(codes.AlreadyExists, "user already exists"))
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
				m.EXPECT().Signup(gomock.Any(), &def.CreateRequest{
					Email:             params.Email,
					EncryptedPassword: params.Password,
				}).Return(nil, status.Error(codes.Internal, "internal error"))
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
