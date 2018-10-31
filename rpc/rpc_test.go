package rpc

import (
	"context"
	"log"
	"net/http"
	"testing"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func TestRpc(t *testing.T) {
	rpc := New(upgrader)

	rpc.RegisterFunc("echo", func(ctx context.Context, req Request) Return {
		return req.Ret("ok").Set("msg", req.Get("msg", ""))
	})

	http.Handle("/ws", rpc)
	log.Println("serve :8100")
	t.Error(http.ListenAndServe(":8100", nil))
}
