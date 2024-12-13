package photo

import (
	"bytes"
	"context"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	mock_repository "go-photo/internal/repository/mocks"
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
		photoFiles       func() []*multipart.FileHeader
		mockBehavior     mockBehavior
		expectedUploaded []string
		expectedError    error
	}{
		{
			name:     "Valid",
			userUUID: "user-id",
			photoFiles: func() []*multipart.FileHeader {
				return []*multipart.FileHeader{
					mockFileHeader("test1.jpg", 100, "content1"),
					mockFileHeader("test2.jpg", 200, "content2"),
				}
			},
			mockBehavior: func(repo *mock_repository.MockPhotoRepository, userUUID string, photoFiles []*multipart.FileHeader) {
				repo.EXPECT().CreateOriginalPhoto(gomock.Any(), gomock.Any()).Return(1, nil).Times(1)
				repo.EXPECT().CreateOriginalPhoto(gomock.Any(), gomock.Any()).Return(2, nil).Times(1)
			},
			expectedUploaded: []string{"test1.jpg", "test2.jpg"},
			expectedError:    nil,
		},
		{
			name:     "File already exists",
			userUUID: "user-id",
			photoFiles: func() []*multipart.FileHeader {
				return []*multipart.FileHeader{
					mockFileHeader("test1.jpg", 100, "content1"),
					mockFileHeader("test1.jpg", 100, "content1"),
				}
			},
			mockBehavior: func(repo *mock_repository.MockPhotoRepository, userUUID string, photoFiles []*multipart.FileHeader) {
				repo.EXPECT().CreateOriginalPhoto(gomock.Any(), gomock.Any())
			},
			expectedUploaded: nil,
			expectedError:    &FileAlreadyExistsError{Filename: "test1.jpg"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storageDir := t.TempDir()
			defer os.RemoveAll(storageDir)

			c := gomock.NewController(t)
			defer c.Finish()

			mockRepo := mock_repository.NewMockPhotoRepository(c)
			tt.mockBehavior(mockRepo, tt.userUUID, tt.photoFiles())

			s := NewService(Deps{StorageFolderPath: storageDir}, mockRepo)

			uploaded, err := s.UploadBatchPhotos(context.Background(), tt.userUUID, tt.photoFiles())
			if tt.expectedError != nil {
				assert.ErrorAs(t, err, &tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedUploaded, uploaded)
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

		err := saveFile(file, destPath)
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
		err := saveFile(file, destPath)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create file")
	})
}

func TestRemovePhotoFromDisk(t *testing.T) {
	t.Run("Remove file successfully", func(t *testing.T) {
		tmpDir := t.TempDir()
		defer os.RemoveAll(tmpDir)

		testFile := filepath.Join(tmpDir, "test.jpg")
		err := os.WriteFile(testFile, []byte("test content"), 0644)
		assert.NoError(t, err)

		err = removePhotoFromDisk(testFile)
		assert.NoError(t, err)

		_, err = os.Stat(testFile)
		assert.Error(t, err)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("Error removing file", func(t *testing.T) {
		nonExistentFile := "/invalid/path/test.jpg"

		err := removePhotoFromDisk(nonExistentFile)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to remove photo file")
	})
}

func TestProcessFile(t *testing.T) {
	type mockBehavior func(repo *mock_repository.MockPhotoRepository, userUUID string)

	tests := []struct {
		name          string
		userUUID      string
		file          func() *multipart.FileHeader
		mockBehavior  mockBehavior
		expectedID    int
		expectedError error
	}{
		{
			name:     "Valid",
			userUUID: "user-uuid",
			file: func() *multipart.FileHeader {
				return mockFileHeader("test.jpg", 100, "content")
			},
			mockBehavior: func(repo *mock_repository.MockPhotoRepository, userUUID string) {
				repo.EXPECT().CreateOriginalPhoto(gomock.Any(), gomock.Any()).Return(123, nil)
			},
			expectedID:    123,
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFolder := t.TempDir()
			defer os.RemoveAll(tmpFolder)

			userFolder := filepath.Join(tmpFolder, tt.userUUID)
			err := os.MkdirAll(userFolder, 0755)
			assert.NoError(t, err)

			c := gomock.NewController(t)
			defer c.Finish()

			mockRepo := mock_repository.NewMockPhotoRepository(c)
			tt.mockBehavior(mockRepo, tt.userUUID)

			s := NewService(Deps{StorageFolderPath: tmpFolder}, mockRepo)

			id, err := s.processFile(context.Background(), tt.userUUID, tt.file(), userFolder)
			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, id)
			}
		})
	}
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
