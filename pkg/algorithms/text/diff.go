package text

// Edit is one line of a diff: Op is '=' (in both), '-' (only in the old
// lines) or '+' (only in the new lines).
type Edit struct {
	Op   byte   `json:"op"`
	Line string `json:"line"`
}

// Diff returns the shortest edit script turning the lines a into the lines
// b, by Myers' O(ND) algorithm (1986): for d = 0, 1, ... find how far each
// diagonal k = x - y reaches with d edits, following runs of equal lines for
// free, until one reaches the end; then walk back through the saved
// frontiers. Keeping the '=' and '-' lines gives a, keeping '=' and '+' gives
// b, and the number of '-' and '+' lines is the smallest possible.
//
// Time: O((n + m) * d) for d edits. Space: O((n + m) * d) for the frontiers.
func Diff(a, b []string) []Edit {
	n, m := len(a), len(b)
	limit := n + m
	off := limit + 1
	v := make([]int, 2*limit+3)
	var frontiers [][]int
	for d := 0; d <= limit; d++ {
		frontiers = append(frontiers, append([]int(nil), v...))
		for k := -d; k <= d; k += 2 {
			var x int
			if k == -d || (k != d && v[off+k-1] < v[off+k+1]) {
				x = v[off+k+1] // down: an insertion
			} else {
				x = v[off+k-1] + 1 // right: a deletion
			}
			y := x - k
			for x < n && y < m && a[x] == b[y] {
				x, y = x+1, y+1
			}
			v[off+k] = x
			if x >= n && y >= m {
				return backtrack(a, b, frontiers, off)
			}
		}
	}
	return nil // not reached: d = n + m always ends
}

func backtrack(a, b []string, frontiers [][]int, off int) []Edit {
	x, y := len(a), len(b)
	var rev []Edit
	for d := len(frontiers) - 1; d >= 0; d-- {
		v := frontiers[d]
		k := x - y
		prevK := k - 1
		if k == -d || (k != d && v[off+k-1] < v[off+k+1]) {
			prevK = k + 1
		}
		prevX := v[off+prevK]
		prevY := prevX - prevK
		for x > prevX && y > prevY {
			rev = append(rev, Edit{'=', a[x-1]})
			x, y = x-1, y-1
		}
		if d > 0 {
			if x == prevX {
				rev = append(rev, Edit{'+', b[y-1]})
			} else {
				rev = append(rev, Edit{'-', a[x-1]})
			}
		}
		x, y = prevX, prevY
	}
	for i, j := 0, len(rev)-1; i < j; i, j = i+1, j-1 {
		rev[i], rev[j] = rev[j], rev[i]
	}
	return rev
}
