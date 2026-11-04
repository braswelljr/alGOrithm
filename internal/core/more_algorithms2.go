package core

import (
	"errors"
	"fmt"
	"math"
	"math/bits"
	"sort"
	"strconv"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/algorithms/compression"
	"github.com/braswelljr/alGOrithm/pkg/algorithms/searching"
	"github.com/braswelljr/alGOrithm/pkg/algorithms/sorting"
	"github.com/braswelljr/alGOrithm/pkg/math/bitwise"
	"github.com/braswelljr/alGOrithm/pkg/math/numbers"
	"github.com/braswelljr/alGOrithm/pkg/math/numerical"
	"github.com/braswelljr/alGOrithm/pkg/math/randomized"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// evenMoreAlgorithmEntries are the third set of algorithms: more sorts and
// searches, number theory, compression, bit tricks, numerical methods and
// randomized algorithms.
func evenMoreAlgorithmEntries() []Entry {
	var out []Entry
	for _, part := range [][]Entry{extraSortEntries(), extraSearchEntries(), extraNumberEntries(),
		compressionEntries(), bitwiseEntries(), numericalEntries(), randomizedEntries()} {
		out = append(out, part...)
	}
	return out
}

// seedOption is the option every seeded entry reads.
var seedOption = Option{"seed", "random seed", 1, 0, 1_000_000}

func extraSortEntries() []Entry {
	mk := func(id, name, file, fn, summary, time, space string, limit int, traced func([]int, trace.Recorder) []int) Entry {
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
	return []Entry{
		mk("tim-sort", "Tim Sort", "tim.go", "Tim", "Find the runs already in order, extend short ones with insertion sort, then merge runs; near O(n) on partly sorted data. Python's and Java's sort.", "O(n log n), O(n) if sorted", "O(n)", 100_000, sorting.TraceTim[int]),
		mk("intro-sort", "Introsort", "intro.go", "Intro", "Quick sort that switches to heap sort when the recursion gets too deep and to insertion sort on small ranges. C++'s std::sort.", "O(n log n)", "O(log n)", 100_000, sorting.TraceIntro[int]),
		mk("bitonic-sort", "Bitonic Sort", "bitonic.go", "Bitonic", "A fixed network of compare-and-swap steps that merges rising-then-falling sequences; every step of a stage can run in parallel.", "O(n log^2 n)", "O(log n)", 100_000, sorting.TraceBitonic[int]),
		mk("tree-sort", "Tree Sort", "tree.go", "Tree", "Insert every value into a binary search tree, then read it back in order.", "O(n log n) average, O(n^2) if sorted", "O(n)", 10_000, sorting.TraceTree[int]),
		mk("stooge-sort", "Stooge Sort", "stooge.go", "Stooge", "Order the ends, then sort the first two thirds, the last two thirds, and the first two thirds again: correct and famously slow.", "O(n^2.71)", "O(log n)", 200, sorting.TraceStooge[int]),
		bogo,
		mk("strand-sort", "Strand Sort", "strand.go", "Strand", "Pull out an increasing strand of the remaining values and merge it into the result, until none remain.", "O(n^2), O(n) if sorted", "O(n)", 20_000, sorting.TraceStrand[int]),
	}
}

func extraSearchEntries() []Entry {
	ex := &Input{Values: []int{3, 9, 14, 21, 21, 21, 30, 38, 42, 51, 64, 70}, Target: 21}
	bound := func(id, name, fn, summary string, traced func([]int, int, trace.Recorder) int) Entry {
		e := entry(id, name, "algorithms/searching/bounds.go", fn)
		e.Summary, e.Time, e.Space, e.Input, e.Traced, e.Params, e.Example = summary, "O(log n)", "O(1)", SortedInts, true, []string{"target"}, ex
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
	sel := func(id, name, fn, summary, time string, traced func([]int, int, trace.Recorder) (int, bool)) Entry {
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
	return []Entry{
		bound("lower-bound", "Lower Bound", "LowerBound", "The first index whose value is at least the target: where it would be inserted, before any equal values.", searching.TraceLowerBound[int]),
		bound("upper-bound", "Upper Bound", "UpperBound", "The first index whose value is greater than the target: the insertion point after any equal values.", searching.TraceUpperBound[int]),
		sel("quickselect", "Quickselect", "Quickselect", "The kth smallest value: partition around a pivot like quick sort, then continue only into the side that holds position k.", "O(n) average, O(n^2) worst", searching.TraceQuickselect[int]),
		sel("median-of-medians", "Median of Medians", "MedianOfMedians", "Quickselect with a pivot chosen as the median of the medians of groups of five, which guarantees linear time.", "O(n)", searching.TraceMedianOfMedians[int]),
	}
}

func extraNumberEntries() []Entry {
	single := func(id, name, file, fn, summary, time, space string, ex int, run func(n int) (Output, error)) Entry {
		e := entry(id, name, "math/numbers/"+file, fn)
		e.Summary, e.Time, e.Space, e.Input, e.Example = summary, time, space, Number, &Input{Values: []int{ex}}
		e.Run = func(in Input, _ trace.Recorder) (Output, error) {
			n, err := one(in)
			if err != nil {
				return Output{}, err
			}
			return run(n)
		}
		e.Call = func(in Input) (string, string) { return "", fmt.Sprintf("numbers.%s(%d)", fn, firstOr(in.Values, 0)) }
		return e
	}

	crt := entry("chinese-remainder", "Chinese Remainder Theorem", "math/numbers/crt.go", "ChineseRemainder")
	crt.Summary, crt.Time, crt.Space, crt.Input = "The x that leaves given remainders for given moduli, combining one congruence at a time; the moduli need not be coprime.", "O(k log M)", "O(1)", Text
	crt.Example = &Input{Text: "2 3\n3 5\n2 7"}
	crt.Run = func(in Input, _ trace.Recorder) (Output, error) {
		r, m, err := parsePairs(in.Text)
		if err != nil {
			return Output{}, inputError(err.Error() + " (one \"remainder modulus\" pair a line)")
		}
		x, mod, err := numbers.ChineseRemainder(r, m)
		if err != nil {
			return Output{}, inputError("no solution: the congruences disagree, a modulus is below 1, or the product is too large")
		}
		out := Output{Number: ptr(x), Values: []int{x, mod}}
		for i := range r {
			out.Lines = append(out.Lines, fmt.Sprintf("%d mod %d = %d", x, m[i], x%m[i]))
		}
		out.Lines = append(out.Lines, fmt.Sprintf("every solution is %d + %d k", x, mod))
		return out, nil
	}
	crt.Call = func(in Input) (string, string) {
		r, m, _ := parsePairs(in.Text)
		return "", fmt.Sprintf("numbers.ChineseRemainder(%s, %s)", ints(r), ints(m))
	}

	phi := single("euler-totient", "Euler's Totient", "totient.go", "EulerTotient", "How many of 1..n share no factor with n: n times (1 - 1/p) for each prime p dividing it.", "O(sqrt n)", "O(1)", 36,
		func(n int) (Output, error) {
			if n < 1 {
				return Output{}, inputError("give n >= 1")
			}
			p := numbers.EulerTotient(n)
			out := Output{Number: ptr(p), Lines: []string{fmt.Sprintf("phi(%d) = %d", n, p)}}
			if n <= 1000 {
				for k := 1; k <= n; k++ {
					if numbers.BinaryGCD(n, k) == 1 {
						out.Values = append(out.Values, k)
					}
				}
			}
			if n <= 100_000 {
				t := numbers.TotientSieve(min(n, 20))
				out.Lines = append(out.Lines, "phi(1..20) by the sieve: "+joinInts(t[1:]))
			}
			return out, nil
		})

	seg := entry("segmented-sieve", "Segmented Sieve", "math/numbers/segmented.go", "SegmentedSieve")
	seg.Summary, seg.Time, seg.Space, seg.Input, seg.Traced = "The primes in a range [lo, hi], sieving one window of about sqrt(hi) numbers at a time with the primes up to sqrt(hi).", "O((hi - lo) log log hi + sqrt hi)", "O(sqrt hi)", TwoNumbers, true
	seg.Example, seg.Limits = &Input{Values: []int{100, 200}}, "hi - lo below 100000"
	seg.Run = func(in Input, r trace.Recorder) (Output, error) {
		lo, hi, err := two(in)
		if err != nil {
			return Output{}, err
		}
		if hi < lo || hi < 2 || hi-lo >= 100_000 {
			return Output{}, inputError("give lo <= hi, hi >= 2 and hi - lo below 100000")
		}
		p := numbers.TraceSegmentedSieve(lo, hi, r)
		return Output{Values: p, Number: ptr(len(p))}, nil
	}
	seg.Call = func(in Input) (string, string) {
		a, b := firstOr(in.Values, 0), 0
		if len(in.Values) > 1 {
			b = in.Values[1]
		}
		return "", fmt.Sprintf("numbers.SegmentedSieve(%d, %d)", a, b)
	}

	div := single("divisors", "Divisors", "divisors.go", "Divisors", "Every divisor of n, found in pairs d and n/d while d*d <= n.", "O(sqrt n)", "O(d(n))", 360,
		func(n int) (Output, error) {
			if n < 1 {
				return Output{}, inputError("give n >= 1")
			}
			d := numbers.Divisors(n)
			sum := 0
			for _, v := range d {
				sum += v
			}
			return Output{Values: d, Number: ptr(len(d)), Summary: map[string]float64{"divisors": float64(len(d)), "sum": float64(sum)}}, nil
		})
	perfect := single("perfect-numbers", "Perfect Numbers", "divisors.go", "PerfectNumbers", "The numbers up to a limit equal to the sum of their smaller divisors (6 = 1 + 2 + 3), with a sieve over divisor sums.", "O(n log n)", "O(n)", 10_000,
		func(n int) (Output, error) {
			if n < 1 || n > 1_000_000 {
				return Output{}, inputError("give a limit from 1 to 1000000")
			}
			p := numbers.PerfectNumbers(n)
			out := Output{Values: p, Number: ptr(len(p))}
			for _, v := range p {
				d := numbers.Divisors(v)
				parts := make([]string, len(d)-1)
				for i, x := range d[:len(d)-1] {
					parts[i] = strconv.Itoa(x)
				}
				out.Lines = append(out.Lines, fmt.Sprintf("%d = %s", v, strings.Join(parts, " + ")))
			}
			return out, nil
		})
	perfect.Limits = "a limit of at most 1000000"
	fib := single("fibonacci-fast-doubling", "Fibonacci (Fast Doubling)", "doubling.go", "FibonacciDoubling", "F(n) from F(2k) = F(k)(2F(k+1) - F(k)) and F(2k+1) = F(k)^2 + F(k+1)^2, reading the bits of n from the top.", "O(log n)", "O(1)", 90,
		func(n int) (Output, error) {
			f, err := numbers.FibonacciDoubling(n)
			if err != nil {
				return Output{}, inputError("give n from 0 to 92")
			}
			out := Output{Number: ptr(f)}
			for s := bits.Len(uint(n)) - 1; s >= 0; s-- {
				k := n >> s
				fk, _ := numbers.FibonacciDoubling(k)
				out.Lines = append(out.Lines, fmt.Sprintf("bits %b: k = %d, F(k) = %d", k, k, fk))
			}
			return out, nil
		})
	fib.Limits = "0 <= n <= 92"
	return []Entry{crt, phi, seg, div, perfect, fib}
}

// textBytes checks a compression entry's text.
func textBytes(in Input, limit int) ([]byte, error) {
	if in.Text == "" {
		return nil, inputError("give the text to compress")
	}
	if len(in.Text) > limit {
		return nil, inputError(fmt.Sprintf("at most %d characters", limit))
	}
	return []byte(in.Text), nil
}

// rowSnapshot draws text as a row of characters.
func rowSnapshot(labels []string) trace.Snapshot {
	s := trace.Snapshot{Layout: "row"}
	for i, l := range labels {
		s.Nodes = append(s.Nodes, trace.Node{ID: i, Label: l, Parent: -1})
	}
	return s
}

func chars(b []byte) []string {
	out := make([]string, len(b))
	for i, c := range b {
		out[i] = string(rune(c))
	}
	return out
}

func compressionEntries() []Entry {
	textEntry := func(id, name, file, fn, summary, time, space, ex string, run func(data []byte, in Input) (Output, error)) Entry {
		e := entry(id, name, "algorithms/compression/"+file, fn)
		e.Summary, e.Time, e.Space, e.Input, e.Example, e.Limits = summary, time, space, Text, &Input{Text: ex}, "at most 2000 characters"
		e.Run = func(in Input, _ trace.Recorder) (Output, error) {
			data, err := textBytes(in, 2000)
			if err != nil {
				return Output{}, err
			}
			return run(data, in)
		}
		e.Call = func(in Input) (string, string) {
			return "", fmt.Sprintf("compression.%s([]byte(%q))", fn, in.Text)
		}
		return e
	}
	rle := textEntry("run-length-encoding", "Run-Length Encoding", "runlength.go", "RunLength", "Write each run of equal characters as the character and a count: great on long runs, twice the size with none.", "O(n)", "O(runs)", "WWWWWWWWWWWWBWWWWWWWWWWWWBBBWWWWWWWWWWWWWW",
		func(data []byte, _ Input) (Output, error) {
			runs := compression.RunLength(data)
			var parts []string
			out := Output{}
			for _, r := range runs {
				parts = append(parts, fmt.Sprintf("%d%c", r.Count, r.Value))
			}
			back, _ := compression.RunLengthDecode(runs)
			out.Text = strings.Join(parts, "")
			out.Lines = []string{fmt.Sprintf("%d runs: %s", len(runs), strings.Join(parts, " ")), "decodes to " + string(back)}
			out.Summary = map[string]float64{"runs": float64(len(runs)), "ratio": compression.Ratio(len(data), 2*len(runs))}
			return out, nil
		})
	lz := textEntry("lz77", "LZ77", "lz77.go", "LZ77", "Replace repeats with (offset, length) pointers back into a sliding window of what came before, plus the next character; the idea under gzip and PNG.", "O(n * window)", "O(tokens)", "abracadabra abracadabra",
		func(data []byte, in Input) (Output, error) {
			w := in.opt("window", 64)
			if w < 1 || w > 4096 {
				return Output{}, inputError("the window is 1 to 4096")
			}
			tokens := compression.LZ77(data, w)
			out := Output{}
			var sofar []byte
			for _, t := range tokens {
				part := []byte{t.Next}
				if t.Length > 0 {
					start := len(sofar) - t.Offset
					part = nil
					for k := 0; k < t.Length; k++ {
						part = append(part, sofar[start+k%t.Offset])
					}
					part = append(part, t.Next)
				}
				sofar = append(sofar, part...)
				label := fmt.Sprintf("(%d, %d, %q)", t.Offset, t.Length, t.Next)
				out.Lines = append(out.Lines, fmt.Sprintf("%s adds %q", label, part))
				out.Steps = append(out.Steps, Step{Op: Op{Name: "token", Value: label}, Result: string(part), Snapshot: rowSnapshot(chars(sofar))})
			}
			out.Number = ptr(len(tokens))
			out.Summary = map[string]float64{"tokens": float64(len(tokens)), "ratio": compression.Ratio(len(data), 3*len(tokens))}
			return out, nil
		})
	lz.Options = []Option{{"window", "window size", 64, 1, 4096}}
	lz.Example.Options = map[string]int{"window": 64}
	lz.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("compression.LZ77([]byte(%q), %d)", in.Text, in.opt("window", 64))
	}
	lzw := textEntry("lzw", "LZW", "lzw.go", "LZW", "Emit the code of the longest phrase already in the dictionary and add that phrase plus the next character as a new code; GIF's compressor.", "O(n)", "O(n)", "TOBEORNOTTOBEORTOBEORNOT",
		func(data []byte, _ Input) (Output, error) {
			codes, added := compression.LZW(data)
			out := Output{Values: codes}
			var dict []string
			for i, p := range added {
				dict = append(dict, fmt.Sprintf("%d=%s", 256+i, p))
				out.Steps = append(out.Steps, Step{Op: Op{Name: "add", Value: p}, Result: strconv.Itoa(256 + i), Snapshot: rowSnapshot(dict)})
				out.Lines = append(out.Lines, fmt.Sprintf("code %d = %q", 256+i, p))
			}
			out.Summary = map[string]float64{"codes": float64(len(codes)), "dictionary entries": float64(len(added)), "ratio (12-bit codes)": compression.Ratio(8*len(data), 12*len(codes))}
			return out, nil
		})
	bwt := textEntry("burrows-wheeler", "Burrows-Wheeler Transform", "bwt.go", "BurrowsWheeler", "Sort every rotation of the text and keep the last column: repeated contexts bunch equal characters together, and the transform can be undone.", "O(n^2 log n) here, O(n) with suffix arrays", "O(n)", "banana",
		func(data []byte, _ Input) (Output, error) {
			if len(data) > 300 {
				return Output{}, inputError("at most 300 characters, so the rotations fit")
			}
			s := string(data)
			last, idx := compression.BurrowsWheeler(s)
			back, _ := compression.InverseBurrowsWheeler(last, idx)
			out := Output{Text: last, Number: ptr(idx)}
			for i, r := range compression.Rotations(s) {
				line := r
				if i == idx {
					line += "   <- the text, row " + strconv.Itoa(i)
				}
				out.Lines = append(out.Lines, line)
			}
			out.Lines = append(out.Lines, fmt.Sprintf("last column %q, row %d; inverting gives %q", last, idx, back))
			return out, nil
		})
	bwt.Limits = "at most 300 characters"
	bwt.Call = func(in Input) (string, string) { return "", fmt.Sprintf("compression.BurrowsWheeler(%q)", in.Text) }
	mtf := textEntry("move-to-front", "Move-to-Front", "mtf.go", "MoveToFront", "Replace each byte with its place in a list and move it to the front, so recently seen bytes become small numbers.", "O(256 n)", "O(n)", "nnbaaa",
		func(data []byte, _ Input) (Output, error) {
			ranks := compression.MoveToFront(data)
			out := Output{}
			zeros := 0
			for i, r := range ranks {
				out.Values = append(out.Values, int(r))
				if r == 0 {
					zeros++
				}
				if i < 50 {
					out.Lines = append(out.Lines, fmt.Sprintf("%q -> %d", data[i], r))
				}
			}
			out.Summary = map[string]float64{"zeros": float64(zeros)}
			return out, nil
		})
	return []Entry{rle, lz, lzw, bwt, mtf}
}

// binary writes x in at least width bits.
func binary(x uint64, width int) string { return fmt.Sprintf("%0*b", width, x) }

func bitwiseEntries() []Entry {
	num := func(id, name, file, fn, summary, time string, ex int, run func(x uint64, in Input, r trace.Recorder) (Output, error)) Entry {
		e := entry(id, name, "math/bitwise/"+file, fn)
		e.Summary, e.Time, e.Space, e.Input, e.Example = summary, time, "O(1)", Number, &Input{Values: []int{ex}}
		e.Run = func(in Input, r trace.Recorder) (Output, error) {
			n, err := one(in)
			if err != nil {
				return Output{}, err
			}
			if n < 0 {
				return Output{}, inputError("give a number >= 0")
			}
			return run(uint64(n), in, r)
		}
		e.Call = func(in Input) (string, string) { return "", fmt.Sprintf("bitwise.%s(%d)", fn, firstOr(in.Values, 0)) }
		return e
	}
	pop := num("popcount", "Population Count (Kernighan)", "bitwise.go", "PopCount", "Count the set bits by clearing the lowest one with x & (x - 1) until none remain: one step per set bit.", "O(set bits)", 2024,
		func(x uint64, _ Input, r trace.Recorder) (Output, error) {
			w := max(bits.Len64(x), 1)
			out := Output{Lines: []string{binary(x, w)}}
			rec := &trace.Log{Limit: 64}
			n := bitwise.TracePopCount(x, tee{rec, r})
			for _, e := range rec.Events {
				x &^= 1 << e.Indices[0]
				out.Lines = append(out.Lines, fmt.Sprintf("%s  cleared bit %d", binary(x, w), e.Indices[0]))
			}
			out.Number = ptr(n)
			return out, nil
		})
	pop.Traced = true
	pow := num("power-of-two", "Power of Two", "power.go", "IsPowerOfTwo", "x is a power of two when x & (x - 1) is zero; the next power of two comes from copying the top bit into every lower bit.", "O(1)", 40,
		func(x uint64, _ Input, _ trace.Recorder) (Output, error) {
			ok := bitwise.IsPowerOfTwo(x)
			next := bitwise.NextPowerOfTwo(x)
			w := max(bits.Len64(x), 1)
			out := Output{Bool: ptr(ok), Lines: []string{"0 is not a power of two"}}
			if x > 0 {
				out.Lines = []string{binary(x, w) + "  x", binary(x-1, w) + "  x - 1", binary(x&(x-1), w) + "  x & (x - 1)"}
			}
			out.Lines = append(out.Lines, fmt.Sprintf("next power of two: %d (%b)", next, next))
			if next != 0 && next <= math.MaxInt64 {
				out.Number = ptr(int(next))
			}
			return out, nil
		})
	gray := num("gray-code", "Gray Code", "gray.go", "GraySequence", "Number the values so neighbors differ in one bit: x ^ (x >> 1), decoded by XOR-folding the shifts.", "O(2^bits)", 3,
		func(b uint64, _ Input, _ trace.Recorder) (Output, error) {
			if b > 8 {
				return Output{}, inputError("give a width of 0 to 8 bits")
			}
			seq := bitwise.GraySequence(int(b))
			out := Output{}
			w := max(int(b), 1)
			for i, g := range seq {
				out.Values = append(out.Values, int(g))
				out.Lines = append(out.Lines, fmt.Sprintf("%3d  %s  gray %s  decodes to %d", i, binary(uint64(i), w), binary(g, w), bitwise.GrayDecode(g)))
			}
			return out, nil
		})
	gray.Limits = "a width of 0 to 8 bits"
	rev := num("reverse-bits", "Reverse Bits", "tricks.go", "ReverseBits", "Reverse the lowest bits of x, moving one bit at a time from the bottom of x to the bottom of the result.", "O(width)", 11,
		func(x uint64, in Input, _ trace.Recorder) (Output, error) {
			w := in.opt("width", 8)
			if w < 1 || w > 62 {
				return Output{}, inputError("the width is 1 to 62")
			}
			y := bitwise.ReverseBits(x, w)
			return Output{Number: ptr(int(y)), Lines: []string{binary(x&(1<<w-1), w) + "  x", binary(y, w) + "  reversed"}}, nil
		})
	rev.Options = []Option{{"width", "width in bits", 8, 1, 62}}
	rev.Example.Options = map[string]int{"width": 8}
	rev.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("bitwise.ReverseBits(%d, %d)", firstOr(in.Values, 0), in.opt("width", 8))
	}
	subsets := num("bit-subsets", "Submask Enumeration", "tricks.go", "BitSubsets", "Every submask of a mask with sub = (sub - 1) & mask, from the mask itself down to 0; the core of subset dynamic programming.", "O(2^k) for k set bits", 13,
		func(x uint64, _ Input, _ trace.Recorder) (Output, error) {
			if bits.OnesCount64(x) > 10 {
				return Output{}, inputError("give a mask with at most 10 bits set")
			}
			out := Output{}
			w := max(bits.Len64(x), 1)
			for _, s := range bitwise.BitSubsets(x) {
				out.Values = append(out.Values, int(s))
				out.Lines = append(out.Lines, binary(s, w))
			}
			out.Number = ptr(len(out.Values))
			return out, nil
		})
	subsets.Limits = "at most 10 bits set"

	single := entry("single-number", "Single Number (XOR)", "math/bitwise/tricks.go", "SingleNumber")
	single.Summary, single.Time, single.Space, single.Input = "The one value without a pair: XOR everything, and every pair cancels itself (v ^ v = 0).", "O(n)", "O(1)", Ints
	single.Example = &Input{Values: []int{4, 1, 2, 1, 2, 7, 7}}
	single.Run = func(in Input, _ trace.Recorder) (Output, error) {
		if err := values(in, "search", 100_000); err != nil {
			return Output{}, err
		}
		out := Output{Number: ptr(bitwise.SingleNumber(in.Values))}
		x := 0
		for i, v := range in.Values {
			x ^= v
			if i < 30 {
				out.Lines = append(out.Lines, fmt.Sprintf("^ %d = %d", v, x))
			}
		}
		return out, nil
	}
	single.Call = func(in Input) (string, string) { return "", "bitwise.SingleNumber(" + ints(in.Values) + ")" }
	return []Entry{pop, pow, gray, rev, single, subsets}
}

// tee records to two recorders.
type tee struct{ a, b trace.Recorder }

func (t tee) Record(e trace.Event) { t.a.Record(e); t.b.Record(e) }

// floats writes ints as a Go []float64 literal.
func floats(v []int) string { return "[]float64{" + strings.TrimPrefix(ints(v), "[]int{") }

func toFloats(v []int) []float64 {
	out := make([]float64, len(v))
	for i, x := range v {
		out[i] = float64(x)
	}
	return out
}

// polyText writes coefficients, constant first, as a polynomial.
func polyText(c []int) string {
	var parts []string
	for i := len(c) - 1; i >= 0; i-- {
		if c[i] == 0 {
			continue
		}
		term := strconv.Itoa(c[i])
		switch i {
		case 0:
		case 1:
			term += "x"
		default:
			term += "x^" + strconv.Itoa(i)
		}
		parts = append(parts, term)
	}
	if len(parts) == 0 {
		return "0"
	}
	return strings.ReplaceAll(strings.Join(parts, " + "), "+ -", "- ")
}

// coefficients checks a polynomial entry's coefficients.
func coefficients(in Input) ([]float64, error) {
	if len(in.Values) == 0 || len(in.Values) > 20 {
		return nil, inputError("give 1 to 20 coefficients, the constant term first")
	}
	return toFloats(in.Values), nil
}

// parseMatrix reads rows of whole numbers, one row a line.
func parseMatrix(text string) ([][]int, error) {
	var m [][]int
	for i, line := range strings.Split(text, "\n") {
		f := splitWords(strings.ReplaceAll(line, "|", " "))
		if len(f) == 0 {
			continue
		}
		row := make([]int, len(f))
		for j, w := range f {
			v, err := strconv.Atoi(w)
			if err != nil {
				return nil, inputError(fmt.Sprintf("line %d: %q is not a whole number", i+1, w))
			}
			row[j] = v
		}
		if len(m) > 0 && len(row) != len(m[0]) {
			return nil, inputError(fmt.Sprintf("line %d: every row needs %d numbers", i+1, len(m[0])))
		}
		m = append(m, row)
	}
	if len(m) == 0 || len(m) > 16 {
		return nil, inputError("give 1 to 16 rows")
	}
	return m, nil
}

// matrixLiteral writes m as a Go [][]int literal.
func matrixLiteral(m [][]int) string {
	rows := make([]string, len(m))
	for i, r := range m {
		rows[i] = "{" + strings.TrimSuffix(strings.TrimPrefix(ints(r), "[]int{"), "}") + "}"
	}
	return "[][]int{" + strings.Join(rows, ", ") + "}"
}

func matrixCall(in Input, call func(m string) string) (string, string) {
	m, err := parseMatrix(in.Text)
	if err != nil {
		return "", ""
	}
	return "", call(matrixLiteral(m))
}

func numericalEntries() []Entry {
	roots := func(id, name, file, fn, summary, time string, opts []Option, solve func(c []float64, in Input) (float64, []float64, error), call func(in Input) string) Entry {
		e := entry(id, name, "math/numerical/"+file, fn)
		e.Summary, e.Time, e.Space, e.Input, e.Options = summary, time, "O(steps)", Ints, opts
		ex := map[string]int{}
		for _, o := range opts {
			ex[o.Name] = o.Default
		}
		e.Example = &Input{Values: []int{-2, 0, 1}, Options: ex}
		e.Run = func(in Input, _ trace.Recorder) (Output, error) {
			c, err := coefficients(in)
			if err != nil {
				return Output{}, err
			}
			root, tried, err := solve(c, in)
			f := numerical.Poly(c)
			out := Output{Lines: []string{"f(x) = " + polyText(in.Values)}}
			for i, x := range tried {
				out.Lines = append(out.Lines, fmt.Sprintf("step %2d  x = %.12g  f(x) = %.3g", i+1, x, f(x)))
			}
			switch {
			case errors.Is(err, numerical.ErrBracket):
				return Output{}, inputError("f(lo) and f(hi) have the same sign, so no root is bracketed")
			case err != nil:
				out.Lines = append(out.Lines, "did not converge: "+err.Error())
			default:
				out.Lines = append(out.Lines, fmt.Sprintf("root %.12g", root))
			}
			out.Summary = map[string]float64{"root": root, "steps": float64(len(tried)), "f(root)": f(root)}
			return out, nil
		}
		e.Call = func(in Input) (string, string) { return "", call(in) }
		return e
	}
	iter := Option{"iterations", "most steps", 60, 1, 1000}
	bis := roots("bisection", "Bisection Method", "roots.go", "Bisection", "Halve an interval whose ends have opposite signs, keeping the half where the sign changes: slow, one bit a step, but sure.", "O(log(width / tol))",
		[]Option{{"lo", "interval start", 0, -1_000_000, 1_000_000}, {"hi", "interval end", 2, -1_000_000, 1_000_000}, iter},
		func(c []float64, in Input) (float64, []float64, error) {
			return numerical.Bisection(numerical.Poly(c), float64(in.opt("lo", 0)), float64(in.opt("hi", 2)), 1e-12, in.opt("iterations", 60))
		},
		func(in Input) string {
			return fmt.Sprintf("numerical.Bisection(numerical.Poly(%s), %d, %d, 1e-12, %d)", floats(in.Values), in.opt("lo", 0), in.opt("hi", 2), in.opt("iterations", 60))
		})
	newton := roots("newton-raphson", "Newton-Raphson Method", "roots.go", "NewtonRaphson", "Follow the tangent line to where it crosses zero, x - f(x)/f'(x); near a root the correct digits double each step.", "O(steps)",
		[]Option{{"x0", "first guess", 1, -1_000_000, 1_000_000}, iter},
		func(c []float64, in Input) (float64, []float64, error) {
			return numerical.NewtonRaphson(numerical.Poly(c), numerical.Poly(numerical.Derivative(c)), float64(in.opt("x0", 1)), 1e-12, in.opt("iterations", 60))
		},
		func(in Input) string {
			c := floats(in.Values)
			return fmt.Sprintf("numerical.NewtonRaphson(numerical.Poly(%s), numerical.Poly(numerical.Derivative(%s)), %d, 1e-12, %d)", c, c, in.opt("x0", 1), in.opt("iterations", 60))
		})
	secant := roots("secant-method", "Secant Method", "roots.go", "Secant", "Newton's method with the derivative replaced by the slope through the last two guesses; no derivative needed.", "O(steps)",
		[]Option{{"x0", "first guess", 1, -1_000_000, 1_000_000}, {"x1", "second guess", 2, -1_000_000, 1_000_000}, iter},
		func(c []float64, in Input) (float64, []float64, error) {
			return numerical.Secant(numerical.Poly(c), float64(in.opt("x0", 1)), float64(in.opt("x1", 2)), 1e-12, in.opt("iterations", 60))
		},
		func(in Input) string {
			return fmt.Sprintf("numerical.Secant(numerical.Poly(%s), %d, %d, 1e-12, %d)", floats(in.Values), in.opt("x0", 1), in.opt("x1", 2), in.opt("iterations", 60))
		})

	integrate := func(id, name, fn, summary, time string, rule func(f func(float64) float64, a, b float64, n int) float64) Entry {
		e := entry(id, name, "math/numerical/integrate.go", fn)
		e.Summary, e.Time, e.Space, e.Input = summary, time, "O(1)", Ints
		e.Options = []Option{{"a", "from", 0, -10_000, 10_000}, {"b", "to", 2, -10_000, 10_000}, {"n", "slices", 4, 1, 1_000_000}}
		e.Example = &Input{Values: []int{1, 0, -3, 2}, Options: map[string]int{"a": 0, "b": 2, "n": 4}}
		e.Run = func(in Input, _ trace.Recorder) (Output, error) {
			c, err := coefficients(in)
			if err != nil {
				return Output{}, err
			}
			a, b, n := float64(in.opt("a", 0)), float64(in.opt("b", 2)), in.opt("n", 4)
			if n < 1 || n > 1_000_000 {
				return Output{}, inputError("give 1 to 1000000 slices")
			}
			got := rule(numerical.Poly(c), a, b, n)
			exact := 0.0
			for i, ci := range c {
				p := float64(i + 1)
				exact += ci * (math.Pow(b, p) - math.Pow(a, p)) / p
			}
			out := Output{Lines: []string{
				fmt.Sprintf("integral of %s from %g to %g", polyText(in.Values), a, b),
				fmt.Sprintf("%s with %d slices: %.12g", name, n, got),
				fmt.Sprintf("exact (from the antiderivative): %.12g", exact),
			}}
			out.Summary = map[string]float64{"estimate": got, "exact": exact, "error": math.Abs(got - exact)}
			return out, nil
		}
		e.Call = func(in Input) (string, string) {
			return "", fmt.Sprintf("numerical.%s(numerical.Poly(%s), %d, %d, %d)", fn, floats(in.Values), in.opt("a", 0), in.opt("b", 2), in.opt("n", 4))
		}
		return e
	}

	gauss := entry("gaussian-elimination", "Gaussian Elimination", "math/numerical/gauss.go", "GaussianElimination")
	gauss.Summary, gauss.Time, gauss.Space, gauss.Input = "Solve A x = b: eliminate below each pivot, swapping in the largest pivot first (partial pivoting), then substitute back.", "O(n^3)", "O(n^2)", Text
	gauss.Example = &Input{Text: "2 1 -1 | 8\n-3 -1 2 | -11\n-2 1 2 | -3"}
	gauss.Run = func(in Input, _ trace.Recorder) (Output, error) {
		m, err := parseMatrix(in.Text)
		if err != nil {
			return Output{}, err
		}
		if len(m[0]) != len(m)+1 {
			return Output{}, inputError("give n rows of n + 1 numbers: the coefficients, then the right-hand side")
		}
		a := make([][]float64, len(m))
		b := make([]float64, len(m))
		for i, row := range m {
			a[i], b[i] = toFloats(row[:len(row)-1]), float64(row[len(row)-1])
		}
		x, steps, err := numerical.GaussianEliminationSteps(a, b)
		if err != nil {
			return Output{}, inputError("the matrix is singular: no unique solution")
		}
		out := Output{Summary: map[string]float64{}}
		for _, s := range steps {
			out.Lines = append(out.Lines, s.Note)
			for _, row := range s.Matrix {
				cells := make([]string, len(row))
				for j, v := range row {
					cells[j] = fmt.Sprintf("%8.4g", v)
				}
				out.Lines = append(out.Lines, "  "+strings.Join(cells[:len(row)-1], " ")+" | "+cells[len(row)-1])
			}
		}
		for i, v := range x {
			name := fmt.Sprintf("x%d", i+1)
			out.Summary[name] = v
			out.Lines = append(out.Lines, fmt.Sprintf("%s = %.10g", name, v))
		}
		return out, nil
	}

	gauss.Call = func(in Input) (string, string) {
		m, err := parseMatrix(in.Text)
		if err != nil || len(m[0]) != len(m)+1 {
			return "", ""
		}
		rows := make([]string, len(m))
		var b []int
		for i, row := range m {
			rows[i] = "{" + strings.TrimPrefix(floats(row[:len(row)-1]), "[]float64{")
			b = append(b, row[len(row)-1])
		}
		return "", fmt.Sprintf("numerical.GaussianElimination([][]float64{%s}, %s)", strings.Join(rows, ", "), floats(b))
	}

	matPow := entry("matrix-power", "Matrix Power", "math/numerical/matrix.go", "MatrixPower")
	matPow.Summary, matPow.Time, matPow.Space, matPow.Input = "A square matrix to the kth power by repeated squaring: O(log k) products. [[1 1] [1 0]]^n holds Fibonacci numbers.", "O(n^3 log k)", "O(n^2)", Text
	matPow.Options = []Option{{"k", "power", 10, 0, 1_000_000}}
	matPow.Example = &Input{Text: "1 1\n1 0", Options: map[string]int{"k": 10}}
	matPow.Run = func(in Input, _ trace.Recorder) (Output, error) {
		m, err := parseMatrix(in.Text)
		if err != nil {
			return Output{}, err
		}
		p, err := numerical.MatrixPower(m, in.opt("k", 10))
		if err != nil {
			return Output{}, inputError("give a square matrix and k >= 0")
		}
		k := in.opt("k", 10)
		return Output{Table: tableOf(p, nil, nil), Lines: []string{
			fmt.Sprintf("k = %d is %b in binary: square once per bit, multiply in once per set bit", k, k),
			"entries wrap around on overflow",
		}}, nil
	}
	matPow.Call = func(in Input) (string, string) {
		return matrixCall(in, func(m string) string { return fmt.Sprintf("numerical.MatrixPower(%s, %d)", m, in.opt("k", 10)) })
	}

	strassen := entry("strassen", "Strassen Multiplication", "math/numerical/matrix.go", "Strassen")
	strassen.Summary, strassen.Time, strassen.Space, strassen.Input = "Multiply square matrices with seven block products instead of eight, recursively: under n^3.", "O(n^2.81)", "O(n^2)", TwoTexts
	strassen.Example = &Input{Text: "1 2 3\n4 5 6\n7 8 9", Other: "9 8 7\n6 5 4\n3 2 1"}
	strassen.Run = func(in Input, _ trace.Recorder) (Output, error) {
		a, err := parseMatrix(in.Text)
		if err != nil {
			return Output{}, err
		}
		b, err := parseMatrix(in.Other)
		if err != nil {
			return Output{}, err
		}
		c, err := numerical.Strassen(a, b)
		if err != nil {
			return Output{}, inputError("give two square matrices of the same size")
		}
		size := 1
		for size < len(a) {
			size <<= 1
		}
		return Output{Table: tableOf(c, nil, nil), Lines: []string{fmt.Sprintf("padded to %d x %d; direct multiplication gives the same", size, size)}}, nil
	}
	strassen.Call = func(in Input) (string, string) {
		a, err1 := parseMatrix(in.Text)
		b, err2 := parseMatrix(in.Other)
		if err1 != nil || err2 != nil {
			return "", ""
		}
		return "", fmt.Sprintf("numerical.Strassen(%s, %s)", matrixLiteral(a), matrixLiteral(b))
	}

	fft := entry("fft", "Fast Fourier Transform", "math/numerical/fft.go", "FFT")
	fft.Summary, fft.Time, fft.Space, fft.Input = "The frequencies in a signal, from the transforms of its even and odd positions combined with twiddle factors (Cooley-Tukey radix 2).", "O(n log n)", "O(n log n)", Ints
	fft.Example, fft.Limits = &Input{Values: []int{1, 1, 1, 1, 0, 0, 0, 0}}, "at most 1024 values, padded with zeros to a power of two"
	fft.Run = func(in Input, _ trace.Recorder) (Output, error) {
		if err := values(in, "transform", 1024); err != nil {
			return Output{}, err
		}
		size := 1
		for size < len(in.Values) {
			size <<= 1
		}
		x := make([]complex128, size)
		for i, v := range in.Values {
			x[i] = complex(float64(v), 0)
		}
		y, _ := numerical.FFT(x)
		out := Output{}
		for k, v := range y {
			if k < 64 {
				out.Lines = append(out.Lines, fmt.Sprintf("X[%d] = %8.4f %+8.4fi   |X| = %.4f", k, real(v), imag(v), math.Hypot(real(v), imag(v))))
			}
		}
		if size > len(in.Values) {
			out.Lines = append(out.Lines, fmt.Sprintf("padded with %d zeros to %d values", size-len(in.Values), size))
		}
		return out, nil
	}
	fft.Call = func(in Input) (string, string) {
		lit := "[]complex128{" + strings.TrimPrefix(ints(in.Values), "[]int{")
		return "x := " + lit + "\n\t", "numerical.FFT(x)"
	}

	horner := entry("horner", "Horner's Rule", "math/numerical/numerical.go", "Horner")
	horner.Summary, horner.Time, horner.Space, horner.Input = "Evaluate a polynomial as a + x(b + x(c + ...)): one multiply and one add per coefficient, no powers.", "O(n)", "O(1)", Ints
	horner.Options = []Option{{"x", "x", 3, -1_000_000, 1_000_000}}
	horner.Example = &Input{Values: []int{5, -3, 0, 2}, Options: map[string]int{"x": 3}}
	horner.Run = func(in Input, _ trace.Recorder) (Output, error) {
		c, err := coefficients(in)
		if err != nil {
			return Output{}, err
		}
		x := float64(in.opt("x", 3))
		out := Output{Lines: []string{fmt.Sprintf("p(x) = %s at x = %g", polyText(in.Values), x)}}
		y := 0.0
		for i := len(c) - 1; i >= 0; i-- {
			prev := y
			y = y*x + c[i]
			out.Lines = append(out.Lines, fmt.Sprintf("%g * %g + %g = %g", prev, x, c[i], y))
		}
		out.Summary = map[string]float64{"p(x)": numerical.Horner(c, x)}
		if y == math.Trunc(y) && math.Abs(y) < 1<<53 {
			out.Number = ptr(int(y))
		}
		return out, nil
	}
	horner.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("numerical.Horner(%s, %d)", floats(in.Values), in.opt("x", 3))
	}

	return []Entry{bis, newton, secant,
		integrate("simpson-integration", "Simpson's Rule", "Simpson", "Integrate by fitting a parabola through each pair of slices (weights 1, 4, 2, ..., 4, 1); exact for cubics.", "O(n)", numerical.Simpson),
		integrate("trapezoid-rule", "Trapezoid Rule", "Trapezoid", "Integrate by joining neighboring points with straight lines and adding the trapezoids' areas; exact for lines.", "O(n)", numerical.Trapezoid),
		gauss, matPow, strassen, fft, horner}
}

// parseEdges reads an undirected graph, one edge a line ("A B"), as node
// names and index pairs.
func parseEdges(text string) ([]string, [][2]int, error) {
	index := map[string]int{}
	var names []string
	node := func(s string) int {
		if i, ok := index[s]; ok {
			return i
		}
		index[s] = len(names)
		names = append(names, s)
		return len(names) - 1
	}
	var edges [][2]int
	for i, line := range strings.Split(text, "\n") {
		f := splitWords(line)
		if len(f) == 0 {
			continue
		}
		if len(f) < 2 || len(f) > 3 {
			return nil, nil, inputError(fmt.Sprintf("line %d: want an edge \"A B\"", i+1))
		}
		edges = append(edges, [2]int{node(f[0]), node(f[1])})
	}
	if len(names) < 2 || len(names) > 60 || len(edges) > 600 {
		return nil, nil, inputError("give 2 to 60 nodes and at most 600 edges")
	}
	return names, edges, nil
}

func edgesLiteral(edges [][2]int) string {
	parts := make([]string, len(edges))
	for i, e := range edges {
		parts[i] = fmt.Sprintf("{%d, %d}", e[0], e[1])
	}
	return "[][2]int{" + strings.Join(parts, ", ") + "}"
}

func randomizedEntries() []Entry {
	res := entry("reservoir-sampling", "Reservoir Sampling", "math/randomized/randomized.go", "ReservoirSampling")
	res.Summary, res.Time, res.Space, res.Input = "Choose k values uniformly from a stream read once: keep the first k, then let the ith replace a random kept one with probability k/i.", "O(n)", "O(k)", Ints
	res.Options = []Option{{"k", "sample size", 3, 1, 100_000}, seedOption}
	res.Example = &Input{Values: []int{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}, Options: map[string]int{"k": 3, "seed": 1}}
	res.Run = func(in Input, _ trace.Recorder) (Output, error) {
		if err := values(in, "sample", 100_000); err != nil {
			return Output{}, err
		}
		k := in.opt("k", 3)
		if k < 1 {
			return Output{}, inputError("k is at least 1")
		}
		s := randomized.ReservoirSampling(in.Values, k, int64(in.opt("seed", 1)))
		return Output{Values: s, Lines: []string{fmt.Sprintf("each value is kept with probability %d/%d", min(k, len(in.Values)), len(in.Values))}}, nil
	}
	res.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("randomized.ReservoirSampling(%s, %d, %d)", ints(in.Values), in.opt("k", 3), in.opt("seed", 1))
	}

	pi := entry("monte-carlo-pi", "Monte Carlo Pi", "math/randomized/montecarlo.go", "MonteCarloPi")
	pi.Summary, pi.Time, pi.Space, pi.Input, pi.Traced = "Throw random points into a square and count those inside the quarter circle: pi is about 4 * inside / all. The error shrinks like 1/sqrt(n).", "O(n)", "O(1)", Simulation, true
	pi.Options = []Option{{"points", "points", 1000, 1, 1_000_000}, seedOption}
	pi.Example, pi.Limits = &Input{Options: map[string]int{"points": 1000, "seed": 1}}, "at most 1000000 points, the first 2000 drawn"
	pi.Run = func(in Input, r trace.Recorder) (Output, error) {
		n := in.opt("points", 1000)
		if n < 1 || n > 1_000_000 {
			return Output{}, inputError("give 1 to 1000000 points")
		}
		est, samples := randomized.MonteCarloPi(n, int64(in.opt("seed", 1)))
		out := Output{Points: make([]Point, len(samples))}
		inside := 0
		for i, s := range samples {
			out.Points[i] = Point{X: int(s.X * 1000), Y: int(s.Y * 1000)}
			if s.Inside {
				inside++
				r.Record(trace.Event{Kind: trace.Mark, Indices: []int{i}, Note: "inside"})
			}
		}
		out.Summary = map[string]float64{"estimate": est, "error": math.Abs(est - math.Pi), "points": float64(n)}
		out.Lines = []string{fmt.Sprintf("pi is about %.6f after %d points (%d of the %d drawn fell inside)", est, n, inside, len(samples))}
		return out, nil
	}
	pi.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("randomized.MonteCarloPi(%d, %d)", in.opt("points", 1000), in.opt("seed", 1))
	}

	karger := entry("karger-min-cut", "Karger's Minimum Cut", "math/randomized/karger.go", "KargerMinCut")
	karger.Summary, karger.Time, karger.Space, karger.Input, karger.Traced = "Merge the ends of random edges until two groups remain; the edges between them are a cut, and the smallest over many trials is probably the minimum.", "O(trials * m)", "O(n + m)", GraphInput, true
	karger.Options = []Option{{"trials", "trials", 50, 1, 10_000}, seedOption}
	karger.Example = &Input{Text: "A B\nA C\nA D\nB C\nB D\nC D\nD E\nC F\nE F\nE G\nE H\nF G\nF H\nG H", Options: map[string]int{"trials": 50, "seed": 1}}
	karger.Run = func(in Input, r trace.Recorder) (Output, error) {
		names, edges, err := parseEdges(in.Text)
		if err != nil {
			return Output{}, err
		}
		trials := in.opt("trials", 50)
		if trials < 1 || trials > 10_000 {
			return Output{}, inputError("give 1 to 10000 trials")
		}
		cut, side, err := randomized.KargerMinCut(len(names), edges, trials, int64(in.opt("seed", 1)))
		if err != nil {
			return Output{}, wrap(err)
		}
		out := Output{Number: ptr(cut), Graph: &GraphOutput{Nodes: names}}
		for i, s := range side {
			g := 1
			if s {
				g = 0
			}
			r.Record(trace.Event{Kind: trace.Mark, Indices: []int{i}, Values: []int{g}, Note: "group"})
		}
		for _, e := range edges {
			out.Graph.Edges = append(out.Graph.Edges, GraphEdge{From: e[0], To: e[1], Weight: 1})
			if side[e[0]] != side[e[1]] {
				out.Lines = append(out.Lines, fmt.Sprintf("cut edge %s - %s", names[e[0]], names[e[1]]))
			}
		}
		out.Lines = append(out.Lines, fmt.Sprintf("smallest cut found in %d trials: %d edges", trials, cut))
		return out, nil
	}
	karger.Call = func(in Input) (string, string) {
		names, edges, err := parseEdges(in.Text)
		if err != nil {
			return "", ""
		}
		return "", fmt.Sprintf("randomized.KargerMinCut(%d, %s, %d, %d)", len(names), edgesLiteral(edges), in.opt("trials", 50), in.opt("seed", 1))
	}

	rq := entry("randomized-quickselect", "Randomized Quickselect", "math/randomized/quickselect.go", "RandomizedQuickselect")
	rq.Summary, rq.Time, rq.Space, rq.Input = "The kth smallest value by quickselect around random pivots: always exact (Las Vegas), linear time expected for every input.", "O(n) expected", "O(n)", Ints
	rq.Options = []Option{{"k", "k (0 is the smallest)", 5, 0, 99_999}, seedOption}
	rq.Example = &Input{Values: []int{38, 27, 43, 3, 9, 82, 10, 27, 55, 1, 64, 19}, Options: map[string]int{"k": 5, "seed": 1}}
	rq.Run = func(in Input, _ trace.Recorder) (Output, error) {
		if err := values(in, "search", 100_000); err != nil {
			return Output{}, err
		}
		k := in.opt("k", 5)
		v, ok := randomized.RandomizedQuickselect(in.Values, k, int64(in.opt("seed", 1)))
		if !ok {
			return Output{}, inputError(fmt.Sprintf("k is 0 to %d", len(in.Values)-1))
		}
		return Output{Number: ptr(v)}, nil
	}
	rq.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("randomized.RandomizedQuickselect(%s, %d, %d)", ints(in.Values), in.opt("k", 5), in.opt("seed", 1))
	}
	return []Entry{res, pi, karger, rq}
}
