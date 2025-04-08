package utils

import "github.com/google/uuid"

func (u *Utils) UUID() string {
    return uuid.NewString()
}
