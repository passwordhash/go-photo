package model

import (
	"go-photo/internal/handler/response/photo"
	"sync"
	"time"
)

type UploadInfoList struct {
	mu      sync.RWMutex
	uploads []UploadInfo
	total   int
}

type UploadInfo struct {
	PhotoID      int
	Filename     string
	UUIDFilename string
	Error        error
	Size         int64
	Height       int
	Width        int
	SavedAt      time.Time
}

func NewUploadInfoList(infos []UploadInfo) *UploadInfoList {
	return &UploadInfoList{
		uploads: infos,
		total:   len(infos),
	}
}

func (il *UploadInfoList) Add(info UploadInfo) {
	il.mu.Lock()
	defer il.mu.Unlock()
	il.uploads = append(il.uploads, info)
	il.total++
}

func (il *UploadInfoList) Get() []UploadInfo {
	il.mu.RLock()
	defer il.mu.RUnlock()
	return il.uploads
}

func (il *UploadInfoList) Total() int {
	il.mu.RLock()
	defer il.mu.RUnlock()
	return il.total
}

func (il *UploadInfoList) ErrorCount() int {
	il.mu.RLock()
	defer il.mu.RUnlock()
	cnt := 0
	for _, upload := range il.uploads {
		if upload.Error != nil {
			cnt++
		}
	}
	return cnt
}

func (il *UploadInfoList) SuccessCount() int {
	il.mu.RLock()
	defer il.mu.RUnlock()
	return il.total - il.ErrorCount()
}

func (il *UploadInfoList) IsAllError() bool {
	il.mu.RLock()
	defer il.mu.RUnlock()
	return il.total == il.ErrorCount()
}

func (il *UploadInfoList) IsSomeError() bool {
	il.mu.RLock()
	defer il.mu.RUnlock()
	errCount := il.ErrorCount()
	return errCount > 0 && errCount < il.total
}

func ToUploadsInfoFromService(uploads []UploadInfo) []photo.UploadInfo {
	uploadsInfo := make([]photo.UploadInfo, 0, len(uploads))
	for _, upload := range uploads {
		uploadsInfo = append(uploadsInfo, photo.UploadInfo{
			PhotoID:  upload.PhotoID,
			Filename: upload.Filename,
			Error:    upload.Error,
		})
	}
	return uploadsInfo
}
