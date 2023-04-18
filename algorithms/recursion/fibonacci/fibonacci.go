package fibonacci

// Fibonacci function - returns the nth number in the Fibonacci sequence, The first two numbers in the sequence are 0 and 1.
//
//	@param n - non-negative integer
//	@return - non-negative integer
//	@complexity - O(2^n)
//	@example - Fibonacci(0) = 0
//	@example - Fibonacci(1) = 1
//	@example - Fibonacci(2) = 1
//	@example - Fibonacci(3) = 2
func Fibonacci(n int) int {
	// check if n is negative
	if n < 0 {
		// return the negative of the absolute value of n
		return Fibonacci(n*-1) * -1 // for negative numbers
	} else if n == 0 || n == 1 { // check if n is 0 or 1
		return n // base case
	}

	// Recur the function with parameters n - 1 and n - 2
	return Fibonacci(n-1) + Fibonacci(n-2)
}
