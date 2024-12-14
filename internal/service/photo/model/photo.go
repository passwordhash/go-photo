package model

import (
	"go-photo/internal/handler/response"
	"sync"
)

type UploadInfoList struct {
	mu       sync.RWMutex
	uploads  []UploadInfo
	total    int
	errCount int
}

type UploadInfo struct {
	PhotoID  int
	Filename string
	Error    error
	Size     int64
}

func (il *UploadInfoList) Add(info UploadInfo) {
	il.mu.Lock()
	defer il.mu.Unlock()
	il.uploads = append(il.uploads, info)
	il.total++
	if info.Error != nil {
		il.errCount++
	}
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

func (il *UploadInfoList) SuccessCount() int {
	il.mu.RLock()
	defer il.mu.RUnlock()
	return il.total - il.errCount
}

func (il *UploadInfoList) IsAllError() bool {
	il.mu.RLock()
	defer il.mu.RUnlock()
	return il.total == il.errCount
}

func (il *UploadInfoList) IsSomeError() bool {
	il.mu.RLock()
	defer il.mu.RUnlock()
	return il.errCount > 0 && il.errCount < il.total
}

func ToUploadsInfoFromService(uploads []UploadInfo) []response.UploadInfo {
	uploadsInfo := make([]response.UploadInfo, 0, len(uploads))
	for _, upload := range uploads {
		uploadsInfo = append(uploadsInfo, response.UploadInfo{
			PhotoID:  upload.PhotoID,
			Filename: upload.Filename,
			Error:    upload.Error,
		})
	}
	return uploadsInfo
}
