package core

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/algorithms/dynamic"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// dynamicEntries is every entry of the dynamic category.
func dynamicEntries() []Entry {
	return []Entry{
		coinChangeEntry(),
		editDistanceEntry(),
		eggDropEntry(),
		knapsack01Entry(),
		lisFastEntry(),
		longestCommonSubsequenceEntry(),
		longestIncreasingSubsequenceEntry(),
		longestPalindromicSubsequenceEntry(),
		matrixChainEntry(),
		rodCuttingEntry(),
		subsetSumEntry(),
		uniquePathsEntry(),
	}
}

func dynamicWords(in Input) error {
	if in.Text == "" && in.Other == "" {
		return inputError("give two words")
	}
	if (len(in.Text)+1)*(len(in.Other)+1) > 250_000 {
		return inputError("the words are too long for the table (at most 250000 cells)")
	}
	return nil
}

func editDistanceEntry() Entry {
	ed := entry("edit-distance", "Edit Distance", "algorithms/dynamic/editdistance.go", "EditDistance")
	ed.Summary, ed.Time, ed.Space, ed.Input, ed.Traced = "The fewest insertions, deletions and substitutions that turn one word into another (Levenshtein, 1965).", "O(n * m)", "O(n * m)", TwoTexts, true
	ed.Example = &Input{Text: "kitten", Other: "sitting"}
	ed.Run = func(in Input, r trace.Recorder) (Output, error) {
		if err := dynamicWords(in); err != nil {
			return Output{}, err
		}
		d, t := dynamic.TraceEditDistance(in.Text, in.Other, r)
		return Output{Number: &d, Table: &TableOutput{Rows: len(t), Cols: len(t[0]), Cells: flatten(t), RowLabels: labelsOf(in.Text), ColLabels: labelsOf(in.Other)}}, nil
	}
	return ed
}

func longestCommonSubsequenceEntry() Entry {
	lcs := entry("longest-common-subsequence", "Longest Common Subsequence", "algorithms/dynamic/lcs.go", "LCS")
	lcs.Summary, lcs.Time, lcs.Space, lcs.Input, lcs.Traced = "The longest sequence of letters found in both words in the same order: the comparison behind diff.", "O(n * m)", "O(n * m)", TwoTexts, true
	lcs.Example = &Input{Text: "ABCBDAB", Other: "BDCABA"}
	lcs.Run = func(in Input, r trace.Recorder) (Output, error) {
		if err := dynamicWords(in); err != nil {
			return Output{}, err
		}
		s, t := dynamic.TraceLCS(in.Text, in.Other, r)
		return Output{Text: s, Number: ptr(len(s)), Table: &TableOutput{Rows: len(t), Cols: len(t[0]), Cells: flatten(t), RowLabels: labelsOf(in.Text), ColLabels: labelsOf(in.Other)}}, nil
	}
	return lcs
}

func coinChangeEntry() Entry {
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
	return cc
}

func longestIncreasingSubsequenceEntry() Entry {
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
	return lis
}

func dynamicKnapParse(in Input) ([]int, []int, error) {
	text := in.Text
	if lines := strings.SplitN(text, "\n", 2); len(lines) == 2 && strings.Contains(strings.ToLower(lines[0]), "weight") {
		text = lines[1] // a header row
	}
	return parsePairs(text)
}

func knapsack01Entry() Entry {
	knap := entry("knapsack-01", "0/1 Knapsack", "algorithms/dynamic/knapsack.go", "Knapsack01")
	knap.Summary, knap.Time, knap.Space, knap.Input, knap.Traced = "Pick items, each at most once, for the most value within a weight limit, filling a table of best values per item and capacity.", "O(n * capacity)", "O(n * capacity)", Text, true
	knap.Options = []Option{{"capacity", "capacity", 10, 0, 100_000}}
	knap.Example = &Input{Text: "weight value\n1 1\n3 4\n4 5\n5 7", Options: map[string]int{"capacity": 7}}
	knap.Run = func(in Input, r trace.Recorder) (Output, error) {
		w, v, err := dynamicKnapParse(in)
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
		w, v, _ := dynamicKnapParse(in)
		return "", fmt.Sprintf("dynamic.Knapsack01(%s, %s, %d)", ints(w), ints(v), in.opt("capacity", 10))
	}
	return knap
}

func matrixChainEntry() Entry {
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
	return mc
}

func rodCuttingEntry() Entry {
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
	return rod
}

func subsetSumEntry() Entry {
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
	return sub
}

func longestPalindromicSubsequenceEntry() Entry {
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
	return lps
}

func uniquePathsEntry() Entry {
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
	return up
}

func eggDropEntry() Entry {
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
	return egg
}

func lisFastEntry() Entry {
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
	return lf
}
