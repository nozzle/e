package e

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseStack(t *testing.T) {
	tests := []struct {
		name        string
		stack       []byte
		fullFnNames []string
		want        []*frame
	}{
		{
			"eliminate e.Wrap stack frames + line # without trailing spaces and args",
			[]byte(`goroutine 54 [running]:
github.com/nozzle/e.wrap(0x1422ac0, 0xc42008c5a0, 0xc42000e060, 0x1, 0x1, 0xc420001800)
	/Users/derek/go/src/github.com/nozzle/e/wrap.go:38 +0x11b
github.com/nozzle/e.Wrap(0x1422ac0, 0xc42008c5a0, 0xc42000e060, 0x1, 0x1, 0x39)
	/Users/derek/go/src/github.com/nozzle/e/wrap.go:19 +0x53
go.nozzle.io/views/billingvw.humanizeCommafDollars(0xc4201fc2d0, 0xc440c00000, 0x2, 0x2)
	/Users/derek/go/src/go.nozzle.io/views/billingvw/statements.go:75 +0xd3
go.nozzle.io/views/billingvw.viewStatementTable(0xc4201fc2d0, 0xc4204947d0, 0x0, 0x0)
	/Users/derek/go/src/go.nozzle.io/views/billingvw/statements.go:64 +0x332
go.nozzle.io/views/billingvw.ViewStatementSlackMessage(0xc4201fc2d0, 0xc4204947d0, 0xaa58df, 0xa, 0xaa3431, 0x7, 0x20)
	/Users/derek/go/src/go.nozzle.io/views/billingvw/statements.go:23 +0x3f
go.nozzle.io/namespaces/billing/billing/internal/billingserver.reconcileStatement(0xc4201fc2d0, 0xc4203e00c3, 0xa, 0x12, 0xc400000012, 0xaa7900, 0xd)
	/Users/derek/go/src/go.nozzle.io/namespaces/billing/billing/internal/billingserver/reconcile_statements.go:165 +0x131
go.nozzle.io/namespaces/billing/billing/internal/billingserver.handleReconcileStatement(0xc4201fc2d0, 0xde5a80, 0xc4202fb200, 0xc420153600, 0x0, 0xc42024b9b0, 0xc42024b9b0, 0x0, 0x0, 0x0)
	/Users/derek/go/src/go.nozzle.io/namespaces/billing/billing/internal/billingserver/reconcile_statements.go:76 +0x138
github.com/nozzle/rtr.route.func1(0xde5a80, 0xc4202fb200, 0xc420153600, 0x0)
	/Users/derek/go/src/github.com/nozzle/rtr/router.go:122 +0x914
go.nozzle.io/vendor/github.com/dimfeld/httptreemux.(*TreeMux).ServeLookupResult(0xc42019f360, 0xde5a80, 0xc4202fb200, 0xc420153600, 0xc8, 0xc4200e70f0, 0x0, 0x0)
	/Users/derek/go/src/go.nozzle.io/vendor/github.com/dimfeld/httptreemux/router.go:245 +0x133`),
			[]string{
				"go.nozzle.io/views/billingvw.humanizeCommafDollars",
				"go.nozzle.io/views/billingvw.viewStatementTable",
				"go.nozzle.io/views/billingvw.ViewStatementSlackMessage",
				"go.nozzle.io/namespaces/billing/billing/internal/billingserver.reconcileStatement",
				"go.nozzle.io/namespaces/billing/billing/internal/billingserver.handleReconcileStatement",
				"github.com/nozzle/rtr.route.func1",
				"go.nozzle.io/vendor/github.com/dimfeld/httptreemux.(*TreeMux).ServeLookupResult",
			},
			stackframes{
				{
					file:  "statements.go",
					path:  "go.nozzle.io/views/billingvw",
					pkg:   "billingvw",
					fn:    "humanizeCommafDollars",
					full:  "go.nozzle.io/views/billingvw.humanizeCommafDollars",
					line:  75,
					class: classApp,
				},
				{
					file:  "statements.go",
					path:  "go.nozzle.io/views/billingvw",
					pkg:   "billingvw",
					fn:    "viewStatementTable",
					full:  "go.nozzle.io/views/billingvw.viewStatementTable",
					line:  64,
					class: classApp,
				},
				{
					file:  "statements.go",
					path:  "go.nozzle.io/views/billingvw",
					pkg:   "billingvw",
					fn:    "ViewStatementSlackMessage",
					full:  "go.nozzle.io/views/billingvw.ViewStatementSlackMessage",
					line:  23,
					class: classApp,
				},
				{
					file:  "reconcile_statements.go",
					path:  "go.nozzle.io/namespaces/billing/billing/internal/billingserver",
					pkg:   "billingserver",
					fn:    "reconcileStatement",
					full:  "go.nozzle.io/namespaces/billing/billing/internal/billingserver.reconcileStatement",
					line:  165,
					class: classApp,
				},
				{
					file:  "reconcile_statements.go",
					path:  "go.nozzle.io/namespaces/billing/billing/internal/billingserver",
					pkg:   "billingserver",
					fn:    "handleReconcileStatement",
					full:  "go.nozzle.io/namespaces/billing/billing/internal/billingserver.handleReconcileStatement",
					line:  76,
					class: classApp,
				},
				{
					file:  "router.go",
					path:  "github.com/nozzle/rtr",
					pkg:   "rtr",
					fn:    "route.func1",
					full:  "github.com/nozzle/rtr.route.func1",
					line:  122,
					class: classVendor,
				},
				{
					file:  "router.go",
					path:  "go.nozzle.io/vendor/github.com/dimfeld/httptreemux",
					pkg:   "httptreemux",
					fn:    "(*TreeMux).ServeLookupResult",
					full:  "go.nozzle.io/vendor/github.com/dimfeld/httptreemux.(*TreeMux).ServeLookupResult",
					line:  245,
					class: classVendor,
				},
			},
		},
		{
			"panic stack trace",
			[]byte(`goroutine 2069872 [running]:
runtime/debug.Stack(0xe3c200, 0xe31920, 0x1566390)
	/usr/local/Cellar/go/1.9.2/libexec/src/runtime/debug/stack.go:24 +0xa7
github.com/nozzle/rtr.route.func1.1(0xc4211e45a0, 0x151dc40, 0xc432a3ec00)
	/Users/derek/go/src/github.com/nozzle/rtr/router.go:84 +0xb7
panic(0xe31920, 0x1566390)
	/usr/local/Cellar/go/1.9.2/libexec/src/runtime/panic.go:491 +0x283
github.com/nozzle/e.(*reportData).report(0xc422c6d3e0, 0xc4211e45a0)
	/Users/derek/go/src/github.com/nozzle/e/report.go:114 +0xcbc
github.com/nozzle/e.middleware(0xc4211e45a0, 0xc4229a0e00, 0x1513880, 0xc42a9c7fb0, 0x0, 0xc42494ef00)
	/Users/derek/go/src/github.com/nozzle/e/handler.go:23
github.com/nozzle/rtr.route.func1(0x151dc40, 0xc432a3ec00, 0xc4229a0e00, 0x0)
	/Users/derek/go/src/github.com/nozzle/rtr/router.go:124 +0x97b
go.nozzle.io/vendor/github.com/dimfeld/httptreemux.(*TreeMux).ServeLookupResult(0xc420182dc0, 0x151dc40, 0xc432a3ec00, 0xc4229a0e00, 0xc8, 0xc4203ae0a0, 0x0, 0x0)
	/Users/derek/go/src/go.nozzle.io/vendor/github.com/dimfeld/httptreemux/router.go:245 +0x133
go.nozzle.io/vendor/github.com/dimfeld/httptreemux.(*TreeMux).ServeHTTP(0xc420182dc0, 0x151dc40, 0xc432a3ec00, 0xc4229a0e00)
	/Users/derek/go/src/go.nozzle.io/vendor/github.com/dimfeld/httptreemux/router.go:266 +0xdb
net/http.(*ServeMux).ServeHTTP(0x1592040, 0x151dc40, 0xc432a3ec00, 0xc4229a0e00)
	/usr/local/Cellar/go/1.9.2/libexec/src/net/http/server.go:2254 +0x130
go.nozzle.io/vendor/google.golang.org/appengine/internal.executeRequestSafely(0xc432a3ec00, 0xc4229a0e00)
	/Users/derek/go/src/go.nozzle.io/vendor/google.golang.org/appengine/internal/api.go:156 +0x77
go.nozzle.io/vendor/google.golang.org/appengine/internal.handleHTTP(0x15208c0, 0xc4210ac000, 0xc4229a0e00)
	/Users/derek/go/src/go.nozzle.io/vendor/google.golang.org/appengine/internal/api.go:124 +0x26a
net/http.HandlerFunc.ServeHTTP(0x1016118, 0x15208c0, 0xc4210ac000, 0xc4229a0e00)
	/usr/local/Cellar/go/1.9.2/libexec/src/net/http/server.go:1918 +0x44
net/http.serverHandler.ServeHTTP(0xc42007c5b0, 0x15208c0, 0xc4210ac000, 0xc4229a0e00)
	/usr/local/Cellar/go/1.9.2/libexec/src/net/http/server.go:2619 +0xb4
net/http.(*conn).serve(0xc4211afe00, 0x15216c0, 0xc4338d0940)
	/usr/local/Cellar/go/1.9.2/libexec/src/net/http/server.go:1801 +0x71d
created by net/http.(*Server).Serve
	/usr/local/Cellar/go/1.9.2/libexec/src/net/http/server.go:2720 +0x288`),
			[]string{
				"github.com/nozzle/rtr.route.func1.1",
				"panic",
				"github.com/nozzle/e.(*reportData).report",
				"github.com/nozzle/e.middleware",
				"github.com/nozzle/rtr.route.func1",
				"go.nozzle.io/vendor/github.com/dimfeld/httptreemux.(*TreeMux).ServeLookupResult",
				"go.nozzle.io/vendor/github.com/dimfeld/httptreemux.(*TreeMux).ServeHTTP",
				"net/http.(*ServeMux).ServeHTTP",
				"go.nozzle.io/vendor/google.golang.org/appengine/internal.executeRequestSafely",
				"go.nozzle.io/vendor/google.golang.org/appengine/internal.handleHTTP",
				"net/http.HandlerFunc.ServeHTTP",
				"net/http.serverHandler.ServeHTTP",
				"net/http.(*conn).serve",
				"net/http.(*Server).Serve",
			},
			stackframes{
				{
					file:  "router.go",
					path:  "github.com/nozzle/rtr",
					pkg:   "rtr",
					fn:    "route.func1.1",
					full:  "github.com/nozzle/rtr.route.func1.1",
					line:  84,
					class: classVendor,
				},
				{
					file:  "panic.go",
					path:  "",
					pkg:   "",
					fn:    "panic",
					full:  "panic",
					line:  491,
					class: classPanic,
				},
				{
					file:  "report.go",
					path:  "github.com/nozzle/e",
					pkg:   "e",
					fn:    "(*reportData).report",
					full:  "github.com/nozzle/e.(*reportData).report",
					line:  114,
					class: classVendor,
				},
				{
					file:  "handler.go",
					path:  "github.com/nozzle/e",
					pkg:   "e",
					fn:    "middleware",
					full:  "github.com/nozzle/e.middleware",
					line:  23,
					class: classVendor,
				},
				{
					file:  "router.go",
					path:  "github.com/nozzle/rtr",
					pkg:   "rtr",
					fn:    "route.func1",
					full:  "github.com/nozzle/rtr.route.func1",
					line:  124,
					class: classVendor,
				},
				{
					file:  "router.go",
					path:  "go.nozzle.io/vendor/github.com/dimfeld/httptreemux",
					pkg:   "httptreemux",
					fn:    "(*TreeMux).ServeLookupResult",
					full:  "go.nozzle.io/vendor/github.com/dimfeld/httptreemux.(*TreeMux).ServeLookupResult",
					line:  245,
					class: classVendor,
				},
				{
					file:  "router.go",
					path:  "go.nozzle.io/vendor/github.com/dimfeld/httptreemux",
					pkg:   "httptreemux",
					fn:    "(*TreeMux).ServeHTTP",
					full:  "go.nozzle.io/vendor/github.com/dimfeld/httptreemux.(*TreeMux).ServeHTTP",
					line:  266,
					class: classVendor,
				},
				{
					file:  "server.go",
					path:  "net/http",
					pkg:   "http",
					fn:    "(*ServeMux).ServeHTTP",
					full:  "net/http.(*ServeMux).ServeHTTP",
					line:  2254,
					class: classStdLib,
				},
				{
					file:  "api.go",
					path:  "go.nozzle.io/vendor/google.golang.org/appengine/internal",
					pkg:   "internal",
					fn:    "executeRequestSafely",
					full:  "go.nozzle.io/vendor/google.golang.org/appengine/internal.executeRequestSafely",
					line:  156,
					class: classVendor,
				},
				{
					file:  "api.go",
					path:  "go.nozzle.io/vendor/google.golang.org/appengine/internal",
					pkg:   "internal",
					fn:    "handleHTTP",
					full:  "go.nozzle.io/vendor/google.golang.org/appengine/internal.handleHTTP",
					line:  124,
					class: classVendor,
				},
				{
					file:  "server.go",
					path:  "net/http",
					pkg:   "http",
					fn:    "HandlerFunc.ServeHTTP",
					full:  "net/http.HandlerFunc.ServeHTTP",
					line:  1918,
					class: classStdLib,
				},
				{
					file:  "server.go",
					path:  "net/http",
					pkg:   "http",
					fn:    "serverHandler.ServeHTTP",
					full:  "net/http.serverHandler.ServeHTTP",
					line:  2619,
					class: classStdLib,
				},
				{
					file:  "server.go",
					path:  "net/http",
					pkg:   "http",
					fn:    "(*conn).serve",
					full:  "net/http.(*conn).serve",
					line:  1801,
					class: classStdLib,
				},
			},
		},
		{
			"panic inside test",
			[]byte(`goroutine 6 [running]:
testing.tRunner.func1(0xc42004c3c0)
	/usr/local/Cellar/go/1.9.2/libexec/src/testing/testing.go:711 +0x2d2
panic(0x15420c0, 0xc4204dec20)
	/usr/local/Cellar/go/1.9.2/libexec/src/runtime/panic.go:491 +0x283
github.com/nozzle/e.(*Err).findFrame(0xc4203a8640, 0xc4202e4000)
	/Users/derek/go/src/github.com/nozzle/e/errors.go:200 +0x70
github.com/nozzle/e.wrap(0x197e420, 0xc4204debc0, 0x0, 0x0, 0x0, 0xc4204bc780)
	/Users/derek/go/src/github.com/nozzle/e/wrap.go:54 +0x6c
github.com/nozzle/e.Wrap(0x197e420, 0xc4204debc0, 0x0, 0x0, 0x0, 0x39)
	/Users/derek/go/src/github.com/nozzle/e/wrap.go:22 +0x58
github.com/nozzle/e.TestWrap.func1(0xc42004c3c0)
	/Users/derek/go/src/github.com/nozzle/e/errors_test.go:39 +0x69
testing.tRunner(0xc42004c3c0, 0xc4204debe0)
	/usr/local/Cellar/go/1.9.2/libexec/src/testing/testing.go:746 +0xd0
created by testing.(*T).Run
	/usr/local/Cellar/go/1.9.2/libexec/src/testing/testing.go:789 +0x2de`),
			[]string{
				"testing.tRunner.func1",
				"panic",
				"github.com/nozzle/e.(*Err).findFrame",
				"github.com/nozzle/e.wrap",
				"github.com/nozzle/e.Wrap",
				"github.com/nozzle/e.TestWrap.func1",
			},
			stackframes{
				{
					file:  "testing.go",
					path:  "testing",
					pkg:   "testing",
					fn:    "tRunner.func1",
					full:  "testing.tRunner.func1",
					line:  711,
					class: classStdLib,
				},
				{
					file:  "panic.go",
					path:  "",
					pkg:   "",
					fn:    "panic",
					full:  "panic",
					line:  491,
					class: classPanic,
				},
				{
					file:  "errors.go",
					path:  "github.com/nozzle/e",
					pkg:   "e",
					fn:    "(*Err).findFrame",
					full:  "github.com/nozzle/e.(*Err).findFrame",
					line:  200,
					class: classVendor,
				},
				{
					file:  "wrap.go",
					path:  "github.com/nozzle/e",
					pkg:   "e",
					fn:    "wrap",
					full:  "github.com/nozzle/e.wrap",
					line:  54,
					class: classVendor,
				},
				{
					file:  "wrap.go",
					path:  "github.com/nozzle/e",
					pkg:   "e",
					fn:    "Wrap",
					full:  "github.com/nozzle/e.Wrap",
					line:  22,
					class: classVendor,
				},
				{
					file:  "errors_test.go",
					path:  "github.com/nozzle/e",
					pkg:   "e",
					fn:    "TestWrap.func1",
					full:  "github.com/nozzle/e.TestWrap.func1",
					line:  39,
					class: classVendor,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseStack(tt.stack)

			// set up an error to check the associate findFrame function
			err := &Err{frames: tt.want}

			// make sure the parsed stack is the same length as the expected stack
			assert.Equal(t, len(tt.want), len(got))
			for i := range got {
				// make sure that the parsed frames match the expected frames
				assert.Equal(t, tt.want[i], got[i])

				// make sure that findFrame works in tandem with parseFrames
				currentFrameIdx := err.findFrame(tt.fullFnNames[i])
				assert.Equal(t, tt.want[i], got[currentFrameIdx])
			}
		})
	}
}
