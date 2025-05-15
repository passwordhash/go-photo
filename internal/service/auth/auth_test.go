package auth

import (
	"context"
	serviceAuthModel "go-photo/internal/service/auth/model"
	"io"
	"log/slog"
	"testing"

	def "github.com/passwordhash/protos/gen/go/go-sso"
	"github.com/stretchr/testify/assert"

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
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedUserUUID, userUUID)
			}
		})
	}
}
