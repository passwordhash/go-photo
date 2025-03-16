package photo

import (
	"bytes"
	"context"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	mock_repository "go-photo/internal/repository/mock"
	serviceErr "go-photo/internal/service/error"
	serviceModel "go-photo/internal/service/photo/model"
	"image"
	"image/color"
	"image/jpeg"
	"io"
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

func TestService_SaveToDatabase(t *testing.T) {
	type mockBehavior func(repo *mock_repository.MockPhotoRepository, ctx context.Context, userUUID string, info serviceModel.UploadInfo)

	tests := []struct {
		name           string
		userUUID       string
		filePath       string
		uploadInfo     serviceModel.UploadInfo
		mockBehavior   mockBehavior
		expectedInfo   serviceModel.UploadInfo
		expectedLogMsg string
	}{
		{
			name:     "Successful Save to Database",
			userUUID: "user-id",
			filePath: "test1.jpg",
			uploadInfo: serviceModel.UploadInfo{
				Filename: "test1.jpg",
				Size:     100,
			},
			mockBehavior: func(repo *mock_repository.MockPhotoRepository, ctx context.Context, userUUID string, info serviceModel.UploadInfo) {
				repo.EXPECT().CreateOriginalPhoto(ctx, gomock.Any()).Return(1, nil).Times(1)
			},
			expectedInfo: serviceModel.UploadInfo{
				Filename: "test1.jpg",
				Size:     100,
				PhotoID:  1,
			},
		},
		{
			name:     "Database Save Error with Rollback",
			userUUID: "user-id",
			filePath: "test1.jpg",
			uploadInfo: serviceModel.UploadInfo{
				Filename: "test1.jpg",
				Size:     100,
			},
			mockBehavior: func(repo *mock_repository.MockPhotoRepository, ctx context.Context, userUUID string, info serviceModel.UploadInfo) {
				repo.EXPECT().CreateOriginalPhoto(ctx, gomock.Any()).Return(0, fmt.Errorf("db error")).Times(1)
			},
			expectedInfo: serviceModel.UploadInfo{
				Filename: "test1.jpg",
				Size:     100,
				Error:    fmt.Errorf("db save error: db error"),
			},
			expectedLogMsg: "Failed to remove file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем временную директорию для хранения файлов
			storageDir := t.TempDir()
			defer os.RemoveAll(storageDir)

			filePath := filepath.Join(storageDir, tt.filePath)
			_ = os.WriteFile(filePath, []byte("file content"), 0644)

			c := gomock.NewController(t)
			defer c.Finish()

			mockRepo := mock_repository.NewMockPhotoRepository(c)
			tt.mockBehavior(mockRepo, context.Background(), tt.userUUID, tt.uploadInfo)

			s := NewService(Deps{StorageFolderPath: storageDir}, mockRepo)

			// Вызываем тестируемый метод
			info := s.saveToDatabase(context.Background(), tt.userUUID, tt.uploadInfo)

			// Проверяем результат
			assert.Equal(t, tt.expectedInfo.Filename, info.Filename)
			assert.Equal(t, tt.expectedInfo.PhotoID, info.PhotoID)
			if tt.expectedInfo.Error != nil {
				assert.Contains(t, info.Error.Error(), tt.expectedInfo.Error.Error())
			} else {
				assert.NoError(t, info.Error)
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

func mockFileHeader(filename string, size int64, content string) *multipart.FileHeader {
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	var imgBuf bytes.Buffer
	if err := jpeg.Encode(&imgBuf, img, nil); err != nil {
		panic(fmt.Sprintf("failed to encode jpeg image: %v", err))
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		panic(fmt.Sprintf("failed to create form file: %v", err))
	}
	_, err = part.Write(imgBuf.Bytes())
	if err != nil {
		panic(fmt.Sprintf("failed to write image to form file: %v", err))
	}
	writer.Close()

	req := &http.Request{Header: http.Header{"Content-Type": {writer.FormDataContentType()}}}
	req.Body = io.NopCloser(body)
	if err = req.ParseMultipartForm(10 << 20); err != nil {
		panic(fmt.Sprintf("failed to parse multipart form: %v", err))
	}

	fileHeaders := req.MultipartForm.File["file"]
	if len(fileHeaders) == 0 {
		panic("no file headers found")
	}

	return fileHeaders[0]
}
