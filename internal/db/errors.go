package db

import "errors"

var (
	ErrInvalidUUID    = errors.New("invalid UUID")
	ErrEntityNotFound = errors.New("entity not found")
)
