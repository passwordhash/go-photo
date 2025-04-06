package photo

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go-photo/internal/model"
	mock_repository "go-photo/internal/repository/mock"
	repoModel "go-photo/internal/repository/photo/model"
	"os"
	"path/filepath"
	"testing"
)

func TestPhotoService_GetPhotoFileByVersionAndToken(t *testing.T) {
	type mockBehavior func(*mock_repository.MockPhotoRepository, string, string)

	tests := []struct {
		name          string
		inputToken    string
		inputVersion  string
		mockBehavior  mockBehavior
		expectedBytes []byte
		expectedError error
	}{
		{
			name:         "Valid",
			inputToken:   "token",
			inputVersion: "original",
			mockBehavior: func(repo *mock_repository.MockPhotoRepository, token string, version string) {
				photoVersion := &repoModel.PhotoVersion{
					Filepath: "test.png",
					Size:     4,
				}
				versionType, _ := model.ParseVersionType(version)
				repo.EXPECT().GetPhotoVersionByToken(gomock.Any(), token, &repoModel.FilterParams{
					VersionType: versionType,
				}).Return(photoVersion, nil)
			},
			expectedBytes: []byte("test"),
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			tmpFileName := "test.png"
			tmpFilePath := filepath.Join(tmpDir, tmpFileName)

			content := []byte("test")
			err := os.WriteFile(tmpFilePath, content, 0644)
			require.NoError(t, err)

			defer os.Remove(tmpFileName)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock_repository.NewMockPhotoRepository(ctrl)
			tt.mockBehavior(mockRepo, tt.inputToken, tt.inputVersion)

			s := NewService(Deps{StorageFolderPath: tmpDir}, mockRepo)

			bytes, err := s.GetPhotoFileByVersionAndToken(context.TODO(), tt.inputToken, tt.inputVersion)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBytes, bytes)
			}
		})
	}
}
