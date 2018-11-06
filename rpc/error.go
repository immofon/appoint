package rpc

type ErrorRetType string

const (
	Internal     ErrorRetType = "internal"
	NotFound     ErrorRetType = "not_found"
	Unauthorized ErrorRetType = "unauthorized"
)
