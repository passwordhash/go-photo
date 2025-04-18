package photos

import (
	"context"
	"errors"
	"fmt"
	"go-photo/internal/config"
	"go-photo/internal/handler/middleware"
	"go-photo/internal/handler/response"
	"go-photo/internal/handler/response/auth"
	photoResp "go-photo/internal/handler/response/photo"
	serviceErr "go-photo/internal/service/error"
	"go-photo/internal/service/photo/model"
	"go-photo/internal/utils"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	FormPhotoFile       = "photo_file"
	FormPhotoBatchFiles = "batch_photo_files"
)

// @Summary Upload photo
// @Description Upload single photo
// @Tags photos
// @Accept multipart/form-data
// @Produce json
// @Security JWTAuth
// @Param photo_file formData file true "Photo file"
// @Success 200 {object} photo.UploadPhotoResponse
// @Failure 400 {object} response.Error "Bad Request."
// @Failure 401 {object} response.Error "Unauthorized."
// @Failure 500 {object} response.Error "Unexpected error occurred."
// @Router /api/v1/photos/ [post]
func (h *handler) uploadPhoto(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c, config.DefaultContextTimeout)
	defer cancel()

	uuid, ok := auth.MustGetUUID(c, middleware.UserUUIDCtx)
	if !ok {
		return
	}

	fileHeader, err := c.FormFile(FormPhotoFile)
	if err != nil {
		response.NewErr(c, http.StatusBadRequest, response.ParamsMissing, err, fmt.Sprintf("No %s in form.", FormPhotoFile))
		return
	}

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if !utils.IsPhoto(ext) {
		response.NewErr(c, http.StatusBadRequest, response.UnsupportedFileType, nil, "Unsupported file type: "+ext)
		return
	}

	photoID, err := h.photoService.UploadPhoto(ctx, uuid, fileHeader)
	if response.HandleError(c, err) {
		return
	}

	response.NewOk(c, photoResp.UploadPhotoResponse{PhotoID: photoID})
}

// @Summary Upload batch photos
// @Description Upload multiple photos
// @Tags photos
// @Accept multipart/form-data
// @Produce json
// @Security JWTAuth
// @Param batch_photo_files formData file true "Batch photo files"
// @Success 200 {object} photo.UploadBatchPhotosResponse
// @Failure 206 {object} photo.UploadBatchPhotosResponse
// @Failure 400 {object} response.Error "Bad Request."
// @Failure 401 {object} response.Error "Unauthorized."
// @Failure 500 {object} response.Error "Unexpected error occurred."
// @Router /api/v1/photos/batch [post]
func (h *handler) uploadBatchPhotos(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c, config.DefaultContextTimeout)
	defer cancel()

	respStatus := http.StatusOK

	uuid, ok := auth.MustGetUUID(c, middleware.UserUUIDCtx)
	if !ok {
		response.NewErr(c, http.StatusUnauthorized, response.Unauthorized, nil, "Try logging in again.")
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		response.NewErr(c, http.StatusBadRequest, response.InvalidRequestParams, err, "No form data.")
		return
	}

	files := form.File[FormPhotoBatchFiles]
	if len(files) == 0 {
		response.NewErr(c, http.StatusBadRequest, response.ParamsMissing, nil, fmt.Sprintf("No %s in form.", FormPhotoBatchFiles))
		return
	}

	if ok, notPhoto := utils.IsAllPhotos(files); !ok {
		response.NewErr(c, http.StatusBadRequest, response.UnsupportedFileType, err, fmt.Sprintf("File %s is not a photo.", notPhoto))
		return
	}

	uploads, err := h.photoService.UploadBatchPhotos(ctx, uuid, files)
	if errors.Is(err, serviceErr.AllFailedError) {
		respStatus = http.StatusBadRequest
	} else if errors.Is(err, serviceErr.ParticalSuccessError) {
		respStatus = http.StatusPartialContent
	} else if response.HandleError(c, err) {
		return
	}

	body := photoResp.UploadBatchPhotosResponse{
		TotalCount:   uploads.Total(),
		SuccessCount: uploads.SuccessCount(),
		UploadInfos:  append(make([]photoResp.UploadInfo, 0), model.ToUploadsInfoFromService(uploads.Get())...),
	}

	c.JSON(respStatus, body)
}

