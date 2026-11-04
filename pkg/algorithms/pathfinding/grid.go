package pathfinding

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
)

// MaxCells is the largest grid Parse and Generate accept.
const MaxCells = 100 * 100

// Grid is a Rows x Cols map, stored row-major: cell i is at row i/Cols,
// column i%Cols.
type Grid struct {
	Rows, Cols int
	// Cost[i] is the cost of entering cell i: 0 for a wall, 1 to 9 otherwise.
	Cost       []int
	Start, End int
}

// Parse reads a grid from text, one row per line: "#" is a wall, "." or a
// space an open cell of cost 1, "1" to "9" an open cell of that cost, "S" the
// start and "E" the end (both cost 1). Short rows are padded with open cells.
// Without an "S" the start is the first open cell; without an "E" the end is
// the last.
func Parse(text string) (*Grid, error) {
	var lines []string
	for _, l := range strings.Split(strings.ReplaceAll(text, "\r", ""), "\n") {
		if strings.TrimSpace(l) != "" {
			lines = append(lines, strings.TrimRight(l, " "))
		}
	}
	if len(lines) == 0 {
		return nil, errors.New("pathfinding: no rows")
	}
	cols := 0
	for _, l := range lines {
		if len(l) > cols {
			cols = len(l)
		}
	}
	if len(lines)*cols > MaxCells {
		return nil, fmt.Errorf("pathfinding: %dx%d is more than %d cells", len(lines), cols, MaxCells)
	}
	g := &Grid{Rows: len(lines), Cols: cols, Cost: make([]int, len(lines)*cols), Start: -1, End: -1}
	for r, l := range lines {
		for c := 0; c < cols; c++ {
			ch := byte('.')
			if c < len(l) {
				ch = l[c]
			}
			i := r*cols + c
			switch {
			case ch == '#':
				g.Cost[i] = 0
			case ch == '.' || ch == ' ':
				g.Cost[i] = 1
			case ch >= '1' && ch <= '9':
				g.Cost[i] = int(ch - '0')
			case ch == 'S' || ch == 's':
				g.Cost[i], g.Start = 1, i
			case ch == 'E' || ch == 'e':
				g.Cost[i], g.End = 1, i
			default:
				return nil, fmt.Errorf("pathfinding: line %d, column %d: %q is not a grid cell (use # . 1-9 S E)", r+1, c+1, ch)
			}
		}
	}
	for i, cost := range g.Cost {
		if cost > 0 && g.Start < 0 {
			g.Start = i
		}
	}
	for i := len(g.Cost) - 1; i >= 0 && g.End < 0; i-- {
		if g.Cost[i] > 0 && i != g.Start {
			g.End = i
		}
	}
	if g.Start < 0 || g.End < 0 {
		return nil, errors.New("pathfinding: needs at least two open cells")
	}
	return g, nil
}

