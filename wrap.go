package e

import (
	"context"
	"errors"
	"runtime"
	"runtime/debug"
	"strings"

	"google.golang.org/api/googleapi"
	"google.golang.org/grpc/codes"
)

// New returns an initialized *Err with the provided string
func New(s string, opts ...WrapOption) *Err {
	return wrap(errors.New(s), opts...)
}

// Wrap adds context to the provided error
func Wrap(err error, opts ...WrapOption) *Err {
	if err == nil {
		return New("you tried to wrap a nil error", Code(codes.DataLoss), Critical(), NotRetriable())
	}
	return wrap(err, opts...)
}

// wrap is sed so Wrap/New have the same function depth
func wrap(rawErr error, opts ...WrapOption) *Err {
	if googleErr, ok := rawErr.(*googleapi.Error); ok {
		// prepend these options in case the caller has set a specific code
		opts = append([]WrapOption{
			Code(CodeFromHTTPStatus(googleErr.Code)),
			With("googleError", googleErr),
		}, opts...)
	}

	// we could simply use errors.As, but since we expect the vast majority of calls to already be
	// wrapped, we'll do a fast path type assertion that doubles as creating the err/wasAlreadyWrapped vars
	err, wasAlreadyWrapped := rawErr.(*Err)
	if !wasAlreadyWrapped {
		wasAlreadyWrapped = errors.As(rawErr, &err) // annoyingly allocates
	}

	switch {
	// if the error is already wrapped and there are no opts to run, exit immediately
	case wasAlreadyWrapped && len(opts) == 0:
		err.currentFrameIdx++
		return err

	// the error hasn't been wrapped, so initialize it
	case !wasAlreadyWrapped:
		err = newErr(rawErr, debug.Stack())

	default:
		// Extract details about the calling function and package
		pc, filepath, line, ok := runtime.Caller(2)
		var fn *runtime.Func
		if ok {
			fn = runtime.FuncForPC(pc)
			err.currentFrameIdx = err.findFrame(fn.Name())

			// exit the switch if the current stack frame was found
			if err.currentFrameIdx != -1 {
				break
			}
		}

		// if we get here, no frame was found, so we'll initialize one
		err.unknownFrames = append(err.unknownFrames, &frame{
			fn:   fn.Name(),
			full: fn.Name(),
			file: filepath,
			line: line,
		})
	}

	// if any custom errors implement our interfaces, add them to the opts
	opts = append(opts, getWrapOptionsFromInterfaces(rawErr)...)

	// execute WrapOptions
	for _, opt := range opts {
		opt(err)
	}

	switch {
	// there's a chicken and the egg problem with the skip frame option not being available
	// until after the stack has been parsed, so this manipulates it if necessary
	case !wasAlreadyWrapped && err.skipFrames > 0:
		err.handleSkipFrameOption()

	// increment the index so we aren't searching every frame every time we go up the stack
	// this also serves the purpose of resetting index to 0 if findFrame failed. This needs
	// to not run after skip frames, because we don't know if we should advance.
	default:
		err.currentFrameIdx++
	}

	return err
}

func newErr(rawErr error, stack []byte) *Err {
	err := &Err{
		rootErr:         rawErr,
		frames:          parseStack(stack),
		currentFrameIdx: 0,
		isRetriable:     true,
		memStats:        &runtime.MemStats{},
		Level:           LevelError,
	}

	// set default error code
	switch {
	case rawErr == context.Canceled:
		err.code = codes.Canceled

	case rawErr == context.DeadlineExceeded:
		err.code = codes.DeadlineExceeded

	case strings.Contains(rawErr.Error(), "context canceled"):
		err.code = codes.Canceled

	case strings.Contains(rawErr.Error(), "deadline exceeded"):
		err.code = codes.DeadlineExceeded

	default:
		err.code = codes.Unknown
	}

	// load memstats for later consumption
	runtime.ReadMemStats(err.memStats)

	err.setTopAppFrame()

	// keep the raw stack trace around
	err.rawStack = stack

	return err
}

func (err *Err) setTopAppFrame() {
	// set the top level app frame, which will be used as the main reporting frame
	// default to the topmost frame in case no app frame is found
	err.topAppFrame = err.frames[0]
	for _, f := range err.frames {
		if err.topAppFrame == nil && f.class == classApp {
			err.topAppFrame = f
		}

		// if any of the frames contain a panic, mark it as such
		if f.class == classPanic {
			err.isPanic = true
		}
	}
}

// since the SkipFrame option doesn't run before the stack is parsed and other
// options, we have to manually edit the stack
func (err *Err) handleSkipFrameOption() {
	switch {
	// the caller can't skip more frames than there are, and there has to be at least
	// one frame left to attach details to
	case err.skipFrames >= len(err.frames):
		With("invalidSkipFrameCount", err.skipFrames)(err)

	// we're now left with a valid number of frames to skip, so we don't have to do
	// any additional length checks to avoid panics
	case err.skipFrames > 0:
		// get the first frame that won't be skipped
		mergeFrame := err.frames[err.skipFrames]

		// copy all the frame details to the new top of the stack
		mergeFrame.vars = err.frames[0].vars
		mergeFrame.msg = err.frames[0].msg
		mergeFrame.errDetails = err.frames[0].errDetails

		// we generally won't be skipping enough frames to make it worthwhile to copy out frame
		// values to let the unreferenced frames get garbage collected sooner
		err.frames = err.frames[err.skipFrames:]

		// since we've manipulated the stack, we need to rerun the top app frame pointer
		err.setTopAppFrame()
	}
}
