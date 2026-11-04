package core

import (
	"fmt"
	"sort"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/algorithms/dynamic"
	"github.com/braswelljr/alGOrithm/pkg/algorithms/pathfinding"
	"github.com/braswelljr/alGOrithm/pkg/algorithms/recursion"
	"github.com/braswelljr/alGOrithm/pkg/algorithms/searching"
	"github.com/braswelljr/alGOrithm/pkg/algorithms/sorting"
	"github.com/braswelljr/alGOrithm/pkg/algorithms/strmatch"
	"github.com/braswelljr/alGOrithm/pkg/math/numbers"
	"github.com/braswelljr/alGOrithm/pkg/security/ciphers"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

var sortExample = &Input{Values: []int{38, 27, 43, 3, 9, 82, 10, 27, 55, 1, 64, 19}}

func sortingEntries() []Entry {
	mk := func(id, name, file, fn, summary, time, space string, traced func([]int, trace.Recorder) []int) Entry {
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
	counting := func(s []int, r trace.Recorder) []int {
		out, err := sorting.TraceCounting(s, r)
		if err != nil {
			panic(inputError("counting sort needs values within a range of about a million; try radix sort"))
		}
		return out
	}
	return []Entry{
		mk("bubble-sort", "Bubble Sort", "bubble.go", "Bubble", "Swap neighbors that are out of order, pass after pass, until a pass swaps nothing.", "O(n^2), O(n) if sorted", "O(1)", sorting.TraceBubble[int]),
		mk("cocktail-shaker-sort", "Cocktail Shaker Sort", "cocktail.go", "Cocktail", "Bubble sort passing forward and then backward, so small values near the end move left quickly.", "O(n^2), O(n) if sorted", "O(1)", sorting.TraceCocktail[int]),
		mk("comb-sort", "Comb Sort", "comb.go", "Comb", "Bubble sort over a gap that shrinks by 1.3 each pass, so small values far down move early.", "O(n^2) worst, about O(n log n) random", "O(1)", sorting.TraceComb[int]),
		mk("cycle-sort", "Cycle Sort", "cycle.go", "Cycle", "Put each value straight into the position it belongs in, so every value is written at most once.", "O(n^2), at most n writes", "O(1)", sorting.TraceCycle[int]),
		mk("gnome-sort", "Gnome Sort", "gnome.go", "Gnome", "Walk forward while neighbors are in order; when they are not, swap them and step back.", "O(n^2), O(n) if sorted", "O(1)", sorting.TraceGnome[int]),
		mk("heap-sort", "Heap Sort", "heap.go", "Heap", "Arrange the values as a max-heap, then move the largest to the end, again and again.", "O(n log n)", "O(1)", sorting.TraceHeap[int]),
		mk("insertion-sort", "Insertion Sort", "insertion.go", "Insertion", "Take each value in turn and move it left past every larger value before it.", "O(n^2), O(n) if sorted", "O(1)", sorting.TraceInsertion[int]),
		mk("merge-sort", "Merge Sort", "merge.go", "Merge", "Sort each half, then merge the two sorted halves; stable.", "O(n log n)", "O(n)", sorting.TraceMerge[int]),
		mk("odd-even-sort", "Odd-Even Sort", "oddeven.go", "OddEven", "Alternate passes compare pairs at even and odd positions; each pass is independent, which suits parallel hardware.", "O(n^2)", "O(1)", sorting.TraceOddEven[int]),
		mk("pancake-sort", "Pancake Sort", "pancake.go", "Pancake", "The only move is flipping a prefix: flip the largest to the top, then down into place.", "O(n^2), at most 2n - 3 flips", "O(1)", sorting.TracePancake[int]),
		mk("quick-sort", "Quick Sort", "quick.go", "Quick", "Move everything smaller than a pivot to its left and larger to its right, then sort each side.", "O(n log n) average, O(n^2) worst", "O(log n)", sorting.TraceQuick[int]),
		mk("selection-sort", "Selection Sort", "selection.go", "Selection", "Find the smallest remaining value and swap it into the next position.", "O(n^2)", "O(1)", sorting.TraceSelection[int]),
		mk("shell-sort", "Shell Sort", "shell.go", "Shell", "Insertion sort over a shrinking sequence of gaps, so values far from their place move early.", "about O(n^(4/3)), O(n^2) worst", "O(1)", sorting.TraceShell[int]),
		mk("bucket-sort", "Bucket Sort", "bucket.go", "Bucket", "Spread the values over about sqrt(n) buckets by range, sort each bucket, and concatenate.", "O(n) average, O(n^2) worst", "O(n)", sorting.TraceBucket),
		mk("counting-sort", "Counting Sort", "counting.go", "Counting", "Count how often each value occurs, then write them out in order; never compares two values.", "O(n + k) for a range of k", "O(n + k)", counting),
		mk("radix-sort", "Radix Sort", "radix.go", "Radix", "Sort by the last digit, then the next, each pass a stable counting sort; never compares two values.", "O(n * k) for k digits", "O(n + 10)", sorting.TraceRadix),
	}
}

func searchingEntries() []Entry {
	ex := &Input{Values: []int{3, 9, 14, 21, 27, 30, 38, 42, 51, 64, 70, 77, 85, 93}, Target: 51}
	mk := func(id, name, file, fn, summary, time string, sorted bool, traced func([]int, int, trace.Recorder) int) Entry {
		e := entry(id, name, "algorithms/searching/"+file, fn)
		e.Summary, e.Time, e.Space, e.Traced, e.Params, e.Example = summary, time, "O(1)", true, []string{"target"}, ex
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
	return []Entry{
		mk("binary-search", "Binary Search", "binary.go", "Binary", "Halve a sorted range until the target is found or the range is empty.", "O(log n)", true, searching.TraceBinary[int]),
		mk("exponential-search", "Exponential Search", "exponential.go", "Exponential", "Double a bound until it passes the target, then binary search that last range.", "O(log i) for a target at i", true, searching.TraceExponential[int]),
		mk("fibonacci-search", "Fibonacci Search", "fibonacci.go", "Fibonacci", "Split a sorted range at Fibonacci numbers instead of halves.", "O(log n)", true, searching.TraceFibonacci[int]),
		mk("interpolation-search", "Interpolation Search", "interpolation.go", "Interpolation", "Guess the target's position from its value, as in a phone book.", "O(log log n) uniform, O(n) worst", true, searching.TraceInterpolation),
		mk("jump-search", "Jump Search", "jump.go", "Jump", "Jump ahead in blocks of sqrt(n), then scan the one block that can hold the target.", "O(sqrt n)", true, searching.TraceJump[int]),
		mk("linear-search", "Linear Search", "linear.go", "Linear", "Look at every value in order until the target is found; needs no sorting.", "O(n)", false, searching.TraceLinear[int]),
		mk("ternary-search", "Ternary Search", "ternary.go", "Ternary", "Split a sorted range into thirds at two midpoints and keep the third that can hold the target.", "O(log n)", true, searching.TraceTernary[int]),
	}
}

func pathfindingEntries() []Entry {
	maze, _ := pathfinding.Generate(pathfinding.Maze, 15, 31, 1)
	ex := &Input{Text: maze.String()}
	mk := func(id, name, file, fn, summary, time string, traced func(*pathfinding.Grid, trace.Recorder) ([]int, int)) Entry {
		e := entry(id, name, "algorithms/pathfinding/"+file, fn)
		e.Summary, e.Time, e.Space, e.Input, e.Traced, e.Example = summary, time, "O(V)", GridInput, true, ex
		e.Limits = fmt.Sprintf("at most %d cells", pathfinding.MaxCells)
		e.Run = func(in Input, r trace.Recorder) (Output, error) {
			g, err := pathfinding.Parse(in.Text)
			if err != nil {
				return Output{}, wrap(err)
			}
			path, cost := traced(g, r)
			return Output{Grid: &GridOutput{Rows: g.Rows, Cols: g.Cols, Cost: g.Cost, Start: g.Start, End: g.End}, Path: path, Number: &cost}, nil
		}
		return e
	}
	return []Entry{
		mk("breadth-first-search", "Breadth-First Search", "breadthfirst.go", "BreadthFirst", "Explore every cell one step away, then two, and so on; the first path to the end has the fewest steps.", "O(V)", pathfinding.TraceBreadthFirst),
		mk("depth-first-search", "Depth-First Search", "depthfirst.go", "DepthFirst", "Follow one direction as far as it goes and back up at dead ends; finds a path, not the shortest.", "O(V)", pathfinding.TraceDepthFirst),
		mk("dijkstra", "Dijkstra's Algorithm", "dijkstra.go", "Dijkstra", "Always explore the cell cheapest to reach from the start; finds the cheapest path (Dijkstra, 1956).", "O(V log V)", pathfinding.TraceDijkstra),
		mk("a-star", "A* Search", "astar.go", "AStar", "Dijkstra's algorithm guided by an estimate of the cost still to go; finds the cheapest path, exploring less.", "O(V log V) worst", pathfinding.TraceAStar),
		mk("greedy-best-first", "Greedy Best-First Search", "greedy.go", "Greedy", "Always explore the cell that looks closest to the end; fast, but not always cheapest.", "O(V log V)", pathfinding.TraceGreedy),
	}
}

func dynamicEntries() []Entry {
	words := func(in Input) error {
		if in.Text == "" && in.Other == "" {
			return inputError("give two words")
		}
		if (len(in.Text)+1)*(len(in.Other)+1) > 250_000 {
			return inputError("the words are too long for the table (at most 250000 cells)")
		}
		return nil
	}
	ed := entry("edit-distance", "Edit Distance", "algorithms/dynamic/editdistance.go", "EditDistance")
	ed.Summary, ed.Time, ed.Space, ed.Input, ed.Traced = "The fewest insertions, deletions and substitutions that turn one word into another (Levenshtein, 1965).", "O(n * m)", "O(n * m)", TwoTexts, true
	ed.Example = &Input{Text: "kitten", Other: "sitting"}
	ed.Run = func(in Input, r trace.Recorder) (Output, error) {
		if err := words(in); err != nil {
			return Output{}, err
		}
		d, t := dynamic.TraceEditDistance(in.Text, in.Other, r)
		return Output{Number: &d, Table: &TableOutput{Rows: len(t), Cols: len(t[0]), Cells: flatten(t), RowLabels: labelsOf(in.Text), ColLabels: labelsOf(in.Other)}}, nil
	}
	lcs := entry("longest-common-subsequence", "Longest Common Subsequence", "algorithms/dynamic/lcs.go", "LCS")
	lcs.Summary, lcs.Time, lcs.Space, lcs.Input, lcs.Traced = "The longest sequence of letters found in both words in the same order: the comparison behind diff.", "O(n * m)", "O(n * m)", TwoTexts, true
	lcs.Example = &Input{Text: "ABCBDAB", Other: "BDCABA"}
	lcs.Run = func(in Input, r trace.Recorder) (Output, error) {
		if err := words(in); err != nil {
			return Output{}, err
		}
		s, t := dynamic.TraceLCS(in.Text, in.Other, r)
		return Output{Text: s, Number: ptr(len(s)), Table: &TableOutput{Rows: len(t), Cols: len(t[0]), Cells: flatten(t), RowLabels: labelsOf(in.Text), ColLabels: labelsOf(in.Other)}}, nil
	}
	cc := entry("coin-change", "Coin Change", "algorithms/dynamic/coinchange.go", "CoinChange")
	cc.Summary, cc.Time, cc.Space, cc.Input, cc.Traced, cc.Params = "The fewest coins that make an amount, building up from every smaller amount.", "O(amount * coins)", "O(amount)", Amount, true, []string{"amount"}
	cc.Example, cc.Limits = &Input{Values: []int{1, 5, 10, 25}, Target: 63}, fmt.Sprintf("amount 0 to %d", dynamic.MaxAmount)
	cc.Run = func(in Input, r trace.Recorder) (Output, error) {
		if err := values(in, "pay with (the coins)", 100); err != nil {
			return Output{}, err
		}
		n, used, t, err := dynamic.TraceCoinChange(in.Values, in.Target, r)
		if err != nil {
			return Output{}, wrap(err)
		}
		return Output{Number: &n, Values: used, Table: &TableOutput{Rows: 1, Cols: len(t), Cells: t, ColLabels: numberLabels(len(t), 0)}}, nil
	}
	lis := entry("longest-increasing-subsequence", "Longest Increasing Subsequence", "algorithms/dynamic/lis.go", "LIS")
	lis.Summary, lis.Time, lis.Space, lis.Input, lis.Traced = "The longest strictly increasing run of values taken in order, from the best run ending at each position.", "O(n^2)", "O(n)", Ints, true
	lis.Example, lis.Limits = &Input{Values: []int{10, 9, 2, 5, 3, 7, 101, 18, 4, 20}}, "at most 2000 values"
	lis.Run = func(in Input, r trace.Recorder) (Output, error) {
		if err := values(in, "scan", 2000); err != nil {
			return Output{}, err
		}
		seq, t := dynamic.TraceLIS(in.Values, r)
		cols := make([]string, len(in.Values))
		for i, v := range in.Values {
			cols[i] = fmt.Sprint(v)
		}
		return Output{Initial: append([]int(nil), in.Values...), Values: seq, Number: ptr(len(seq)), Table: &TableOutput{Rows: 1, Cols: len(t), Cells: t, ColLabels: cols}}, nil
	}
	return []Entry{ed, lcs, cc, lis}
}

func strmatchEntries() []Entry {
	ex := &Input{Text: "abracadabra, abracadabra", Other: "abra"}
	mk := func(id, name, file, fn, summary, time, space string, traced func(string, string, trace.Recorder) []int) Entry {
		e := entry(id, name, "algorithms/strmatch/"+file, fn)
		e.Summary, e.Time, e.Space, e.Input, e.Traced, e.Params, e.Example = summary, time, space, Pattern, true, []string{"pattern"}, ex
		e.Limits = "text up to 10000 characters"
		e.Run = func(in Input, r trace.Recorder) (Output, error) {
			if in.Text == "" || in.Other == "" {
				return Output{}, inputError("give a text and a pattern")
			}
			if len(in.Text) > 10_000 {
				return Output{}, inputError("the text is longer than 10000 characters")
			}
			m := traced(in.Text, in.Other, r)
			if m == nil {
				m = []int{}
			}
			return Output{Text: in.Text, Matches: m, Number: ptr(len(m))}, nil
		}
		return e
	}
	ana := entry("anagram", "Anagram", "algorithms/strmatch/anagram.go", "Anagram")
	ana.Summary, ana.Time, ana.Space, ana.Input = "Whether two phrases use exactly the same letters, ignoring case and spaces.", "O(n)", "O(1)", TwoTexts
	ana.Example = &Input{Text: "Dormitory", Other: "dirty room"}
	ana.Run = func(in Input, _ trace.Recorder) (Output, error) {
		if in.Text == "" || in.Other == "" {
			return Output{}, inputError("give two words")
		}
		return Output{Bool: ptr(strmatch.Anagram(in.Text, in.Other))}, nil
	}
	return []Entry{
		mk("naive-search", "Naive String Search", "naive.go", "Naive", "Try every starting position and compare letter by letter.", "O(n * m)", "O(1)", strmatch.TraceNaive),
		mk("knuth-morris-pratt", "Knuth-Morris-Pratt", "kmp.go", "KMP", "After a mismatch, keep the longest part of the match that still fits, so no text is read twice (1977).", "O(n + m)", "O(m)", strmatch.TraceKMP),
		mk("rabin-karp", "Rabin-Karp", "rabinkarp.go", "RabinKarp", "Compare a rolling hash of each window with the pattern's hash, and letters only when they agree (1987).", "O(n + m) average", "O(1)", strmatch.TraceRabinKarp),
		ana,
	}
}

func recursionEntries() []Entry {
	num := func(id, name, file, fn, summary, time, space string, lo, hi int, f func(int) int) Entry {
		e := entry(id, name, "algorithms/recursion/"+file, fn)
		e.Summary, e.Time, e.Space, e.Input, e.Limits = summary, time, space, Number, fmt.Sprintf("%d <= n <= %d", lo, hi)
		e.Example = &Input{Values: []int{min(hi, 10)}}
		e.Run = func(in Input, _ trace.Recorder) (Output, error) {
			n, err := one(in)
			if err != nil {
				return Output{}, err
			}
			if n < lo || n > hi {
				return Output{}, inputError(fmt.Sprintf("%s takes %d <= n <= %d", id, lo, hi))
			}
			return Output{Number: ptr(f(n))}, nil
		}
		return e
	}
	pair := func(id, name, fn, summary string, f func(int, int) int) Entry {
		e := entry(id, name, "algorithms/recursion/gcd.go", fn)
		e.Summary, e.Time, e.Space, e.Input, e.Example = summary, "O(log min(a, b))", "O(log min(a, b))", TwoNumbers, &Input{Values: []int{1071, 462}}
		e.Run = func(in Input, _ trace.Recorder) (Output, error) {
			a, b, err := two(in)
			if err != nil {
				return Output{}, err
			}
			return Output{Number: ptr(f(a, b))}, nil
		}
		return e
	}
	fact := num("factorial", "Factorial", "factorial.go", "Factorial", "n! = n * (n-1) * ... * 1, computed recursively.", "O(n)", "O(n)", 0, 20, func(n int) int { v, _ := recursion.Factorial(n); return v })
	ack := entry("ackermann", "Ackermann Function", "algorithms/recursion/ackermann.go", "Ackermann")
	ack.Summary, ack.Time, ack.Space, ack.Input, ack.Limits = "A total recursive function that grows faster than any primitive recursive one.", "faster than exponential", "O(A(m, n))", TwoNumbers, "m <= 3, n <= 10"
	ack.Example = &Input{Values: []int{2, 3}}
	ack.Run = func(in Input, _ trace.Recorder) (Output, error) {
		m, n, err := two(in)
		if err != nil {
			return Output{}, err
		}
		if m < 0 || n < 0 || m > 3 || n > 10 {
			return Output{}, inputError("ackermann takes 0 <= m <= 3 and 0 <= n <= 10; larger values do not finish")
		}
		return Output{Number: ptr(recursion.Ackermann(m, n))}, nil
	}
	hanoi := entry("tower-of-hanoi", "Tower of Hanoi", "algorithms/recursion/hanoi.go", "Hanoi")
	hanoi.Summary, hanoi.Time, hanoi.Space, hanoi.Input, hanoi.Limits = "Move n disks from peg A to C, never a larger disk on a smaller one: 2^n - 1 moves.", "O(2^n)", "O(n)", Number, "1 <= n <= 12"
	hanoi.Example = &Input{Values: []int{4}}
	hanoi.Run = func(in Input, _ trace.Recorder) (Output, error) {
		n, err := one(in)
		if err != nil {
			return Output{}, err
		}
		if n < 1 || n > 12 {
			return Output{}, inputError("tower-of-hanoi takes 1 to 12 disks")
		}
		return Output{Lines: recursion.Hanoi(n), Number: &n}, nil
	}
	pal := entry("palindrome", "Palindrome", "algorithms/recursion/palindrome.go", "Palindrome")
	pal.Summary, pal.Time, pal.Space, pal.Input, pal.Example = "Whether a word reads the same backward: its ends match and its middle is a palindrome.", "O(n)", "O(n)", Text, &Input{Text: "racecar"}
	pal.Run = func(in Input, _ trace.Recorder) (Output, error) {
		if in.Text == "" {
			return Output{}, inputError("give a word")
		}
		return Output{Bool: ptr(recursion.Palindrome(in.Text))}, nil
	}
	perm := entry("permutations", "Permutations", "algorithms/recursion/palindrome.go", "Permutations")
	perm.Summary, perm.Time, perm.Space, perm.Input, perm.Limits = "Every ordering of the values, by fixing each first and permuting the rest (backtracking).", "O(n * n!)", "O(n * n!)", Ints, "at most 7 values"
	perm.Example = &Input{Values: []int{1, 2, 3}}
	perm.Run = func(in Input, _ trace.Recorder) (Output, error) {
		if err := values(in, "permute", 7); err != nil {
			return Output{}, err
		}
		var lines []string
		for _, p := range recursion.Permutations(in.Values) {
			lines = append(lines, joinInts(p))
		}
		return Output{Lines: lines, Number: ptr(len(lines))}, nil
	}
	return []Entry{
		fact,
		num("fibonacci", "Fibonacci (plain recursion)", "fibonacci.go", "Fibonacci", "F(n) = F(n-1) + F(n-2) by plain recursion: exponential, because the same values are computed again and again.", "O(1.618^n)", "O(n)", 0, 35, recursion.Fibonacci),
		num("fibonacci-memo", "Fibonacci (memoized)", "fibonacci.go", "FibonacciMemo", "The same recursion, remembering each F(k) once computed: linear.", "O(n)", "O(n)", 0, 90, recursion.FibonacciMemo),
		num("mccarthy-91", "McCarthy 91", "ackermann.go", "McCarthy91", "n - 10 above 100, else M(M(n + 11)): 91 for every n <= 101.", "O(101 - n)", "O(101 - n)", -1000, 1000, recursion.McCarthy91),
		pair("gcd", "Euclidean GCD", "GCD", "The greatest common divisor by Euclid's algorithm: gcd(a, b) = gcd(b, a mod b).", recursion.GCD),
		pair("lcm", "Least Common Multiple", "LCM", "The smallest number both divide: a / gcd(a, b) * b.", recursion.LCM),
		ack, hanoi, pal, perm,
	}
}

func numberEntries() []Entry {
	sieve := entry("sieve-of-eratosthenes", "Sieve of Eratosthenes", "math/numbers/sieve.go", "Sieve")
	sieve.Summary, sieve.Time, sieve.Space, sieve.Input, sieve.Traced = "Cross out the multiples of each prime from its square; what is never crossed out is prime.", "O(n log log n)", "O(n)", Number, true
	sieve.Example, sieve.Limits = &Input{Values: []int{120}}, fmt.Sprintf("2 <= n <= %d", numbers.MaxSieve)
	sieve.Run = func(in Input, r trace.Recorder) (Output, error) {
		n, err := one(in)
		if err != nil {
			return Output{}, err
		}
		if n < 2 || n > numbers.MaxSieve {
			return Output{}, inputError(fmt.Sprintf("the sieve takes 2 <= n <= %d", numbers.MaxSieve))
		}
		p := numbers.TraceSieve(n, r)
		initial := make([]int, n+1)
		for i := range initial {
			initial[i] = i
		}
		return Output{Initial: initial, Values: p, Number: ptr(len(p))}, nil
	}
	boolean := func(id, name, fn, summary, time string, f func(int) bool) Entry {
		e := entry(id, name, "math/numbers/numbers.go", fn)
		e.Summary, e.Time, e.Space, e.Input, e.Example = summary, time, "O(1)", Number, &Input{Values: []int{97}}
		e.Run = func(in Input, _ trace.Recorder) (Output, error) {
			n, err := one(in)
			if err != nil {
				return Output{}, err
			}
			return Output{Bool: ptr(f(n))}, nil
		}
		return e
	}
	fac := entry("prime-factors", "Prime Factorization", "math/numbers/numbers.go", "Factorize")
	fac.Summary, fac.Time, fac.Space, fac.Input, fac.Example = "The prime factors of n, by trial division.", "O(sqrt n)", "O(log n)", Number, &Input{Values: []int{360}}
	fac.Run = func(in Input, _ trace.Recorder) (Output, error) {
		n, err := one(in)
		if err != nil {
			return Output{}, err
		}
		return Output{Values: numbers.Factorize(n)}, nil
	}
	pow := entry("fast-power", "Fast Exponentiation", "math/numbers/numbers.go", "Power")
	pow.Summary, pow.Time, pow.Space, pow.Input, pow.Options = "base^exp mod m by squaring: O(log exp) multiplications instead of exp.", "O(log exp)", "O(1)", TwoNumbers, []Option{{"mod", "modulus", 1_000_000_007, 1, 2_000_000_000}}
	pow.Example = &Input{Values: []int{3, 200}}
	pow.Run = func(in Input, _ trace.Recorder) (Output, error) {
		b, e, err := two(in)
		if err != nil {
			return Output{}, err
		}
		m := in.opt("mod", 1_000_000_007)
		if e < 0 || m < 1 || m > 2_000_000_000 {
			return Output{}, inputError("give exp >= 0 and 1 <= mod <= 2000000000")
		}
		return Output{Number: ptr(numbers.Power(b, e, m))}, nil
	}
	return []Entry{
		sieve, fac, pow,
		boolean("prime", "Primality", "Prime", "Whether n is prime, by trial division up to sqrt(n).", "O(sqrt n)", numbers.Prime),
		boolean("even", "Even", "Even", "Whether n is even.", "O(1)", numbers.Even),
	}
}

func cipherEntries() []Entry {
	text := func(in Input) error {
		if in.Text == "" {
			return inputError("give the text to encrypt")
		}
		return nil
	}
	caesar := entry("caesar-cipher", "Caesar Cipher", "security/ciphers/caesar.go", "Caesar")
	caesar.Summary, caesar.Time, caesar.Space, caesar.Input, caesar.Params = "Shift every letter forward by a fixed number of places (A becomes D for 3).", "O(n)", "O(n)", Text, []string{"shift"}
	caesar.Example = &Input{Text: "Attack at dawn", Shift: 3}
	caesar.Run = func(in Input, _ trace.Recorder) (Output, error) {
		if err := text(in); err != nil {
			return Output{}, err
		}
		return Output{Text: ciphers.Caesar(in.Text, in.Shift)}, nil
	}
	rot := entry("rot13", "ROT13", "security/ciphers/caesar.go", "ROT13")
	rot.Summary, rot.Time, rot.Space, rot.Input, rot.Example = "Shift every letter by 13, so applying it twice gives the text back.", "O(n)", "O(n)", Text, &Input{Text: "Why did the chicken cross the road?"}
	rot.Run = func(in Input, _ trace.Recorder) (Output, error) {
		if err := text(in); err != nil {
			return Output{}, err
		}
		return Output{Text: ciphers.ROT13(in.Text)}, nil
	}
	vig := entry("vigenere-cipher", "Vigenère Cipher", "security/ciphers/vigenere.go", "Vigenere")
	vig.Summary, vig.Time, vig.Space, vig.Input, vig.Params = "Shift each letter by the letter of a repeating keyword under it.", "O(n)", "O(n)", Text, []string{"key"}
	vig.Example = &Input{Text: "ATTACKATDAWN", Key: "LEMON"}
	vig.Run = func(in Input, _ trace.Recorder) (Output, error) {
		if err := text(in); err != nil {
			return Output{}, err
		}
		if strings.TrimSpace(in.Key) == "" {
			return Output{}, inputError("give a keyword")
		}
		return Output{Text: ciphers.Vigenere(in.Text, in.Key)}, nil
	}
	hill := entry("hill-cipher", "Hill Cipher", "security/ciphers/hill.go", "Hill")
	hill.Summary, hill.Time, hill.Space, hill.Input = "Multiply blocks of letters by a key matrix modulo 26; decrypt with its inverse.", "O(n * k)", "O(n)", Text
	hill.Options = []Option{{"size", "key size", 2, 1, 5}, {"seed", "key seed", 1, 0, 1_000_000}}
	hill.Example = &Input{Text: "Help me, Hill", Options: map[string]int{"size": 2, "seed": 1}}
	hill.Run = func(in Input, _ trace.Recorder) (Output, error) {
		if err := text(in); err != nil {
			return Output{}, err
		}
		k := in.opt("size", 2)
		if k < 1 || k > 5 {
			return Output{}, inputError("the key size is 1 to 5")
		}
		key := ciphers.HillKey(k, int64(in.opt("seed", 1)))
		enc, _ := ciphers.Hill(in.Text, key)
		dec, _ := ciphers.HillDecrypt(enc, key)
		cells := flatten(key)
		return Output{Text: enc, Lines: []string{"key " + fmt.Sprint(key), "decrypts to " + dec}, Table: &TableOutput{Rows: k, Cols: k, Cells: cells}}, nil
	}
	return []Entry{caesar, rot, vig, hill}
}
