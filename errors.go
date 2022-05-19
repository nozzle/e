package e

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/getsentry/sentry-go"
	"github.com/golang/protobuf/proto"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	// verify that Err conforms to the GRPC status interface
	_ interface {
		GRPCStatus() *status.Status
	} = (*Err)(nil)
)

// Err defines an error with an optional stacktrace and contextual data
type Err struct {
	rootErr error

	// raw data captured at the moment the error was initially wrapped
	rawStack []byte
	memStats *runtime.MemStats

	// frames that weren't skipped/truncated from the rawStack parse
	frames stackframes
	// if findFrame fails, frame data is stored here. This typically
	// only happens when wrapping across goroutine boundaries, as that
	// leaves the
	unknownFrames   stackframes
	currentFrameIdx int
	topAppFrame     *frame

	// contextual data set from WrapOptions
	Level       Level
	code        codes.Code
	tags        map[string]string
	isRetriable bool
	noReport    bool
	isPanic     bool
	isInfra     bool
	skipFrames  int

	// contextual data set from ReportOptions
	fromHandler  bool
	shouldWait   bool
	sentryClient *sentry.Client
}

type stackframes []*frame

type frame struct {
	msg   string
	file  string
	path  string
	pkg   string
	fn    string
	full  string
	line  int
	class class
	vars  map[string]interface{}
	// errDetails contains proto messages containing detailed descriptions
	// ex. errdetails.FieldViolations, errdetails.Help, errdetails.PreconditionFailure.
	errDetails []proto.Message
}

type class int

const (
	// this frame is application logic
	classApp = 1
	// this frame is our internal pkg
	classPkg = 2
	// this frame is vendored code
	classVendor = 3
	// this frame is from the stdlib
	classStdLib = 4
	// this frame is a panic
	classPanic = 5
)

// Level represents an error severity level
type Level int

const (
	// LevelCritical = 1
	LevelCritical Level = 1
	// LevelError = 1
	LevelError Level = 2
	// LevelWarning = 1
	LevelWarning Level = 3
)

// String fulfills the Stringer interface
func (err *Err) String() string {
	return err.Error()
}

type causer interface {
	Cause() error
}

type unwrapper interface {
	Unwrap() error
}

// Cause returns the root cause of the error without the extra wrapped data.
// Supports our own errors, Dave Cheney's pkg/errors, and anything that implements
// either the Cause() error or Unwrap() error interface.
func Cause(rawErr error) error {
	switch err := rawErr.(type) {
	case *Err:
		return err.rootErr
	case unwrapper:
		return Cause(err.Unwrap())
	case causer:
		return Cause(err.Cause())
	default:
		return rawErr
	}
}

// Unwrap fulfills the stdlib Unwrap interface
func (err *Err) Unwrap() error {
	return err.rootErr
}

// Error fulfills the error interface
func (err *Err) Error() string {
	buf := getBuffer()
	defer putBuffer(buf)

	// print out the simple error string for easy consumption
	buf.WriteString(err.rootErr.Error())
	buf.WriteByte('\n')
	buf.WriteByte('\n')
	buf.WriteString("Error Code: ")
	buf.WriteString(err.code.String())
	buf.WriteByte('\n')

	// print out a link to where the error occurred
	buf.WriteString(err.githubURL())
	buf.WriteByte('\n')
	buf.WriteByte('\n')

	// callout that this task isn't retriable
	if !err.isRetriable {
		buf.WriteString("--- NOT RETRIABLE ---\n\n")
	}

	// print out our organized stack trace with any added context
	buf.WriteString(err.frames.String())

	// extra reporting if unknown frames were found
	if len(err.unknownFrames) > 0 {
		buf.WriteString("\n\n--- UNKNOWN STACK FRAMES FOUND ---")
		buf.WriteString(err.unknownFrames.String())
	}

	// spew out the full error with type info if it is anything but errors.New()
	errStr := fmt.Sprintf("%#v", err.rootErr)
	if !strings.HasPrefix(errStr, "&errors.errorString") {
		buf.WriteString("\nErr: ")
		buf.WriteString(spew.Sdump(err.rootErr))
	}

	return buf.String()
}

// Location returns the path/file/line where the first New/Wrap happened
func Location(rawErr error) (string, string, int) {
	err, ok := rawErr.(*Err)
	if !ok || len(err.frames) == 0 {
		return "", "", 0
	}

	f := err.frames[0]

	return f.path, f.file, f.line
}

// CodeFromHTTPStatus does the inverse of HTTPStatusFromCode. It's a lossy conversion as there isn't a 1 to 1 mapping
func CodeFromHTTPStatus(statusCode int) codes.Code {
	switch statusCode {
	case http.StatusOK:
		return codes.OK
	case http.StatusRequestTimeout:
		return codes.Canceled
	case http.StatusInternalServerError:
		return codes.Unknown
	case http.StatusBadRequest:
		return codes.InvalidArgument
	case http.StatusGatewayTimeout:
		return codes.DeadlineExceeded
	case http.StatusNotFound:
		return codes.NotFound
	case http.StatusConflict:
		return codes.AlreadyExists
	case http.StatusForbidden:
		return codes.PermissionDenied
	case http.StatusUnauthorized:
		return codes.Unauthenticated
	case http.StatusTooManyRequests:
		return codes.ResourceExhausted
	case http.StatusNotImplemented:
		return codes.Unimplemented
	case http.StatusServiceUnavailable:
		return codes.Unavailable
	default:
		return codes.Unknown
	}
}

