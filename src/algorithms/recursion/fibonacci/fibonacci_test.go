package fibonacci

import "testing"

func TestFibonacci(t *testing.T) {
	// test fibonacci(0)
	x := 0
	y := Fibonacci(x)
	if y != 0 {
		t.Error("Expected 0 for Fibonacci(0), got", y)
	}
}
