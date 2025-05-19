package tests

import (
	"fmt"
	"go-photo/internal/handler/request"
	authResp "go-photo/internal/handler/response/auth"
	photoResp "go-photo/internal/handler/response/photo"
	"net/http"
	"strconv"
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

func TestAuthUploadImageAndPublicate_HappyPath(t *testing.T) {
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

	var uploadResp photoResp.UploadPhotoResponse
	file := gofakeit.ImagePng(300, 200)
	e.POST("/api/v1/photos/").
		WithHeader("Authorization", "Bearer "+loginResp.Token).
		WithMultipart().
		WithFileBytes(uploadFormFieldName, "img.png", file).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Decode(&uploadResp)

	var pubResp photoResp.PublishPhotoResponse
	pubURL := fmt.Sprintf("/api/v1/photos/%s/publicate", strconv.Itoa(uploadResp.PhotoID))
	e.POST(pubURL).
		WithHeader("Authorization", "Bearer "+loginResp.Token).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Decode(&pubResp)

	fmt.Println("publicate resp: ", pubResp.PublicToken)
}

func TestAuthUploadImageAndPublicate_ByNotOwner(t *testing.T) {
	e := httpexpect.Default(t, u.String())

	email1 := gofakeit.Email()
	pass1 := randomFakePassword()

	email2 := gofakeit.Email()
	pass2 := randomFakePassword()

	// Auth first user
	e.POST("/api/v1/auth/register").
		WithJSON(request.AuthRegister{
			Email:    email1,
			Password: pass1,
		}).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		ContainsKey("user_uuid").
		Decode(&authResp.Register{})

	var loginResp1 authResp.Login
	e.POST("/api/v1/auth/login").
		WithJSON(request.AuthLogin{
			Email:    email1,
			Password: pass1,
		}).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Decode(&loginResp1)

	// Auth second user
	e.POST("/api/v1/auth/register").
		WithJSON(request.AuthRegister{
			Email:    email2,
			Password: pass2,
		}).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		ContainsKey("user_uuid").
		Decode(&authResp.Register{})

	var loginResp2 authResp.Login
	e.POST("/api/v1/auth/login").
		WithJSON(request.AuthLogin{
			Email:    email2,
			Password: pass2,
		}).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Decode(&loginResp2)

	// First user upload image
	var uploadResp photoResp.UploadPhotoResponse
	file := gofakeit.ImagePng(300, 200)
	e.POST("/api/v1/photos/").
		WithHeader("Authorization", "Bearer "+loginResp1.Token).
		WithMultipart().
		WithFileBytes(uploadFormFieldName, "img.png", file).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Decode(&uploadResp)

	// Second user try to publicate image
	pubURL := fmt.Sprintf("/api/v1/photos/%s/publicate", strconv.Itoa(uploadResp.PhotoID))
	e.POST(pubURL).
		WithHeader("Authorization", "Bearer "+loginResp2.Token).
		Expect().
		Status(http.StatusForbidden)
}

func TestAuthUploadImage_Unauthorized(t *testing.T) {
	e := httpexpect.Default(t, u.String())

	file := gofakeit.ImagePng(300, 200)

	e.POST("/api/v1/photos/").
		WithMultipart().
		WithFileBytes(uploadFormFieldName, "img.png", file).
		Expect().
		Status(http.StatusUnauthorized)
}