// Code returns the error code
func (err *Err) Code() codes.Code {
	return err.code
}

// CodeFromError returns the error code from an error if there is one.
// Returns OK if err is nil and Unknown if it isn't a wrapped error.
func CodeFromError(err error) codes.Code {
	if err == nil {
		return codes.OK
	}

	typedErr, ok := err.(*Err)
	if !ok {
		return codes.Unknown
	}

	return typedErr.code
}

// IsRetriable returns whether or not a retry might succeed
func (err *Err) IsRetriable() bool {
	return err.isRetriable
}

// githubURL returns a direct link to the line where the first New/Wrap happened
// https://github.com/nozzle/nozzle/blob/ab15cc7c4cf7ed21fcaa8957ef390dbdbc7e1816/namespaces/publish/rankings_client.go#L11
func (err *Err) githubURL() string {
	if len(err.frames) == 0 {
		return ""
	}

	buf := getBuffer()
	defer putBuffer(buf)

	d := err.frames[0]
	buf.WriteString("https://github.com/nozzle/e/blob/")
	buf.WriteString("main")
	buf.WriteString(strings.TrimPrefix(d.path, "github.com/nozzle/e"))
	buf.WriteByte('/')
	buf.WriteString(d.file)
	buf.WriteString("#L")
	buf.WriteString(strconv.Itoa(d.line))

	return buf.String()
}

func (err *Err) currentFrame() *frame {
	// don't panic if the current frame is unknown from the stack parse
	if err.currentFrameIdx == -1 {
		return err.unknownFrames[len(err.unknownFrames)-1]
	}

	return err.frames[err.currentFrameIdx]
}

func (fs stackframes) String() string {
	buf := getBuffer()
	defer putBuffer(buf)

	buf.WriteString("-------------------------------\n")
	for i := range fs {
		buf.WriteString(fs[i].String())
		buf.WriteString("\n")
	}
	return buf.String()
}

const maxLogStringLen = 500

func (f frame) String() string {
	buf := getBuffer()
	defer putBuffer(buf)

	// write the user provided message if there is one
	if f.msg != "" {
		buf.WriteString("*** ")
		buf.WriteString(f.msg)
		buf.WriteByte('\n')
	}

	// write the file, line and function name
	// e.g. errors_test.go:103 - TestRecursiveWrap.func1
	buf.WriteString(fmt.Sprintf("%s.%s", f.pkg, f.fn))
	buf.WriteString(fmt.Sprintf("\n %s/%s:%d", f.path, f.file, f.line))

	for k, v := range f.vars {
		switch t := v.(type) {
		case fmt.Stringer:
			str := t.String()
			// limit the max length of the string to be logged
			if len(str) > maxLogStringLen {
				str = str[:maxLogStringLen]
			}
			buf.WriteString(fmt.Sprintf("\n + %s: %#v", k, str))

		case string:
			// limit the max length of the string to be logged
			if len(t) > maxLogStringLen {
				t = t[:maxLogStringLen]
			}
			buf.WriteString(fmt.Sprintf("\n + %s: %#v", k, t))

		case []byte:
			// limit the max length of the string to be logged
			if len(t) > maxLogStringLen {
				t = t[:maxLogStringLen]
			}
			buf.WriteString(fmt.Sprintf("\n + %s: %#v", k, string(t)))

		default:
			buf.WriteString(fmt.Sprintf("\n + %s: %#v", k, v))
		}
	}

	// if we have error details include those in the message
	for _, detail := range f.errDetails {
		buf.WriteByte('\n')
		switch t := detail.(type) {
		case *errdetails.BadRequest:
			buf.WriteString(" + Bad Request:")
			for _, violation := range t.FieldViolations {
				buf.WriteString(fmt.Sprintf("\n\t%s: %s", violation.Field, violation.Description))
			}

		default:
			buf.WriteString(" + Unknown ErrDetail: ")
			buf.WriteString(t.String())
		}
	}

	buf.WriteString("\n-------------------------------")

	return buf.String()
}

// NewErrRequiredArg will build a structured error with errdetails
func NewErrRequiredArg(c context.Context, fields ...string) error {
	opts := []WrapOption{
		Code(codes.InvalidArgument),
	}

	for _, field := range fields {
		opts = append(opts, FieldViolation(field, field+" is required"), With("required field", field))
	}

	return New("required argument is missing", opts...)
}

// NewErrForbiddenArg creates an error with not allowed fields
func NewErrForbiddenArg(c context.Context, fields ...string) error {
	opts := []WrapOption{
		Code(codes.InvalidArgument),
	}

	for _, field := range fields {
		opts = append(opts, FieldViolation(field, field+" is not allowed"), With("disallowed field", field))
	}

	return New("forbidden argument", opts...)
}

// NewErrInvalidArg will return an error with an invalid field and why it was invalid
func NewErrInvalidArg(c context.Context, fields ...string) error {
	if len(fields)%2 == 1 {
		panic("odd number of arguments passed")
	}

	opts := []WrapOption{
		Code(codes.InvalidArgument),
	}

	for len(fields) > 0 {
		field, description := fields[0], fields[1]
		opts = append(opts, FieldViolation(field, description), With(field, description))
		fields = fields[2:]
	}

	return New("invalid argument", opts...)
}
