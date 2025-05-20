package tests

import (
	"net/http"
	"testing"

	"github.com/gavv/httpexpect/v2"
)

func TestCheckAccessRights_NotOwner(t *testing.T) {
	e := httpexpect.Default(t, u.String())

	// Работа с несколькими пользователями
	_, _, token1 := registerAndLogin(e)
	_, _, token2 := registerAndLogin(e)

	// Первый пользователь загружает фото
	photoID := uploadPhoto(e, token1)

	// Второй пользователь пытается получить версии фото
	e.GET("/api/v1/photos/{photo_id}/versions", photoID).
		WithHeader("Authorization", "Bearer "+token2).
		Expect().
		Status(http.StatusForbidden).
		JSON().
		Object().
		ContainsKey("error").
		ContainsKey("message")

	// Второй пользователь пытается опубликовать фото
	e.POST("/api/v1/photos/{photo_id}/publish", photoID).
		WithHeader("Authorization", "Bearer "+token2).
		Expect().
		Status(http.StatusForbidden).
		JSON().
		Object().
		ContainsKey("error").
		ContainsKey("message")

	// Второй пользователь пытается удалить публикацию
	e.DELETE("/api/v1/photos/{photo_id}/unpublish", photoID).
		WithHeader("Authorization", "Bearer "+token2).
		Expect().
		Status(http.StatusForbidden).
		JSON().
		Object().
		ContainsKey("error").
		ContainsKey("message")
}
