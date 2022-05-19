package e

import (
	"errors"
	"testing"

	"github.com/maxatome/go-testdeep/td"
	"google.golang.org/grpc/codes"
)

func TestWrap(t *testing.T) {
	// makes it easier to change test cases when line numbers change
	const testLine = 139

	tests := []struct {
		name   string
		opts   []WrapOption
		want   *Err
		output string
	}{
		{
			"test",
			nil,
			&Err{
				code:    codes.Unknown,
				rootErr: errors.New("some error"),
				frames: []*frame{
					{
						file:  "wrap_test.go",
						path:  "github.com/nozzle/e",
						pkg:   "e",
						fn:    "wrapHelper4",
						full:  "github.com/nozzle/e.wrapHelper4",
						line:  testLine + 23 + 5,
						class: classVendor,
						msg:   "skip msg",
						vars: map[string]interface{}{
							"helper": "final",
						},
					},
					{
						file:  "wrap_test.go",
						path:  "github.com/nozzle/e",
						pkg:   "e",
						fn:    "wrapHelper3",
						full:  "github.com/nozzle/e.wrapHelper3",
						line:  testLine + 19 + 5,
						class: classVendor,
						msg:   "some message",
						vars: map[string]interface{}{
							"helper": 3,
						},
					},
					{
						file:  "wrap_test.go",
						path:  "github.com/nozzle/e",
						pkg:   "e",
						fn:    "wrapHelper2",
						full:  "github.com/nozzle/e.wrapHelper2",
						line:  testLine + 14 + 5,
						class: classVendor,
						vars: map[string]interface{}{
							"helper": 2,
						},
					},
					{
						file:  "wrap_test.go",
						path:  "github.com/nozzle/e",
						pkg:   "e",
						fn:    "wrapHelper1",
						full:  "github.com/nozzle/e.wrapHelper1",
						line:  testLine + 9 + 5,
						class: classVendor,
						vars: map[string]interface{}{
							"helper": 1,
						},
					},
					{
						file:  "wrap_test.go",
						path:  "github.com/nozzle/e",
						pkg:   "e",
						fn:    "TestWrap.func1",
						full:  "github.com/nozzle/e.TestWrap.func1",
						line:  testLine,
						class: classVendor,
					},
				},
				topAppFrame: &frame{
					file:  "wrap_test.go",
					path:  "github.com/nozzle/e",
					pkg:   "e",
					fn:    "wrapHelper4",
					full:  "github.com/nozzle/e.wrapHelper4",
					line:  testLine + 23 + 5,
					class: classVendor,
					msg:   "skip msg",
					vars: map[string]interface{}{
						"helper": "final",
					},
				},
				skipFrames:  1,
				isRetriable: false,
				Level:       LevelCritical,
			},
			`some error

Error Code: Unknown
https://github.com/nozzle/e/blob/main/wrap_test.go#L167

--- NOT RETRIABLE ---

-------------------------------
*** skip msg
e.wrapHelper4
 github.com/nozzle/e/wrap_test.go:167
 + helper: "final"
-------------------------------
*** some message
e.wrapHelper3
 github.com/nozzle/e/wrap_test.go:163
 + helper: 3
-------------------------------
e.wrapHelper2
 github.com/nozzle/e/wrap_test.go:158
 + helper: 2
-------------------------------
e.wrapHelper1
 github.com/nozzle/e/wrap_test.go:153
 + helper: 1
-------------------------------
e.TestWrap.func1
 github.com/nozzle/e/wrap_test.go:139
-------------------------------
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Wrap(wrapHelper1(), tt.opts...)

			// make the error comparable
			got.rawStack = nil
			got.memStats = nil
			got.currentFrameIdx = 0

			td.Cmp(t, got, tt.want)
			td.Cmp(t, got.Error(), tt.output)
		})
	}
}

func wrapHelper1() error {
	err := wrapHelper2()
	return Wrap(err, With("helper", 1), Critical())
}

func wrapHelper2() error {
	err := wrapHelper3()
	return Wrap(err, With("helper", 2), NotRetriable())
}

func wrapHelper3() error {
	return Wrap(wrapHelper4(), Msg("some message"), With("helper", 3), Critical())
}

func wrapHelper4() error {
	err := wrapHelperSkipped()
	return Wrap(err)
}

func wrapHelperSkipped() error {
	return New("some error", With("helper", "final"), Msg("skip msg"), SkipFrames(1))
}
