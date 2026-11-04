package dynamic

import "github.com/braswelljr/alGOrithm/pkg/trace"

// EditDistance returns the Levenshtein distance between a and b (Vladimir
// Levenshtein, 1965): the fewest insertions, deletions and substitutions that
// turn a into b, compared byte by byte. It also returns the table it filled: table[i][j] is the distance between a[:i] and
// b[:j].
//
// Time: O(len(a) * len(b)).
// Space: O(len(a) * len(b)) for the table.
func EditDistance(a, b string) (int, [][]int) { return editDistance(a, b, trace.Nop{}) }

// TraceEditDistance is EditDistance, reporting each cell written (Set, Indices[0] is
// i*(len(b)+1)+j) and the cells of one cheapest sequence of edits (Mark
// "path").
func TraceEditDistance(a, b string, r trace.Recorder) (int, [][]int) { return editDistance(a, b, r) }

func editDistance(a, b string, r trace.Recorder) (int, [][]int) {
	rows, cols := len(a)+1, len(b)+1
	t := make([][]int, rows)
	set := func(i, j, v int) {
		t[i][j] = v
		r.Record(trace.Event{Kind: trace.Set, Indices: []int{i*cols + j}, Values: []int{v}})
	}
	for i := range t {
		t[i] = make([]int, cols)
	}
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			switch {
			case i == 0:
				set(i, j, j) // insert j letters
			case j == 0:
				set(i, j, i) // delete i letters
			default:
				cost := 1
				if a[i-1] == b[j-1] {
					cost = 0
				}
				r.Record(trace.Event{Kind: trace.Compare, Indices: []int{(i-1)*cols + j, i*cols + j - 1, (i-1)*cols + j - 1}})
				set(i, j, min3(t[i-1][j]+1, t[i][j-1]+1, t[i-1][j-1]+cost))
			}
		}
	}
	r.Record(trace.Event{Kind: trace.Mark, Indices: editBacktrack(t, a, b), Note: "path"})
	return t[rows-1][cols-1], t
}

// editBacktrack walks from the last cell to the first along cells that explain
// each value, and returns their indexes.
func editBacktrack(t [][]int, a, b string) []int {
	cols := len(b) + 1
	i, j := len(a), len(b)
	path := []int{i*cols + j}
	for i > 0 || j > 0 {
		switch {
		case i > 0 && j > 0 && a[i-1] == b[j-1] && t[i][j] == t[i-1][j-1]:
			i, j = i-1, j-1
		case i > 0 && j > 0 && t[i][j] == t[i-1][j-1]+1:
			i, j = i-1, j-1
		case i > 0 && t[i][j] == t[i-1][j]+1:
			i--
		default:
			j--
		}
		path = append(path, i*cols+j)
	}
	return path
}

func min3(a, b, c int) int {
	if b < a {
		a = b
	}
	if c < a {
		a = c
	}
	return a
}
