package factorial

// Factorial returns the factorial of n
// Factorial(0) = 1
// Factorial(1) = 1
// Factorial(2) = 2 * 1

func Factorial(number int) int {
	if number < 2 {
		return 1
	}
	return number * Factorial(number-1)
}
