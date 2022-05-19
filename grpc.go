package e

import (
	"context"
	"log"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc/status"
)

// GRPCStatus converts the error into a grpc compatible status
func (err *Err) GRPCStatus() *status.Status {
	// if originating error is a status.Status return that otherwise create a new one
	st, ok := status.FromError(err.rootErr)
	if !ok {
		st = status.New(err.code, err.rootErr.Error())
	}

	stWithDetails, withDetailErr := st.WithDetails(err.Details()...)
	// if there was an error appending details just send out the standard status without them
	if withDetailErr != nil {
		log.Println("couldn't create status with details " + withDetailErr.Error())
		return st
	}

	return stWithDetails
}

// StatusToError converts a status to a nozzle Err.
func StatusToError(c context.Context, st *status.Status) error {
	// convert the status to a internal Err
	e := New(st.Message(), Code(st.Code()))
	f := e.currentFrame()
	details := st.Details()
	f.errDetails = make([]proto.Message, 0, len(details))
	for _, detail := range details {
		switch t := detail.(type) {
		case proto.Message:
			f.errDetails = append(f.errDetails, t)
		case error:
			// there was error parsing the proto message log it out
			log.Println("unable to parse proto message " + t.Error())
		default:
			// if we don't know the type add to the error as an unknown detail
			With("unknown detail", t)(e)
		}
	}

	return e
}

// Details returns a slice of google.golang.org/genproto/googleapis/rpc/errdetails set on an Err.
func (err *Err) Details() []proto.Message {
	var details []proto.Message
	for _, f := range err.frames {
		details = append(details, f.errDetails...)
	}

	return details
}
