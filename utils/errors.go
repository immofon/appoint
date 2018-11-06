package utils

import "errors"

var (
	ErrInternal = errors.New("internal")
	ErrNotFound = errors.New("not_found")
)
