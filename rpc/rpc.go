package rpc

import (
	"context"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/immofon/appoint/log"
	"github.com/immofon/appoint/utils"
)

type Request struct {
	Id   string            `json:"id"`
	Func string            `json:"func"`
	Argv map[string]string `json:"argv"`
}

func (req Request) Get(key string, defaultv string) string {
	v, ok := req.Argv[key]
	if !ok {
		return defaultv
	}
	return v
}

func (req Request) Ret(status string) Return {
	return Return{
		Id:      req.Id,
		Status:  status,
		Details: make(map[string]string),
	}

}

type Return struct {
	Id            string            `json:"id"`
	Status        string            `json:"status"`
	Details       map[string]string `json:"details"`
	UpdateContext func(context.Context) context.Context
}

func (ret Return) Set(key, value string) Return {
	ret.Details[key] = value
	return ret
}
func (ret Return) Get(key, defaultv string) string {
	v, ok := ret.Details[key]
	if !ok {
		return defaultv
	}
	return v
}
func (ret Return) Has(key string) bool {
	_, ok := ret.Details[key]
	return ok
}

const InnerPrefix = "__$"

func (ret Return) SetInner(key, value string) Return {
	ret.Details[key] = InnerPrefix + value
	return ret
}
func (ret Return) GetInnert(key, defaultv string) string {
	v, ok := ret.Details[InnerPrefix+key]
	if !ok {
		return defaultv
	}
	return v
}

func (ret *Return) RemoveInnerDetails() {
	for k, _ := range ret.Details {
		if strings.HasPrefix(k, InnerPrefix) {
			delete(ret.Details, k)
		}
	}
}

type Handler interface {
	RPCHandle(ctx context.Context, req Request) Return
}

type HandleFunc func(ctx context.Context, req Request) Return

func (hdf HandleFunc) RPCHandle(ctx context.Context, req Request) Return {
	return hdf(ctx, req)
}

type RPC struct {
	sync.Mutex
	upgrader websocket.Upgrader
	funcs    map[string]Handler // key: function name
}

func New(upgrader websocket.Upgrader) *RPC {
	return &RPC{
		upgrader: upgrader,
		funcs:    make(map[string]Handler, 0),
	}
}

func (rpc *RPC) Register(name string, handler Handler) {
	rpc.Lock()
	defer rpc.Unlock()
	if handler != nil {
		rpc.funcs[name] = handler
	} else {
		delete(rpc.funcs, name)
	}
}
func (rpc *RPC) RegisterFunc(name string, fn HandleFunc) {
	rpc.Register(name, fn)
}

func (rpc *RPC) Call(ctx context.Context, req Request) Return {
	rpc.Lock()
	defer rpc.Unlock()

	handler, ok := rpc.funcs[req.Func]
	if !ok {
		return req.Ret(utils.ErrNotFound.Error()).Set("__debug", "rpc method ["+req.Func+"] was not deined")
	}

	return handler.RPCHandle(ctx, req)
}

func (rpc *RPC) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := rpc.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.E(err).Error()
		return
	}
	defer conn.Close()

	ctx := context.Background()
	for {
		var req Request
		if err := conn.ReadJSON(&req); err != nil {
			log.E(err).Error()
			return
		}

		ret := rpc.Call(ctx, req)
		if ret.UpdateContext != nil {
			ctx = ret.UpdateContext(ctx)
		}

		ret.RemoveInnerDetails()

		if err := conn.WriteJSON(ret); err != nil {
			log.E(err).Error()
			return
		}
	}
}
