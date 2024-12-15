package user

import (
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	serviceErr "go-photo/internal/service/error"
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
		request       def.LoginRequest
		mockBehavior  mockBehavior
		expectedToken string
		expectedError error
	}{
		{
			name:     "Valid",
			email:    "john@doe.ru",
			password: "password",
			request: def.LoginRequest{
				Email:             "john@doe.ru",
				EncryptedPassword: "password",
			},
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
			request: def.LoginRequest{
				Email:             "john@doe.ru",
				EncryptedPassword: "password",
			},
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
			request: def.LoginRequest{
				Email:             "john@doe.ru",
				EncryptedPassword: "password",
			},
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
