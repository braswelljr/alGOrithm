package lcm

import "testing"

// TestLCM ...
func TestLCM(t *testing.T) {
	v := LCM(6, 8)
	if v != 24 {
		t.Error("Expected 24, got ", v)
	}
	t.Log("LCM(6, 8) = ", v, " passed")
}
