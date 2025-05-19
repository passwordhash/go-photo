package tests

import (
	"fmt"
	"go-photo/internal/handler/request"
	authResp "go-photo/internal/handler/response/auth"
	"net/http"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/gavv/httpexpect/v2"
)

const (
	uploadFormFieldName = "photo_file"
)

func TestAuthUploadImage_HappyPath(t *testing.T) {
	e := httpexpect.Default(t, u.String())

	email := gofakeit.Email()
	pass := randomFakePassword()

	e.POST("/api/v1/auth/register").
		WithJSON(request.AuthRegister{
			Email:    email,
			Password: pass,
		}).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		ContainsKey("user_uuid").
		Decode(&authResp.Register{})

	var loginResp authResp.Login
	e.POST("/api/v1/auth/login").
		WithJSON(request.AuthLogin{
			Email:    email,
			Password: pass,
		}).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Decode(&loginResp)

	fmt.Println("token: ", loginResp.Token)

	file := gofakeit.ImagePng(300, 200)
	e.POST("/api/v1/photos/").
		WithHeader("Authorization", "Bearer "+loginResp.Token).
		WithMultipart().
		WithFileBytes(uploadFormFieldName, "img.png", file).
		Expect().
		Status(http.StatusOK)
}
