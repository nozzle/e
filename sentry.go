package e

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
)

var (
	sentryClient         *sentry.Client
	sentryCriticalClient *sentry.Client
	sentryInfraClient    *sentry.Client

	// this initializes all the environment variables to ship with errors
	envVars = func() []envVar {
		environ := os.Environ()
		vars := make([]envVar, 0, len(environ))

		for _, e := range environ {
			pair := strings.SplitN(e, "=", 2)
			ev := envVar{k: pair[0]}
			if len(pair) > 1 {
				ev.v = pair[1]
			}
			vars = append(vars, ev)
		}

		return vars
	}()
)

type envVar struct {
	k, v string
}

// SetSentryClient stores a global reference to a sentry client.
// Should only be called once at app startup.
func SetSentryClient(c context.Context, cl *sentry.Client) {
	sentryClient = cl
}

// SetSentryCriticalClient stores a global reference to a sentry client used only for critical errors.
// Should only be called once at app startup.
func SetSentryCriticalClient(c context.Context, cl *sentry.Client) {
	sentryCriticalClient = cl
}

// SetSentryInfraClient stores a global reference to a sentry client used only for infrastructure errors.
// Should only be called once at app startup.
func SetSentryInfraClient(c context.Context, cl *sentry.Client) {
	sentryInfraClient = cl
}

func (err *Err) reportToSentry(c context.Context) {
	var cl *sentry.Client
	switch {
	case err.sentryClient != nil:
		cl = err.sentryClient

	case err.isInfra:
		cl = sentryInfraClient

	case err.Level == LevelCritical:
		cl = sentryCriticalClient

	default:
		cl = sentryClient
	}

	if cl == nil {
		log.Println("error reported to Sentry without valid client")
		return
	}

	cl.CaptureEvent(err.sentryEvent(c), nil, nil)
	if err.shouldWait {
		cl.Flush(2 * time.Second)
	}
}

func (err *Err) sentryEvent(c context.Context) *sentry.Event {
	var msg string
	for _, f := range err.frames {
		if f.msg != "" {
			msg = f.msg
			break
		}
	}

	var req *sentry.Request
	// activeRequest := internal.Request(c)
	// if activeRequest != nil {
	// 	req = sentry.NewRequest(activeRequest)
	// }

	return &sentry.Event{
		Extra:   err.sentryExtra(),
		Level:   err.sentryLevel(),
		Tags:    err.sentryTags(c),
		Message: msg,
		User:    sentry.User{
			// ID: strconv.FormatInt(ctx.UserID(c), 10),
		},
		Request: req,
		Exception: []sentry.Exception{{
			Type:       err.rootErr.Error(),
			Value:      err.githubURL(),
			Stacktrace: err.sentryStacktrace(),
		}},
	}
}

func (err *Err) sentryTags(c context.Context) map[string]string {
	// set default error reporting tags
	errTags := make(map[string]string, 15)

	// set user tags first, so that they can't overwrite built in tags
	for k, v := range err.tags {
		setTagIfNotEmpty(errTags, k, v)
	}

	// set env vars
	// setTagIfNotEmpty(errTags, "dataDomain", env.DataDomain)
	// setTagIfNotEmpty(errTags, "app", env.App)
	// setTagIfNotEmpty(errTags, "imageTag", env.ImageTag)
	// setTagIfNotEmpty(errTags, "node", env.Node)
	// setTagIfNotEmpty(errTags, "region", env.Region)
	// setTagIfNotEmpty(errTags, "zone", env.Zone)

	// set current user / workspace context
	// setTagIfNotEmpty(errTags, "workspaceID", strconv.FormatInt(internal.WorkspaceID(c), 10))

	// uID := internal.UserID(c)
	// if uID != 0 {
	// 	errTags["userID"] = strconv.FormatInt(uID, 10)
	// }

	// set main error fields
	setBoolTag(errTags, "handlerErr", err.fromHandler)
	setBoolTag(errTags, "isRetriable", err.isRetriable)
	setBoolTag(errTags, "shouldWait", err.shouldWait)
	setBoolTag(errTags, "hasUnknownFrames", len(err.unknownFrames) > 0)
	setBoolTag(errTags, "isPanic", err.isPanic)
	setTagIfNotEmpty(errTags, "errCode", err.code.String())

	switch err.Level {
	case LevelCritical:
		errTags["level"] = "critical"
	case LevelError:
		errTags["level"] = "error"
	case LevelWarning:
		errTags["level"] = "warning"
	}

	// add originating stack frame tags
	setTagIfNotEmpty(errTags, "file", err.topAppFrame.file)
	setTagIfNotEmpty(errTags, "path", err.topAppFrame.path)
	setTagIfNotEmpty(errTags, "pkg", err.topAppFrame.pkg)
	setTagIfNotEmpty(errTags, "fn", err.topAppFrame.fn)
	setTagIfNotEmpty(errTags, "line", strconv.Itoa(err.topAppFrame.line))

	// add additional request fields if available
	var activeRequest *http.Request
	// activeRequest := internal.Request(c)
	if activeRequest != nil {
		u := activeRequest.URL.String()
		// don't include querystring parameters in the requestPath
		questionIdx := strings.IndexByte(u, '?')
		if questionIdx == -1 {
			questionIdx = len(u)
		}

		setTagIfNotEmpty(errTags, "requestPath", u[:questionIdx])
		setTagIfNotEmpty(errTags, "rawQuery", activeRequest.URL.RawQuery)
		setTagIfNotEmpty(errTags, "method", activeRequest.Method)
		setTagIfNotEmpty(errTags, "userAgent", activeRequest.UserAgent())
		setTagIfNotEmpty(errTags, "referrer", activeRequest.Referer())
		setTagIfNotEmpty(errTags, "remoteIP", strings.TrimSuffix(activeRequest.RemoteAddr, ":80"))
	}

	// add message information to tags if available
	// queue, msgID, tryCount := internal.MessageDetails(c)
	// if queue != "" {
	// 	setTagIfNotEmpty(errTags, "queue", queue)
	// 	setTagIfNotEmpty(errTags, "msgID", strconv.FormatInt(msgID, 10))
	// 	setTagIfNotEmpty(errTags, "tryCount", strconv.FormatInt(int64(tryCount), 10))
	// }

	return errTags
}

