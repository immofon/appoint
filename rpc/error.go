package rpc

type ErrorRetType string

const (
	Internal     ErrorRetType = "internal"
	NotFound     ErrNotFound  = "not_found"
	Unauthorized ErrorRetType = "unauthorized"
)
