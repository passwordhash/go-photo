package photos

import (
	"context"
	"errors"
	"fmt"
	"go-photo/internal/config"
	"go-photo/internal/handler/middleware"
	"go-photo/internal/handler/response"
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
// @Success 200 {object} response.UploadPhotoResponse
// @Failure 400 {object} response.Error "Bad Request."
// @Failure 401 {object} response.Error "Unauthorized."
// @Failure 500 {object} response.Error "Unexpected error occurred."
// @Router /api/v1/photos/upload [post]
func (h *handler) uploadPhoto(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c, config.DefaultContextTimeout)
	defer cancel()

	uuid, ok := response.MustGetUUID(c, middleware.UserUUIDCtx)
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

	response.NewOk(c, response.UploadPhotoResponse{PhotoID: photoID})
}

// @Summary Upload batch photos
// @Description Upload multiple photos
// @Tags photos
// @Accept multipart/form-data
// @Produce json
// @Security JWTAuth
// @Param batch_photo_files formData file true "Batch photo files"
// @Success 200 {object} response.UploadBatchPhotosResponse
// @Failure 206 {object} response.UploadBatchPhotosResponse
// @Failure 400 {object} response.Error "Bad Request."
// @Failure 401 {object} response.Error "Unauthorized."
// @Failure 500 {object} response.Error "Unexpected error occurred."
// @Router /api/v1/photos/upload/batch [post]
func (h *handler) uploadBatchPhotos(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c, config.DefaultContextTimeout)
	defer cancel()

	respStatus := http.StatusOK

	uuid, ok := response.MustGetUUID(c, middleware.UserUUIDCtx)
	if !ok {
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

	body := response.UploadBatchPhotosResponse{
		TotalCount:   uploads.Total(),
		SuccessCount: uploads.SuccessCount(),
		UploadInfos:  append(make([]response.UploadInfo, 0), model.ToUploadsInfoFromService(uploads.Get())...),
	}

	c.JSON(respStatus, body)
}

// TODO: documetation
func (h *handler) getPhotoVersions(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c, config.DefaultContextTimeout)
	defer cancel()

	// TODO: сделать проверку на права доступа к фото
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		response.NewErr(c, http.StatusBadRequest, response.InvalidRequestParams, err, "Invalid photo id.")
		return
	}

	version, err := h.photoService.GetPhotoVersions(ctx, id)
	if err != nil {
		response.NewErr(c, http.StatusInternalServerError, response.InternalServerError, err, "Failed to get photo versions.")
		return
	}
	if response.HandleError(c, err) {
		return
	}

	c.JSON(200, gin.H{
		"versions": version,
	})
}
