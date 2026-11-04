package recursion

// Ackermann returns the Ackermann function A(m, n): A(0, n) = n + 1,
// A(m, 0) = A(m-1, 1), A(m, n) = A(m-1, A(m, n-1)). It is total and computable
// but not primitive recursive: it grows faster than any function a fixed
// number of nested loops can compute. A(4, 2) already has 19729 digits, so
// keep m <= 3.
//
// Time: grows faster than exponentially. Space: O(A(m, n)) recursion depth.
func Ackermann(m, n int) int {
	switch {
	case m == 0:
		return n + 1
	case n == 0:
		return Ackermann(m-1, 1)
	}
	return Ackermann(m-1, Ackermann(m, n-1))
}

// McCarthy91 returns John McCarthy's 91 function: n - 10 for n > 100, and
// M(M(n + 11)) otherwise, which is 91 for every n <= 101 (a puzzle for
// proving recursive programs correct).
//
// Time: O(101 - n) for n <= 100. Space: the same, for the recursion.
func McCarthy91(n int) int {
	if n > 100 {
		return n - 10
	}
	return McCarthy91(McCarthy91(n + 11))
}
