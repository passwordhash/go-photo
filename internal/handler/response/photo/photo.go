package photo

type UploadPhotoResponse struct {
	PhotoID int `json:"photo_id"`
}

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

type GetPhotoVersionsResponse struct {
	Versions []PhotoVersion `json:"versions"`
}

type PhotoVersion struct {
	PhotoID     int    `json:"photo_id"`
	VersionType string `json:"version_type"`
	Filepath    string `json:"filepath"`
	Size        int64  `json:"size"`
	Height      int    `json:"height"`
	Width       int    `json:"width"`
	SavedAt     string `json:"saved_at"`
}

type PublishPhotoResponse struct {
	PublicToken string `json:"public_token"`
}
