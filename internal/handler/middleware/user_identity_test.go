package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	serviceUserModel "go-photo/internal/service/user/model"
)

func TestMiddleware_UserIdIdentity(t *testing.T) {
	gin.SetMode(gin.TestMode)

	dummyVerify := func(ctx context.Context, token string) (serviceUserModel.TokenPayload, error) {
		return serviceUserModel.TokenPayload{UserUUID: "shouldNotBeCalled"}, nil
	}

	verifyInvalid := func(ctx context.Context, token string) (serviceUserModel.TokenPayload, error) {
		return serviceUserModel.TokenPayload{}, errors.New("invalid token")
	}

	verifyValid := func(ctx context.Context, token string) (serviceUserModel.TokenPayload, error) {
		return serviceUserModel.TokenPayload{UserUUID: "12345"}, nil
	}

	tests := []struct {
		name                string
		authHeader          string
		verifyFn            VerifyTokenFunc
		expectedStatusCode  int
		expectedBodyContent string
	}{
		{
			name:                "Пустой заголовок Authorization",
			authHeader:          "",
			verifyFn:            dummyVerify,
			expectedStatusCode:  http.StatusUnauthorized,
			expectedBodyContent: "Auth header is empty.",
		},
		{
			name:                "Неверный формат заголовка (не Bearer)",
			authHeader:          "Basic token",
			verifyFn:            dummyVerify,
			expectedStatusCode:  http.StatusUnauthorized,
			expectedBodyContent: "Bearer token is invalid.",
		},
		{
			name:                "Пустой токен в заголовке Bearer",
			authHeader:          "Bearer ",
			verifyFn:            dummyVerify,
			expectedStatusCode:  http.StatusUnauthorized,
			expectedBodyContent: "Token is empty.",
		},
		{
			name:                "Некорректный токен",
			authHeader:          "Bearer invalidtoken",
			verifyFn:            verifyInvalid,
			expectedStatusCode:  http.StatusUnauthorized,
			expectedBodyContent: "Token is invalid or user cannot be found.",
		},
		{
			name:                "Корректный токен",
			authHeader:          "Bearer validtoken",
			verifyFn:            verifyValid,
			expectedStatusCode:  http.StatusOK,
			expectedBodyContent: "12345",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(UserIdentity(tt.verifyFn))
			router.GET("/", func(c *gin.Context) {
				if userUUID, exists := c.Get(UserUUIDCtx); exists {
					c.JSON(http.StatusOK, gin.H{"user_uuid": userUUID})
				} else {
					c.JSON(http.StatusOK, gin.H{"error": "no user identity"})
				}
			})

			req := httptest.NewRequest("GET", "/", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatusCode, rec.Code, "Неверный код ответа")
			body := rec.Body.String()

			if tt.expectedStatusCode != http.StatusOK {
				assert.Contains(t, body, tt.expectedBodyContent, "Тело ответа должно содержать сообщение об ошибке")
			} else {
				var resp map[string]interface{}
				err := json.Unmarshal(rec.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBodyContent, resp["user_uuid"])
			}
		})
	}
}
