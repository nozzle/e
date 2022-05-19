package e

import (
	"errors"
	"testing"

	"google.golang.org/grpc/codes"
)

//nolint:errcheck
func BenchmarkInitialWrap(b *testing.B) {
	b.ReportAllocs()
	err := errors.New("my special error")
	for n := 0; n < b.N; n++ {
		Wrap(err)
	}
}

//nolint:errcheck
func BenchmarkAlreadyWrappedNoOpts(b *testing.B) {
	b.ReportAllocs()
	err := New("my special error")
	for n := 0; n < b.N; n++ {
		Wrap(err)
	}
}

//nolint:errcheck
func BenchmarkAlreadyWrappedWithVars(b *testing.B) {
	b.ReportAllocs()
	err := New("my special error")
	for n := 0; n < b.N; n++ {
		Wrap(err, With("foo", "bar"), Critical(), Code(codes.FailedPrecondition))
	}
}
