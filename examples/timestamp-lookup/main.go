// Command timestamp-lookup finds the log entries in a time window with
// binary search on sorted timestamps: two O(log n) searches bound the
// window, the way time-series databases and log viewers seek.
package main

import (
	"fmt"

	"github.com/braswelljr/alGOrithm/pkg/algorithms/searching"
)

func main() {
	// Unix seconds of log lines, already sorted as logs are appended in order.
	ts := []int{1700000000, 1700000004, 1700000010, 1700000031, 1700000045, 1700000060, 1700000062, 1700000090, 1700000120}
	from, to := 1700000030, 1700000065
	lo := searching.LowerBound(ts, from) // first >= from
	hi := searching.UpperBound(ts, to)   // first > to
	fmt.Printf("%d entries between %d and %d: %v\n", hi-lo, from, to, ts[lo:hi])
	fmt.Printf("exact match for 1700000045 at index %d\n", searching.Binary(ts, 1700000045))
}
