package core

import (
	"errors"
	"fmt"

	"github.com/braswelljr/alGOrithm/pkg/algorithms/sorting"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// sortingEntries is every entry of the sorting category.
func sortingEntries() []Entry {
	return []Entry{
		bitonicSortEntry(),
		bogoSortEntry(),
		bubbleSortEntry(),
		bucketSortEntry(),
		cocktailShakerSortEntry(),
		combSortEntry(),
		countingSortEntry(),
		cycleSortEntry(),
		gnomeSortEntry(),
		heapSortEntry(),
		insertionSortEntry(),
		introSortEntry(),
		mergeSortEntry(),
		oddEvenSortEntry(),
		pancakeSortEntry(),
		quickSortEntry(),
		radixSortEntry(),
		selectionSortEntry(),
		shellSortEntry(),
		stoogeSortEntry(),
		strandSortEntry(),
		timSortEntry(),
		treeSortEntry(),
	}
}

var sortExample = &Input{Values: []int{38, 27, 43, 3, 9, 82, 10, 27, 55, 1, 64, 19}}

func sortingMk(id, name, file, fn, summary, time, space string, traced func([]int, trace.Recorder) []int) Entry {
	e := entry(id, name, "algorithms/sorting/"+file, fn)
	e.Summary, e.Time, e.Space, e.Input, e.Traced, e.Example = summary, time, space, Ints, true, sortExample
	e.Limits = "at most 100000 values"
	e.Run = func(in Input, r trace.Recorder) (Output, error) {
		if err := values(in, "sort", 100_000); err != nil {
			return Output{}, err
		}
		out := traced(append([]int(nil), in.Values...), r)
		return Output{Initial: append([]int(nil), in.Values...), Values: out}, nil
	}
	return e
}

func sortingCounting(s []int, r trace.Recorder) []int {
	out, err := sorting.TraceCounting(s, r)
	if err != nil {
		panic(inputError("counting sort needs values within a range of about a million; try radix sort"))
	}
	return out
}

func bubbleSortEntry() Entry {
	return sortingMk("bubble-sort", "Bubble Sort", "bubble.go", "Bubble", "Swap neighbors that are out of order, pass after pass, until a pass swaps nothing.", "O(n^2), O(n) if sorted", "O(1)", sorting.TraceBubble[int])
}

func cocktailShakerSortEntry() Entry {
	return sortingMk("cocktail-shaker-sort", "Cocktail Shaker Sort", "cocktail.go", "Cocktail", "Bubble sort passing forward and then backward, so small values near the end move left quickly.", "O(n^2), O(n) if sorted", "O(1)", sorting.TraceCocktail[int])
}

func combSortEntry() Entry {
	return sortingMk("comb-sort", "Comb Sort", "comb.go", "Comb", "Bubble sort over a gap that shrinks by 1.3 each pass, so small values far down move early.", "O(n^2) worst, about O(n log n) random", "O(1)", sorting.TraceComb[int])
}

func cycleSortEntry() Entry {
	return sortingMk("cycle-sort", "Cycle Sort", "cycle.go", "Cycle", "Put each value straight into the position it belongs in, so every value is written at most once.", "O(n^2), at most n writes", "O(1)", sorting.TraceCycle[int])
}

func gnomeSortEntry() Entry {
	return sortingMk("gnome-sort", "Gnome Sort", "gnome.go", "Gnome", "Walk forward while neighbors are in order; when they are not, swap them and step back.", "O(n^2), O(n) if sorted", "O(1)", sorting.TraceGnome[int])
}

func heapSortEntry() Entry {
	return sortingMk("heap-sort", "Heap Sort", "heap.go", "Heap", "Arrange the values as a max-heap, then move the largest to the end, again and again.", "O(n log n)", "O(1)", sorting.TraceHeap[int])
}

func insertionSortEntry() Entry {
	return sortingMk("insertion-sort", "Insertion Sort", "insertion.go", "Insertion", "Take each value in turn and move it left past every larger value before it.", "O(n^2), O(n) if sorted", "O(1)", sorting.TraceInsertion[int])
}

func mergeSortEntry() Entry {
	return sortingMk("merge-sort", "Merge Sort", "merge.go", "Merge", "Sort each half, then merge the two sorted halves; stable.", "O(n log n)", "O(n)", sorting.TraceMerge[int])
}

func oddEvenSortEntry() Entry {
	return sortingMk("odd-even-sort", "Odd-Even Sort", "oddeven.go", "OddEven", "Alternate passes compare pairs at even and odd positions; each pass is independent, which suits parallel hardware.", "O(n^2)", "O(1)", sorting.TraceOddEven[int])
}

func pancakeSortEntry() Entry {
	return sortingMk("pancake-sort", "Pancake Sort", "pancake.go", "Pancake", "The only move is flipping a prefix: flip the largest to the top, then down into place.", "O(n^2), at most 2n - 3 flips", "O(1)", sorting.TracePancake[int])
}

func quickSortEntry() Entry {
	return sortingMk("quick-sort", "Quick Sort", "quick.go", "Quick", "Move everything smaller than a pivot to its left and larger to its right, then sort each side.", "O(n log n) average, O(n^2) worst", "O(log n)", sorting.TraceQuick[int])
}

func selectionSortEntry() Entry {
	return sortingMk("selection-sort", "Selection Sort", "selection.go", "Selection", "Find the smallest remaining value and swap it into the next position.", "O(n^2)", "O(1)", sorting.TraceSelection[int])
}

func shellSortEntry() Entry {
	return sortingMk("shell-sort", "Shell Sort", "shell.go", "Shell", "Insertion sort over a shrinking sequence of gaps, so values far from their place move early.", "about O(n^(4/3)), O(n^2) worst", "O(1)", sorting.TraceShell[int])
}

func bucketSortEntry() Entry {
	return sortingMk("bucket-sort", "Bucket Sort", "bucket.go", "Bucket", "Spread the values over about sqrt(n) buckets by range, sort each bucket, and concatenate.", "O(n) average, O(n^2) worst", "O(n)", sorting.TraceBucket)
}

func countingSortEntry() Entry {
	return sortingMk("counting-sort", "Counting Sort", "counting.go", "Counting", "Count how often each value occurs, then write them out in order; never compares two values.", "O(n + k) for a range of k", "O(n + k)", sortingCounting)
}

func radixSortEntry() Entry {
	return sortingMk("radix-sort", "Radix Sort", "radix.go", "Radix", "Sort by the last digit, then the next, each pass a stable counting sort; never compares two values.", "O(n * k) for k digits", "O(n + 10)", sorting.TraceRadix)
}

func hybridSortMk(id, name, file, fn, summary, time, space string, limit int, traced func([]int, trace.Recorder) []int) Entry {
	e := entry(id, name, "algorithms/sorting/"+file, fn)
	e.Summary, e.Time, e.Space, e.Input, e.Traced, e.Example = summary, time, space, Ints, true, sortExample
	e.Limits = fmt.Sprintf("at most %d values", limit)
	e.Run = func(in Input, r trace.Recorder) (Output, error) {
		if err := values(in, "sort", limit); err != nil {
			return Output{}, err
		}
		out := traced(append([]int(nil), in.Values...), r)
		return Output{Initial: append([]int(nil), in.Values...), Values: out}, nil
	}
	return e
}

func bogoSortEntry() Entry {
	bogo := entry("bogo-sort", "Bogo Sort", "algorithms/sorting/bogo.go", "Bogo")
	bogo.Summary, bogo.Time, bogo.Space, bogo.Input, bogo.Traced = "Shuffle until sorted: a joke and a lesson, about n! shuffles expected, so it gives up after a set number of tries.", "O(n * n!) expected, unbounded worst", "O(1)", Ints, true
	bogo.Options = []Option{seedOption, {"tries", "most shuffles", 10_000, 1, 1_000_000}}
	bogo.Example, bogo.Limits = &Input{Values: []int{3, 1, 4, 1, 5}, Options: map[string]int{"seed": 1, "tries": 10_000}}, "at most 8 values"
	bogo.Run = func(in Input, r trace.Recorder) (Output, error) {
		if err := values(in, "sort", 8); err != nil {
			return Output{}, err
		}
		tries := in.opt("tries", 10_000)
		if tries < 1 || tries > 1_000_000 {
			return Output{}, inputError("tries is 1 to 1000000")
		}
		out, err := sorting.TraceBogo(append([]int(nil), in.Values...), int64(in.opt("seed", 1)), tries, r)
		if errors.Is(err, sorting.ErrGaveUp) {
			return Output{}, inputError(fmt.Sprintf("bogo sort gave up after %d shuffles; try fewer values or more tries", tries))
		}
		return Output{Initial: append([]int(nil), in.Values...), Values: out}, nil
	}
	bogo.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("sorting.Bogo(%s, %d, %d)", ints(in.Values), in.opt("seed", 1), in.opt("tries", 10_000))
	}
	return bogo
}

