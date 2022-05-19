package e

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/getsentry/sentry-go"
)

// Flush sends all buffered errors to servers before returning. Should always be
// called before program exit to ensure no lost errors.
func Flush(c context.Context) {
	wg := sync.WaitGroup{}

	if errorClient != nil {
		wg.Add(1)
		go func() {
			errorClient.Flush()
			wg.Done()
		}()
	}

	if sentryClient != nil {
		wg.Add(1)
		go func() {
			completed := sentryClient.Flush(10 * time.Second)
			if !completed {
				log.Println("sentryClient flush did NOT complete")
			}
			wg.Done()
		}()
	}

	if sentryCriticalClient != nil {
		wg.Add(1)
		go func() {
			completed := sentryCriticalClient.Flush(10 * time.Second)
			if !completed {
				log.Println("sentryCriticalClient flush did NOT complete")
			}
			wg.Done()
		}()
	}

	if sentryInfraClient != nil {
		wg.Add(1)
		go func() {
			completed := sentryInfraClient.Flush(10 * time.Second)
			if !completed {
				log.Println("sentryInfraClient flush did NOT complete")
			}
			wg.Done()
		}()
	}

	wg.Wait()
}

// A ReportOption let you determine report behavior
type ReportOption func(err *Err)

// Wait forces your process to wait synchronously until the report is sent, instead
// of queuing and asynchronously reporting which is the default behavior
func Wait() ReportOption {
	return func(err *Err) {
		err.shouldWait = true
	}
}

// ReportIsHandler sets the report as having come via a system handler
// Should not be set outside pkg
func ReportIsHandler() ReportOption {
	return func(err *Err) {
		err.fromHandler = true
	}
}

// ReportWithSentryClient reports the error with a custom Sentry client
func ReportWithSentryClient(cl *sentry.Client) ReportOption {
	return func(err *Err) {
		err.sentryClient = cl
	}
}

// Report ships your error when you aren't returning through a handler
func (err *Err) Report(c context.Context, opts ...ReportOption) {
	for _, opt := range opts {
		opt(err)
	}

	if !err.shouldReport(c) {
		return
	}

	// ship data to Sentry and Google Error Reporting, waiting if necessary
	wg := sync.WaitGroup{}

	// send to Sentry
	wg.Add(1)
	go func() {
		err.reportToSentry(c)
		wg.Done()
	}()

	// send to Google Error Reporting
	if errorClient != nil {
		wg.Add(1)
		go func() {
			err.reportToStackdriver(c)
			wg.Done()
		}()
	}

	wg.Wait()
}

func (err *Err) shouldReport(c context.Context) bool {
	switch {
	case err.rootErr == nil:
		return true
	case err.noReport:
		return false
	default:
		return true
	}
}
