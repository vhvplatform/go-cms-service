package repository

import "errors"

// Common repository errors
var (
	ErrNotFound = errors.New("not found")
	ErrDuplicate = errors.New("duplicate entry")
)
