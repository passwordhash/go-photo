package response

type UploadBatchPhotosResponse struct {
	TotalCount   int          `json:"total_count"`
	SuccessCount int          `json:"success_count"`
	UploadInfos  []UploadInfo `json:"upload_infos"`
}

type UploadInfo struct {
	PhotoID  int    `json:"photo_id,omitempty"`
	Filename string `json:"filename"`
	Error    error  `json:"error,omitempty"`
}
