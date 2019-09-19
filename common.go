package locker

import (
	"context"
	"fmt"
	"log"
)

// LogFunc ...
type LogFunc func(ctx context.Context, args ...interface{})

// NewContextFunc ...
type NewContextFunc func(ctx context.Context) context.Context

// default new context func
func defaultNewCtx(ctx context.Context) context.Context {
	return ctx
}

// default log func
func defaultLogfunc(ctx context.Context, args ...interface{}) {
	var s string
	if len(args) > 1 {
		if format, ok := args[0].(string); ok {
			s = fmt.Sprintf(format, args[1:]...)
		}
	} else {
		s = fmt.Sprint(args...)
	}
	s = "[INFO]" + s
	log.Output(2, s)
}

// default log error func
func defaultLogErrfunc(ctx context.Context, args ...interface{}) {
	var s string
	if len(args) > 1 {
		if format, ok := args[0].(string); ok {
			s = fmt.Sprintf(format, args[1:]...)
		}
	} else {
		s = fmt.Sprint(args...)
	}
	s = "[ERROR]" + s
	log.Output(2, s)
}

// SetLogFunc 设置输出普通日志的方法
func (r *RedisLocker) SetLogFunc(f LogFunc) {
	r.log = f
}

// SetLogErrorFunc 设置输出错误日志的方法
func (r *RedisLocker) SetLogErrorFunc(f LogFunc) {
	r.logErr = f
}

// SetDoEvalFunc 设置eval方法，请自行包装，适配各种redis library
func (r *RedisLocker) SetDoEvalFunc(f EvalFunc) {
	r.doEval = f
}
