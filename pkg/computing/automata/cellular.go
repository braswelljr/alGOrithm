package automata

import (
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// Elementary runs an elementary cellular automaton (Stephen Wolfram's
// numbering): a row of cells, each alive or dead, and at every generation a
// cell's next state is bit number (left*4 + self*2 + right) of rule. Cells
// beyond the edges are dead. It returns gens+1 rows, the first being start.
// Rule 30 makes chaos from one cell; rule 110 can compute anything.
//
// Time: O(gens * width). Space: the same.
func Elementary(rule int, start []bool, gens int) [][]bool {
	return elementary(rule, start, gens, trace.Nop{})
}

// TraceElementary is Elementary, reporting a Set for every cell of every
// row (Indices[0] row*width + column, Values[0] 1 alive or 0 dead), and a
// Mark "generation" (Values[0] its number) after each row.
func TraceElementary(rule int, start []bool, gens int, r trace.Recorder) [][]bool {
	return elementary(rule, start, gens, r)
}

func elementary(rule int, start []bool, gens int, r trace.Recorder) [][]bool {
	w := len(start)
	rows := [][]bool{append([]bool(nil), start...)}
	emit := func(row int, cells []bool) {
		for i, c := range cells {
			r.Record(trace.Event{Kind: trace.Set, Indices: []int{row*w + i}, Values: []int{b2i(c)}})
		}
		r.Record(trace.Event{Kind: trace.Mark, Values: []int{row}, Note: "generation"})
	}
	emit(0, rows[0])
	for g := 1; g <= gens; g++ {
		prev, next := rows[g-1], make([]bool, w)
		for i := range next {
			l := i > 0 && prev[i-1]
			rt := i+1 < w && prev[i+1]
			next[i] = rule>>(b2i(l)*4+b2i(prev[i])*2+b2i(rt))&1 == 1
		}
		rows = append(rows, next)
		emit(g, next)
	}
	return rows
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

// Center returns a row of width dead cells with the middle one alive: the
// usual start for an elementary automaton.
func Center(width int) []bool {
	row := make([]bool, width)
	if width > 0 {
		row[width/2] = true
	}
	return row
}

// ParseLife reads a Game of Life board: one row a line, "#", "O", "*" or
// "1" alive and anything else dead. Short rows are padded with dead cells.
func ParseLife(text string) ([][]bool, error) {
	var board [][]bool
	width := 0
	for _, line := range strings.Split(strings.TrimRight(text, "\n"), "\n") {
		line = strings.TrimRight(line, "\r")
		row := make([]bool, 0, len(line))
		for _, c := range line {
			row = append(row, c == '#' || c == 'O' || c == '*' || c == '1')
		}
		width = max(width, len(row))
		board = append(board, row)
	}
	if len(board) == 0 || width == 0 {
		return nil, inputErr("give a board, one row a line")
	}
	for i := range board {
		for len(board[i]) < width {
			board[i] = append(board[i], false)
		}
	}
	return board, nil
}

// Life runs Conway's Game of Life (John Conway, 1970) for gens generations:
// a live cell with two or three live neighbors lives on, a dead cell with
// exactly three comes alive, and every other cell is dead next generation.
// Cells beyond the board are dead. It returns gens+1 boards, the first being
// start.
//
// Time: O(gens * rows * cols). Space: the same.
func Life(start [][]bool, gens int) [][][]bool { return life(start, gens, trace.Nop{}) }

// TraceLife is Life, reporting a Set of every cell of the first board, then
// one Set for each cell that changes (Indices[0] row*cols + column, Values[0]
// 1 born or 0 died), and a Mark "generation" (Values[0] its number) after
// the first board and after each generation.
func TraceLife(start [][]bool, gens int, r trace.Recorder) [][][]bool {
	return life(start, gens, r)
}

func life(start [][]bool, gens int, r trace.Recorder) [][][]bool {
	rows := len(start)
	if rows == 0 {
		return [][][]bool{{}}
	}
	cols := len(start[0])
	copyBoard := func(b [][]bool) [][]bool {
		out := make([][]bool, len(b))
		for i := range b {
			out[i] = append([]bool(nil), b[i]...)
		}
		return out
	}
	boards := [][][]bool{copyBoard(start)}
	for i := range start {
		for j := range start[i] {
			r.Record(trace.Event{Kind: trace.Set, Indices: []int{i*cols + j}, Values: []int{b2i(start[i][j])}})
		}
	}
	r.Record(trace.Event{Kind: trace.Mark, Values: []int{0}, Note: "generation"})
	for g := 1; g <= gens; g++ {
		prev := boards[g-1]
		next := copyBoard(prev)
		for i := 0; i < rows; i++ {
			for j := 0; j < cols; j++ {
				n := 0
				for di := -1; di <= 1; di++ {
					for dj := -1; dj <= 1; dj++ {
						y, x := i+di, j+dj
						if (di != 0 || dj != 0) && y >= 0 && y < rows && x >= 0 && x < cols && prev[y][x] {
							n++
						}
					}
				}
				next[i][j] = n == 3 || (n == 2 && prev[i][j])
				if next[i][j] != prev[i][j] {
					r.Record(trace.Event{Kind: trace.Set, Indices: []int{i*cols + j}, Values: []int{b2i(next[i][j])}})
				}
			}
		}
		boards = append(boards, next)
		r.Record(trace.Event{Kind: trace.Mark, Values: []int{g}, Note: "generation"})
	}
	return boards
}

// Alive counts the live cells of a board.
func Alive(board [][]bool) int {
	n := 0
	for _, row := range board {
		for _, c := range row {
			n += b2i(c)
		}
	}
	return n
}
