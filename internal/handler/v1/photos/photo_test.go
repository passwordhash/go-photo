package photos

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go-photo/internal/handler/response"
	serviceErr "go-photo/internal/service/error"
	mock_service "go-photo/internal/service/mocks"
	serviceModel "go-photo/internal/service/photo/model"
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
					Return(0, &serviceErr.FileAlreadyExistsError{Filename: "test.jpg"}).
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

	defaultUploads := createDefaultUploads(3)

	tests := []struct {
		name               string
		userUUID           string
		multipartBody      func() (*bytes.Buffer, string)
		mockBehavior       mockBehavior
		expectedStatusCode int
		expectedResponse   any
	}{
		{
			name:     "Valid",
			userUUID: "123e4567-e89b-12d3-a456-426614174000",
			multipartBody: func() (*bytes.Buffer, string) {
				return createMultipartBody(3, "tt%d.jpg", "fake image data")
			},
			mockBehavior: func(s *mock_service.MockPhotoService, userUUID string, files []*multipart.FileHeader) {
				s.EXPECT().
					UploadBatchPhotos(gomock.Any(), userUUID, gomock.Any()).
					Return(&defaultUploads, nil).
					Times(1)
			},
			expectedStatusCode: 200,
			expectedResponse: response.UploadBatchPhotosResponse{
				TotalCount:   3,
				SuccessCount: 3,
				UploadInfos:  serviceModel.ToUploadsInfoFromService(defaultUploads.Get()),
			},
		},
		{
			name:     "Second file is not a photo",
			userUUID: "123e4567-e89b-12d3-a456-426614174000",
			multipartBody: func() (*bytes.Buffer, string) {
				return createMultipartBodyMixed(
					[]string{"test1.jpg", "test2.txt"},
					[]string{"fake image data", "not an image"},
				)
			},
			mockBehavior:       func(s *mock_service.MockPhotoService, userUUID string, files []*multipart.FileHeader) {},
			expectedStatusCode: 400,
			expectedResponse: map[string]interface{}{
				"message": "unsupported file type: test2.txt",
			},
		},
		{
			name:     "Partical success",
			userUUID: "123e4567-e89b-12d3-a456-426614174000",
			multipartBody: func() (*bytes.Buffer, string) {
				return createMultipartBody(3, "tt%d.jpg", "fake image data")
			},
			mockBehavior: func(s *mock_service.MockPhotoService, userUUID string, files []*multipart.FileHeader) {
				uploads := createPartialUploads()
				s.EXPECT().
					UploadBatchPhotos(gomock.Any(), userUUID, gomock.Any()).
					Return(uploads, serviceErr.ParticalSuccessError).
					Times(1)
			},
			expectedStatusCode: 206,
			expectedResponse: response.UploadBatchPhotosResponse{
				TotalCount:   3,
				SuccessCount: 2,
				UploadInfos:  serviceModel.ToUploadsInfoFromService(createPartialUploads().Get()),
			},
		},
		{
			name:     "All failed",
			userUUID: "123e4567-e89b-12d3-a456-426614174000",
			multipartBody: func() (*bytes.Buffer, string) {
				return createMultipartBody(2, "tt%d.jpg", "fake image data")
			},
			mockBehavior: func(s *mock_service.MockPhotoService, userUUID string, files []*multipart.FileHeader) {
				uploads := createFailedUploads()
				s.EXPECT().
					UploadBatchPhotos(gomock.Any(), userUUID, gomock.Any()).
					Return(uploads, serviceErr.AllFailedError).
					Times(1)
			},
			expectedStatusCode: 400,
			expectedResponse: response.UploadBatchPhotosResponse{
				TotalCount:   2,
				SuccessCount: 0,
				UploadInfos:  serviceModel.ToUploadsInfoFromService(createFailedUploads().Get()),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockPhotoService := mock_service.NewMockPhotoService(ctrl)
			tt.mockBehavior(mockPhotoService, tt.userUUID, nil)

			h := NewPhotosHandler(mockPhotoService)

			r := gin.New()
			r.POST("/uploadBatch", h.uploadBatchPhotos)

			w := httptest.NewRecorder()
			body, contentType := tt.multipartBody()
			req := httptest.NewRequest("POST", "/uploadBatch", body)
			req.Header.Set("Content-Type", contentType)

			r.ServeHTTP(w, req)

			content, err := json.Marshal(tt.expectedResponse)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedStatusCode, w.Code)
			assert.JSONEq(t, string(content), w.Body.String())
		})
	}
}

// Вспомогательные функции

func createMultipartBody(count int, filenamePattern, content string) (*bytes.Buffer, string) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for i := 1; i <= count; i++ {
		fileWriter, _ := writer.CreateFormFile(FormPhotoBatchFiles, fmt.Sprintf(filenamePattern, i))
		fileWriter.Write([]byte(content))
	}

	writer.Close()
	return body, writer.FormDataContentType()
}

func createMultipartBodyMixed(filenames, contents []string) (*bytes.Buffer, string) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for i, filename := range filenames {
		fileWriter, _ := writer.CreateFormFile(FormPhotoBatchFiles, filename)
		fileWriter.Write([]byte(contents[i]))
	}

	writer.Close()
	return body, writer.FormDataContentType()
}

func createDefaultUploads(count int) serviceModel.UploadInfoList {
	uploads := serviceModel.UploadInfoList{}
	for i := 1; i <= count; i++ {
		uploads.Add(serviceModel.UploadInfo{
			PhotoID:  i,
			Filename: fmt.Sprintf("tt%d.jpg", i),
		})
	}
	return uploads
}

func createPartialUploads() *serviceModel.UploadInfoList {
	return serviceModel.NewUploadInfoList([]serviceModel.UploadInfo{
		{PhotoID: 1, Filename: "tt1.jpg"},
		{PhotoID: 0, Filename: "tt2.jpg", Error: serviceErr.DbError},
		{PhotoID: 3, Filename: "tt3.jpg"},
	})
}

func createFailedUploads() *serviceModel.UploadInfoList {
	return serviceModel.NewUploadInfoList([]serviceModel.UploadInfo{
		{PhotoID: 0, Filename: "tt1.jpg", Error: serviceErr.DbError},
		{PhotoID: 0, Filename: "tt2.jpg", Error: serviceErr.ServiceError},
	})
}
