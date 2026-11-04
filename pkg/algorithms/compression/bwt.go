package compression

import "sort"

// BurrowsWheeler returns the Burrows-Wheeler transform of s (Michael Burrows
// and David Wheeler, 1994): sort every rotation of s, and take the last
// character of each in order, plus the row where s itself landed. Equal
// contexts sort together, so the last column groups repeated characters
// ("banana" becomes "nnbaaa"), which run-length and move-to-front coding then
// compress well; bzip2 is built on it. The transform is reversible given the
// index. It returns "", -1 for an empty s.
//
// Time: O(n^2 log n) with this direct rotation sort; suffix arrays make it
// O(n). Space: O(n).
func BurrowsWheeler(s string) (last string, index int) {
	if s == "" {
		return "", -1
	}
	rows := rotationOrder(s)
	out := make([]byte, len(s))
	for i, r := range rows {
		out[i] = s[(r+len(s)-1)%len(s)]
		if r == 0 {
			index = i
		}
	}
	return string(out), index
}

// Rotations returns the rotations of s in sorted order, the table
// BurrowsWheeler reads its last column from.
//
// Time: O(n^2 log n). Space: O(n^2).
func Rotations(s string) []string {
	rows := rotationOrder(s)
	out := make([]string, len(rows))
	for i, r := range rows {
		out[i] = s[r:] + s[:r]
	}
	return out
}

// rotationOrder returns the start offsets of the rotations of s, sorted.
func rotationOrder(s string) []int {
	n := len(s)
	rows := make([]int, n)
	for i := range rows {
		rows[i] = i
	}
	sort.SliceStable(rows, func(a, b int) bool {
		for k := 0; k < n; k++ {
			x, y := s[(rows[a]+k)%n], s[(rows[b]+k)%n]
			if x != y {
				return x < y
			}
		}
		return false
	})
	return rows
}

// InverseBurrowsWheeler rebuilds the string from the last column and index
// with the last-to-first mapping: the kth copy of a character in the last
// column is the kth copy of it in the first column (the sorted characters),
// so following that mapping walks the string backward. It returns ErrCorrupt
// when index is out of range.
//
// Time: O(n log n) for the sort. Space: O(n).
func InverseBurrowsWheeler(last string, index int) (string, error) {
	n := len(last)
	if n == 0 && index == -1 {
		return "", nil
	}
	if index < 0 || index >= n {
		return "", ErrCorrupt
	}
	// first[i] is the row of the first column holding the same character
	// copy as last[i]: a stable sort of positions by character.
	order := make([]int, n)
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(a, b int) bool { return last[order[a]] < last[order[b]] })
	lf := make([]int, n) // last-column position to first-column row
	for row, pos := range order {
		lf[pos] = row
	}
	out := make([]byte, n)
	row := index
	for k := n - 1; k >= 0; k-- {
		out[k] = last[row]
		row = lf[row]
	}
	return string(out), nil
}
