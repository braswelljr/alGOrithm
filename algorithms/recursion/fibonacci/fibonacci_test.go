package fibonacci

import "testing"

func TestFibonacci(t *testing.T) {
	// test fibonacci(0)
	x := 9
	y := Fibonacci(x)
	t.Log("The fibonacci sequence of the", x, "th term is", y)
}
