package factorial

// Factorial function - is a function that multiplies a number by every number below it.
//
//	@param number - non-negative integer
//	@return - non-negative integer
//	@complexity - O(n)
//	@example - Factorial(5) = 5 * 4 * 3 * 2 * 1 = 120
//	@example - Factorial(0) = 1
func Factorial(number int) int {
	// return 1 as the factorial of 0 or 1
	if number < 2 {
		return 1
	}
	// Recur the function with parameter number - 1
	return number * Factorial(number-1)
}
