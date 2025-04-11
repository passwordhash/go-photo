package error

import (
	"errors"
	"fmt"
)

var (
	NotFoundError = errors.New("not found")

	BeginTxError  = errors.New("failed to begin transaction")
	CommitTxError = errors.New("failed to commit transaction")

	InsertError = errors.New("failed to insert")

	ConflictError = errors.New("conflict")

	InvalidParamsError = errors.New("invalid params")
	NilParamsError     = fmt.Errorf("%w: params are nil", InvalidParamsError)
)
