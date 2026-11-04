// Command sort-and-trace sorts a few values with quick sort, then sorts them
// again while recording every step to count comparisons and swaps.
package main

import (
	"fmt"

	"github.com/braswelljr/alGOrithm/pkg/algorithms/sorting"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

func main() {
	values := []int{5, 3, 9, 1, 7, 2}
	fmt.Println("sorted:", sorting.Quick(values))

	var log trace.Log
	sorting.TraceQuick(values, &log)
	fmt.Printf("comparisons: %d, swaps: %d, events: %d\n", log.Count(trace.Compare), log.Count(trace.Swap), log.Total())
}
