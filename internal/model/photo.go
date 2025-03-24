package model

import "time"

type PhotoVersionType string

const (
	Original  PhotoVersionType = "original"
	Thumbnail PhotoVersionType = "thumbnail"
	Preview                    = "preview"
)

type Photo struct {
	ID         int
	UserUUID   string
	Filename   string
	Versions   []PhotoVersion
	UploadedAt time.Time
}

type PhotoVersion struct {
	ID          int
	PhotoID     int
	VersionType PhotoVersionType
	Filepath    string
	Size        int64
	Height      int
	Width       int
	SavedAt     time.Time
}
