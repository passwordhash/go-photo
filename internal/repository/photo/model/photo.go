package model

import (
	"database/sql"
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

type CreateOriginalPhotoParams struct {
	UserUUID string
	Filename string
	Filepath string
	Size     int64
	Height   int
	Width    int
	SavedAt  time.Time
}

func (p *CreateOriginalPhotoParams) IsValid() bool {
	return p.UserUUID != "" && p.Filename != "" && p.Filepath != "" && p.Size > 0 && p.Height > 0 && p.Width > 0 && !p.SavedAt.IsZero()
}
