package ackermann

// Ackermann function - is a classic example of a total computable function that is not primitive recursive.
//
//	@param m - non-negative integer
//	@param n - non-negative integer
//	@return - non-negative integer
func Ackermann(m, n int) int {

	// if m == 0 then return n + 1
	if m == 0 {
		return n + 1
	}

	// if n == 0 then return Ackermann(m - 1, 1)
	if n == 0 {
		return Ackermann(m-1, 1)
	}

	return Ackermann(m-1, Ackermann(m, n-1))
}
