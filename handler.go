package e

import (
	"context"
	"errors"
	"log"
	"runtime/debug"

	"google.golang.org/grpc/codes"
)

type recoverData struct {
	shouldReport bool
	shouldLog    bool
	shouldPanic  bool
	fn           func(c context.Context, err *Err)
}

// RecoverOption lets you add additional context to a panic handler
type RecoverOption func(c context.Context, rd *recoverData)

// RecoverNoReport doesn't report the panic
func RecoverNoReport() RecoverOption {
	return func(c context.Context, rd *recoverData) {
		rd.shouldReport = false
	}
}

// RecoverNoLog doesn't log the panic
func RecoverNoLog() RecoverOption {
	return func(c context.Context, rd *recoverData) {
		rd.shouldLog = false
	}
}

// RecoverNoPanic doesn't propagate the panic
func RecoverNoPanic() RecoverOption {
	return func(c context.Context, rd *recoverData) {
		rd.shouldPanic = false
	}
}

// RecoverFunc runs before returning / propagating the panic
// This function provides access to the error created from the panic, save it if you want to use or return it
func RecoverFunc(fn func(c context.Context, err *Err)) RecoverOption {
	return func(c context.Context, rd *recoverData) {
		rd.fn = fn
	}
}

// Recover is intended to be used for recovering from panics. Should only be temporary when
// looking for panics, shouldn't be used long-term in production
// This must be used directly as the panic function so it cannot return an error
// defer e.Recover(c)
func Recover(c context.Context, opts ...RecoverOption) {
	rec := recover()
	if rec == nil {
		return
	}

	var rawErr error
	// find out exactly what the error was and set err
	switch x := rec.(type) {
	case string:
		rawErr = errors.New(x)
	case []byte:
		rawErr = errors.New(string(x))
	case error:
		rawErr = x
	default:
		rawErr = errors.New("unknown panic")
	}

	rd := &recoverData{
		shouldReport: true,
		shouldLog:    true,
		shouldPanic:  true,
	}
	for _, opt := range opts {
		opt(c, rd)
	}

	err := newErr(rawErr, debug.Stack())
	err.Level = LevelCritical
	err.code = codes.Internal
	err.fromHandler = false
	if c.Err() != nil {
		err = wrap(err, With("c.Err()", c.Err().Error()))
	}

	if rd.shouldReport {
		err.Report(c)
	}

	if rd.shouldLog {
		log.Println(err.Error())
	}

	// make sure we don't lose any logs due to this panic
	// l.Flush(c)

	// execute user function
	if rd.fn != nil {
		// the error with the stack trace is passed in here
		// handle the error in your function if you want to use it
		rd.fn(c, err)
	}

	if rd.shouldPanic {
		panic(err)
	}
}
