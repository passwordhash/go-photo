package photos

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go-photo/internal/handler/middleware"
	"go-photo/internal/handler/response"
	serviceErr "go-photo/internal/service/error"
	mockservice "go-photo/internal/service/mock"
	serviceModel "go-photo/internal/service/photo/model"
	serviceUserModel "go-photo/internal/service/user/model"
	"io/ioutil"
	"mime/multipart"
	"net/http/httptest"
	"testing"
)

func TestHandler_uploadPhoto(t *testing.T) {
	type mockBehavior func(s *mockservice.MockPhotoService, userUUID string, file multipart.File, filename string)

	tests := []struct {
		name                 string
		userUUID             string
		multipartBody        func() (*bytes.Buffer, string)
		mockBehavior         mockBehavior
		expectedStatusCode   int
		expectedResponseBody any
	}{
		{
			name:     "Valid",
			userUUID: "123e4567-e89b-12d3-a456-426614174000",
			multipartBody: func() (*bytes.Buffer, string) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)

				fileWriter, _ := writer.CreateFormFile(FormPhotoFile, "tt.jpg")
				fileWriter.Write([]byte("fake image data"))

				writer.Close()
				return body, writer.FormDataContentType()
			},
			mockBehavior: func(s *mockservice.MockPhotoService, userUUID string, file multipart.File, filename string) {
				s.EXPECT().
					UploadPhoto(gomock.Any(), userUUID, gomock.Any()).
					Return(123, nil).
					Times(1)
			},
			expectedStatusCode:   200,
			expectedResponseBody: `{"photo_id":123}`,
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
			mockBehavior:       func(s *mockservice.MockPhotoService, userUUID string, file multipart.File, filename string) {},
			expectedStatusCode: 400,
			expectedResponseBody: response.Error{
				Error: response.ParamsMissing,
			},
		},
		{
			name:     "File is not a photo",
			userUUID: "123e4567-e89b-12d3-a456-426614174000",
			multipartBody: func() (*bytes.Buffer, string) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)

				fileWriter, _ := writer.CreateFormFile(FormPhotoFile, "tt.txt")
				fileWriter.Write([]byte("fake image data"))

				writer.Close()
				return body, writer.FormDataContentType()
			},
			mockBehavior:       func(s *mockservice.MockPhotoService, userUUID string, file multipart.File, filename string) {},
			expectedStatusCode: 400,
			expectedResponseBody: response.Error{
				Error: response.UnsupportedFileType,
			},
		},
		{
			name:     "Internal error",
			userUUID: "123e4567-e89b-12d3-a456-426614174000",
			multipartBody: func() (*bytes.Buffer, string) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)

				fileWriter, _ := writer.CreateFormFile(FormPhotoFile, "tt.jpg")
				fileWriter.Write([]byte("fake image data"))

				writer.Close()
				return body, writer.FormDataContentType()
			},
			mockBehavior: func(s *mockservice.MockPhotoService, userUUID string, file multipart.File, filename string) {
				s.EXPECT().
					UploadPhoto(gomock.Any(), userUUID, gomock.Any()).
					Return(0, assert.AnError).
					Times(1)
			},
			expectedStatusCode: 500,
			expectedResponseBody: response.Error{
				Error: response.InternalServerError,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockPhotoService := mockservice.NewMockPhotoService(ctrl)
			tt.mockBehavior(mockPhotoService, tt.userUUID, nil, "tt.jpg")

			mockTokenService := mockservice.NewMockTokenService(ctrl)

			h := NewHandler(mockPhotoService, mockTokenService)

			r := gin.New()
			gin.DefaultWriter = ioutil.Discard

			r.Use(middleware.UserIdentity(func(ctx context.Context, token string) (serviceUserModel.TokenPayload, error) {
				if token == "valid-token" {
					return serviceUserModel.TokenPayload{UserUUID: tt.userUUID}, nil
				}
				return serviceUserModel.TokenPayload{}, errors.New("invalid token")
			}))
			r.POST("/upload", h.uploadPhoto)

			w := httptest.NewRecorder()
			body, contentType := tt.multipartBody()
			req := httptest.NewRequest("POST", "/upload", body)
			req.Header.Set("Content-Type", contentType)
			req.Header.Set("Authorization", "Bearer valid-token")

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatusCode, w.Code)
			switch tt.expectedResponseBody.(type) {
			case response.Error:
				var resp response.Error
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResponseBody.(response.Error).Error, resp.Error)
			default:
				assert.JSONEq(t, tt.expectedResponseBody.(string), w.Body.String())
			}
		})
	}
}

func TestHandler_uploadBatchPhotos(t *testing.T) {
	type mockBehavior func(s *mockservice.MockPhotoService, userUUID string, files []*multipart.FileHeader)

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
			mockBehavior: func(s *mockservice.MockPhotoService, userUUID string, files []*multipart.FileHeader) {
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
			mockBehavior:       func(s *mockservice.MockPhotoService, userUUID string, files []*multipart.FileHeader) {},
			expectedStatusCode: 400,
			expectedResponse: response.Error{
				Error: response.UnsupportedFileType,
			},
		},
		{
			name:     "Partical success",
			userUUID: "123e4567-e89b-12d3-a456-426614174000",
			multipartBody: func() (*bytes.Buffer, string) {
				return createMultipartBody(3, "tt%d.jpg", "fake image data")
			},
			mockBehavior: func(s *mockservice.MockPhotoService, userUUID string, files []*multipart.FileHeader) {
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
			mockBehavior: func(s *mockservice.MockPhotoService, userUUID string, files []*multipart.FileHeader) {
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

			mockPhotoService := mockservice.NewMockPhotoService(ctrl)
			tt.mockBehavior(mockPhotoService, tt.userUUID, nil)

			mockTokenService := mockservice.NewMockTokenService(ctrl)

			h := NewHandler(mockPhotoService, mockTokenService)

			r := gin.New()
			r.Use(middleware.UserIdentity(func(ctx context.Context, token string) (serviceUserModel.TokenPayload, error) {
				if token == "valid-token" {
					return serviceUserModel.TokenPayload{UserUUID: tt.userUUID}, nil
				}
				return serviceUserModel.TokenPayload{}, errors.New("invalid token")
			}))
			r.POST("/uploadBatch", h.uploadBatchPhotos)

			w := httptest.NewRecorder()
			body, contentType := tt.multipartBody()
			req := httptest.NewRequest("POST", "/uploadBatch", body)
			req.Header.Set("Content-Type", contentType)
			req.Header.Set("Authorization", "Bearer valid-token")

			r.ServeHTTP(w, req)

			content, err := json.Marshal(tt.expectedResponse)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedStatusCode, w.Code)

			switch tt.expectedResponse.(type) {
			case response.Error:
				var resp response.Error
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResponse.(response.Error).Error, resp.Error)
			default:
				assert.JSONEq(t, string(content), w.Body.String())
			}
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