func timSortEntry() Entry {
	return hybridSortMk("tim-sort", "Tim Sort", "tim.go", "Tim", "Find the runs already in order, extend short ones with insertion sort, then merge runs; near O(n) on partly sorted data. Python's and Java's sort.", "O(n log n), O(n) if sorted", "O(n)", 100_000, sorting.TraceTim[int])
}

func introSortEntry() Entry {
	return hybridSortMk("intro-sort", "Introsort", "intro.go", "Intro", "Quick sort that switches to heap sort when the recursion gets too deep and to insertion sort on small ranges. C++'s std::sort.", "O(n log n)", "O(log n)", 100_000, sorting.TraceIntro[int])
}

func bitonicSortEntry() Entry {
	return hybridSortMk("bitonic-sort", "Bitonic Sort", "bitonic.go", "Bitonic", "A fixed network of compare-and-swap steps that merges rising-then-falling sequences; every step of a stage can run in parallel.", "O(n log^2 n)", "O(log n)", 100_000, sorting.TraceBitonic[int])
}

func treeSortEntry() Entry {
	return hybridSortMk("tree-sort", "Tree Sort", "tree.go", "Tree", "Insert every value into a binary search tree, then read it back in order.", "O(n log n) average, O(n^2) if sorted", "O(n)", 10_000, sorting.TraceTree[int])
}

func stoogeSortEntry() Entry {
	return hybridSortMk("stooge-sort", "Stooge Sort", "stooge.go", "Stooge", "Order the ends, then sort the first two thirds, the last two thirds, and the first two thirds again: correct and famously slow.", "O(n^2.71)", "O(log n)", 200, sorting.TraceStooge[int])
}

func strandSortEntry() Entry {
	return hybridSortMk("strand-sort", "Strand Sort", "strand.go", "Strand", "Pull out an increasing strand of the remaining values and merge it into the result, until none remain.", "O(n^2), O(n) if sorted", "O(n)", 20_000, sorting.TraceStrand[int])
}
