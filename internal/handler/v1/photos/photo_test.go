package photos

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	mock_service "go-photo/internal/service/mocks"
	"go-photo/internal/service/photo"
	"io/ioutil"
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
			expectedResponseBody: `{"message":"unsupported file type"}`,
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
			gin.DefaultWriter = ioutil.Discard
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

func TestHandler_uploadBatchPhotos(t *testing.T) {
	type mockBehavior func(s *mock_service.MockPhotoService, userUUID string, files []*multipart.FileHeader)

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

				for i := 1; i <= 3; i++ {
					fileWriter, _ := writer.CreateFormFile(FormPhotoBatchFiles, fmt.Sprintf("test%d.jpg", i))
					fileWriter.Write([]byte("fake image data"))
				}

				writer.Close()
				return body, writer.FormDataContentType()
			},
			mockBehavior: func(s *mock_service.MockPhotoService, userUUID string, files []*multipart.FileHeader) {
				s.EXPECT().
					UploadBatchPhotos(gomock.Any(), userUUID, gomock.Any()).
					Return([]string{"test1.jpg", "test2.jpg", "test3.jpg"}, nil).
					Times(1)
			},
			expectedStatusCode:   200,
			expectedResponseBody: `{"status":"ok","total_count":3,"success_count":3,"uploaded_photos":["test1.jpg","test2.jpg","test3.jpg"]}`,
		},
		{
			name:     "Second file is not a photo",
			userUUID: "123e4567-e89b-12d3-a456-426614174000",
			multipartBody: func() (*bytes.Buffer, string) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)

				// Первый файл - фото
				fileWriter1, _ := writer.CreateFormFile(FormPhotoBatchFiles, "test1.jpg")
				fileWriter1.Write([]byte("fake image data"))

				// Второй файл - не фото
				fileWriter2, _ := writer.CreateFormFile(FormPhotoBatchFiles, "test2.txt")
				fileWriter2.Write([]byte("not an image"))

				writer.Close()
				return body, writer.FormDataContentType()
			},
			mockBehavior:         func(s *mock_service.MockPhotoService, userUUID string, files []*multipart.FileHeader) {},
			expectedStatusCode:   400,
			expectedResponseBody: `{"message":"unsupported file type: test2.txt"}`,
		},
		{
			name:     "Second file already exists",
			userUUID: "123e4567-e89b-12d3-a456-426614174000",
			multipartBody: func() (*bytes.Buffer, string) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)

				// Первый файл - фото
				fileWriter1, _ := writer.CreateFormFile(FormPhotoBatchFiles, "test1.jpg")
				fileWriter1.Write([]byte("fake image data"))

				// Второй файл - фото, но уже существует
				fileWriter2, _ := writer.CreateFormFile(FormPhotoBatchFiles, "test2.jpg")
				fileWriter2.Write([]byte("fake image data"))

				writer.Close()
				return body, writer.FormDataContentType()
			},
			mockBehavior: func(s *mock_service.MockPhotoService, userUUID string, files []*multipart.FileHeader) {
				s.EXPECT().
					UploadBatchPhotos(gomock.Any(), userUUID, gomock.Any()).
					Return([]string{"test1.jpg"}, &photo.FileAlreadyExistsError{Filename: "test2.jpg"}).
					Times(1)
			},
			expectedStatusCode:   200,
			expectedResponseBody: `{"status":"partial_ok","total_count":2,"success_count":1,"uploaded_photos":["test1.jpg"],"error":"file with name 'test2.jpg' already exists in the folder"}`,
		},
		{
			name:     "Internal error",
			userUUID: "123e4567-e89b-12d3-a456-426614174000",
			multipartBody: func() (*bytes.Buffer, string) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)

				fileWriter, _ := writer.CreateFormFile(FormPhotoBatchFiles, "test.jpg")
				fileWriter.Write([]byte("fake image data"))

				writer.Close()
				return body, writer.FormDataContentType()
			},
			mockBehavior: func(s *mock_service.MockPhotoService, userUUID string, files []*multipart.FileHeader) {
				s.EXPECT().
					UploadBatchPhotos(gomock.Any(), userUUID, gomock.Any()).
					Return(nil, assert.AnError).
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
			test.mockBehavior(mockPhotoService, test.userUUID, nil)

			h := NewPhotosHandler(mockPhotoService)

			r := gin.New()
			r.POST("/uploadBatch", h.uploadBatchPhotos)

			w := httptest.NewRecorder()
			body, contentType := test.multipartBody()
			req := httptest.NewRequest("POST", "/uploadBatch", body)
			req.Header.Set("Content-Type", contentType)

			r.ServeHTTP(w, req)

			assert.Equal(t, test.expectedStatusCode, w.Code)
			assert.JSONEq(t, test.expectedResponseBody, w.Body.String())
		})
	}
}