// String prints the grid in the form Parse reads.
func (g *Grid) String() string {
	var b strings.Builder
	for i, cost := range g.Cost {
		switch {
		case i == g.Start:
			b.WriteByte('S')
		case i == g.End:
			b.WriteByte('E')
		case cost == 0:
			b.WriteByte('#')
		case cost == 1:
			b.WriteByte('.')
		default:
			b.WriteByte(byte('0' + cost))
		}
		if i%g.Cols == g.Cols-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

// Neighbors returns the open cells next to i, in the order up, right, down,
// left.
func (g *Grid) Neighbors(i int) []int {
	r, c := i/g.Cols, i%g.Cols
	out := make([]int, 0, 4)
	for _, d := range [4][2]int{{-1, 0}, {0, 1}, {1, 0}, {0, -1}} {
		nr, nc := r+d[0], c+d[1]
		if nr >= 0 && nr < g.Rows && nc >= 0 && nc < g.Cols && g.Cost[nr*g.Cols+nc] > 0 {
			out = append(out, nr*g.Cols+nc)
		}
	}
	return out
}

// Distance is the Manhattan distance between cells a and b: the fewest steps
// between them with no walls, a heuristic that never overestimates.
func (g *Grid) Distance(a, b int) int {
	dr, dc := a/g.Cols-b/g.Cols, a%g.Cols-b%g.Cols
	if dr < 0 {
		dr = -dr
	}
	if dc < 0 {
		dc = -dc
	}
	return dr + dc
}

// PathTo follows prev back from end and returns the path start to end, or nil
// when end was not reached (prev[end] is -1 and end is not the start).
func PathTo(prev []int, start, end int) []int {
	if end != start && prev[end] < 0 {
		return nil
	}
	var path []int
	for at := end; at >= 0; at = prev[at] {
		path = append(path, at)
		if at == start {
			break
		}
	}
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	return path
}

// Cost of a path: the sum of the costs of entering each cell after the first.
func (g *Grid) PathCost(path []int) int {
	total := 0
	for _, c := range path[min(1, len(path)):] {
		total += g.Cost[c]
	}
	return total
}

// Kinds of generated grid.
const (
	Maze     = "maze"     // a perfect maze (recursive backtracker): one path between any two cells
	Open     = "open"     // no walls
	Walls    = "walls"    // about 28% random walls
	Weighted = "weighted" // no walls, random costs 1 to 9
)

// Kinds lists every kind Generate makes.
var Kinds = []string{Maze, Walls, Weighted, Open}

// Generate makes a rows x cols grid of the given kind from seed, start at the
// top left and end at the bottom right. The same arguments always give the
// same grid. A maze needs odd sizes and rounds even ones down.
func Generate(kind string, rows, cols int, seed int64) (*Grid, error) {
	if rows < 2 || cols < 2 || rows*cols > MaxCells {
		return nil, fmt.Errorf("pathfinding: size %dx%d is outside 2x2..%d cells", rows, cols, MaxCells)
	}
	r := rand.New(rand.NewSource(seed))
	if kind == Maze {
		if rows%2 == 0 {
			rows--
		}
		if cols%2 == 0 {
			cols--
		}
		if rows < 3 || cols < 3 {
			return nil, errors.New("pathfinding: a maze needs at least 3x3")
		}
	}
	g := &Grid{Rows: rows, Cols: cols, Cost: make([]int, rows*cols)}
	switch kind {
	case Open:
		for i := range g.Cost {
			g.Cost[i] = 1
		}
	case Walls:
		for i := range g.Cost {
			if r.Intn(100) >= 28 {
				g.Cost[i] = 1
			}
		}
	case Weighted:
		for i := range g.Cost {
			g.Cost[i] = 1 + r.Intn(9)
		}
	case Maze:
		carve(g, r)
	default:
		return nil, fmt.Errorf("pathfinding: unknown kind %q (one of %s)", kind, strings.Join(Kinds, ", "))
	}
	g.Start, g.End = 0, rows*cols-1
	if kind == Maze {
		g.Start, g.End = cols+1, (rows-2)*cols+cols-2
	}
	g.Cost[g.Start], g.Cost[g.End] = 1, 1
	return g, nil
}

// carve cuts a perfect maze into an all-wall pathfinding: cells at odd row and
// column are rooms, and a depth-first walk knocks down the wall between each
// room and an unvisited neighbor.
func carve(g *Grid, r *rand.Rand) {
	start := g.Cols + 1
	g.Cost[start] = 1
	stack := []int{start}
	for len(stack) > 0 {
		cur := stack[len(stack)-1]
		cr, cc := cur/g.Cols, cur%g.Cols
		var next [][2]int // room, wall between
		for _, d := range [4][2]int{{-2, 0}, {0, 2}, {2, 0}, {0, -2}} {
			nr, nc := cr+d[0], cc+d[1]
			if nr > 0 && nr < g.Rows-1 && nc > 0 && nc < g.Cols-1 && g.Cost[nr*g.Cols+nc] == 0 {
				next = append(next, [2]int{nr*g.Cols + nc, (cr+d[0]/2)*g.Cols + cc + d[1]/2})
			}
		}
		if len(next) == 0 {
			stack = stack[:len(stack)-1]
			continue
		}
		pick := next[r.Intn(len(next))]
		g.Cost[pick[1]], g.Cost[pick[0]] = 1, 1
		stack = append(stack, pick[0])
	}
}
