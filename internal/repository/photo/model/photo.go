package model

import (
	"database/sql"
	"go-photo/internal/model"
	"time"
)

type Photo struct {
	ID         int           `db:"id"`
	UserUUID   string        `db:"user_uuid"`
	Filename   string        `db:"filename"`
	UploadedAt *sql.NullTime `db:"uploaded_at"`
}

type PhotoVersion struct {
	ID          int            `db:"id"`
	PhotoID     int            `db:"photo_id"`
	VersionType sql.NullString `db:"version_type"`
	Filepath    string         `db:"filepath"`
	Size        int64          `db:"size"`
	Height      int            `db:"height"`
	Width       int            `db:"width"`
	SavedAt     *sql.NullTime  `db:"saved_at"`
}

type PublishedPhotoInfo struct {
	ID          int          `db:"id"`
	PublishedAt sql.NullTime `db:"published_at"`
	PublicToken string       `db:"public_token"`
}

type PhotoWithPhotoVersion struct {
	PhotoID     int            `db:"photo_id"`
	UserUUID    string         `db:"user_uuid"`
	Filename    string         `db:"filename"`
	UploadedAt  *sql.NullTime  `db:"uploaded_at"`
	VersionType sql.NullString `db:"version_type"`
	VersionID   int            `db:"version_id"`
	Filepath    string         `db:"filepath"`
	Size        int64          `db:"size"`
	Height      int            `db:"height"`
	Width       int            `db:"width"`
	SavedAt     *sql.NullTime  `db:"saved_at"`
}

type CreateOriginalPhotoParams struct {
	UserUUID string
	Filename string
	Filepath string
	Size     int64
	Height   int
	Width    int
	SavedAt  time.Time
}

type FilterParams struct {
	VersionType model.PhotoVersionType `db:"version_type"`
}

// TODO: сделать обход по полям с помощью reflect
func (f *FilterParams) MapToArgs(params map[string]interface{}) string {
	addQuery := ""

	if f.VersionType != "" {
		addQuery += " AND version_type = :version_type"
		params["version_type"] = f.VersionType
	}

	return addQuery
}

func (p *CreateOriginalPhotoParams) IsValid() bool {
	return p.UserUUID != "" && p.Filename != "" && p.Filepath != "" && p.Size > 0 && p.Height > 0 && p.Width > 0 && !p.SavedAt.IsZero()
}
