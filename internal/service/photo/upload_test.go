package photo

import (
	"bytes"
	"context"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	mock_repository "go-photo/internal/repository/mocks"
	serviceErr "go-photo/internal/service/error"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

func TestService_UploadBatchPhotos(t *testing.T) {
	type mockBehavior func(repo *mock_repository.MockPhotoRepository, userUUID string, photoFiles []*multipart.FileHeader)

	tests := []struct {
		name             string
		userUUID         string
		files            func() []*multipart.FileHeader
		mockBehavior     mockBehavior
		expectedUploaded []string
		expectedError    error
	}{
		{
			name:     "Valid",
			userUUID: "user-id",
			files: func() []*multipart.FileHeader {
				return []*multipart.FileHeader{
					mockFileHeader("test1.jpg", 100, "file content"),
					mockFileHeader("test2.jpg", 100, "file content"),
				}
			},
			mockBehavior: func(repo *mock_repository.MockPhotoRepository, userUUID string, photoFiles []*multipart.FileHeader) {
				for i := range photoFiles {
					repo.EXPECT().CreateOriginalPhoto(gomock.Any(), gomock.Any()).Return(i+1, nil).Times(1)
				}
			},
			expectedUploaded: []string{"test1.jpg", "test2.jpg"},
			expectedError:    nil,
		},
		{
			name:     "Disk Save Error",
			userUUID: "user-id",
			files: func() []*multipart.FileHeader {
				return []*multipart.FileHeader{
					mockFileHeader("test1.jpg", 100, "file content"),
					mockFileHeader("test2.jpg", 100, "file content"),
				}
			},
			mockBehavior: func(repo *mock_repository.MockPhotoRepository, userUUID string, photoFiles []*multipart.FileHeader) {
				repo.EXPECT().CreateOriginalPhoto(gomock.Any(), gomock.Any()).Times(1)
				repo.EXPECT().CreateOriginalPhoto(gomock.Any(), gomock.Any()).Times(1).Return(0, serviceErr.DbError)
			},
			expectedUploaded: []string{"test1.jpg"},
			expectedError:    serviceErr.ParticalSuccessError,
		},
		{
			name:     "DB Save Error",
			userUUID: "user-id",
			files: func() []*multipart.FileHeader {
				return []*multipart.FileHeader{
					mockFileHeader("test1.jpg", 100, "file content"),
					mockFileHeader("test2.jpg", 100, "file content"),
				}
			},
			mockBehavior: func(repo *mock_repository.MockPhotoRepository, userUUID string, photoFiles []*multipart.FileHeader) {
				repo.EXPECT().CreateOriginalPhoto(gomock.Any(), gomock.Any()).
					Return(0, fmt.Errorf("db error")).Times(2)
			},
			expectedUploaded: nil,
			expectedError:    serviceErr.AllFailedError,
		},
		{
			name:     "Partial Error",
			userUUID: "user-id",
			files: func() []*multipart.FileHeader {
				return []*multipart.FileHeader{
					mockFileHeader("test1.jpg", 100, "file content"),
					mockFileHeader("test2.jpg", 100, "file content"),
					mockFileHeader("test3.jpg", 100, "file content"),
				}
			},
			mockBehavior: func(repo *mock_repository.MockPhotoRepository, userUUID string, photoFiles []*multipart.FileHeader) {
				repo.EXPECT().CreateOriginalPhoto(gomock.Any(), gomock.Any()).Return(1, nil).Times(1)
				repo.EXPECT().CreateOriginalPhoto(gomock.Any(), gomock.Any()).Return(0, fmt.Errorf("db error")).Times(1)
				repo.EXPECT().CreateOriginalPhoto(gomock.Any(), gomock.Any()).Return(2, nil).Times(1)
			},
			expectedUploaded: []string{"test1.jpg", "test3.jpg"},
			expectedError:    serviceErr.ParticalSuccessError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storageDir := t.TempDir()
			defer os.RemoveAll(storageDir)

			c := gomock.NewController(t)
			defer c.Finish()

			mockRepo := mock_repository.NewMockPhotoRepository(c)
			tt.mockBehavior(mockRepo, tt.userUUID, tt.files())

			s := NewService(Deps{StorageFolderPath: storageDir}, mockRepo)

			uploaded, err := s.UploadBatchPhotos(context.Background(), tt.userUUID, tt.files())

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				var uploadedFiles []string
				for _, file := range uploaded.Get() {
					uploadedFiles = append(uploadedFiles, file.Filename)
				}
				assert.ElementsMatch(t, tt.expectedUploaded, uploadedFiles)
			}
		})
	}
}

func TestEnsureUserFolder(t *testing.T) {
	t.Run("Create directory successfully", func(t *testing.T) {
		tmpDir := t.TempDir()
		defer os.RemoveAll(tmpDir)

		userUUID := "test-user"
		expectedPath := filepath.Join(tmpDir, userUUID)

		actualPath, err := ensureUserFolder(tmpDir, userUUID)
		assert.NoError(t, err)
		assert.Equal(t, expectedPath, actualPath)

		info, err := os.Stat(expectedPath)
		assert.NoError(t, err)
		assert.True(t, info.IsDir())
	})

	t.Run("Error creating directory", func(t *testing.T) {
		invalidDir := "/invalid/path"

		_, err := ensureUserFolder(invalidDir, "test-user")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create directory")
	})
}

func TestSaveFile(t *testing.T) {
	t.Run("Save file successfully", func(t *testing.T) {
		tmpDir := t.TempDir()
		defer os.RemoveAll(tmpDir)

		file := mockFileHeader("test.jpg", 100, "file content")
		destPath := filepath.Join(tmpDir, "test.jpg")

		err := saveFileToDisk(file, destPath)
		assert.NoError(t, err)

		_, err = os.Stat(destPath)
		assert.NoError(t, err)

		content, err := os.ReadFile(destPath)
		assert.NoError(t, err)
		assert.Equal(t, "file content", string(content))
	})

	t.Run("Error opening file", func(t *testing.T) {
		file := mockFileHeader("test.jpg", 100, "")

		destPath := "/invalid/path/test.jpg"
		err := saveFileToDisk(file, destPath)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create file")
	})
}

func mockFileHeader(filename string, size int64, content string) *multipart.FileHeader {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		panic(fmt.Sprintf("failed to create form file: %v", err))
	}
	_, _ = part.Write([]byte(content)) // Пишем содержимое файл

	writer.Close()

	req := &http.Request{Header: http.Header{"Content-Type": {writer.FormDataContentType()}}}
	req.Body = ioutil.NopCloser(body)

	err = req.ParseMultipartForm(10 << 20) // Парсим с лимитом размера
	if err != nil {
		panic(fmt.Sprintf("failed to parse multipart form: %v", err))
	}

	fileHeaders := req.MultipartForm.File["file"]
	if len(fileHeaders) == 0 {
		panic("no file headers found")
	}

	return fileHeaders[0]
}
