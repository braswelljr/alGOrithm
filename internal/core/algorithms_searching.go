package core

import (
	"fmt"
	"sort"

	"github.com/braswelljr/alGOrithm/pkg/algorithms/searching"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// searchingEntries is every entry of the searching category.
func searchingEntries() []Entry {
	return []Entry{
		binarySearchEntry(),
		exponentialSearchEntry(),
		fibonacciSearchEntry(),
		interpolationSearchEntry(),
		jumpSearchEntry(),
		linearSearchEntry(),
		lowerBoundEntry(),
		medianOfMediansEntry(),
		quickselectEntry(),
		ternarySearchEntry(),
		upperBoundEntry(),
	}
}

var searchingEx = &Input{Values: []int{3, 9, 14, 21, 27, 30, 38, 42, 51, 64, 70, 77, 85, 93}, Target: 51}

func searchingMk(id, name, file, fn, summary, time string, sorted bool, traced func([]int, int, trace.Recorder) int) Entry {
	e := entry(id, name, "algorithms/searching/"+file, fn)
	e.Summary, e.Time, e.Space, e.Traced, e.Params, e.Example = summary, time, "O(1)", true, []string{"target"}, searchingEx
	e.Input = SortedInts
	if !sorted {
		e.Input = AnyInts
	}
	e.Run = func(in Input, r trace.Recorder) (Output, error) {
		if err := values(in, "search", 100_000); err != nil {
			return Output{}, err
		}
		list := append([]int(nil), in.Values...)
		if sorted {
			sort.Ints(list)
		}
		return Output{Initial: list, Index: ptr(traced(list, in.Target, r))}, nil
	}
	return e
}

func binarySearchEntry() Entry {
	return searchingMk("binary-search", "Binary Search", "binary.go", "Binary", "Halve a sorted range until the target is found or the range is empty.", "O(log n)", true, searching.TraceBinary[int])
}

func exponentialSearchEntry() Entry {
	return searchingMk("exponential-search", "Exponential Search", "exponential.go", "Exponential", "Double a bound until it passes the target, then binary search that last range.", "O(log i) for a target at i", true, searching.TraceExponential[int])
}

func fibonacciSearchEntry() Entry {
	return searchingMk("fibonacci-search", "Fibonacci Search", "fibonacci.go", "Fibonacci", "Split a sorted range at Fibonacci numbers instead of halves.", "O(log n)", true, searching.TraceFibonacci[int])
}

func interpolationSearchEntry() Entry {
	return searchingMk("interpolation-search", "Interpolation Search", "interpolation.go", "Interpolation", "Guess the target's position from its value, as in a phone book.", "O(log log n) uniform, O(n) worst", true, searching.TraceInterpolation)
}

func jumpSearchEntry() Entry {
	return searchingMk("jump-search", "Jump Search", "jump.go", "Jump", "Jump ahead in blocks of sqrt(n), then scan the one block that can hold the target.", "O(sqrt n)", true, searching.TraceJump[int])
}

func linearSearchEntry() Entry {
	return searchingMk("linear-search", "Linear Search", "linear.go", "Linear", "Look at every value in order until the target is found; needs no sorting.", "O(n)", false, searching.TraceLinear[int])
}

func ternarySearchEntry() Entry {
	return searchingMk("ternary-search", "Ternary Search", "ternary.go", "Ternary", "Split a sorted range into thirds at two midpoints and keep the third that can hold the target.", "O(log n)", true, searching.TraceTernary[int])
}

var selectionEx = &Input{Values: []int{3, 9, 14, 21, 21, 21, 30, 38, 42, 51, 64, 70}, Target: 21}

func selectionBound(id, name, fn, summary string, traced func([]int, int, trace.Recorder) int) Entry {
	e := entry(id, name, "algorithms/searching/bounds.go", fn)
	e.Summary, e.Time, e.Space, e.Input, e.Traced, e.Params, e.Example = summary, "O(log n)", "O(1)", SortedInts, true, []string{"target"}, selectionEx
	e.Run = func(in Input, r trace.Recorder) (Output, error) {
		if err := values(in, "search", 100_000); err != nil {
			return Output{}, err
		}
		list := append([]int(nil), in.Values...)
		sort.Ints(list)
		i := traced(list, in.Target, r)
		line := fmt.Sprintf("%d goes at index %d", in.Target, i)
		if i == len(list) {
			line += ", after every value"
		}
		return Output{Initial: list, Index: ptr(i), Lines: []string{line}}, nil
	}
	return e
}

func selectionSel(id, name, fn, summary, time string, traced func([]int, int, trace.Recorder) (int, bool)) Entry {
	e := entry(id, name, "algorithms/searching/selection.go", fn)
	e.Summary, e.Time, e.Space, e.Input, e.Traced = summary, time, "O(n)", Ints, true
	e.Options = []Option{{"k", "k (0 is the smallest)", 5, 0, 99_999}}
	e.Example = &Input{Values: []int{38, 27, 43, 3, 9, 82, 10, 27, 55, 1, 64, 19}, Options: map[string]int{"k": 5}}
	e.Run = func(in Input, r trace.Recorder) (Output, error) {
		if err := values(in, "search", 100_000); err != nil {
			return Output{}, err
		}
		k := in.opt("k", 5)
		v, ok := traced(in.Values, k, r)
		if !ok {
			return Output{}, inputError(fmt.Sprintf("k is 0 to %d", len(in.Values)-1))
		}
		return Output{Initial: append([]int(nil), in.Values...), Number: ptr(v),
			Lines: []string{fmt.Sprintf("the value at index %d of the sorted list is %d", k, v)}}, nil
	}
	e.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("searching.%s(%s, %d)", fn, ints(in.Values), in.opt("k", 5))
	}
	return e
}

func lowerBoundEntry() Entry {
	return selectionBound("lower-bound", "Lower Bound", "LowerBound", "The first index whose value is at least the target: where it would be inserted, before any equal values.", searching.TraceLowerBound[int])
}

func upperBoundEntry() Entry {
	return selectionBound("upper-bound", "Upper Bound", "UpperBound", "The first index whose value is greater than the target: the insertion point after any equal values.", searching.TraceUpperBound[int])
}

func quickselectEntry() Entry {
	return selectionSel("quickselect", "Quickselect", "Quickselect", "The kth smallest value: partition around a pivot like quick sort, then continue only into the side that holds position k.", "O(n) average, O(n^2) worst", searching.TraceQuickselect[int])
}

func medianOfMediansEntry() Entry {
	return selectionSel("median-of-medians", "Median of Medians", "MedianOfMedians", "Quickselect with a pivot chosen as the median of the medians of groups of five, which guarantees linear time.", "O(n)", searching.TraceMedianOfMedians[int])
}
