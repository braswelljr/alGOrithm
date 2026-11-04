package core

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/algorithms/dynamic"
	"github.com/braswelljr/alGOrithm/pkg/algorithms/strmatch"
	"github.com/braswelljr/alGOrithm/pkg/math/numbers"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// moreAlgorithmEntries are the second set of dynamic-programming, string and
// number entries.
func moreAlgorithmEntries() []Entry {
	return append(append(moreDynamicEntries(), moreStrmatchEntries()...), moreNumberEntries()...)
}

// splitWords reads words separated by spaces, commas or new lines.
func splitWords(s string) []string {
	return strings.FieldsFunc(s, func(r rune) bool { return r == ' ' || r == ',' || r == '\n' || r == '\t' || r == '\r' })
}

// parsePairs reads "a b" pairs, one a line, as two lists.
func parsePairs(text string) ([]int, []int, error) {
	var a, b []int
	for i, line := range strings.Split(text, "\n") {
		f := splitWords(line)
		if len(f) == 0 {
			continue
		}
		if len(f) != 2 {
			return nil, nil, inputError(fmt.Sprintf("line %d: want two whole numbers", i+1))
		}
		x, err1 := strconv.Atoi(f[0])
		y, err2 := strconv.Atoi(f[1])
		if err1 != nil || err2 != nil {
			return nil, nil, inputError(fmt.Sprintf("line %d: want two whole numbers", i+1))
		}
		a, b = append(a, x), append(b, y)
	}
	if len(a) == 0 {
		return nil, nil, inputError("give at least one line")
	}
	return a, b, nil
}

func tableOf(t [][]int, rowLabels, colLabels []string) *TableOutput {
	if len(t) == 0 {
		return &TableOutput{}
	}
	return &TableOutput{Rows: len(t), Cols: len(t[0]), Cells: flatten(t), RowLabels: rowLabels, ColLabels: colLabels}
}

