package e

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRecover(t *testing.T) {
	tests := []struct {
		name        string
		opts        []RecoverOption
		rawErr      error
		shouldPanic bool
	}{
		{
			"recover panic",
			[]RecoverOption{RecoverNoLog(), RecoverNoReport(), RecoverNoPanic()},
			errors.New("I'm an error"),
			false,
		},
		{
			"propagate panic",
			[]RecoverOption{RecoverNoLog(), RecoverNoReport()},
			errors.New("I'm an error"),
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFn := func() {
				defer Recover(context.Background(), tt.opts...)
				panic(tt.rawErr)
			}

			if tt.shouldPanic {
				assert.Panics(t, testFn)
			} else {
				assert.NotPanics(t, testFn)
			}
		})
	}
}
