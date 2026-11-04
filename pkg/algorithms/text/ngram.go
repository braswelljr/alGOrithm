package text

// NGrams returns the character n-grams of s in order, overlapping: "abcd"
// with n = 2 gives ab, bc, cd. It returns nil when n < 1 or s is shorter
// than n.
//
// Time and space: O(n * len(s)).
func NGrams(s string, n int) []string {
	if n < 1 || len(s) < n {
		return nil
	}
	out := make([]string, 0, len(s)-n+1)
	for i := 0; i+n <= len(s); i++ {
		out = append(out, s[i:i+n])
	}
	return out
}

// Jaccard returns the Jaccard similarity of the sets of character n-grams of
// a and b, |A and B| / |A or B|, from 0 (nothing shared) to 1 (the same set),
// and the n-grams they share. Two texts with no n-grams at all are 1.
//
// Time and space: O(n * (len(a) + len(b))).
func Jaccard(a, b string, n int) (float64, []string) {
	set := func(s string) map[string]bool {
		m := map[string]bool{}
		for _, g := range NGrams(s, n) {
			m[g] = true
		}
		return m
	}
	sa, sb := set(a), set(b)
	if len(sa) == 0 && len(sb) == 0 {
		return 1, nil
	}
	var shared []string
	for _, g := range NGrams(a, n) {
		if sb[g] && sa[g] {
			shared = append(shared, g)
			sa[g] = false // count each shared n-gram once
		}
	}
	union := len(set(a)) + len(sb) - len(shared)
	return float64(len(shared)) / float64(union), shared
}
