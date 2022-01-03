package stack

import "testing"

func TestStackInit(t *testing.T) {
	s := New()
	if s.Size() != 0 {
		t.Errorf("Expected size to be 0, but got %d", s.Size())
	}
}
