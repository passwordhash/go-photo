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
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedUploaded, uploaded)
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
