package e

import "google.golang.org/grpc/codes"

// ErrVars adds all the returned vars to the current stack frame,
// overriding any existing keys
type ErrVars interface {
	ErrVars() map[string]interface{}
}

// ErrTags adds all the returned tags, overriding any existing keys
type ErrTags interface {
	ErrTags() map[string]string
}

// ErrMsg sets the error message, overriding any existing message
type ErrMsg interface {
	ErrMsg() string
}

// ErrCode sets the error code, overriding any existing code
type ErrCode interface {
	ErrCode() codes.Code
}

// ErrLevel sets the error level, adding a var if the level is invalid
type ErrLevel interface {
	ErrLevel() Level
}

// ErrRetriable determines if the error is retriable
type ErrRetriable interface {
	ErrRetriable() bool
}

// ErrReportable sets whether the error should be reported
type ErrReportable interface {
	ErrShouldReport() bool
}

// ErrInfra determines whether an error is infrastructure related
type ErrInfra interface {
	ErrIsInfra() bool
}

func getWrapOptionsFromInterfaces(v interface{}) []WrapOption {
	var opts []WrapOption

	if interfacer, ok := v.(ErrVars); ok {
		for k, v := range interfacer.ErrVars() {
			opts = append(opts, With(k, v))
		}
	}

	if interfacer, ok := v.(ErrTags); ok {
		for k, v := range interfacer.ErrTags() {
			opts = append(opts, Tag(k, v))
		}
	}

	if interfacer, ok := v.(ErrMsg); ok {
		opts = append(opts, Msg(interfacer.ErrMsg()))
	}

	if interfacer, ok := v.(ErrCode); ok {
		opts = append(opts, Code(interfacer.ErrCode()))
	}

	if interfacer, ok := v.(ErrRetriable); ok {
		if !interfacer.ErrRetriable() {
			opts = append(opts, NotRetriable())
		}
	}

	if interfacer, ok := v.(ErrReportable); ok {
		if !interfacer.ErrShouldReport() {
			opts = append(opts, NoReport())
		}
	}

	if interfacer, ok := v.(ErrInfra); ok {
		if !interfacer.ErrIsInfra() {
			opts = append(opts, Infra())
		}
	}

	if interfacer, ok := v.(ErrLevel); ok {
		switch interfacer.ErrLevel() {
		case LevelWarning:
			opts = append(opts, Warning())
		case LevelError:
			opts = append(opts, Error())
		case LevelCritical:
			opts = append(opts, Critical())
		default:
			opts = append(opts, With("invalidLevel", interfacer.ErrLevel()))
		}
	}

	return opts
}
