package e

import (
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
)

func TestWrapGoroutine(t *testing.T) {
	// makes it easier to change test cases when line numbers change
	const testLine = 95

	tests := []struct {
		name   string
		opts   []WrapOption
		want   *Err
		output string
	}{
		{
			"test",
			[]WrapOption{With("k", "v"), NotRetriable(), Critical()},
			&Err{
				rootErr: errors.New("some error"),
				code:    codes.Unknown,
				frames: []*frame{
					{
						file:  "wrap_goroutine_test.go",
						path:  "github.com/nozzle/e",
						pkg:   "e",
						fn:    "wrapGoroutineHelper",
						full:  "github.com/nozzle/e.wrapGoroutineHelper",
						line:  testLine + 22,
						class: classVendor,
						vars: map[string]interface{}{
							"helper": "goroutine",
						},
					},
					{
						file:  "wrap_goroutine_test.go",
						path:  "github.com/nozzle/e",
						pkg:   "e",
						fn:    "TestWrapGoroutine.func1.1",
						full:  "github.com/nozzle/e.TestWrapGoroutine.func1.1",
						line:  testLine,
						class: classVendor,
						vars: map[string]interface{}{
							"k": "v",
						},
					},
				},
				topAppFrame: &frame{
					file:  "wrap_goroutine_test.go",
					path:  "github.com/nozzle/e",
					pkg:   "e",
					fn:    "wrapGoroutineHelper",
					full:  "github.com/nozzle/e.wrapGoroutineHelper",
					line:  testLine + 22,
					class: classVendor,
					vars: map[string]interface{}{
						"helper": "goroutine",
					},
				},
				isRetriable: false,
				Level:       LevelCritical,
			},
			`some error

Error Code: Unknown
https://github.com/nozzle/e/blob/main/wrap_goroutine_test.go#L117

--- NOT RETRIABLE ---

-------------------------------
e.wrapGoroutineHelper
 github.com/nozzle/e/wrap_goroutine_test.go:117
 + helper: "goroutine"
-------------------------------
e.TestWrapGoroutine.func1.1
 github.com/nozzle/e/wrap_goroutine_test.go:95
 + k: "v"
-------------------------------
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got *Err

			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				err := wrapGoroutineHelper()
				got = Wrap(err, tt.opts...)
				wg.Done()
			}()
			wg.Wait()

			// make the error comparable
			got.rawStack = nil
			got.memStats = nil
			got.currentFrameIdx = 0

			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.output, got.Error())
			assert.Equal(t, len(got.frames), len(tt.want.frames))
			for i := range got.frames {
				assert.Equal(t, tt.want.frames[i], got.frames[i])
			}
		})
	}
}

func wrapGoroutineHelper() error {
	return New("some error", With("helper", "goroutine"))
}
