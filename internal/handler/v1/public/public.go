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

// @Summary Get public photo by token
// @Description Get public photo by token
// @Tags public
// @Accept json
// @Produce image/jpeg
// @Param publicToken path string true "Public token of photo"
// @Param version query string false "Version of photo" default(original)
// @Success 200 {file} string "image/jpeg"
// @Failure 400 {object} response.Error "Version type is not valid."
// @Failure 404 {object} response.Error "Photo not found."
// @Failure 500 {object} response.Error "Unexpected error occurred."
// @Router /p/{publicToken} [get]
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
