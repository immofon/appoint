package utils

import "errors"

var (
	ErrInternal     = errors.New("internal")
	ErrNotFound     = errors.New("not_found")
	ErrRequireLogin = errors.New("require_login")
	ErrOp           = errors.New("op")
)
