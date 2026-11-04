package core

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/algorithms/backtracking"
	"github.com/braswelljr/alGOrithm/pkg/algorithms/dynamic"
	"github.com/braswelljr/alGOrithm/pkg/algorithms/greedy"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

const (
	intervalsExample = "name start end\na 1 4\nb 3 5\nc 0 6\nd 5 7\ne 3 9\nf 5 9\ng 6 10\nh 8 11\ni 8 12\nj 2 14\nk 12 16"
	itemsExample     = "name weight value\ngold 10 60\nsilver 20 100\nbronze 30 120"
	jobsExample      = "name deadline profit\na 2 100\nb 1 19\nc 2 27\nd 1 25\ne 3 15"
	sudokuExample    = "53..7....\n6..195...\n.98....6.\n8...6...3\n4..8.3..1\n7...2...6\n.6....28.\n...419..5\n....8..79"
)

// parseCall writes "name, err := pkg.Parse(`text`)" setup for snippets.
func parseCall(v, fn, text string) string {
	return fmt.Sprintf("%s, err := %s(%s)\n\tif err != nil {\n\t\tpanic(err)\n\t}\n\t", v, fn, "`"+text+"`")
}

func greedyEntries() []Entry {
	act := entry("activity-selection", "Activity Selection", "algorithms/greedy/activity.go", "ActivitySelection")
	act.Summary, act.Time, act.Space, act.Input, act.Traced = "The most activities that do not overlap: always take the one that ends first.", "O(n log n)", "O(n)", Text, true
	act.Example = &Input{Text: intervalsExample}
	act.Run = func(in Input, r trace.Recorder) (Output, error) {
		acts, err := greedy.ParseIntervals(in.Text)
		if err != nil {
			return Output{}, wrap(err)
		}
		chosen := greedy.TraceActivitySelection(acts, r)
		taken := map[int]bool{}
		var names []string
		for _, i := range chosen {
			taken[i] = true
			names = append(names, acts[i].Name)
		}
		var spans []Span
		for i, a := range acts {
			state := "idle"
			if taken[i] {
				state = "running"
			}
			spans = append(spans, Span{Lane: a.Name, Label: a.Name, Start: a.Start, End: a.End, State: state})
		}
		return Output{Timeline: spans, Number: ptr(len(chosen)), Lines: []string{"chosen: " + strings.Join(names, ", ")}}, nil
	}
	act.Call = func(in Input) (string, string) {
		return parseCall("acts", "greedy.ParseIntervals", in.Text), "greedy.ActivitySelection(acts)"
	}

	part := entry("interval-partitioning", "Interval Partitioning", "algorithms/greedy/partition.go", "IntervalPartitioning")
	part.Summary, part.Time, part.Space, part.Input, part.Traced = "Give every interval a room so overlapping ones never share, with the fewest rooms.", "O(n log n)", "O(n)", Text, true
	part.Example = &Input{Text: intervalsExample}
	part.Run = func(in Input, r trace.Recorder) (Output, error) {
		iv, err := greedy.ParseIntervals(in.Text)
		if err != nil {
			return Output{}, wrap(err)
		}
		room, n := greedy.TraceIntervalPartitioning(iv, r)
		var spans []Span
		for i, x := range iv {
			spans = append(spans, Span{Lane: "room " + strconv.Itoa(room[i]+1), Label: x.Name, Start: x.Start, End: x.End, State: "running"})
		}
		sort.SliceStable(spans, func(a, b int) bool { return spans[a].Lane < spans[b].Lane })
		return Output{Timeline: spans, Number: &n, Summary: map[string]float64{"rooms": float64(n)}}, nil
	}
	part.Call = func(in Input) (string, string) {
		return parseCall("iv", "greedy.ParseIntervals", in.Text), "greedy.IntervalPartitioning(iv)"
	}

	huff := entry("huffman-coding", "Huffman Coding", "algorithms/greedy/huffman.go", "HuffmanCoding")
	huff.Summary, huff.Time, huff.Space, huff.Input, huff.Traced = "The shortest prefix code for a text: merge the two rarest symbols, again and again (Huffman, 1952).", "O(n + k log k)", "O(n + k)", Text, true
	huff.Example = &Input{Text: "abracadabra alakazam"}
	huff.Limits = fmt.Sprintf("text up to %d characters", greedy.MaxHuffmanText)
	huff.Run = func(in Input, r trace.Recorder) (Output, error) {
		h, err := greedy.TraceHuffmanCoding(in.Text, r)
		if err != nil {
			return Output{}, wrap(err)
		}
		out := Output{Text: h.Encoded, Number: &h.Bits, Summary: map[string]float64{
			"bits": float64(h.Bits), "bits uncoded (8 per byte)": float64(h.OriginalBits),
			"compression ratio": float64(h.Bits) / float64(h.OriginalBits),
		}}
		chars := make([]int, 0, len(h.Codes))
		for c := range h.Codes {
			chars = append(chars, int(c))
		}
		sort.Ints(chars)
		for _, c := range chars {
			out.Lines = append(out.Lines, fmt.Sprintf("%q = %s", greedy.CharLabel(byte(c)), h.Codes[byte(c)]))
		}
		for i, f := range h.Merges {
			op, res := Op{Name: "merge", Value: strconv.Itoa(i)}, plural(len(rootsOf(f)), "tree")+" left"
			if i == 0 {
				op, res = Op{Name: "start"}, fmt.Sprintf("one tree per symbol (%d)", len(rootsOf(f)))
			}
			out.Steps = append(out.Steps, Step{Op: op, Result: res, Snapshot: f})
		}
		return out, nil
	}
	huff.Call = func(in Input) (string, string) { return "", fmt.Sprintf("greedy.HuffmanCoding(%q)", in.Text) }

	knap := entry("fractional-knapsack", "Fractional Knapsack", "algorithms/greedy/knapsack.go", "FractionalKnapsack")
	knap.Summary, knap.Time, knap.Space, knap.Input, knap.Traced = "The most value in a knapsack when items can be split: best value per weight first.", "O(n log n)", "O(n)", Text, true
	knap.Options = []Option{{"capacity", "capacity", 50, 0, 1_000_000}}
	knap.Example = &Input{Text: itemsExample, Options: map[string]int{"capacity": 50}}
	knap.Run = func(in Input, r trace.Recorder) (Output, error) {
		items, err := greedy.ParseItems(in.Text)
		if err != nil {
			return Output{}, wrap(err)
		}
		v, taken := greedy.TraceFractionalKnapsack(items, in.opt("capacity", 50), r)
		out := Output{Summary: map[string]float64{"value carried": v}}
		for i, f := range taken {
			if f > 0 {
				out.Lines = append(out.Lines, fmt.Sprintf("%s: %.0f%% (weight %.4g, value %.4g)", items[i].Name, f*100, f*float64(items[i].Weight), f*float64(items[i].Value)))
			}
		}
		return out, nil
	}
	knap.Call = func(in Input) (string, string) {
		return parseCall("items", "greedy.ParseItems", in.Text), fmt.Sprintf("greedy.FractionalKnapsack(items, %d)", in.opt("capacity", 50))
	}

	jobs := entry("job-sequencing", "Job Sequencing with Deadlines", "algorithms/greedy/jobs.go", "JobSequencing")
	jobs.Summary, jobs.Time, jobs.Space, jobs.Input, jobs.Traced = "The most profit from unit-time jobs with deadlines: richest first, each in the latest free slot.", "O(n log n + n * d)", "O(d)", Text, true
	jobs.Example = &Input{Text: jobsExample}
	jobs.Run = func(in Input, r trace.Recorder) (Output, error) {
		js, err := greedy.ParseJobs(in.Text)
		if err != nil {
			return Output{}, wrap(err)
		}
		slots, profit := greedy.TraceJobSequencing(js, r)
		var spans []Span
		var order []string
		for s, j := range slots {
			if j < 0 {
				spans = append(spans, Span{Lane: "schedule", Label: "idle", Start: s, End: s + 1, State: "idle"})
				continue
			}
			spans = append(spans, Span{Lane: "schedule", Label: js[j].Name, Start: s, End: s + 1, State: "running"})
			order = append(order, js[j].Name)
		}
		return Output{Timeline: spans, Number: &profit, Lines: []string{"order: " + strings.Join(order, ", ")}}, nil
	}
	jobs.Call = func(in Input) (string, string) {
		return parseCall("jobs", "greedy.ParseJobs", in.Text), "greedy.JobSequencing(jobs)"
	}

	coin := entry("greedy-coin-change", "Greedy Coin Change", "algorithms/greedy/coins.go", "CoinChange")
	coin.Summary, coin.Time, coin.Space, coin.Input, coin.Traced, coin.Params = "Pay with the largest coin that fits, again and again; optimal for some coin systems and not for others.", "O(k log k + amount / smallest)", "O(k)", Amount, true, []string{"amount"}
	coin.Example = &Input{Values: []int{1, 3, 4}, Target: 6}
	coin.Run = func(in Input, r trace.Recorder) (Output, error) {
		if err := values(in, "pay with (the coins)", 100); err != nil {
			return Output{}, err
		}
		if in.Target < 0 || in.Target > dynamic.MaxAmount {
			return Output{}, inputError(fmt.Sprintf("the amount is 0 to %d", dynamic.MaxAmount))
		}
		used, ok := greedy.TraceCoinChange(in.Values, in.Target, r)
		best, bestCoins, _, err := dynamic.CoinChange(in.Values, in.Target)
		if err != nil {
			return Output{}, wrap(err)
		}
		out := Output{Values: used, Bool: &ok, Summary: map[string]float64{"greedy coins": float64(len(used)), "fewest possible": float64(best)}}
		switch {
		case !ok && best < 0:
			out.Lines = []string{fmt.Sprintf("%d cannot be paid with these coins", in.Target)}
		case !ok:
			out.Lines = []string{fmt.Sprintf("greedy gets stuck, but %v pays it", bestCoins)}
		case len(used) == best:
			out.Lines = []string{fmt.Sprintf("greedy %v is optimal here", used)}
		default:
			out.Lines = []string{fmt.Sprintf("greedy uses %d coins %v; the optimum %v uses %d", len(used), used, bestCoins, best)}
		}
		return out, nil
	}
	coin.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("greedy.CoinChange(%s, %d)", ints(in.Values), in.Target)
	}
	return []Entry{act, part, huff, knap, jobs, coin}
}