func moreDynamicEntries() []Entry {
	knap := entry("knapsack-01", "0/1 Knapsack", "algorithms/dynamic/knapsack.go", "Knapsack01")
	knap.Summary, knap.Time, knap.Space, knap.Input, knap.Traced = "Pick items, each at most once, for the most value within a weight limit, filling a table of best values per item and capacity.", "O(n * capacity)", "O(n * capacity)", Text, true
	knap.Options = []Option{{"capacity", "capacity", 10, 0, 100_000}}
	knap.Example = &Input{Text: "weight value\n1 1\n3 4\n4 5\n5 7", Options: map[string]int{"capacity": 7}}
	knapParse := func(in Input) ([]int, []int, error) {
		text := in.Text
		if lines := strings.SplitN(text, "\n", 2); len(lines) == 2 && strings.Contains(strings.ToLower(lines[0]), "weight") {
			text = lines[1] // a header row
		}
		return parsePairs(text)
	}
	knap.Run = func(in Input, r trace.Recorder) (Output, error) {
		w, v, err := knapParse(in)
		if err != nil {
			return Output{}, err
		}
		best, chosen, t, err := dynamic.TraceKnapsack01(w, v, in.opt("capacity", 10), r)
		if err != nil {
			return Output{}, wrap(err)
		}
		rows := []string{"none"}
		for i := range w {
			rows = append(rows, fmt.Sprintf("item %d (w%d v%d)", i+1, w[i], v[i]))
		}
		return Output{Number: &best, Values: chosen, Table: tableOf(t, rows, numberLabels(len(t[0]), 0))}, nil
	}
	knap.Call = func(in Input) (string, string) {
		w, v, _ := knapParse(in)
		return "", fmt.Sprintf("dynamic.Knapsack01(%s, %s, %d)", ints(w), ints(v), in.opt("capacity", 10))
	}

	mc := entry("matrix-chain", "Matrix Chain Multiplication", "algorithms/dynamic/matrixchain.go", "MatrixChain")
	mc.Summary, mc.Time, mc.Space, mc.Input, mc.Traced = "The cheapest order to multiply a chain of matrices (matrix i is d[i] x d[i+1]), trying every split of every sub-chain.", "O(n^3)", "O(n^2)", Ints, true
	mc.Example = &Input{Values: []int{40, 20, 30, 10, 30}}
	mc.Run = func(in Input, r trace.Recorder) (Output, error) {
		cost, order, t, err := dynamic.TraceMatrixChain(in.Values, r)
		if err != nil {
			return Output{}, inputError("give 2 to 201 dimensions, all at least 1")
		}
		names := make([]string, len(t))
		for i := range names {
			names[i] = "A" + strconv.Itoa(i+1)
		}
		return Output{Number: &cost, Text: order, Table: tableOf(t, names, names)}, nil
	}
	mc.Call = func(in Input) (string, string) { return "", "dynamic.MatrixChain(" + ints(in.Values) + ")" }

	rod := entry("rod-cutting", "Rod Cutting", "algorithms/dynamic/rodcutting.go", "RodCutting")
	rod.Summary, rod.Time, rod.Space, rod.Input, rod.Traced = "Cut a rod into pieces for the most money, where piece length i sells for price i, from the best answer for every shorter rod.", "O(n^2)", "O(n)", Ints, true
	rod.Example = &Input{Values: []int{1, 5, 8, 9, 10, 17, 17, 20}}
	rod.Run = func(in Input, r trace.Recorder) (Output, error) {
		best, cuts, t, err := dynamic.TraceRodCutting(in.Values, r)
		if err != nil {
			return Output{}, inputError("give 1 to 10000 prices, the price of each length from 1 up")
		}
		return Output{Number: &best, Values: cuts, Table: &TableOutput{Rows: 1, Cols: len(t), Cells: t, ColLabels: numberLabels(len(t), 0)}}, nil
	}
	rod.Call = func(in Input) (string, string) { return "", "dynamic.RodCutting(" + ints(in.Values) + ")" }

	sub := entry("subset-sum", "Subset Sum", "algorithms/dynamic/subsetsum.go", "SubsetSum")
	sub.Summary, sub.Time, sub.Space, sub.Input, sub.Traced, sub.Params = "Whether some of the values add up to exactly the target, from which sums the first i values can make.", "O(n * target)", "O(n * target)", Amount, true, []string{"amount"}
	sub.Example = &Input{Values: []int{3, 34, 4, 12, 5, 2}, Target: 9}
	sub.Run = func(in Input, r trace.Recorder) (Output, error) {
		ok, chosen, t, err := dynamic.TraceSubsetSum(in.Values, in.Target, r)
		if err != nil {
			return Output{}, inputError("give non-negative values and a target from 0 up (table at most 1000000 cells)")
		}
		rows := []string{"none"}
		for _, v := range in.Values {
			rows = append(rows, "+"+strconv.Itoa(v))
		}
		return Output{Bool: &ok, Values: chosen, Table: tableOf(t, rows, numberLabels(in.Target+1, 0))}, nil
	}
	sub.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("dynamic.SubsetSum(%s, %d)", ints(in.Values), in.Target)
	}

	lps := entry("longest-palindromic-subsequence", "Longest Palindromic Subsequence", "algorithms/dynamic/palindrome.go", "LongestPalindromicSubsequence")
	lps.Summary, lps.Time, lps.Space, lps.Input, lps.Traced = "The longest subsequence that reads the same backward, from the answer for every inner range.", "O(n^2)", "O(n^2)", Text, true
	lps.Example = &Input{Text: "character"}
	lps.Run = func(in Input, r trace.Recorder) (Output, error) {
		s, t, err := dynamic.TraceLongestPalindromicSubsequence(in.Text, r)
		if err != nil || in.Text == "" {
			return Output{}, inputError("give a word of 1 to 1000 letters")
		}
		letters := labelsOf(in.Text)[1:]
		return Output{Text: s, Number: ptr(len(s)), Table: tableOf(t, letters, letters)}, nil
	}
	lps.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("dynamic.LongestPalindromicSubsequence(%q)", in.Text)
	}

	up := entry("unique-paths", "Unique Paths", "algorithms/dynamic/uniquepaths.go", "UniquePaths")
	up.Summary, up.Time, up.Space, up.Input, up.Traced = "Count the paths from the top-left to the bottom-right moving right or down around obstacles (#): each cell sums the cells above and to its left.", "O(rows * cols)", "O(rows * cols)", Text, true
	up.Example = &Input{Text: "....\n.#..\n...#\n#...\n...."}
	up.Run = func(in Input, r trace.Recorder) (Output, error) {
		n, t, err := dynamic.TraceUniquePaths(in.Text, r)
		if err != nil {
			return Output{}, inputError("give a grid of up to 30 x 30 cells, # for an obstacle")
		}
		return Output{Number: &n, Table: tableOf(t, nil, nil)}, nil
	}
	up.Call = func(in Input) (string, string) { return "", "dynamic.UniquePaths(" + "`" + in.Text + "`" + ")" }

	egg := entry("egg-drop", "Egg Drop", "algorithms/dynamic/eggdrop.go", "EggDrop")
	egg.Summary, egg.Time, egg.Space, egg.Input, egg.Traced = "The fewest drops that always find the highest safe floor with a given number of eggs, from how many floors t drops can cover.", "O(eggs * drops)", "O(eggs * drops)", TwoNumbers, true
	egg.Example = &Input{Values: []int{2, 36}}
	egg.Run = func(in Input, r trace.Recorder) (Output, error) {
		e, f, err := two(in)
		if err != nil {
			return Output{}, inputError("give the eggs and the floors")
		}
		n, t, err := dynamic.TraceEggDrop(e, f, r)
		if err != nil {
			return Output{}, inputError("give 1 to 64 eggs and 0 to 1000000 floors")
		}
		cols := numberLabels(e+1, 0)
		for i := range cols {
			cols[i] += " eggs"
		}
		rows := numberLabels(len(t), 0)
		for i := range rows {
			rows[i] += " drops"
		}
		return Output{Number: &n, Table: tableOf(t, rows, cols)}, nil
	}
	egg.Call = func(in Input) (string, string) {
		a, b := firstOr(in.Values, 1), 0
		if len(in.Values) > 1 {
			b = in.Values[1]
		}
		return "", fmt.Sprintf("dynamic.EggDrop(%d, %d)", a, b)
	}

	lf := entry("lis-fast", "Longest Increasing Subsequence (n log n)", "algorithms/dynamic/lisfast.go", "LISFast")
	lf.Summary, lf.Time, lf.Space, lf.Input, lf.Traced = "Patience sorting: keep the smallest tail of every run length and binary search where each value goes.", "O(n log n)", "O(n)", Ints, true
	lf.Example = &Input{Values: []int{10, 9, 2, 5, 3, 7, 101, 18, 4, 20}}
	lf.Run = func(in Input, r trace.Recorder) (Output, error) {
		if err := values(in, "scan", 100_000); err != nil {
			return Output{}, err
		}
		n, seq, tails := dynamic.TraceLISFast(in.Values, r)
		return Output{Initial: append([]int(nil), in.Values...), Values: seq, Number: &n,
			Table: &TableOutput{Rows: 1, Cols: len(tails), Cells: tails, ColLabels: numberLabels(len(tails), 1)}}, nil
	}
	lf.Call = func(in Input) (string, string) { return "", "dynamic.LISFast(" + ints(in.Values) + ")" }

	return []Entry{knap, mc, rod, sub, lps, up, egg, lf}
}

