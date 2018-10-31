package rpc

import (
	"context"
)

type _context_type_id int // type:string

func WithId(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, _context_type_id(0), id)
}

func GetId(ctx context.Context) string {
	id, ok := ctx.Value(_context_type_id(0)).(string)
	if ok {
		return id
	}
	return ""
}

type _context_type_authed int // type: bool

func WithAuthed(ctx context.Context, authed bool) context.Context {
	return context.WithValue(ctx, _context_type_authed(0), authed)
}

func IsAuthed(ctx context.Context) bool {
	authed, ok := ctx.Value(_context_type_authed(0)).(bool)
	if ok {
		return authed
	}
	return false
}
