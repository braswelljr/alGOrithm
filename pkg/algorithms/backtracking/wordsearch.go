package backtracking

import (
	"fmt"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// Board is a grid of letters, row by row.
type Board [][]byte

// ParseBoard reads rows of letters, one row a line (spaces ignored), all the
// same length.
func ParseBoard(text string) (Board, error) {
	var b Board
	for _, line := range strings.Split(text, "\n") {
		row := strings.ReplaceAll(strings.TrimSpace(line), " ", "")
		if row == "" {
			continue
		}
		if len(b) > 0 && len(row) != len(b[0]) {
			return nil, fmt.Errorf("%w: every row needs %d letters; row %d has %d", ErrInput, len(b[0]), len(b)+1, len(row))
		}
		b = append(b, []byte(row))
	}
	if len(b) == 0 || len(b)*len(b[0]) > 400 {
		return nil, fmt.Errorf("%w: give 1 to 400 letters", ErrInput)
	}
	return b, nil
}

// WordSearch finds word on the board as a path of letters, each next to the
// previous one (up, down, left or right), no cell used twice: start from each
// cell matching the first letter, step to neighbors matching the next, and
// back up at a dead end. It returns the cells of the path, row*cols+col, and
// whether the word is there.
//
// Time: O(cells * 4^len(word)) in the worst case. Space: O(len(word)).
func WordSearch(b Board, word string) ([]int, bool) { return wordSearch(b, word, nop) }

// TraceWordSearch is WordSearch, reporting each cell tried (Visit) and the
// path so far after every step (Mark "path").
func TraceWordSearch(b Board, word string, r trace.Recorder) ([]int, bool) {
	return wordSearch(b, word, rec{r})
}

func wordSearch(b Board, word string, x rec) ([]int, bool) {
	if len(b) == 0 || word == "" {
		return nil, false
	}
	rows, cols := len(b), len(b[0])
	used := make([]bool, rows*cols)
	var path []int
	var walk func(r, c, k int) bool
	walk = func(r, c, k int) bool {
		if r < 0 || r >= rows || c < 0 || c >= cols || used[r*cols+c] {
			return false
		}
		x.r.Record(trace.Event{Kind: trace.Visit, Indices: []int{r*cols + c}})
		if b[r][c] != word[k] {
			return false
		}
		used[r*cols+c] = true
		path = append(path, r*cols+c)
		x.mark("path", path...)
		if k == len(word)-1 {
			return true
		}
		for _, d := range [4][2]int{{-1, 0}, {0, 1}, {1, 0}, {0, -1}} {
			if walk(r+d[0], c+d[1], k+1) {
				return true
			}
		}
		used[r*cols+c] = false
		path = path[:len(path)-1]
		x.mark("path", path...)
		return false
	}
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			if walk(r, c, 0) {
				return path, true
			}
		}
	}
	return nil, false
}
