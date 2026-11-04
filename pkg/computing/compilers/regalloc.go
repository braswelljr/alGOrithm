package compilers

import (
	"fmt"
	"slices"
	"strings"
)

// Allocation is the result of register allocation: the variables in the
// order they first appear, the variables live after each instruction, the
// interference edges (as indexes into Vars), each variable's register (or
// -1 when it is spilled to memory), and each step in words.
type Allocation struct {
	Vars     []string   `json:"vars"`
	Live     [][]string `json:"live"`
	Edges    [][2]int   `json:"edges"`
	Register []int      `json:"register"`
	Spilled  []string   `json:"spilled"`
	Steps    []string   `json:"steps"`
}

type tacLine struct {
	def  string
	uses []string
	move bool // "x = y": x and y may share a register
}

func isName(s string) bool {
	return s != "" && isLetter(s[0])
}

// parseTAC reads straight-line three-address code: "x = a + b", "x = a",
// "x = -a", "x = 5", and "print x" or "return x" (uses only).
func parseTAC(code []string) ([]tacLine, error) {
	var out []tacLine
	for i, raw := range code {
		f := strings.Fields(raw)
		if len(f) == 0 {
			continue
		}
		var l tacLine
		switch {
		case len(f) >= 3 && f[1] == "=" && isName(f[0]):
			l.def = f[0]
			rest := f[2:]
			if len(rest) != 1 && len(rest) != 3 {
				return nil, fmt.Errorf("%w: line %d: want x = a, or x = a op b", ErrInput, i+1)
			}
			for k, w := range rest {
				w = strings.TrimLeft(w, "-!")
				if k != 1 && isName(w) {
					l.uses = append(l.uses, w)
				}
			}
			l.move = len(rest) == 1 && isName(rest[0])
		case len(f) == 2 && (f[0] == "print" || f[0] == "return" || f[0] == "use"):
			if isName(f[1]) {
				l.uses = []string{f[1]}
			}
		default:
			return nil, fmt.Errorf("%w: line %d: %q is not x = a op b, print x or return x", ErrInput, i+1, raw)
		}
		out = append(out, l)
	}
	return out, nil
}

// AllocateRegisters assigns k registers to the variables of straight-line
// three-address code by graph coloring (Gregory Chaitin, 1982):
//
//  1. Liveness: walking backward, a variable is live from its last use up
//     to its definition.
//  2. Interference: a variable defined while another is live cannot share
//     its register, so they get an edge.
//  3. Simplify: a variable with fewer than k neighbors can always be given
//     a register, so set it aside; if none is left, set aside the one with
//     the most neighbors as a spill candidate.
//  4. Select: take them back in reverse, each getting the lowest register
//     its neighbors do not use; one that finds none is spilled to memory.
//
// liveOut names the variables still needed after the code.
//
// Time: O(n * v + v^2) for n instructions and v variables.
// Space: O(v^2) for the interference graph.
func AllocateRegisters(code []string, k int, liveOut []string) (Allocation, error) {
	if k < 1 {
		return Allocation{}, fmt.Errorf("%w: at least one register", ErrInput)
	}
	lines, err := parseTAC(code)
	if err != nil {
		return Allocation{}, err
	}
	ra := newRegAlloc(lines, liveOut, k)
	ra.liveness(lines, liveOut)
	ra.selectRegisters(ra.simplify())
	return ra.a, nil
}

// regAlloc is one allocation in progress: the result so far, each
// variable's index, and the interference graph as a matrix.
type regAlloc struct {
	a       Allocation
	k       int
	index   map[string]int
	adj     [][]bool
	removed []bool
}

// newRegAlloc names every variable in the order it first appears.
func newRegAlloc(lines []tacLine, liveOut []string, k int) *regAlloc {
	ra := &regAlloc{k: k, index: map[string]int{}}
	for _, l := range lines {
		for _, u := range l.uses {
			ra.add(u)
		}
		if l.def != "" {
			ra.add(l.def)
		}
	}
	for _, v := range liveOut {
		ra.add(v)
	}
	n := len(ra.a.Vars)
	ra.adj, ra.removed = make([][]bool, n), make([]bool, n)
	for i := range ra.adj {
		ra.adj[i] = make([]bool, n)
	}
	return ra
}

func (ra *regAlloc) add(v string) {
	if _, ok := ra.index[v]; !ok {
		ra.index[v] = len(ra.a.Vars)
		ra.a.Vars = append(ra.a.Vars, v)
	}
}

