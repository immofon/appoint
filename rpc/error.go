package rpc

type ErrorRetType string

const (
	ErrInternal     ErrorRetType = "internal"
	ErrNotFound     ErrNotFound  = "not_found"
	ErrUnauthorized ErrorRetType = "unauthorized"
)