func moreStrmatchEntries() []Entry {
	ex := &Input{Text: "here is a simple example of a pattern in a sample text", Other: "ample"}
	match := func(id, name, file, fn, summary, time, space string, traced func(string, string, trace.Recorder) []int) Entry {
		e := entry(id, name, "algorithms/strmatch/"+file, fn)
		e.Summary, e.Time, e.Space, e.Input, e.Traced, e.Params, e.Example = summary, time, space, Pattern, true, []string{"pattern"}, ex
		e.Run = func(in Input, r trace.Recorder) (Output, error) {
			if in.Text == "" || in.Other == "" || len(in.Text) > 10_000 {
				return Output{}, inputError("give a text of up to 10000 characters and a pattern")
			}
			m := traced(in.Text, in.Other, r)
			if m == nil {
				m = []int{}
			}
			return Output{Text: in.Text, Matches: m, Number: ptr(len(m))}, nil
		}
		e.Call = func(in Input) (string, string) { return "", fmt.Sprintf("strmatch.%s(%q, %q)", fn, in.Text, in.Other) }
		return e
	}
	z := match("z-algorithm", "Z-Algorithm", "zalgorithm.go", "ZSearch", "The Z array of pattern + separator + text: wherever a value reaches the pattern's length, it matches; the Z-box reuses earlier work.", "O(n + m)", "O(n + m)", strmatch.TraceZSearch)
	z.Func = "ZSearch"

	aho := entry("aho-corasick", "Aho-Corasick", "algorithms/strmatch/ahocorasick.go", "AhoCorasick")
	aho.Summary, aho.Time, aho.Space, aho.Input, aho.Traced, aho.Params = "Find many patterns in one pass: a trie of the patterns with failure links, so a mismatch falls back without re-reading the text (1975).", "O(n + patterns + matches)", "O(total pattern length)", Pattern, true, []string{"pattern"}
	aho.Example = &Input{Text: "ushers say she sells seashells", Other: "he she his hers sea"}
	aho.Run = func(in Input, r trace.Recorder) (Output, error) {
		pats := splitWords(in.Other)
		if in.Text == "" || len(pats) == 0 || len(in.Text) > 10_000 || len(pats) > 100 {
			return Output{}, inputError("give a text of up to 10000 characters and 1 to 100 patterns")
		}
		ms := strmatch.TraceAhoCorasick(in.Text, pats, r)
		out := Output{Text: in.Text, Matches: []int{}, Number: ptr(len(ms))}
		for _, m := range ms {
			out.Matches = append(out.Matches, m.Start)
			out.Lines = append(out.Lines, fmt.Sprintf("%q at %d", pats[m.Pattern], m.Start))
		}
		out.Steps = []Step{{Op: Op{Name: "trie", Value: strings.Join(pats, " ")}, Result: "the patterns' trie, with failure links", Snapshot: strmatch.AhoCorasickTrie(pats)}}
		return out, nil
	}
	aho.Call = func(in Input) (string, string) {
		q := make([]string, 0)
		for _, p := range splitWords(in.Other) {
			q = append(q, strconv.Quote(p))
		}
		return "", fmt.Sprintf("strmatch.AhoCorasick(%q, []string{%s})", in.Text, strings.Join(q, ", "))
	}

	man := entry("manacher", "Manacher's Algorithm", "algorithms/strmatch/manacher.go", "Manacher")
	man.Summary, man.Time, man.Space, man.Input, man.Traced = "The longest palindromic substring in linear time: grow each center's radius from its mirror inside the rightmost palindrome so far (1975).", "O(n)", "O(n)", Text, true
	man.Example = &Input{Text: "forgeeksskeegfor"}
	man.Run = func(in Input, r trace.Recorder) (Output, error) {
		if in.Text == "" || len(in.Text) > 10_000 {
			return Output{}, inputError("give a text of 1 to 10000 characters")
		}
		start, n, radius := strmatch.TraceManacher(in.Text, r)
		labels := make([]string, len(radius))
		for i := range labels {
			labels[i] = "#"
			if i%2 == 1 {
				labels[i] = string(in.Text[i/2])
			}
		}
		return Output{Text: in.Text, Matches: []int{start}, Number: &n, Lines: []string{in.Text[start : start+n]},
			Table: &TableOutput{Rows: 1, Cols: len(radius), Cells: radius, ColLabels: labels}}, nil
	}
	man.Call = func(in Input) (string, string) { return "", fmt.Sprintf("strmatch.Manacher(%q)", in.Text) }

	sa := entry("suffix-array", "Suffix Array", "algorithms/strmatch/suffixarray.go", "SuffixArray")
	sa.Summary, sa.Time, sa.Space, sa.Input = "Every suffix in sorted order by prefix doubling, with the LCP array of neighbors' common prefixes (Kasai).", "O(n log^2 n)", "O(n)", Text
	sa.Example = &Input{Text: "banana"}
	sa.Run = func(in Input, _ trace.Recorder) (Output, error) {
		if in.Text == "" || len(in.Text) > 5000 {
			return Output{}, inputError("give a text of 1 to 5000 characters")
		}
		arr, lcp := strmatch.SuffixArray(in.Text)
		t := &TableOutput{Rows: len(arr), Cols: 2, ColLabels: []string{"start", "lcp"}}
		out := Output{Table: t, Values: arr}
		for i, s := range arr {
			t.Cells = append(t.Cells, s, lcp[i])
			t.RowLabels = append(t.RowLabels, strconv.Itoa(i))
			suffix := in.Text[s:]
			if len(suffix) > 60 {
				suffix = suffix[:60] + "..."
			}
			out.Lines = append(out.Lines, fmt.Sprintf("%3d  %s", s, suffix))
		}
		return out, nil
	}
	sa.Call = func(in Input) (string, string) { return "", fmt.Sprintf("strmatch.SuffixArray(%q)", in.Text) }

	return []Entry{
		match("boyer-moore-horspool", "Boyer-Moore-Horspool", "horspool.go", "BoyerMooreHorspool", "Compare each window from its right end and skip ahead by the bad-character shift, often many letters at once (1980).", "O(n / m) best, O(n * m) worst", "O(256)", strmatch.TraceBoyerMooreHorspool),
		z, aho, man, sa,
	}
}

