package recursion

// Palindrome reports whether s reads the same backward, byte by byte: a string
// of length 0 or 1 is one, and a longer one is when its ends match and its
// middle is.
//
// Time: O(n). Space: O(n) for the recursion.
func Palindrome(s string) bool {
	if len(s) < 2 {
		return true
	}
	return s[0] == s[len(s)-1] && Palindrome(s[1:len(s)-1])
}

// Permutations returns every ordering of s, by fixing each element first and
// permuting the rest (backtracking). The caller's slice is not modified.
//
// Time: O(n * n!). Space: O(n * n!) for the result.
func Permutations[T any](s []T) [][]T {
	work := append([]T(nil), s...)
	var out [][]T
	var permute func(k int)
	permute = func(k int) {
		if k == len(work) {
			out = append(out, append([]T(nil), work...))
			return
		}
		for i := k; i < len(work); i++ {
			work[k], work[i] = work[i], work[k]
			permute(k + 1)
			work[k], work[i] = work[i], work[k]
		}
	}
	permute(0)
	return out
}
