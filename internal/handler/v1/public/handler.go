package public

import (
	"errors"
	"github.com/gin-gonic/gin"
	"go-photo/internal/handler/response"
	serviceErr "go-photo/internal/service/error"
	"net/http"
)

const (
	publicPhotoParam = "publicToken"
)

const (
	versionQueryParam        = "version"
	versionQueryParamDefault = "original"
)

func (h *handler) getPublicPhoto(c *gin.Context) {
	tokenParam := c.Param(publicPhotoParam)

	versionQuery := c.DefaultQuery(versionQueryParam, versionQueryParamDefault)

	imgData, err := h.photoService.GetPhotoFileByVersionAndToken(c, tokenParam, versionQuery)
	if errors.Is(err, serviceErr.PhotoNotFoundError) {
		response.NewErr(c, http.StatusNotFound, response.PhotoNotFound, err, "Photo not found by token and version")
		return
	}
	if errors.Is(err, serviceErr.InvalidVersionTypeError) {
		response.NewErr(c, http.StatusBadRequest, response.InvalidReqestsQueryParams, err, "Invalid version type")
		return
	}
	if response.HandleError(c, err) {
		return
	}

	c.Data(200, "image/jpeg", imgData)
}
