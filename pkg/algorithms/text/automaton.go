package text

// DistinctSubstrings returns how many different non-empty substrings s has,
// and the number of states of its suffix automaton, the smallest automaton
// that accepts exactly the suffixes of s. Each state stands for a set of
// substrings whose lengths run from len(link)+1 to len, so the count is the
// sum of len - len(link) over the states.
//
// Time: O(n) amortized (O(n log sigma) with map transitions). Space: O(n).
func DistinctSubstrings(s string) (count, states int) {
	type state struct {
		length, link int
		next         map[byte]int
	}
	st := []state{{link: -1, next: map[byte]int{}}}
	last := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		cur := len(st)
		st = append(st, state{length: st[last].length + 1, next: map[byte]int{}})
		p := last
		for p != -1 {
			if _, ok := st[p].next[c]; ok {
				break
			}
			st[p].next[c] = cur
			p = st[p].link
		}
		switch {
		case p == -1:
			st[cur].link = 0
		case st[st[p].next[c]].length == st[p].length+1:
			st[cur].link = st[p].next[c]
		default:
			q := st[p].next[c]
			clone := len(st)
			next := make(map[byte]int, len(st[q].next))
			for k, v := range st[q].next {
				next[k] = v
			}
			st = append(st, state{length: st[p].length + 1, link: st[q].link, next: next})
			for p != -1 && st[p].next[c] == q {
				st[p].next[c] = clone
				p = st[p].link
			}
			st[q].link, st[cur].link = clone, clone
		}
		last = cur
	}
	for v := 1; v < len(st); v++ {
		count += st[v].length - st[st[v].link].length
	}
	return count, len(st)
}
