package photos

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
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
)

const (
	FormPhotoFile       = "photo_file"
	FormPhotoBatchFiles = "batch_photo_files"
)

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

	c.JSON(200, gin.H{
		"status": "ok",
		"id":     photoID,
	})
}

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

func (h *handler) getPhotoVersions(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c, config.DefaultContextTimeout)
	defer cancel()

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
