package mccarthy91

import "testing"

func TestMcCarthy91(t *testing.T) {
	// Test #1
	str := "kenneth"
	lim := 16
	mc := McCarthy91(str, lim)
	t.Log("McCarthy91(\"kenneth\", 5) = ", mc)
}
