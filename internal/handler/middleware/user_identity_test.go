package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-photo/internal/lib/jwt"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestMiddleware_UserIdIdentity(t *testing.T) {
	gin.SetMode(gin.TestMode)

	dummyVerify := func(ctx context.Context, token, secret string) (*jwt.Claims, error) {
		return &jwt.Claims{UserUUID: "shouldNotBeCalled"}, nil
	}

	verifyInvalid := func(ctx context.Context, token, secret string) (*jwt.Claims, error) {
		return &jwt.Claims{}, errors.New("invalid token")
	}

	verifyValid := func(ctx context.Context, token, secret string) (*jwt.Claims, error) {
		return &jwt.Claims{UserUUID: "12345"}, nil
	}

	tests := []struct {
		name                string
		authHeader          string
		verifyFn            jwt.VerifyTokenFunc
		expectedStatusCode  int
		expectedBodyContent string
	}{
		{
			name:                "Valid",
			authHeader:          "Bearer validtoken",
			verifyFn:            verifyValid,
			expectedStatusCode:  http.StatusOK,
			expectedBodyContent: "12345",
		},
		{
			name:                "Empty auth header",
			authHeader:          "",
			verifyFn:            dummyVerify,
			expectedStatusCode:  http.StatusUnauthorized,
			expectedBodyContent: "Auth header is empty.",
		},
		{
			name:                "Invalid header format",
			authHeader:          "Basic token",
			verifyFn:            dummyVerify,
			expectedStatusCode:  http.StatusUnauthorized,
			expectedBodyContent: "Bearer token is invalid.",
		},
		{
			name:                "Empty token",
			authHeader:          "Bearer ",
			verifyFn:            dummyVerify,
			expectedStatusCode:  http.StatusUnauthorized,
			expectedBodyContent: "Token is empty.",
		},
		{
			name:                "Invalid token",
			authHeader:          "Bearer invalidtoken",
			verifyFn:            verifyInvalid,
			expectedStatusCode:  http.StatusUnauthorized,
			expectedBodyContent: "Token is invalid or user cannot be found.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(UserIdentity(tt.verifyFn, "some-secret"))

			router.GET("/", func(c *gin.Context) {
				if userUUID, exists := c.Get(UserUUIDKey); exists {
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
