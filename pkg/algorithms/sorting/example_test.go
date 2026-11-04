package sorting_test

import (
	"fmt"

	"github.com/braswelljr/alGOrithm/pkg/algorithms/sorting"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

func ExampleQuick() {
	fmt.Println(sorting.Quick([]int{5, 3, 9, 1}))
	fmt.Println(sorting.Quick([]string{"pear", "apple", "fig"}))
	// Output:
	// [1 3 5 9]
	// [apple fig pear]
}

func ExampleTraceBubble() {
	var log trace.Log
	sorting.TraceBubble([]int{3, 2, 1}, &log)
	fmt.Println(log.Count(trace.Compare), "comparisons,", log.Count(trace.Swap), "swaps")
	// Output: 3 comparisons, 3 swaps
}
