package public

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go-photo/internal/handler/response"
	serviceErr "go-photo/internal/service/error"
	mock_service "go-photo/internal/service/mock"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler_getPublicPhoto(t *testing.T) {
	type mockBehavior func(s *mock_service.MockPhotoService, token string, versionQuery string)

	tests := []struct {
		name                 string
		token                string
		versionQuery         string
		mockBehavior         mockBehavior
		expectedStatusCode   int
		expectedContentType  string
		expectedResponseBody any
	}{
		{
			name:         "Valid",
			token:        "valid-token",
			versionQuery: "original",
			mockBehavior: func(s *mock_service.MockPhotoService, token string, versionQuery string) {
				s.EXPECT().
					GetPhotoFileByVersionAndToken(gomock.Any(), token, versionQuery).
					Return([]byte("test-data"), nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedContentType:  "image/jpeg",
			expectedResponseBody: []byte("test-data"),
		},
		{
			name:         "Valid default version",
			token:        "valid-token",
			versionQuery: "",
			mockBehavior: func(s *mock_service.MockPhotoService, token string, versionQuery string) {
				s.EXPECT().
					GetPhotoFileByVersionAndToken(gomock.Any(), token, versionQuery).
					Return([]byte("test-data"), nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedContentType:  "image/jpeg",
			expectedResponseBody: []byte("test-data"),
		},
		{
			name:         "Valid preview version",
			token:        "valid-token",
			versionQuery: "",
			mockBehavior: func(s *mock_service.MockPhotoService, token string, versionQuery string) {
				s.EXPECT().
					GetPhotoFileByVersionAndToken(gomock.Any(), token, versionQuery).
					Return([]byte("test-data"), nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedContentType:  "image/jpeg",
			expectedResponseBody: []byte("test-data"),
		},
		{
			name:         "Invalid Version",
			token:        "valid-token",
			versionQuery: "invalid-version",
			mockBehavior: func(s *mock_service.MockPhotoService, token string, versionQuery string) {
				s.EXPECT().
					GetPhotoFileByVersionAndToken(gomock.Any(), token, versionQuery).
					Return(nil, serviceErr.InvalidVersionTypeError)
			},
			expectedStatusCode:  http.StatusBadRequest,
			expectedContentType: "application/json",
			expectedResponseBody: response.Error{
				Error: response.InvalidReqestsQueryParams,
			},
		},
		{
			name:         "Public photo not found",
			token:        "valid-token",
			versionQuery: "original",
			mockBehavior: func(s *mock_service.MockPhotoService, token string, versionQuery string) {
				s.EXPECT().
					GetPhotoFileByVersionAndToken(gomock.Any(), token, versionQuery).
					Return(nil, serviceErr.PhotoNotFoundError)
			},
			expectedStatusCode:  http.StatusNotFound,
			expectedContentType: "application/json",
			expectedResponseBody: response.Error{
				Error: response.PhotoNotFound,
			},
		},
		{
			name:         "Internal Server Error",
			token:        "valid-token",
			versionQuery: "original",
			mockBehavior: func(s *mock_service.MockPhotoService, token string, versionQuery string) {
				s.EXPECT().
					GetPhotoFileByVersionAndToken(gomock.Any(), token, versionQuery).
					Return(nil, serviceErr.UnexpectedError)
			},
			expectedStatusCode:  http.StatusInternalServerError,
			expectedContentType: "application/json",
			expectedResponseBody: response.Error{
				Error: response.InternalServerError,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockPhotoService := mock_service.NewMockPhotoService(ctrl)
			tt.mockBehavior(mockPhotoService, tt.token, tt.versionQuery)

			h := NewHandler(mockPhotoService)

			r := gin.New()
			gin.DefaultWriter = io.Discard
			r.GET("/p/:publicToken", h.getPublicPhoto)

			w := httptest.NewRecorder()
			url := fmt.Sprintf("/p/%s?version=%s", tt.token, tt.versionQuery)
			req := httptest.NewRequest("GET", url, nil)

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatusCode, w.Code)
			assert.Contains(t, w.Header().Get("Content-Type"), tt.expectedContentType)
			switch tt.expectedResponseBody.(type) {
			case []byte:
				body := w.Body.String()
				assert.Contains(t, body, string(tt.expectedResponseBody.([]byte)))
			case response.Error:
				var resp response.Error
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResponseBody.(response.Error).Error, resp.Error)
			}
		})
	}
}
