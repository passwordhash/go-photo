package error

import (
	"errors"
	"fmt"
)

var (
	NotFoundError = errors.New("not found")
	PhotoNotFound = fmt.Errorf("photo %w", NotFoundError)

	BeginTxError  = errors.New("failed to begin transaction")
	CommitTxError = errors.New("failed to commit transaction")

	InsertError = errors.New("failed to insert")

	InvalidParamsError = errors.New("invalid params")
	NilParamsError     = fmt.Errorf("%w: params are nil", InvalidParamsError)
)
