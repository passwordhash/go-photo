package auth

import (
	"context"
	serviceAuthModel "go-photo/internal/service/auth/model"
	serviceErr "go-photo/internal/service/error"
	"io"
	"log/slog"
	"testing"

	def "github.com/passwordhash/protos/gen/go/go-sso"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/golang/mock/gomock"
	"github.com/passwordhash/protos/mocks"
)

const (
	appName   = "test-app"
	appSecret = "test-secret"
)

func TestAuthService_Register(t *testing.T) {
	type mockBehavior func(authAPI *mocks.MockAuthClient, params serviceAuthModel.RegisterParams)

	tests := []struct {
		name             string
		inputParams      serviceAuthModel.RegisterParams
		mockBehavior     mockBehavior
		expectedUserUUID string
		expectedError    error
	}{
		{
			name: "Valid",
			inputParams: serviceAuthModel.RegisterParams{
				Email:    "john@doe.com",
				Password: "password",
			},
			mockBehavior: func(authAPI *mocks.MockAuthClient, params serviceAuthModel.RegisterParams) {
				authAPI.EXPECT().Register(gomock.Any(), &def.RegisterRequest{
					Email:    params.Email,
					Password: params.Password,
				}).Return(&def.RegisterResponse{
					UserUuid: "user-id",
				}, nil).Times(1)
			},
			expectedUserUUID: "user-id",
			expectedError:    nil,
		},
		{
			name: "User Already Exists",
			inputParams: serviceAuthModel.RegisterParams{
				Email:    "john@doe.com",
				Password: "password",
			},
			mockBehavior: func(authAPI *mocks.MockAuthClient, params serviceAuthModel.RegisterParams) {
				authAPI.EXPECT().Register(gomock.Any(), &def.RegisterRequest{
					Email:    params.Email,
					Password: params.Password,
				}).Return(nil, status.Error(codes.AlreadyExists, "user already exists")).Times(1)
			},
			expectedUserUUID: "",
			expectedError:    serviceErr.UserAlreadyExistsError,
		},
		{
			name:        "Internal Error",
			inputParams: serviceAuthModel.RegisterParams{},
			mockBehavior: func(authAPI *mocks.MockAuthClient, params serviceAuthModel.RegisterParams) {
				authAPI.EXPECT().Register(gomock.Any(), &def.RegisterRequest{
					Email:    params.Email,
					Password: params.Password,
				}).Return(nil, status.Error(codes.Internal, "internal error")).Times(1)
			},
			expectedUserUUID: "",
			expectedError:    serviceErr.UnexpectedError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			authAPI := mocks.NewMockAuthClient(ctrl)

			log := slog.New(slog.NewTextHandler(io.Discard, nil))

			tt.mockBehavior(authAPI, tt.inputParams)

			s := New(log, authAPI, appName, appSecret)

			userUUID, err := s.Register(context.Background(), tt.inputParams)
			if tt.expectedError != nil {
				assert.Error(t, err)
				// assert.Equal(t, tt.expectedError, err)
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedUserUUID, userUUID)
			}
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	type mockBehavior func(authAPI *mocks.MockAuthClient, email, password string)

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
			email:    "joh@doe.com",
			password: "password",
			mockBehavior: func(authAPI *mocks.MockAuthClient, email string, password string) {
				authAPI.EXPECT().Login(gomock.Any(), &def.LoginRequest{
					Email:    email,
					Password: password,
					AppName:  appName,
				}).Return(&def.LoginResponse{
					Token: "valid-token",
				}, nil).Times(1)
			},
			expectedToken: "valid-token",
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			authAPI := mocks.NewMockAuthClient(ctrl)

			log := slog.New(slog.NewTextHandler(io.Discard, nil))

			tt.mockBehavior(authAPI, tt.email, tt.password)

			s := New(log, authAPI, appName, appSecret)

			token, err := s.Login(context.Background(), tt.email, tt.password)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedToken, token)
			}
		})
	}
}
