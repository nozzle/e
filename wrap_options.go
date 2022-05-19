package e

import (
	"encoding/json" //nolint:depguard // this is just for json.RawMessage, and there are import cycles with pkg/json

	jsoniter "github.com/json-iterator/go"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
)

// WrapOption lets you add context
type WrapOption func(err *Err)

//
// error scoped wrap options: typically the last value wins
//

// Code sets the error code. This is only set once per error, so if called
// multiple times, the last call persists.
// Panics if an invalid code is given.
func Code(code codes.Code) WrapOption {
	return func(err *Err) {
		if code == codes.OK {
			With("invalidOkCode", code)(err)
			return
		}

		if code == codes.InvalidArgument {
			err.isRetriable = false
		}

		err.code = code
	}
}

// Critical sets the importance of the error, overwriting the previous level
func Critical() WrapOption {
	return func(err *Err) {
		err.Level = LevelCritical
	}
}

// Error sets the importance of the error, overwriting the previous level
func Error() WrapOption {
	return func(err *Err) {
		err.Level = LevelError
	}
}

// Warning sets the importance of the error, overwriting the previous level
func Warning() WrapOption {
	return func(err *Err) {
		err.Level = LevelWarning
	}
}

// Tag adds a tag to the error. It will overwrite previous tags with the same key.
// This is error scoped, not frame scoped.
func Tag(key, val string) WrapOption {
	return func(err *Err) {
		if err.tags == nil {
			err.tags = make(map[string]string)
		}
		err.tags[key] = val
	}
}

// NotRetriable denotes that this error should be retried
func NotRetriable() WrapOption {
	return func(err *Err) {
		err.isRetriable = false
	}
}

// NoReport denotes that this error should not be reported to Sentry
func NoReport() WrapOption {
	return func(err *Err) {
		err.noReport = true
	}
}

// Infra designates an error specifically as infrastructure related, not app related
func Infra() WrapOption {
	return func(err *Err) {
		err.isInfra = true
	}
}

// Stack lets you provide a stacktrace as opposed to capturing it internally
func Stack(stack []byte) WrapOption {
	return func(err *Err) {
		err.rawStack = stack
	}
}

// Override lets you override the rootErr with a new error
func Override(overrideErr error) WrapOption {
	return func(err *Err) {
		err.rootErr = overrideErr
	}
}

// SkipFrames works like runtime.Caller, where the top N frames aren't included
// in the stack trace. This should rarely be used, except for in cases where we
// are normalizing errors, but don't want all stacks to originate from a single
// line.
// All WrapOptions will be merged into the new top of the stack var, so
// that they aren't lost. If skipFrames is > 1, any WrapOptions in between
// the first wrap will be applied to an unknown frame, since when they execute,
// that frame will already have been removed. If you really need to skip multiple
// frames, you should pass those options up to the primary wrap site.
func SkipFrames(skipFrames int) WrapOption {
	return func(err *Err) {
		err.skipFrames = skipFrames
	}
}

//
// frame scoped wrap options: typically the last value wins
//

// Msg lets you assign a message to the error
func Msg(msg string) WrapOption {
	return func(err *Err) {
		f := err.currentFrame()
		f.msg = msg
	}
}

// With lets you attach a variable
func With(k string, v interface{}) WrapOption {
	return func(err *Err) {
		// if v implements any of our interfaces, we'll execute those options inline,
		// then immediately return, since the caller chose to take control of the error output
		opts := getWrapOptionsFromInterfaces(v)
		if len(opts) > 0 {
			// execute the options before returning
			for _, opt := range opts {
				opt(err)
			}

			return
		}

		f := err.currentFrame()

		if f.vars == nil {
			f.vars = make(map[string]interface{})
		}

		// []byte is generally useless to attach, so we'll convert it
		// to a json.RawMessage or string
		var b []byte
		switch t := v.(type) {
		case []byte:
			b = t
		case *[]byte:
			if t != nil {
				b = *t
			}
		}
		if b != nil {
			if jsoniter.Valid(b) {
				v = json.RawMessage(b)
			} else {
				v = string(b)
			}
		}

		f.vars[k] = v
	}
}

// FieldViolation adds a error details for a field violation.
func FieldViolation(field, description string) WrapOption {
	return func(err *Err) {
		f := err.currentFrame()
		fieldViolation := &errdetails.BadRequest_FieldViolation{
			Field:       field,
			Description: description,
		}

		// look for existing bad request and set the field violation on it
		var found bool
		for _, detail := range f.errDetails {
			switch t := detail.(type) {
			case *errdetails.BadRequest:
				t.FieldViolations = append(t.FieldViolations, fieldViolation)
				found = true
			}
		}

		// if none was found create one and add to details
		if !found {
			f.errDetails = append(f.errDetails, &errdetails.BadRequest{
				FieldViolations: []*errdetails.BadRequest_FieldViolation{fieldViolation},
			})
		}
	}
}
