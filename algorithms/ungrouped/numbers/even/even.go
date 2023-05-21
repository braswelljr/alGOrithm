package even

// Even - returns true if n is even, false otherwise.
//
//	@param n - the number to check
//	@return true if n is even, false otherwise
func Even(number int) bool {
	// return true if n is divisible by 2
	return number%2 == 0
}
