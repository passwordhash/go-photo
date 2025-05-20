package photo

import (
	"context"
	repoErr "go-photo/internal/repository/error"
	mock_repository "go-photo/internal/repository/mock"
	repoModel "go-photo/internal/repository/photo/model"
	serviceErr "go-photo/internal/service/error"
	mock_service "go-photo/internal/service/mock"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestService_PublishPhoto(t *testing.T) {
	type mockBehavior func(s *mock_service.MockPhotoService, r *mock_repository.MockPhotoRepository, userUUID string, photoID int)
	tests := []struct {
		name          string
		userUUID      string
		photoID       int
		mockBehavior  mockBehavior
		expectedToken string
		expectedError error
	}{
		{
			name:     "Valid",
			userUUID: "user-uuid",
			photoID:  1,
			mockBehavior: func(s *mock_service.MockPhotoService, r *mock_repository.MockPhotoRepository, userUUID string, photoID int) {
				r.EXPECT().PhotoByID(gomock.Any(), photoID).Return(&repoModel.Photo{
					ID:       1,
					UserUUID: userUUID,
				}, nil).Times(1)

				r.EXPECT().CreatePhotoPublishedInfo(gomock.Any(), photoID).Return("token", nil).Times(1)
			},
			expectedToken: "token",
			expectedError: nil,
		},
		{
			name:     "Photo Not Found",
			userUUID: "user-uuid",
			photoID:  1,
			mockBehavior: func(s *mock_service.MockPhotoService, r *mock_repository.MockPhotoRepository, userUUID string, photoID int) {
				r.EXPECT().PhotoByID(gomock.Any(), photoID).Return(nil, repoErr.NotFoundError).Times(1)
			},
			expectedToken: "",
			expectedError: serviceErr.PhotoNotFoundError,
		},

		{
			name:     "Photo Not Owned",
			userUUID: "user-uuid",
			photoID:  1,
			mockBehavior: func(s *mock_service.MockPhotoService, r *mock_repository.MockPhotoRepository, userUUID string, photoID int) {
				r.EXPECT().PhotoByID(gomock.Any(), photoID).Return(&repoModel.Photo{
					ID:       1,
					UserUUID: "other-user-uuid",
				}, nil).Times(1)
			},
			expectedToken: "",
			expectedError: serviceErr.AccessDeniedError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock_repository.NewMockPhotoRepository(ctrl)
			tt.mockBehavior(mock_service.NewMockPhotoService(ctrl), mockRepo, tt.userUUID, tt.photoID)

			s := NewService(Deps{StorageFolderPath: ""}, mockRepo, nil)

			token, err := s.PublishPhoto(context.TODO(), tt.userUUID, tt.photoID)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedToken, token)
			}
		})
	}

}

func TestService_UnpublishPhoto(t *testing.T) {
	type mockBehavior func(s *mock_service.MockPhotoService, r *mock_repository.MockPhotoRepository, userUUID string, photoID int)
	tests := []struct {
		name          string
		userUUID      string
		photoID       int
		mockBehavior  mockBehavior
		expectedError error
	}{
		{
			name:     "Valid",
			userUUID: "user-uuid",
			photoID:  1,
			mockBehavior: func(s *mock_service.MockPhotoService, r *mock_repository.MockPhotoRepository, userUUID string, photoID int) {
				r.EXPECT().PhotoByID(gomock.Any(), photoID).Return(&repoModel.Photo{
					ID:       1,
					UserUUID: userUUID,
				}, nil).Times(1)

				r.EXPECT().DeletePhotoPublishedInfo(gomock.Any(), photoID).Return(nil).Times(1)
			},
			expectedError: nil,
		},
		{
			name:     "Photo Not Found",
			userUUID: "user-uuid",
			photoID:  1,
			mockBehavior: func(s *mock_service.MockPhotoService, r *mock_repository.MockPhotoRepository, userUUID string, photoID int) {
				r.EXPECT().PhotoByID(gomock.Any(), photoID).Return(nil, repoErr.NotFoundError).Times(1)
			},
			expectedError: serviceErr.PhotoNotFoundError,
		},

		{
			name:     "Photo Not Owned",
			userUUID: "user-uuid",
			photoID:  1,
			mockBehavior: func(s *mock_service.MockPhotoService, r *mock_repository.MockPhotoRepository, userUUID string, photoID int) {
				r.EXPECT().PhotoByID(gomock.Any(), photoID).Return(&repoModel.Photo{
					ID:       1,
					UserUUID: "other-user-uuid",
				}, nil).Times(1)
			},
			expectedError: serviceErr.AccessDeniedError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock_repository.NewMockPhotoRepository(ctrl)
			tt.mockBehavior(mock_service.NewMockPhotoService(ctrl), mockRepo, tt.userUUID, tt.photoID)
			s := NewService(Deps{StorageFolderPath: ""}, mockRepo, nil)
			err := s.UnpublishPhoto(context.TODO(), tt.userUUID, tt.photoID)
			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
