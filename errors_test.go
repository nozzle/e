package e

import (
	"context"
	"strings"
	"testing"
)

// Test_printErrorDetails will verify that an error with error details will output
// the details in the Error() output
func Test_printErrorDetails(t *testing.T) {
	const errorDetailOutput = `workspaceId: workspaceId 0 is invalid`
	err := NewErrInvalidArg(context.Background(), "workspaceId", "workspaceId 0 is invalid", "name", "test is not a valid name")

	// verify error string contains the error detail
	if !strings.Contains(err.Error(), errorDetailOutput) {
		t.Fatal("error didn't contain the error details")
	}
}
