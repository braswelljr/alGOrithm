package prime

// Prime - returns true if n is prime, false otherwise.
//
//	@param n - the number to check
//	@return true if n is prime, false otherwise
func Prime(number int) bool {
	// return false if n is less than 2
	if number < 2 {
		return false
	}

	// check if n is divisible by 2 or 3
	for i := 2; i*i <= number; i++ {
		// if n is divisible by 2 or 3, it is not prime
		if number%i == 0 {
			return false
		}
	}

	// otherwise, n is prime
	return true
}
