package photos

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	mock_service "go-photo/internal/service/mocks"
	"go-photo/internal/service/photo"
	"mime/multipart"
	"net/http/httptest"
	"testing"
)

func TestHandler_uploadPhoto(t *testing.T) {
	type mockBehavior func(s *mock_service.MockPhotoService, userUUID string, file multipart.File, filename string)

	tests := []struct {
		name                 string
		userUUID             string
		multipartBody        func() (*bytes.Buffer, string)
		mockBehavior         mockBehavior
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:     "Valid",
			userUUID: "123e4567-e89b-12d3-a456-426614174000",
			multipartBody: func() (*bytes.Buffer, string) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)

				fileWriter, _ := writer.CreateFormFile(FormPhotoFile, "test.jpg")
				fileWriter.Write([]byte("fake image data"))

				writer.Close()
				return body, writer.FormDataContentType()
			},
			mockBehavior: func(s *mock_service.MockPhotoService, userUUID string, file multipart.File, filename string) {
				s.EXPECT().
					UploadPhoto(gomock.Any(), userUUID, gomock.Any()).
					Return(123, nil).
					Times(1)
			},
			expectedStatusCode:   200,
			expectedResponseBody: `{"status":"ok","id":123}`,
		},
		{
			name:     "File not found",
			userUUID: "123e4567-e89b-12d3-a456-426614174000",
			multipartBody: func() (*bytes.Buffer, string) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				writer.Close()
				return body, writer.FormDataContentType()
			},
			mockBehavior:         func(s *mock_service.MockPhotoService, userUUID string, file multipart.File, filename string) {},
			expectedStatusCode:   400,
			expectedResponseBody: `{"message":"file not found"}`,
		},
		{
			name:     "File is not a photo",
			userUUID: "123e4567-e89b-12d3-a456-426614174000",
			multipartBody: func() (*bytes.Buffer, string) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)

				fileWriter, _ := writer.CreateFormFile(FormPhotoFile, "test.txt")
				fileWriter.Write([]byte("fake image data"))

				writer.Close()
				return body, writer.FormDataContentType()
			},
			mockBehavior:         func(s *mock_service.MockPhotoService, userUUID string, file multipart.File, filename string) {},
			expectedStatusCode:   400,
			expectedResponseBody: `{"message":"file is not a photo"}`,
		},
		{
			name:     "File with the same name already exists",
			userUUID: "123e4567-e89b-12d3-a456-426614174000",
			multipartBody: func() (*bytes.Buffer, string) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)

				fileWriter, _ := writer.CreateFormFile(FormPhotoFile, "test.jpg")
				fileWriter.Write([]byte("fake image data"))

				writer.Close()
				return body, writer.FormDataContentType()
			},
			mockBehavior: func(s *mock_service.MockPhotoService, userUUID string, file multipart.File, filename string) {
				s.EXPECT().
					UploadPhoto(gomock.Any(), userUUID, gomock.Any()).
					Return(0, &photo.FileAlreadyExistsError{Filename: "test.jpg"}).
					Times(1)
			},
			expectedStatusCode:   400,
			expectedResponseBody: `{"message":"file with the same name already exists"}`,
		},
		{
			name:     "Internal error",
			userUUID: "123e4567-e89b-12d3-a456-426614174000",
			multipartBody: func() (*bytes.Buffer, string) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)

				fileWriter, _ := writer.CreateFormFile(FormPhotoFile, "test.jpg")
				fileWriter.Write([]byte("fake image data"))

				writer.Close()
				return body, writer.FormDataContentType()
			},
			mockBehavior: func(s *mock_service.MockPhotoService, userUUID string, file multipart.File, filename string) {
				s.EXPECT().
					UploadPhoto(gomock.Any(), userUUID, gomock.Any()).
					Return(0, assert.AnError).
					Times(1)
			},
			expectedStatusCode:   500,
			expectedResponseBody: `{"message":"internal server error"}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockPhotoService := mock_service.NewMockPhotoService(ctrl)
			test.mockBehavior(mockPhotoService, test.userUUID, nil, "test.jpg")

			h := NewPhotosHandler(mockPhotoService)

			r := gin.New()
			r.POST("/upload", h.uploadPhoto)

			w := httptest.NewRecorder()
			body, contentType := test.multipartBody()
			req := httptest.NewRequest("POST", "/upload", body)
			req.Header.Set("Content-Type", contentType)

			r.ServeHTTP(w, req)

			assert.Equal(t, test.expectedStatusCode, w.Code)
			assert.JSONEq(t, test.expectedResponseBody, w.Body.String())
		})
	}
}
