package rpc

import (
	"context"
	"fmt"

	"github.com/immofon/appoint/log"
)

type SetContextKV func(k, v interface{})

func WithSetContextValue(parent context.Context, fn SetContextKV) context.Context {
	return context.WithValue(parent, SetContextKV(nil), fn)
}

func GetSetContextValue(ctx context.Context) SetContextKV {
	skv, ok := ctx.Value(SetContextKV(nil)).(SetContextKV)
	if ok && skv != nil {
		return skv
	} else {
		return func(k, v interface{}) {
			log.L().
				WithField("key", fmt.Sprint(k)).
				WithField("value", fmt.Sprint(v)).
				Error("try to set k,v to context")
		}
	}
}
