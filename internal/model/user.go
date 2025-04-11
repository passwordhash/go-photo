package model

import "time"

type User struct {
	UUID         string
	Email        string
	PasswordHash string
	IsVerified   bool
	CreatedAt    time.Time
}
