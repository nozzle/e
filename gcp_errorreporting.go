package e

import (
	"context"

	er "cloud.google.com/go/errorreporting"
)

var errorClient *er.Client

// SetStackdriverErrorReportingClient enables reporting to GCP.
// Should only be called once at app startup.
func SetStackdriverErrorReportingClient(c context.Context, cl *er.Client) error {
	errorClient = cl
	return nil
}

func (err *Err) reportToStackdriver(c context.Context) {
	entry := er.Entry{
		Error: err.rootErr,
		// Req:   internal.Request(c),
		Stack: err.rawStack,
	}
	errorClient.Report(entry)
}
