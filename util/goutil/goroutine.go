package goutil

import (
	"context"
	"fmt"
	"rock_review/util/xlogger"
	"runtime/debug"
)

func SafeGo(ctx context.Context, f func()) {
	go SafeDo(ctx, f)
}

func SafeDo(ctx context.Context, f func()) error {
	var err error

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("[panic recovered] stack: %s; panic: %v", string(debug.Stack()), r)
			xlogger.ErrorF(ctx, "[panic recovered] stack: %s; panic: %v", string(debug.Stack()), r)
		}
	}()
	f()

	return err
}