// @Summary Get photo versions
// @Description Get all versions of a photo
// @Tags photos
// @Produce json
// @Security JWTAuth
// @Param id path int true "Photo ID"
// @Success 200 {object} photo.GetPhotoVersionsResponse
// @Failure 400 {object} response.Error "Bad Request."
// @Failure 401 {object} response.Error "Unauthorized."
// @Failure 403 {object} response.Error "Access denied."
// @Failure 404 {object} response.Error "Photo not found."
// @Failure 500 {object} response.Error "Unexpected error occurred."
// @Router /api/v1/photos/{id}/versions [get]
func (h *handler) getPhotoVersions(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c, config.DefaultContextTimeout)
	defer cancel()

	uuid, ok := auth.MustGetUUID(c, middleware.UserUUIDCtx)
	if !ok {
		response.NewErr(c, http.StatusUnauthorized, response.Unauthorized, nil, "Try logging in again.")
		return
	}

	idParam := c.Param("id")
	photoID, err := strconv.Atoi(idParam)
	if err != nil {
		response.NewErr(c, http.StatusBadRequest, response.InvalidRequestParams, err, "Invalid photo id.")
		return
	}

	versions, err := h.photoService.GetPhotoVersions(ctx, uuid, photoID)
	if errors.Is(err, serviceErr.PhotoNotFoundError) {
		response.NewErr(c, http.StatusNotFound, response.PhotoNotFound, err, "Photo not found.")
		return
	}
	if errors.Is(err, serviceErr.AccessDeniedError) {
		response.NewErr(c, http.StatusForbidden, response.Forbidden, err, "You do not have access to this photo.")
		return
	}
	if response.HandleError(c, err) {
		return
	}

	response.NewOk(c, photoResp.GetPhotoVersionsResponse{
		Versions: photoResp.ToPhotoVersionsFromModel(versions),
	})
}

// @Summary Publish photo
// @Description Make a photo public
// @Tags photos
// @Produce json
// @Security JWTAuth
// @Param id path int true "Photo ID"
// @Success 200 {object} photo.PublishPhotoResponse
// @Failure 400 {object} response.Error "Bad Request."
// @Failure 401 {object} response.Error "Unauthorized."
// @Failure 403 {object} response.Error "Access denied."
// @Failure 404 {object} response.Error "Photo not found."
// @Failure 409 {object} response.Error "Photo already published."
// @Failure 500 {object} response.Error "Unexpected error occurred."
// @Router /api/v1/photos/{id}/publicate [post]
func (h *handler) publishPhoto(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c, config.DefaultContextTimeout)
	defer cancel()

	userUUID, ok := auth.MustGetUUID(c, middleware.UserUUIDCtx)
	if !ok {
		response.NewErr(c, http.StatusUnauthorized, response.Unauthorized, nil, "Try logging in again.")
		return
	}

	idParam := c.Param("id")
	photoID, err := strconv.Atoi(idParam)
	if err != nil {
		response.NewErr(c, http.StatusBadRequest, response.InvalidRequestParams, err, "Invalid photo id.")
		return
	}

	publicToken, err := h.photoService.PublishPhoto(ctx, userUUID, photoID)
	if errors.Is(err, serviceErr.PhotoNotFoundError) {
		response.NewErr(c, http.StatusNotFound, response.PhotoNotFound, err, "Photo not found.")
		return
	}
	if errors.Is(err, serviceErr.AlreadyExists) {
		response.New(c, http.StatusNoContent, "Photo already published.")
		return
	}
	if response.HandleError(c, err) {
		return
	}

	response.NewOk(c, photoResp.PublishPhotoResponse{
		PublicToken: publicToken,
	})
}

// @Summary Unpublicate photo
// @Description Unpublicate a photo by ID
// @Tags photos
// @Produce json
// @Security JWTAuth
// @Param id path int true "Photo ID"
// @Success 200 {object} nil
// @Failure 400 {object} response.Error "Bad Request."
// @Failure 401 {object} response.Error "Unauthorized."
// @Failure 403 {object} response.Error "Access denied."
// @Failure 404 {object} response.Error "Photo not found or already unpublished."
// @Failure 500 {object} response.Error "Unexpected error occurred."
// @Router /api/v1/photos/{id}/unpublicate [delete]
func (h *handler) unpublicatePhoto(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c, config.DefaultContextTimeout)
	defer cancel()

	userUUID, ok := auth.MustGetUUID(c, middleware.UserUUIDCtx)
	if !ok {
		response.NewErr(c, http.StatusUnauthorized, response.Unauthorized, nil, "Try logging in again.")
		return
	}

	idParam := c.Param("id")
	photoID, err := strconv.Atoi(idParam)
	if err != nil {
		response.NewErr(c, http.StatusBadRequest, response.InvalidRequestParams, err, "Invalid photo id.")
		return
	}

	err = h.photoService.UnpublishPhoto(ctx, userUUID, photoID)
	if errors.Is(err, serviceErr.PhotoNotFoundError) {
		response.NewErr(c, http.StatusNotFound, response.NotFound, err, "Photo not found or already unpublished.")
		return
	}
	if response.HandleError(c, err) {
		return
	}

	response.NewOk(c, nil)
}
