package error

import "errors"

var (
	ServiceError = errors.New("service error")
	DbError      = errors.New("db error")
)
