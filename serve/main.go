package main

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/immofon/appoint"
	"github.com/immofon/appoint/account"
	"github.com/immofon/appoint/log"
	"github.com/immofon/appoint/rpc"
	"github.com/immofon/appoint/utils"
	bolt "go.etcd.io/bbolt"
)

func main() {
	start()
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func start() {
	log.TextMode()

	// open db
	db, err := bolt.Open("appoint.bolt", 0x600, &bolt.Options{
		Timeout: time.Second * 1,
	})
	if err != nil {
		log.E(err).Error()
		return
	}
	defer db.Close()

	// prepare buckets
	if err := account.Prepare(db); err != nil {
		log.E(err).Error("prepare account")
	}
	if err := appoint.Prepare(db); err != nil {
		log.E(err).Error("prepare appoint")
	}

	updateDate := func() {
		db.Update(func(tx *bolt.Tx) error {
			appoint.UpdateData(tx)
			return nil
		})
	}

	updateDate()

	//test something
	//Sometest(db)

	// rpc methods registe
	r := rpc.New(upgrader)
	r.RegisterFunc("login", func(ctx context.Context, req rpc.Request) rpc.Return {
		u_account := req.Get("account", "")
		u_password := req.Get("password", "")

		var user account.User
		var role appoint.Role
		err := db.View(func(tx *bolt.Tx) error {
			var err error
			user, err = account.Auth(tx, u_account, u_password)
			role = appoint.GetRole(tx, user.Id)
			return err
		})
		if err != nil {
			return ErrorRet(err, req)
		}

		return req.Ret("ok").
			Set("id", user.Id).
			Set("name", user.Name).
			Set("role", string(role)).
			SetUpdateContext(func(ctx context.Context) context.Context {
				return rpc.WithId(ctx, user.Id)
			})
	})

	// student
	r.RegisterFunc("appointment.status",
		func(ctx context.Context, req rpc.Request) rpc.Return {
			id := rpc.GetId(ctx)

			var role appoint.Role
			db.View(func(tx *bolt.Tx) error {
				role = appoint.GetRole(tx, id)
				return nil
			})

			if role != appoint.Role_Student {
				return ErrorRet(utils.ErrInternal, req)
			}

			status := appoint.GetData().UserStatus[id]
			if status == "" {
				return ErrorRet(utils.ErrInternal, req)
			}

			return req.Ret("ok").
				Set("status", string(status))
		})
	// listen

	http.Handle("/ws", r)
	log.L().Info("serve :8100")
	if err := http.ListenAndServe(":8100", nil); err != nil {
		log.E(err).Error()
	}
}

func ErrorRet(err error, req rpc.Request) rpc.Return {
	switch err {
	case nil:
		panic("err should not nil")
	case account.ErrUnvalid:
		return req.Ret(string(rpc.Unauthorized))
	case utils.ErrInternal:
		return req.Ret(string(rpc.Internal))
	case utils.ErrNotFound:
		return req.Ret(string(rpc.NotFound))

		// ADD TO HERE
	default:
		log.E(err).Error("not implement this type for ErrorRet")
		return req.Ret(string(rpc.Internal))
	}
}
