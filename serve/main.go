package main

import (
	"context"
	"fmt"
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
	db, err := bolt.Open("appoint.bolt", 0600, &bolt.Options{
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
	register_require_auth_student := func(method string, fn func(ctx context.Context, req rpc.Request) rpc.Return) {
		r.RegisterFunc(method,
			func(ctx context.Context, req rpc.Request) rpc.Return {
				id := rpc.GetId(ctx)

				if id == "" {
					return ErrorRet(account.ErrUnvalid, req)
				}

				var role appoint.Role
				db.View(func(tx *bolt.Tx) error {
					role = appoint.GetRole(tx, id)
					return nil
				})

				if role != appoint.Role_Student {
					return ErrorRet(utils.ErrInternal, req)
				}

				return fn(ctx, req)
			})
	}

	register_require_auth_student("appointment.student.status", func(ctx context.Context, req rpc.Request) rpc.Return {
		id := rpc.GetId(ctx)
		status := appoint.GetData().UserStatus[id]
		if status == "" {
			return ErrorRet(utils.ErrInternal, req)
		}

		return req.Ret("ok").
			Set("status", string(status))
	})

	register_require_auth_student("appointment.student.time_ranges", func(ctx context.Context, req rpc.Request) rpc.Return {
		ret := req.Ret("ok")
		err := db.View(func(tx *bolt.Tx) error {
			return appoint.EachTimeRange(tx, func(tr appoint.TimeRange) error {
				if tr.Status == appoint.Status_Enable {
					ret = ret.Set(appoint.TimeRangeId(tr), fmt.Sprintf("%v:%v:%s", tr.From, tr.To, tr.Teacher))
				}
				return nil
			})
		})
		if err != nil {
			return ErrorRet(err, req)
		}
		return ret
	})
	register_require_auth_student("appointment.student.appoint", func(ctx context.Context, req rpc.Request) rpc.Return {
		id := rpc.GetId(ctx)
		status := appoint.GetData().UserStatus[id]

		if status != appoint.UserState_Unappointed {
			return ErrorRet(utils.ErrInternal, req)
		}

		tr_id := req.Get("tr_id", "")

		err := db.Update(func(tx *bolt.Tx) error {
			defer appoint.UpdateData(tx)

			return appoint.UpdateTimeRange(tx, tr_id, func(tr appoint.TimeRange) (appoint.TimeRange, error) {
				if tr.Student != "" || tr.Status != appoint.Status_Enable {
					return tr, utils.ErrOp
				}

				tr.Student = id
				tr.Status = appoint.Status_Disable
				return tr, nil
			})
		})
		if err != nil {
			return ErrorRet(err, req)
		}
		return req.Ret("ok")
	})

	// teacher
	register_require_auth_teacher := func(method string, fn func(ctx context.Context, req rpc.Request) rpc.Return) {
		r.RegisterFunc(method,
			func(ctx context.Context, req rpc.Request) rpc.Return {
				id := rpc.GetId(ctx)

				if id == "" {
					return ErrorRet(account.ErrUnvalid, req)
				}

				var role appoint.Role
				db.View(func(tx *bolt.Tx) error {
					role = appoint.GetRole(tx, id)
					return nil
				})

				if role != appoint.Role_Teacher {
					return ErrorRet(utils.ErrInternal, req)
				}

				return fn(ctx, req)
			})
	}

	register_require_auth_teacher("appointment.teacher.trs@canOperate", func(ctx context.Context, req rpc.Request) rpc.Return {
		id := rpc.GetId(ctx)
		ret := req.Ret("ok")
		err := db.View(func(tx *bolt.Tx) error {
			return appoint.EachTimeRange(tx, func(tr appoint.TimeRange) error {
				if tr.Teacher != id {
					return nil
				}

				if tr.Operable() {
					ret = ret.Set(appoint.TimeRangeId(tr), fmt.Sprintf("%v:%v:%v", tr.From, tr.To, tr.Status))
				}
				return nil
			})
		})
		if err != nil {
			return ErrorRet(err, req)
		}
		return ret
	})

	register_require_auth_teacher("appointment.teacher.tr.close", func(ctx context.Context, req rpc.Request) rpc.Return {
		id := rpc.GetId(ctx)
		tr_id := req.Get("tr_id", "")
		err := db.Update(func(tx *bolt.Tx) error {
			return appoint.UpdateTimeRange(tx, tr_id, func(tr appoint.TimeRange) (appoint.TimeRange, error) {
				if !tr.Operable() || tr.Status != appoint.Status_Enable || tr.Teacher != id {
					log.E(err).Error()
					return tr, utils.ErrInternal
				}

				tr.Status = appoint.Status_Disable
				return tr, nil
			})
		})
		if err != nil {
			return ErrorRet(err, req)
		}
		return req.Ret("ok")
	})
	register_require_auth_teacher("appointment.teacher.tr.open", func(ctx context.Context, req rpc.Request) rpc.Return {
		id := rpc.GetId(ctx)
		tr_id := req.Get("tr_id", "")
		err := db.Update(func(tx *bolt.Tx) error {
			return appoint.UpdateTimeRange(tx, tr_id, func(tr appoint.TimeRange) (appoint.TimeRange, error) {
				if !tr.Operable() || tr.Status != appoint.Status_Disable || tr.Teacher != id {
					log.E(err).Error()
					return tr, utils.ErrInternal
				}

				tr.Status = appoint.Status_Enable
				return tr, nil
			})
		})
		if err != nil {
			return ErrorRet(err, req)
		}
		return req.Ret("ok")
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
	case utils.ErrOp:
		return req.Ret(string(rpc.Op))

		// ADD TO HERE
	default:
		log.E(err).Error("not implement this type for ErrorRet")
		return req.Ret(string(rpc.Internal))
	}
}