func setTagIfNotEmpty(m map[string]string, k, v string) {
	if v != "" {
		m[k] = v
	}
}

func setBoolTag(m map[string]string, k string, v bool) {
	if v {
		m[k] = "true"
	} else {
		m[k] = "false"
	}
}

func (err *Err) sentryLevel() sentry.Level {
	switch err.Level {
	case LevelCritical:
		return sentry.LevelFatal
	case LevelError:
		return sentry.LevelError
	case LevelWarning:
		return sentry.LevelWarning
	default:
		panic("invalid level:" + err.Error())
	}
}

// sentryStacktrace converts our stack representation to Sentry's. We store it top
// down, but Sentry expects it bottom first.
func (err *Err) sentryStacktrace() *sentry.Stacktrace {
	sentryFrames := make([]sentry.Frame, 0, len(err.unknownFrames)+len(err.frames))

	for i := len(err.unknownFrames) - 1; i >= 0; i-- {
		sentryFrames = append(sentryFrames, frameToSentryFrame(err.unknownFrames[i]))
	}

	for i := len(err.frames) - 1; i >= 0; i-- {
		sentryFrames = append(sentryFrames, frameToSentryFrame(err.frames[i]))
	}

	return &sentry.Stacktrace{
		Frames: sentryFrames,
	}
}

func frameToSentryFrame(f *frame) sentry.Frame {
	sentryFrame := sentry.Frame{
		Function: f.fn,
		Package:  f.pkg,
		Filename: f.file,
		AbsPath:  f.path,
		Lineno:   f.line,
		InApp:    f.class == classApp || f.class == classPkg,
		Vars:     f.vars,
	}

	// if a custom message has been added to the frame, add it to the vars
	if f.msg != "" {
		// if the frame vars map hasn't been initialized, make a new one to avoid a panic
		if sentryFrame.Vars == nil {
			sentryFrame.Vars = make(map[string]interface{})
		}
		sentryFrame.Vars["_Msg"] = f.msg
	}

	return sentryFrame
}

// sentryExtra adds extra non-tagged data to an error
func (err *Err) sentryExtra() map[string]interface{} {
	m := map[string]interface{}{
		"stackDepth": len(err.frames),

		"MemStats.General.Alloc":       err.memStats.Alloc,
		"MemStats.General.TotalAlloc":  err.memStats.TotalAlloc,
		"MemStats.General.Sys":         err.memStats.Sys,
		"MemStats.General.Mallocs":     err.memStats.Mallocs,
		"MemStats.General.Frees":       err.memStats.Frees,
		"MemStats.General.LiveObjects": err.memStats.Mallocs - err.memStats.Frees,

		"MemStats.Heap.Alloc":    err.memStats.HeapAlloc,
		"MemStats.Heap.Sys":      err.memStats.HeapSys,
		"MemStats.Heap.Idle":     err.memStats.HeapIdle,
		"MemStats.Heap.Inuse":    err.memStats.HeapInuse,
		"MemStats.Heap.Released": err.memStats.HeapReleased,
		"MemStats.Heap.Objects":  err.memStats.HeapObjects,

		"MemStats.GC.NextGC":       err.memStats.NextGC,
		"MemStats.GC.LastGC":       err.memStats.LastGC,
		"MemStats.GC.PauseTotalNs": err.memStats.PauseTotalNs,
		"MemStats.GC.NumGC":        err.memStats.NumGC,
		"MemStats.GC.CPUFraction":  fmt.Sprintf("%.3f%%", err.memStats.GCCPUFraction*100),
	}

	for _, ev := range envVars {
		m["env."+ev.k] = ev.v
	}

	return m
}
