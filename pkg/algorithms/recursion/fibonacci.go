package recursion

// Fibonacci returns the nth Fibonacci number, F(0) = 0, F(1) = 1, F(n) =
// F(n-1) + F(n-2), by plain recursion: the textbook example of exponential
// blow-up, because F(n-2) is computed again inside F(n-1). For negative n it
// returns -F(-n).
//
// Time: O(phi^n), about O(1.618^n). Space: O(n) for the recursion.
func Fibonacci(n int) int {
	if n < 0 {
		return -Fibonacci(-n)
	}
	if n < 2 {
		return n
	}
	return Fibonacci(n-1) + Fibonacci(n-2)
}

// FibonacciMemo is Fibonacci with memoization: each F(k) is computed once and
// remembered, which turns the same recursion linear.
//
// Time: O(n). Space: O(n).
func FibonacciMemo(n int) int {
	if n < 0 {
		return -FibonacciMemo(-n)
	}
	memo := map[int]int{0: 0, 1: 1}
	var f func(int) int
	f = func(k int) int {
		if v, ok := memo[k]; ok {
			return v
		}
		memo[k] = f(k-1) + f(k-2)
		return memo[k]
	}
	return f(n)
}

// FibonacciIterative computes the same number bottom-up with two variables,
// the dynamic-programming form of the recursion.
//
// Time: O(n). Space: O(1).
func FibonacciIterative(n int) int {
	if n < 0 {
		return -FibonacciIterative(-n)
	}
	a, b := 0, 1
	for i := 0; i < n; i++ {
		a, b = b, a+b
	}
	return a
}
