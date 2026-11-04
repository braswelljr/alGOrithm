// Package greedy holds greedy algorithms, one file each: activity selection,
// interval partitioning, Huffman coding, the fractional knapsack, job
// sequencing with deadlines, and greedy coin change. A greedy algorithm takes
// the choice that looks best right now and never revisits it; each function's
// documentation says why that is optimal for its problem, and the coin change
// shows a case where it is not.
//
// Each X has a TraceX that reports its choices to a trace.Recorder: Compare
// for each candidate weighed, Mark "chosen" (Values[0] the room, slot or
// percent where that applies) for each one taken.
//
//	acts, _ := greedy.ParseIntervals("a 1 4\nb 3 5\nc 5 7")
//	chosen := greedy.ActivitySelection(acts) // [0 2]
package greedy

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// ErrInput is returned for input a greedy algorithm cannot take.
var ErrInput = errors.New("greedy: invalid input")

// MaxItems bounds the inputs the parsers accept.
const MaxItems = 10_000

type rec struct{ r trace.Recorder }

func (x rec) compare(i ...int) { x.r.Record(trace.Event{Kind: trace.Compare, Indices: i}) }

func (x rec) chosen(i int, value ...int) {
	x.r.Record(trace.Event{Kind: trace.Mark, Indices: []int{i}, Values: value, Note: "chosen"})
}

var nop = rec{trace.Nop{}}

// parseLines reads lines of "name n1 n2 ...", fields separated by spaces,
// tabs or commas, with exactly n numbers after the name. Blank lines and
// lines starting with # are skipped, and a first line whose numbers are not
// numbers is taken as a header.
func parseLines(text string, n int, what string) (names []string, nums [][]int, err error) {
	for i, line := range strings.Split(text, "\n") {
		f := strings.FieldsFunc(line, func(r rune) bool { return r == ' ' || r == '\t' || r == ',' || r == '\r' })
		if len(f) == 0 || strings.HasPrefix(f[0], "#") {
			continue
		}
		if len(f) != n+1 {
			return nil, nil, fmt.Errorf("%w: line %d: want %s", ErrInput, i+1, what)
		}
		row := make([]int, n)
		bad := 0
		for j, s := range f[1:] {
			v, e := strconv.Atoi(s)
			if e != nil {
				bad++
			}
			row[j] = v
		}
		if bad == n && len(names) == 0 {
			continue // a header
		}
		if bad > 0 {
			return nil, nil, fmt.Errorf("%w: line %d: want %s, with whole numbers", ErrInput, i+1, what)
		}
		names, nums = append(names, f[0]), append(nums, row)
		if len(names) > MaxItems {
			return nil, nil, fmt.Errorf("%w: more than %d lines", ErrInput, MaxItems)
		}
	}
	if len(names) == 0 {
		return nil, nil, fmt.Errorf("%w: no lines of %s", ErrInput, what)
	}
	return names, nums, nil
}
