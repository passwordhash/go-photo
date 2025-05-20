package middleware

import (
	"encoding/json"
	"go-photo/internal/model"
	serviceErr "go-photo/internal/service/error"
	mock_service "go-photo/internal/service/mock"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestMiddleware_PhotoOwner(t *testing.T) {
	gin.SetMode(gin.TestMode)

	type mockBehavior func(s mock_service.MockPhotoService, photoID int, userUUID string)

	tests := []struct {
		name                string
		photoID             int
		userUUID            string
		mockBehavior        mockBehavior
		expectedStatus      int
		expectedBodyContent string
	}{
		{
			name:     "Valid",
			photoID:  1,
			userUUID: "12345",
			mockBehavior: func(s mock_service.MockPhotoService, photoID int, userUUID string) {
				s.EXPECT().PhotoByID(gomock.Any(), photoID).
					Return(&model.Photo{UserUUID: userUUID}, nil)
			},
			expectedStatus:      http.StatusOK,
			expectedBodyContent: "12345",
		},
		{
			name:     "Photo not found",
			photoID:  0,
			userUUID: "12345",
			mockBehavior: func(s mock_service.MockPhotoService, photoID int, userUUID string) {
				s.EXPECT().PhotoByID(gomock.Any(), photoID).
					Return(nil, serviceErr.PhotoNotFoundError)
			},
			expectedStatus:      http.StatusNotFound,
			expectedBodyContent: "Photo not found",
		},
		{
			name:     "Photo not owner",
			photoID:  1,
			userUUID: "12345",
			mockBehavior: func(s mock_service.MockPhotoService, photoID int, userUUID string) {
				s.EXPECT().PhotoByID(gomock.Any(), photoID).
					Return(&model.Photo{UserUUID: "67890"}, nil)
			},
			expectedStatus:      http.StatusForbidden,
			expectedBodyContent: "You are not the owner of this photo",
		},
		{
			name:     "Empty user UUID",
			photoID:  1,
			userUUID: "",
			mockBehavior: func(s mock_service.MockPhotoService, photoID int, userUUID string) {
			},
			expectedStatus:      http.StatusUnauthorized,
			expectedBodyContent: "User not authorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			router := gin.New()

			s := mock_service.NewMockPhotoService(ctrl)

			tt.mockBehavior(*s, tt.photoID, tt.userUUID)

			router.Use(func(c *gin.Context) {
				c.Set(UserUUIDKey, tt.userUUID)
			})

			router.Use(PhotoOwner(s))

			router.GET("/photos/:id", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"user_uuid": tt.userUUID})
			})

			req := httptest.NewRequest(http.MethodGet, "/photos/"+strconv.Itoa(tt.photoID), nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code, "Неверный код ответа")
			body := rec.Body.String()

			if tt.expectedStatus != http.StatusOK {
				assert.Contains(t, body, tt.expectedBodyContent, "Неверное содержимое ответа")
			} else {
				var resp map[string]interface{}
				err := json.Unmarshal(rec.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBodyContent, resp["user_uuid"])
			}
		})
	}
}
