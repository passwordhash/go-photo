package model

import (
	"fmt"
	"time"
)

type PhotoVersionType string

const (
	Original  PhotoVersionType = "original"
	Thumbnail PhotoVersionType = "thumbnail"
	Preview                    = "preview"
)

func ParseVersionType(version string) (PhotoVersionType, error) {
	switch version {
	case "original":
		return Original, nil
	case "thumbnail":
		return Thumbnail, nil
	case "preview":
		return Preview, nil
	default:
		return "", fmt.Errorf("invalid version type: %s", version)
	}
}

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
