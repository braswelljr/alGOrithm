package hashing_test

import (
	"testing"

	"github.com/braswelljr/alGOrithm/pkg/security/hashing"
)

// TestRobinHoodLen checks that Len counts distinct keys: a new key adds one, a
// key already there adds none.
func TestRobinHoodLen(t *testing.T) {
	r := hashing.NewRobinHood(16)
	if r.Len() != 0 {
		t.Fatalf("a new table holds %d keys", r.Len())
	}
	for _, k := range []int{3, 19, 35, 3, 7} {
		r.Insert(k)
	}
	if r.Len() != 4 {
		t.Errorf("Len() = %d after inserting 3, 19, 35, 3, 7; want 4", r.Len())
	}
}
