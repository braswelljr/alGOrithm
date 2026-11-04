package strmatch

// Anagram reports whether a and b use exactly the same letters the same
// number of times, ignoring case and anything that is not a letter or digit
// ("Dormitory" and "dirty room").
//
// Time: O(n). Space: O(1) (a count per byte value).
func Anagram(a, b string) bool {
	var count [256]int
	add := func(s string, d int) {
		for i := 0; i < len(s); i++ {
			c := s[i]
			if c >= 'A' && c <= 'Z' {
				c += 'a' - 'A'
			}
			if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c >= 0x80 {
				count[c] += d
			}
		}
	}
	add(a, 1)
	add(b, -1)
	for _, n := range count {
		if n != 0 {
			return false
		}
	}
	return true
}
