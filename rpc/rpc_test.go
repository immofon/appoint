package rpc

import (
	"context"
	"net/http"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/immofon/appoint/log"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func TestRpc(t *testing.T) {
	log.TextMode()
	rpc := New(upgrader)

	rpc.RegisterFunc("login", func(ctx context.Context, req Request) Return {
		log.L().Info("rpc.call " + "login")
		account := req.Get("account", "")

		//TODO auth

		return req.Ret("ok").
			Set("account", account).
			SetUpdateContext(func(ctx context.Context) context.Context {
				return WithId(ctx, account)
			})
	})

	rpc.RegisterFunc("logout", func(ctx context.Context, req Request) Return {
		return req.Ret("ok").
			SetUpdateContext(func(ctx context.Context) context.Context {
				return WithId(ctx, "")
			})
	})

	rpc.RegisterFunc("self", func(ctx context.Context, req Request) Return {
		id := GetId(ctx)
		if id == "" {
			return req.Ret("err").Set("err", "require-auth")
		}

		return req.Ret("ok").Set("id", id)
	})

	http.Handle("/ws", rpc)
	log.L().Info("serve :8100")
	t.Error(http.ListenAndServe(":8100", nil))
}
