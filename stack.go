package e

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"strings"

	"github.com/maruel/panicparse/v2/stack"
)

// findFrame loops through the existing stack frames to find the one that matches the function name.
// We can't compare line numbers because the stacktrace reports the line for the function,
// while the Wrap call reports the source line of the Wrap. This is a coarse find
func (err *Err) findFrame(fullFnName string) int {
	var f *frame
	for i := err.currentFrameIdx; i < len(err.frames); i++ {
		f = err.frames[i]
		if f.full == fullFnName {
			return i
		}
	}

	return -1
}

func parseStack(rawStack []byte) []*frame {
	ppStack := getPanicParseStack(rawStack)
	frames := make([]*frame, 0, len(ppStack.Calls))

	for _, call := range ppStack.Calls {
		// determine whether or not to skip frames that just clutter the stacktrace
		if shouldSkipFrame(&call, len(frames)) {
			continue
		}

		// once we reach the router, grpc layer, or root worker, we can eliminate the rest of the stack
		// checking the length of the frames prevents false positives, in case the error originates
		// in a frame that typically causes truncation. As of 2021-02-24, this only happens in testing.
		if shouldTruncateStack(&call) && len(frames) > 1 {
			break
		}

		frames = append(frames, &frame{
			file:  call.SrcName,
			path:  call.Func.ImportPath,
			pkg:   call.Func.DirName,
			fn:    call.Func.Name,
			full:  call.Func.Complete,
			line:  call.Line,
			class: getFrameClass(&call),
		})
	}

	return frames
}

// getPanicParseStack expects a stdlib stacktrace from runtime.Stack or debug.Stack and returns
// the parsed stack object. It is a convenience function wrapping ScanSnapshot.
func getPanicParseStack(rawStack []byte) stack.Stack {
	s, _, err := stack.ScanSnapshot(bytes.NewReader(rawStack), ioutil.Discard, stack.DefaultOpts())
	if err != nil && err != io.EOF {
		panic(err)
	}

	if len(s.Goroutines) > 1 {
		panic(errors.New("provided stacktrace had more than one goroutine"))
	}

	return s.Goroutines[0].Signature.Stack
}

func getFrameClass(call *stack.Call) class {
	// set the frame type
	switch {
	case call.Func.Name == "panic":
		return classPanic

	case call.Location == stack.Stdlib:
		return classStdLib

	case strings.HasPrefix(call.Func.ImportPath, "go.nozzle.io/pkg"):
		return classPkg

	case strings.HasPrefix(call.Func.ImportPath, "go.nozzle.io/vendor"):
		return classVendor

	case strings.HasPrefix(call.Func.ImportPath, "go.nozzle.io"):
		return classApp

	case call.Location == stack.GoPkg || call.Location == stack.GoMod:
		return classVendor

	case !strings.Contains(call.Func.ImportPath, "."):
		return classStdLib

	default:
		return classVendor
	}
}

//nolint:gocyclo
func shouldSkipFrame(call *stack.Call, frameCount int) bool {
	// for now, we're only skipping frames at the top of the stack
	if frameCount > 0 {
		return false
	}

	// skip the wrap functions in this package if they are at the top of the stack
	if call.Func.DirName == "e" &&
		(call.Func.Name == "wrap" || call.Func.Name == "Wrap" || call.Func.Name == "New" || call.Func.Name == "StatusToError") {
		return true
	}

	// skip the runtime debug frame
	if call.Func.ImportPath == "runtime/debug" && call.Func.Name == "Stack" {
		return true
	}

	// skip extra client frames for grpc calls
	if call.Func.DirName == "ngrpc" && call.Func.Name == "contextClientUnaryInterceptor" {
		return true
	}

	if call.Func.DirName == "grpc" && call.Func.Name == "(*ClientConn).Invoke" {
		return true
	}

	// skip MySQL error normalization
	if call.Func.DirName == "mysql" &&
		(call.Func.Name == "normalizeError" ||
			call.Func.Name == "(*DB).ExecContext" || call.Func.Name == "(*DB).QueryContext" || call.Func.Name == "(*DB).QueryRowContext" ||
			call.Func.Name == "(*Tx).ExecContext" || call.Func.Name == "(*Tx).QueryContext" || call.Func.Name == "(*Tx).QueryRowContext" ||
			call.Func.Name == "(*Stmt).ExecContext" || call.Func.Name == "(*Stmt).QueryContext" || call.Func.Name == "(*Stmt).QueryRowContext") {
		return true
	}

	return false
}

func shouldTruncateStack(call *stack.Call) bool {
	switch call.Func.Complete {
	case "go.nozzle.io/pkg/router.route.func1",
		"go.nozzle.io/pkg/ngrpc.execHandler",
		"go.nozzle.io/pkg/workers.(*worker).execUserFn",
		"github.com/spf13/cobra.(*Command).execute",
		"testing.tRunner":
		return true

	default:
		return false
	}
}
