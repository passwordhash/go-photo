package tests

import (
	"go-photo/internal/handler/request"
	authResp "go-photo/internal/handler/response/auth"
	photoResp "go-photo/internal/handler/response/photo"
	photoHandler "go-photo/internal/handler/v1/photos"
	"net/http"
	"net/url"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/gavv/httpexpect/v2"
)

const (
	host = "localhost:8080"

	passDefaultLen = 12
)

var u = url.URL{
	Scheme: "http",
	Host:   host,
}

func registerUser(e *httpexpect.Expect, email, password string) (userUUID string) {
	var resp authResp.Register
	e.POST("/api/v1/auth/register").
		WithJSON(request.AuthRegister{
			Email:    email,
			Password: password,
		}).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		ContainsKey("user_uuid").
		Decode(&resp)
	return resp.UserUUID
}

func loginUser(e *httpexpect.Expect, email, password string) (token string) {
	var resp authResp.Login
	e.POST("/api/v1/auth/login").
		WithJSON(request.AuthLogin{
			Email:    email,
			Password: password,
		}).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Decode(&resp)
	return resp.Token
}

func registerAndLogin(e *httpexpect.Expect) (email, password, token string) {
	email = gofakeit.Email()
	password = randomFakePassword()
	registerUser(e, email, password)
	token = loginUser(e, email, password)
	return
}

func uploadPhoto(e *httpexpect.Expect, token string) int {
	var resp photoResp.UploadPhotoResponse
	file := gofakeit.ImagePng(300, 200)
	e.POST("/api/v1/photos/").
		WithHeader("Authorization", "Bearer "+token).
		WithMultipart().
		WithFileBytes(photoHandler.FormPhotoFile, "img.png", file).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Decode(&resp)
	return resp.PhotoID
}

func randomFakePassword() string {
	return gofakeit.Password(true, true, true, true, false, passDefaultLen)
}
