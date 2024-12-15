package auth

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go-photo/internal/handler/response"
	serviceErr "go-photo/internal/service/error"
	mock_service "go-photo/internal/service/mocks"
	serviceUserModel "go-photo/internal/service/user/model"
	"net/http/httptest"
	"testing"
)

func TestHandler_login(t *testing.T) {
	type mockBehavior func(s *mock_service.MockUserService, email, password string)

	tests := []struct {
		name               string
		inputBody          string
		email              string
		password           string
		mockBehavior       mockBehavior
		expectedStatusCode int
		expectedResponse   any
	}{
		{
			name:      "Valid",
			inputBody: `{"email":"test@mail.ru","password":"password"}`,
			email:     "test@mail.ru",
			password:  "password",
			mockBehavior: func(s *mock_service.MockUserService, email, password string) {
				s.EXPECT().Login(gomock.Any(), email, password).Return("accessToken", nil).Times(1)
			},
			expectedStatusCode: 200,
			expectedResponse: response.Login{
				Token: "accessToken",
			},
		},
		{
			name: "Invalid Request Body",
			inputBody: `{"email":"
			email:     "
			password:  "`,
			mockBehavior: func(s *mock_service.MockUserService, email, password string) {
				s.EXPECT().Login(gomock.Any(), email, password).Times(0)
			},
			expectedStatusCode: 400,
			expectedResponse: response.Error{
				Error: response.InvalidRequestParams,
			},
		},
		{
			name:      "Empty Request Body",
			inputBody: `{}`,
			mockBehavior: func(s *mock_service.MockUserService, email, password string) {
				s.EXPECT().Login(gomock.Any(), email, password).Times(0)
			},
			expectedStatusCode: 400,
			expectedResponse: response.Error{
				Error: response.InvalidRequestParams,
			},
		},
		{
			name:      "Invalid Credentials",
			inputBody: `{"email":"test@mail.ru","password":"wrongpassword"}`,
			email:     "test@mail.ru",
			password:  "wrongpassword",
			mockBehavior: func(s *mock_service.MockUserService, email, password string) {
				s.EXPECT().Login(gomock.Any(), email, password).Return("", serviceErr.UserNotFoundError).Times(1)
			},
			expectedStatusCode: 401,
			expectedResponse: response.Error{
				Error: response.InvalidCredentials,
			},
		},
		{
			name:      "Internal Service Error",
			inputBody: `{"email":"test@mail.ru","password":"password"}`,
			email:     "test@mail.ru",
			password:  "password",
			mockBehavior: func(s *mock_service.MockUserService, email, password string) {
				s.EXPECT().Login(gomock.Any(), email, password).Return("", serviceErr.InternalError).Times(1)
			},
			expectedStatusCode: 500,
			expectedResponse: response.Error{
				Error: response.InternalServerError,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			mockPhotoService := mock_service.NewMockUserService(c)
			tt.mockBehavior(mockPhotoService, tt.email, tt.password)

			h := NewHandler(mockPhotoService)

			r := gin.New()
			r.POST("/login", h.login)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/login", bytes.NewBufferString(tt.inputBody))

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatusCode, w.Code)

			content, err := json.Marshal(tt.expectedResponse)
			assert.NoError(t, err)

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

func TestHandler_register(t *testing.T) {
	type mockBehavior func(s *mock_service.MockUserService, params serviceUserModel.RegisterParams)

	tests := []struct {
		name               string
		inputBody          string
		params             serviceUserModel.RegisterParams
		mockBehavior       mockBehavior
		expectedStatusCode int
		expectedResponse   any
	}{
		{
			name:      "Valid",
			inputBody: `{"email":"john@doe.com","password":"password"}`,
			params: serviceUserModel.RegisterParams{
				Email:    "john@doe.com",
				Password: "password",
			},
			mockBehavior: func(s *mock_service.MockUserService, params serviceUserModel.RegisterParams) {
				s.EXPECT().Register(gomock.Any(), params).Return(serviceUserModel.RegisterInfo{
					UserUUID: "user-id",
					Token:    "jwt-token",
				}, nil).Times(1)
			},
			expectedStatusCode: 200,
			expectedResponse: response.Register{
				UserUUID: "user-id",
				Token:    "jwt-token",
			},
		},
		{
			name:               "Invalid Request Body",
			inputBody:          `{"email": "email@mail.ru"`,
			mockBehavior:       func(s *mock_service.MockUserService, params serviceUserModel.RegisterParams) {},
			expectedStatusCode: 400,
			expectedResponse: response.Error{
				Error: response.InvalidRequestParams,
			},
		},
		{
			name:               "Empty Request Body",
			inputBody:          `{}`,
			mockBehavior:       func(s *mock_service.MockUserService, params serviceUserModel.RegisterParams) {},
			expectedStatusCode: 400,
			expectedResponse: response.Error{
				Error: response.InvalidRequestParams,
			},
		},
		{
			name:      "User Already Exists",
			inputBody: `{"email":"john@doe.ru","password":"password"}`,
			params: serviceUserModel.RegisterParams{
				Email:    "john@doe.ru",
				Password: "password",
			},
			mockBehavior: func(s *mock_service.MockUserService, params serviceUserModel.RegisterParams) {
				s.EXPECT().Register(gomock.Any(), params).Return(serviceUserModel.RegisterInfo{}, serviceErr.UserAlreadyExistsError).Times(1)
			},
			expectedStatusCode: 409,
			expectedResponse: response.Error{
				Error: response.UserAlreadyExists,
			},
		},
		{
			name:      "Internal Service Error",
			inputBody: `{"email":"john@doe.ru", "password":"password"}`,
			params: serviceUserModel.RegisterParams{
				Email:    "john@doe.ru",
				Password: "password",
			},
			mockBehavior: func(s *mock_service.MockUserService, params serviceUserModel.RegisterParams) {
				s.EXPECT().Register(gomock.Any(), params).Return(serviceUserModel.RegisterInfo{}, serviceErr.InternalError).Times(1)
			},
			expectedStatusCode: 500,
			expectedResponse: response.Error{
				Error: response.InternalServerError,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			mockPhotoService := mock_service.NewMockUserService(c)
			tt.mockBehavior(mockPhotoService, tt.params)

			h := NewHandler(mockPhotoService)

			r := gin.New()
			r.POST("/register", h.register)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/register", bytes.NewBufferString(tt.inputBody))

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatusCode, w.Code)

			content, err := json.Marshal(tt.expectedResponse)
			assert.NoError(t, err)

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
