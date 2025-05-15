package tests

import (
	"go-photo/internal/handler/request"
	authReq "go-photo/internal/handler/request"
	authResp "go-photo/internal/handler/response/auth"
	"net/http"
	"net/url"
	"testing"

	"github.com/brianvoe/gofakeit/v6"

	"github.com/gavv/httpexpect/v2"
)

const passDefaultLen = 12

const (
	host = "localhost:8080"
)

func TestRegisterLogin_HappyPath(t *testing.T) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
	}

	e := httpexpect.Default(t, u.String())

	email := gofakeit.Email()
	pass := randomFakePassword()

	e.POST("/api/v1/auth/register").
		WithJSON(authReq.AuthRegister{
			Email:    email,
			Password: pass,
		}).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		ContainsKey("user_uuid").
		Decode(&authResp.Register{})

	e.POST("/api/v1/auth/login").
		WithJSON(request.AuthLogin{
			Email:    email,
			Password: pass,
		}).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Decode(&authResp.Login{})
}

func randomFakePassword() string {
	return gofakeit.Password(true, true, true, true, false, passDefaultLen)
}