func (ra *regAlloc) link(x, y string) {
	i, j := ra.index[x], ra.index[y]
	if i != j && !ra.adj[i][j] {
		ra.adj[i][j], ra.adj[j][i] = true, true
		ra.a.Edges = append(ra.a.Edges, [2]int{min(i, j), max(i, j)})
	}
}

// liveness is steps 1 and 2: walk backward keeping the live set, and link a
// definition to everything live across it (except the source of a move).
func (ra *regAlloc) liveness(lines []tacLine, liveOut []string) {
	live := map[string]bool{}
	for _, v := range liveOut {
		live[v] = true
	}
	ra.a.Live = make([][]string, len(lines))
	for i := len(lines) - 1; i >= 0; i-- {
		l := lines[i]
		ra.a.Live[i] = sortedKeys(live)
		if l.def != "" {
			for _, v := range sortedKeys(live) { // sorted, so the edges come out the same every run
				if v != l.def && (!l.move || v != l.uses[0]) {
					ra.link(l.def, v)
				}
			}
			delete(live, l.def)
		}
		for _, u := range l.uses {
			live[u] = true
		}
	}
	entry := sortedKeys(live) // used before defined: they hold values on entry
	for i := range entry {
		for j := i + 1; j < len(entry); j++ {
			ra.link(entry[i], entry[j])
		}
	}
	ra.a.Steps = append(ra.a.Steps, fmt.Sprintf("liveness and interference: %d variables, %d edges", len(ra.a.Vars), len(ra.a.Edges)))
}

// degree counts i's neighbors not yet set aside.
func (ra *regAlloc) degree(i int) int {
	d := 0
	for j := range ra.a.Vars {
		if ra.adj[i][j] && !ra.removed[j] {
			d++
		}
	}
	return d
}

// simplify is step 3: set aside variables with fewer than k neighbors, or
// the busiest one as a spill candidate, and return them in that order.
func (ra *regAlloc) simplify() []int {
	var stack []int
	for len(stack) < len(ra.a.Vars) {
		pick := ra.colorable()
		if pick >= 0 {
			ra.a.Steps = append(ra.a.Steps, fmt.Sprintf("simplify: %s has %d neighbors, fewer than %d: set it aside", ra.a.Vars[pick], ra.degree(pick), ra.k))
		} else {
			pick = ra.busiest()
			ra.a.Steps = append(ra.a.Steps, fmt.Sprintf("simplify: every variable left has %d or more neighbors; set aside %s (%d neighbors) as a spill candidate", ra.k, ra.a.Vars[pick], ra.degree(pick)))
		}
		ra.removed[pick] = true
		stack = append(stack, pick)
	}
	return stack
}

// colorable is the first variable left with fewer than k neighbors, or -1.
func (ra *regAlloc) colorable() int {
	for i := range ra.a.Vars {
		if !ra.removed[i] && ra.degree(i) < ra.k {
			return i
		}
	}
	return -1
}

// busiest is the variable left with the most neighbors.
func (ra *regAlloc) busiest() int {
	pick := -1
	for i := range ra.a.Vars {
		if !ra.removed[i] && (pick < 0 || ra.degree(i) > ra.degree(pick)) {
			pick = i
		}
	}
	return pick
}

// selectRegisters is step 4: take the variables back in reverse, each
// getting the lowest register no neighbor holds, or spilling.
func (ra *regAlloc) selectRegisters(stack []int) {
	ra.a.Register = make([]int, len(ra.a.Vars))
	for i := range ra.a.Register {
		ra.a.Register[i] = -2 // not yet colored
	}
	for i := len(stack) - 1; i >= 0; i-- {
		v := stack[i]
		ra.a.Register[v] = ra.freeRegister(v)
		if ra.a.Register[v] < 0 {
			ra.a.Spilled = append(ra.a.Spilled, ra.a.Vars[v])
			ra.a.Steps = append(ra.a.Steps, fmt.Sprintf("select: %s: every register is taken by a neighbor; spill it to memory", ra.a.Vars[v]))
		} else {
			ra.a.Steps = append(ra.a.Steps, fmt.Sprintf("select: %s gets r%d", ra.a.Vars[v], ra.a.Register[v]))
		}
	}
}

// freeRegister is the lowest register none of v's neighbors holds, or -1.
func (ra *regAlloc) freeRegister(v int) int {
	used := make([]bool, ra.k)
	for j := range ra.a.Vars {
		if ra.adj[v][j] && ra.a.Register[j] >= 0 {
			used[ra.a.Register[j]] = true
		}
	}
	for r := 0; r < ra.k; r++ {
		if !used[r] {
			return r
		}
	}
	return -1
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	slices.Sort(out)
	return out
}