func moreNumberEntries() []Entry {
	pair := func(id, name, file, fn, summary, time string, ex []int, run func(a, b int) (Output, error)) Entry {
		e := entry(id, name, "math/numbers/"+file, fn)
		e.Summary, e.Time, e.Space, e.Input, e.Example = summary, time, "O(1)", TwoNumbers, &Input{Values: ex}
		e.Run = func(in Input, _ trace.Recorder) (Output, error) {
			a, b, err := two(in)
			if err != nil {
				return Output{}, err
			}
			return run(a, b)
		}
		e.Call = func(in Input) (string, string) {
			a, b := firstOr(in.Values, 0), 0
			if len(in.Values) > 1 {
				b = in.Values[1]
			}
			return "", fmt.Sprintf("numbers.%s(%d, %d)", fn, a, b)
		}
		return e
	}
	single := func(id, name, file, fn, summary, time string, ex int, run func(n int) (Output, error)) Entry {
		e := entry(id, name, "math/numbers/"+file, fn)
		e.Summary, e.Time, e.Space, e.Input, e.Example = summary, time, "O(1)", Number, &Input{Values: []int{ex}}
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
	egcd := pair("extended-gcd", "Extended Euclidean Algorithm", "modular.go", "ExtendedGCD", "gcd(a, b) and the x, y with a*x + b*y = gcd: Euclid, carrying the coefficients along.", "O(log min(a, b))", []int{240, 46},
		func(a, b int) (Output, error) {
			g, x, y := numbers.ExtendedGCD(a, b)
			return Output{Number: &g, Values: []int{g, x, y}, Lines: []string{fmt.Sprintf("%d * %d + %d * %d = %d", a, x, b, y, g)}}, nil
		})
	inv := pair("mod-inverse", "Modular Inverse", "modular.go", "ModInverse", "The x with a*x = 1 (mod m), from the extended Euclidean algorithm; none when a and m share a factor.", "O(log m)", []int{3, 11},
		func(a, m int) (Output, error) {
			x, ok := numbers.ModInverse(a, m)
			out := Output{Bool: &ok}
			if ok {
				out.Number = &x
				out.Lines = []string{fmt.Sprintf("%d * %d = 1 (mod %d)", a, x, m)}
			} else {
				out.Lines = []string{fmt.Sprintf("%d has no inverse modulo %d", a, m)}
			}
			return out, nil
		})
	bgcd := pair("binary-gcd", "Binary GCD (Stein)", "modular.go", "BinaryGCD", "The greatest common divisor with only shifts and subtraction, taking out shared factors of two first.", "O(log^2 n) bit operations", []int{1071, 462},
		func(a, b int) (Output, error) { return Output{Number: ptr(numbers.BinaryGCD(a, b))}, nil })
	mr := single("miller-rabin", "Miller-Rabin Primality", "modular.go", "MillerRabin", "Write n-1 = d * 2^s and test bases: a composite fails for one; with twelve bases it is exact for every 64-bit n.", "O(k log^3 n)", 1000000007,
		func(n int) (Output, error) {
			p, w := numbers.MillerRabin(n)
			out := Output{Bool: &p}
			if w > 0 {
				out.Lines = []string{fmt.Sprintf("base %d proves %d composite", w, n)}
			}
			return out, nil
		})
	kara := entry("karatsuba", "Karatsuba Multiplication", "math/numbers/karatsuba.go", "Karatsuba")
	kara.Summary, kara.Time, kara.Space, kara.Input = "Multiply big numbers with three half-size products instead of four (1960).", "O(n^1.585)", "O(n)", TwoTexts
	kara.Example = &Input{Text: "3141592653589793238462643383279502884197169399375105820974944592", Other: "2718281828459045235360287471352662497757247093699959574966967627"}
	kara.Run = func(in Input, _ trace.Recorder) (Output, error) {
		p, err := numbers.Karatsuba(in.Text, in.Other)
		if err != nil {
			return Output{}, inputError("give two non-negative whole numbers of up to 100000 digits")
		}
		return Output{Text: p, Lines: []string{fmt.Sprintf("%d digits", len(p))}}, nil
	}
	kara.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("numbers.Karatsuba(%q, %q)", in.Text, in.Other)
	}
	collatz := single("collatz", "Collatz Sequence", "sequences.go", "Collatz", "Halve even numbers, take 3n+1 of odd ones, until 1: every start tried so far gets there.", "O(steps)", 27,
		func(n int) (Output, error) {
			seq := numbers.Collatz(n)
			if seq == nil {
				return Output{}, inputError("give n >= 1")
			}
			return Output{Values: seq, Initial: seq, Number: ptr(len(seq) - 1)}, nil
		})
	pascal := single("pascal-triangle", "Pascal's Triangle", "sequences.go", "PascalTriangle", "Each value is the sum of the two above it, so row n holds the binomial coefficients C(n, k).", "O(rows^2)", 10,
		func(n int) (Output, error) {
			rows := numbers.PascalTriangle(n)
			if rows == nil {
				return Output{}, inputError(fmt.Sprintf("give 1 to %d rows", numbers.MaxPascalRows))
			}
			t := &TableOutput{Rows: n, Cols: n, RowLabels: numberLabels(n, 0), ColLabels: numberLabels(n, 0), Blank: ptr(-1)}
			for _, row := range rows {
				cells := make([]int, n)
				for k := range cells {
					cells[k] = -1
				}
				copy(cells, row)
				t.Cells = append(t.Cells, cells...)
			}
			return Output{Table: t}, nil
		})
	newton := single("newton-sqrt", "Integer Square Root (Newton)", "sequences.go", "NewtonSqrt", "x = (x + n/x) / 2 until it stops shrinking: the digits right double each step.", "O(log n)", 1000000,
		func(n int) (Output, error) {
			r, steps := numbers.NewtonSqrt(n)
			if r < 0 {
				return Output{}, inputError("give n >= 0")
			}
			return Output{Number: &r, Values: steps}, nil
		})
	fib := single("fibonacci-matrix", "Fibonacci (matrix power)", "sequences.go", "FibonacciMatrix", "F(n) from the nth power of [[1 1] [1 0]] by repeated squaring: O(log n) instead of n steps.", "O(log n)", 90,
		func(n int) (Output, error) {
			v, err := numbers.FibonacciMatrix(n)
			if err != nil {
				return Output{}, inputError("give 0 <= n <= 92")
			}
			return Output{Number: &v}, nil
		})
	base := entry("base-convert", "Base Conversion", "math/numbers/sequences.go", "BaseConvert")
	base.Summary, base.Time, base.Space, base.Input = "Write a number of any length in another base from 2 to 36.", "O(n^2)", "O(n)", Text
	base.Options = []Option{{"from", "from base", 10, 2, 36}, {"to", "to base", 2, 2, 36}}
	base.Example = &Input{Text: "2026", Options: map[string]int{"from": 10, "to": 2}}
	base.Run = func(in Input, _ trace.Recorder) (Output, error) {
		s, err := numbers.BaseConvert(in.Text, in.opt("from", 10), in.opt("to", 2))
		if err != nil {
			return Output{}, inputError("give a number written in the from base (2 to 36)")
		}
		return Output{Text: s}, nil
	}
	base.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("numbers.BaseConvert(%q, %d, %d)", in.Text, in.opt("from", 10), in.opt("to", 2))
	}
	return []Entry{egcd, inv, bgcd, mr, kara, collatz, pascal, newton, fib, base}
}