// plural writes "1 tree" or "3 trees".
func plural(n int, word string) string {
	if n == 1 {
		return "1 " + word
	}
	return strconv.Itoa(n) + " " + word + "s"
}

// rootsOf returns the root nodes of a forest snapshot.
func rootsOf(s trace.Snapshot) []trace.Node {
	var out []trace.Node
	for _, n := range s.Nodes {
		if n.Parent < 0 {
			out = append(out, n)
		}
	}
	return out
}

func backtrackingEntries() []Entry {
	queens := entry("n-queens", "N-Queens", "algorithms/backtracking/nqueens.go", "NQueens")
	queens.Summary, queens.Time, queens.Space, queens.Input, queens.Traced = "Place n queens so none attacks another: a queen per row, backing up when a row has no safe column.", "O(n!) worst", "O(n)", Number, true
	queens.Example, queens.Limits = &Input{Values: []int{8}}, fmt.Sprintf("1 <= n <= %d", backtracking.MaxQueens)
	queens.Run = func(in Input, r trace.Recorder) (Output, error) {
		n, err := one(in)
		if err != nil {
			return Output{}, err
		}
		count, first, err := backtracking.TraceNQueens(n, r)
		if err != nil {
			return Output{}, wrap(err)
		}
		t := &TableOutput{Rows: n, Cols: n, Cells: make([]int, n*n), RowLabels: numberLabels(n, 1), ColLabels: numberLabels(n, 1)}
		for row, c := range first {
			t.Cells[row*n+c] = 1
		}
		return Output{Table: t, Number: &count, Summary: map[string]float64{"solutions": float64(count)}}, nil
	}
	queens.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("backtracking.NQueens(%d)", firstOr(in.Values, 8))
	}

	sudoku := entry("sudoku-solver", "Sudoku Solver", "algorithms/backtracking/sudoku.go", "SolveSudoku")
	sudoku.Summary, sudoku.Time, sudoku.Space, sudoku.Input, sudoku.Traced = "Fill the most constrained empty cell with each allowed digit, undoing it when the rest cannot be filled.", "exponential worst, fast with pruning", "O(81)", Text, true
	sudoku.Example = &Input{Text: sudokuExample}
	sudoku.Run = func(in Input, r trace.Recorder) (Output, error) {
		g, err := backtracking.ParseSudoku(in.Text)
		if err != nil {
			return Output{}, wrap(err)
		}
		solved, ok := backtracking.TraceSolveSudoku(g, r)
		t := &TableOutput{Rows: 9, Cols: 9, Cells: make([]int, 81), RowLabels: numberLabels(9, 1), ColLabels: numberLabels(9, 1)}
		for i := range t.Cells {
			t.Cells[i] = solved[i/9][i%9]
		}
		return Output{Table: t, Bool: &ok, Text: solved.String()}, nil
	}
	sudoku.Call = func(in Input) (string, string) {
		return parseCall("g", "backtracking.ParseSudoku", in.Text), "backtracking.SolveSudoku(g)"
	}

	ss := entry("subset-sum-backtracking", "Subset Sum (backtracking)", "algorithms/backtracking/subsets.go", "SubsetSum")
	ss.Summary, ss.Time, ss.Space, ss.Input, ss.Traced, ss.Params = "Take or leave each value in turn, backing up when the sum overshoots, until it hits the target.", "O(2^n) worst", "O(n)", AnyInts, true, []string{"target"}
	ss.Example, ss.Limits = &Input{Values: []int{3, 34, 4, 12, 5, 2}, Target: 9}, "at most 40 values, none negative"
	ss.Run = func(in Input, r trace.Recorder) (Output, error) {
		if err := values(in, "choose from", 40); err != nil {
			return Output{}, err
		}
		idx, ok, err := backtracking.TraceSubsetSum(in.Values, in.Target, r)
		if err != nil {
			return Output{}, wrap(err)
		}
		out := Output{Initial: append([]int(nil), in.Values...), Bool: &ok}
		if ok {
			var picked []string
			for _, i := range idx {
				picked = append(picked, strconv.Itoa(in.Values[i]))
			}
			out.Lines = []string{fmt.Sprintf("%s = %d", strings.Join(picked, " + "), in.Target)}
		}
		return out, nil
	}
	ss.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("backtracking.SubsetSum(%s, %d)", ints(in.Values), in.Target)
	}

	cs := entry("combination-sum", "Combination Sum", "algorithms/backtracking/subsets.go", "CombinationSum")
	cs.Summary, cs.Time, cs.Space, cs.Input, cs.Params = "Every multiset of candidates (each reusable) that sums to the target.", "exponential in target / smallest", "O(target)", AnyInts, []string{"target"}
	cs.Example, cs.Limits = &Input{Values: []int{2, 3, 6, 7}, Target: 7}, fmt.Sprintf("target up to 10000, at most %d results", backtracking.MaxCombinations)
	cs.Run = func(in Input, _ trace.Recorder) (Output, error) {
		if err := values(in, "combine", 100); err != nil {
			return Output{}, err
		}
		combos, err := backtracking.CombinationSum(in.Values, in.Target)
		if err != nil {
			return Output{}, wrap(err)
		}
		var lines []string
		for _, c := range combos {
			parts := make([]string, len(c))
			for i, v := range c {
				parts[i] = strconv.Itoa(v)
			}
			lines = append(lines, strings.Join(parts, " + "))
		}
		return Output{Lines: lines, Number: ptr(len(combos))}, nil
	}
	cs.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("backtracking.CombinationSum(%s, %d)", ints(in.Values), in.Target)
	}

	color := entry("graph-coloring", "Graph Coloring", "algorithms/backtracking/coloring.go", "GraphColoring")
	color.Summary, color.Time, color.Space, color.Input, color.Traced = "Give each node one of m colors so no neighbors match, backing up when a node has none left.", "O(m^n) worst", "O(n)", GraphInput, true
	color.Options = []Option{{"colors", "colors", 3, 1, 10}}
	color.Example = &Input{Text: "WA NT\nWA SA\nNT SA\nNT Q\nSA Q\nSA NSW\nSA V\nQ NSW\nNSW V\nT", Options: map[string]int{"colors": 3}}
	color.Run = func(in Input, r trace.Recorder) (Output, error) {
		g, err := backtracking.ParseGraph(in.Text)
		if err != nil {
			return Output{}, wrap(err)
		}
		m := in.opt("colors", 3)
		colors, ok := backtracking.TraceGraphColoring(g, m, r)
		out := Output{Graph: &GraphOutput{Nodes: g.Nodes}, Bool: &ok}
		for a, nbs := range g.Adj {
			for _, b := range nbs {
				if a < b {
					out.Graph.Edges = append(out.Graph.Edges, GraphEdge{From: a, To: b, Weight: 1})
				}
			}
		}
		if ok {
			for n, c := range colors {
				out.Lines = append(out.Lines, fmt.Sprintf("%s: color %d", g.Nodes[n], c+1))
			}
		} else {
			out.Lines = []string{fmt.Sprintf("%d colors are not enough", m)}
		}
		return out, nil
	}
	color.Call = func(in Input) (string, string) {
		return parseCall("g", "backtracking.ParseGraph", in.Text), fmt.Sprintf("backtracking.GraphColoring(g, %d)", in.opt("colors", 3))
	}

	word := entry("word-search", "Word Search", "algorithms/backtracking/wordsearch.go", "WordSearch")
	word.Summary, word.Time, word.Space, word.Input, word.Traced = "Find a word as a path of neighboring letters, no cell twice, backing up at dead ends.", "O(cells * 4^len)", "O(len)", TwoTexts, true
	word.Example, word.Limits = &Input{Text: "ABCE\nSFCS\nADEE", Other: "ABCCED"}, "a board of at most 400 letters"
	word.Run = func(in Input, r trace.Recorder) (Output, error) {
		b, err := backtracking.ParseBoard(in.Text)
		if err != nil {
			return Output{}, wrap(err)
		}
		if strings.TrimSpace(in.Other) == "" {
			return Output{}, inputError("give the word to find")
		}
		path, ok := backtracking.TraceWordSearch(b, strings.TrimSpace(in.Other), r)
		rows, cols := len(b), len(b[0])
		g := &GridOutput{Rows: rows, Cols: cols, Cost: make([]int, rows*cols), Start: -1, End: -1}
		for i := range g.Cost {
			g.Cost[i] = 1
			g.Labels = append(g.Labels, string(b[i/cols][i%cols]))
		}
		if ok {
			g.Start, g.End = path[0], path[len(path)-1]
		}
		lines := make([]string, rows)
		for i, row := range b {
			lines[i] = string(row)
		}
		return Output{Grid: g, Path: path, Bool: &ok, Lines: lines}, nil
	}
	word.Call = func(in Input) (string, string) {
		return parseCall("b", "backtracking.ParseBoard", in.Text), fmt.Sprintf("backtracking.WordSearch(b, %q)", in.Other)
	}
	return []Entry{queens, sudoku, ss, cs, color, word}
}
