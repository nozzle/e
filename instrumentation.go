package e

import (
	"sync"

	"go.opencensus.io/tag"
)

const (
	maxErrorMutators   = 10
	maxErrorMutatorLen = 250
)

var (
	mu                = sync.Mutex{}
	errKeys           = make(map[string]tag.Mutator, maxErrorMutators)
	mutatorErrorOther = tag.Insert(tag.MustNewKey("error"), "other")
)

// MutatorFromError keeps a package global map of error strings, capped at
// maxErrorMutators (10) to prevent unbounded cardinality, and has one of three outcomes:
// 1. returns an already created mutator from the map
// 2. returns an "other" mutator if the map is full
// 3. returns a newly created mutator and stores it in the map
func MutatorFromError(err error) tag.Mutator {
	// get the root error text and truncate it if necessary
	reportErr := Cause(err).Error()
	if len(reportErr) > maxErrorMutatorLen {
		reportErr = reportErr[:maxErrorMutatorLen]
	}

	mu.Lock()
	defer mu.Unlock()

	// check for an existing mutator for this error text and return if found
	mut, ok := errKeys[reportErr]
	if ok {
		return mut
	}

	// if we have already reached max cardinality, return "other"
	if len(errKeys) > maxErrorMutators {
		return mutatorErrorOther
	}

	// create a new mutator, insert it into the map, then return it
	// mut = tag.Insert(nstats.KeyError, reportErr)
	errKeys[reportErr] = mut
	return mut
}
