package photo

import (
	"context"
	"errors"
	"github.com/golang/mock/gomock"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go-photo/internal/model"
	mock_repository "go-photo/internal/repository/mock"
	repoModel "go-photo/internal/repository/photo/model"
	serviceErr "go-photo/internal/service/error"
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
					PhotoID:      123,
					UUIDFilename: "test.png",
					Size:         4,
				}
				versionType, _ := model.ParseVersionType(version)

				repo.EXPECT().GetPhotoVersionByToken(gomock.Any(), token, &repoModel.FilterParams{
					VersionType: versionType,
				}).Return(photoVersion, nil)

				repo.EXPECT().GetPhotoByID(gomock.Any(), photoVersion.PhotoID).
						Return(&repoModel.Photo{
							UserUUID: "some-user-uuid",
						}, nil)
			},
			expectedBytes: []byte("test"),
			expectedError: nil,
		},
		{
			name:         "Valid - small thumbnail version",
			inputToken:   "thumb_token",
			inputVersion: "thumbnail",
			mockBehavior: func(repo *mock_repository.MockPhotoRepository, token string, version string) {
				photoVersion := &repoModel.PhotoVersion{
					UUIDFilename: "test.png",
					Size:         4,
				}
				versionType, _ := model.ParseVersionType(version)
				repo.EXPECT().GetPhotoVersionByToken(gomock.Any(), token, &repoModel.FilterParams{
					VersionType: versionType,
				}).Return(photoVersion, nil)

				repo.EXPECT().GetPhotoByID(gomock.Any(), photoVersion.PhotoID).
						Return(&repoModel.Photo{
							UserUUID: "some-user-uuid",
						}, nil)
			},
			expectedBytes: []byte("test"),
			expectedError: nil,
		},
		{
			name:         "Valid - preview version",
			inputToken:   "preview_token",
			inputVersion: "preview",
			mockBehavior: func(repo *mock_repository.MockPhotoRepository, token string, version string) {
				photoVersion := &repoModel.PhotoVersion{
					UUIDFilename: "test.png",
					Size:         4,
				}
				versionType, _ := model.ParseVersionType(version)
				repo.EXPECT().GetPhotoVersionByToken(gomock.Any(), token, &repoModel.FilterParams{
					VersionType: versionType,
				}).Return(photoVersion, nil)

				repo.EXPECT().GetPhotoByID(gomock.Any(), photoVersion.PhotoID).
						Return(&repoModel.Photo{
							UserUUID: "some-user-uuid",
						}, nil)
			},
			expectedBytes: []byte("test"),
			expectedError: nil,
		},
		{
			name:         "Invalid Version",
			inputToken:   "token",
			inputVersion: "unknown_version",
			mockBehavior: func(repo *mock_repository.MockPhotoRepository, token string, version string) {
			},
			expectedBytes: nil,
			expectedError: serviceErr.InvalidVersionTypeError,
		},
		{
			name:         "Repo returns error",
			inputToken:   "token",
			inputVersion: "original",
			mockBehavior: func(repo *mock_repository.MockPhotoRepository, token string, version string) {
				versionType, _ := model.ParseVersionType(version)
				repo.EXPECT().GetPhotoVersionByToken(gomock.Any(), token, &repoModel.FilterParams{
					VersionType: versionType,
				}).Return(nil, errors.New("db error"))
			},
			expectedBytes: nil,
			expectedError: serviceErr.UnexpectedError,
		},
		{
			name:         "File not found",
			inputToken:   "token",
			inputVersion: "original",
			mockBehavior: func(repo *mock_repository.MockPhotoRepository, token string, version string) {
				photoVersion := &repoModel.PhotoVersion{
					UUIDFilename: "nonexistent.png",
					Size:         10,
				}
				versionType, _ := model.ParseVersionType(version)
				repo.EXPECT().GetPhotoVersionByToken(gomock.Any(), token, &repoModel.FilterParams{
					VersionType: versionType,
				}).Return(photoVersion, nil)

				repo.EXPECT().GetPhotoByID(gomock.Any(), photoVersion.PhotoID).
						Return(&repoModel.Photo{
							UserUUID: "some-user-uuid",
						}, nil)
			},
			expectedBytes: nil,
			expectedError: serviceErr.UnexpectedError,
		},
		{
			name:         "Error reading file",
			inputToken:   "token",
			inputVersion: "original",
			mockBehavior: func(repo *mock_repository.MockPhotoRepository, token string, version string) {
				versionType, _ := model.ParseVersionType(version)
				photoVersion := &repoModel.PhotoVersion{
					UUIDFilename: "truncated_file.png",
					Size:         1000, // deliberately large size
				}
				repo.EXPECT().GetPhotoVersionByToken(gomock.Any(), token, &repoModel.FilterParams{
					VersionType: versionType,
				}).Return(photoVersion, nil)

				repo.EXPECT().GetPhotoByID(gomock.Any(), photoVersion.PhotoID).
						Return(&repoModel.Photo{
							UserUUID: "some-user-uuid",
						}, nil)
			},
			expectedBytes: nil,
			expectedError: serviceErr.UnexpectedError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userUUID := "some-user-uuid"
			tmpFileName := "test.png"

			tmpDir := t.TempDir()
			usrDir := filepath.Join(tmpDir, userUUID)
			assert.NoError(t, os.MkdirAll(usrDir, os.ModePerm|os.ModeDir))

			tmpFilePath := filepath.Join(usrDir, tmpFileName)
			log.Println(tmpFilePath)

			content := []byte("test")
			err := os.WriteFile(tmpFilePath, content, os.ModePerm|os.ModeDir)
			require.NoError(t, err)

			defer os.Remove(tmpFileName)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock_repository.NewMockPhotoRepository(ctrl)
			tt.mockBehavior(mockRepo, tt.inputToken, tt.inputVersion)

			s := NewService(Deps{StorageFolderPath: tmpDir}, mockRepo, nil)

			bytes, err := s.GetPhotoFileByVersionAndToken(context.TODO(), tt.inputToken, tt.inputVersion)
			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBytes, bytes)
			}
		})
	}
}
