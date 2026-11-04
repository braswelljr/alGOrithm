package core

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/algorithms/backtracking"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// backtrackingEntries is every entry of the backtracking category.
func backtrackingEntries() []Entry {
	return []Entry{
		combinationSumEntry(),
		graphColoringEntry(),
		nQueensEntry(),
		subsetSumBacktrackingEntry(),
		sudokuSolverEntry(),
		wordSearchEntry(),
	}
}

func nQueensEntry() Entry {
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
	return queens
}

func sudokuSolverEntry() Entry {
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
	return sudoku
}

func subsetSumBacktrackingEntry() Entry {
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
	return ss
}

func combinationSumEntry() Entry {
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
	return cs
}

func graphColoringEntry() Entry {
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
	return color
}

func wordSearchEntry() Entry {
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
	return word
}
