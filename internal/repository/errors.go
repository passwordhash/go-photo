package repository

import (
	"errors"
	"fmt"
)

var (
	NotFoundError = errors.New("not found")
	PhotoNotFound = fmt.Errorf("photo %w", NotFoundError)
)
