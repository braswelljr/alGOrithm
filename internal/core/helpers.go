package core

import (
	"fmt"
	"path"
	"slices"
	"strconv"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/algorithms/graphs"
	"github.com/braswelljr/alGOrithm/pkg/os/concurrency"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

func one(in Input) (int, error) {
	if len(in.Values) != 1 {
		return 0, inputError("give one number")
	}
	return in.Values[0], nil
}

func two(in Input) (int, int, error) {
	if len(in.Values) != 2 {
		return 0, 0, inputError("give two numbers")
	}
	return in.Values[0], in.Values[1], nil
}

func values(in Input, what string, limit int) error {
	if len(in.Values) == 0 {
		return inputError("give the values to " + what)
	}
	if limit > 0 && len(in.Values) > limit {
		return inputError(fmt.Sprintf("at most %d values", limit))
	}
	return nil
}

func ptr[T any](v T) *T { return &v }

// ParseOps reads operations written as "push 4, pop, push 7" (commas,
// semicolons or new lines between them).
func ParseOps(s string) []Op {
	var ops []Op
	for _, part := range strings.FieldsFunc(s, func(r rune) bool { return r == ',' || r == ';' || r == '\n' }) {
		f := strings.Fields(part)
		if len(f) == 0 {
			continue
		}
		op := Op{Name: strings.ToLower(f[0])}
		if len(f) > 1 {
			op.Value = strings.Join(f[1:], " ")
		}
		ops = append(ops, op)
	}
	return ops
}

// opInt reads an operation's value as a whole number.
func opInt(op Op) (int, error) {
	v, err := strconv.Atoi(strings.TrimSpace(op.Value))
	if err != nil {
		return 0, inputError(fmt.Sprintf("%s needs a whole number, not %q", op.Name, op.Value))
	}
	return v, nil
}

// opPair reads an operation's value as two whole numbers ("3 4").
func opPair(op Op) ([]int, error) {
	const n = 2
	f := strings.Fields(op.Value)
	if len(f) != n {
		return nil, inputError(fmt.Sprintf("%s needs %d whole numbers", op.Name, n))
	}
	out := make([]int, n)
	for i, s := range f {
		v, err := strconv.Atoi(s)
		if err != nil {
			return nil, inputError(fmt.Sprintf("%s needs whole numbers, not %q", op.Name, op.Value))
		}
		out[i] = v
	}
	return out, nil
}

func labelsOf(s string) []string {
	out := make([]string, len(s)+1)
	for i := 0; i < len(s); i++ {
		out[i+1] = string(s[i])
	}
	return out
}

func numberLabels(n, from int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = strconv.Itoa(from + i)
	}
	return out
}

func flatten(t [][]int) []int {
	var out []int
	for _, row := range t {
		out = append(out, row...)
	}
	return out
}

func joinInts(s []int) string {
	parts := make([]string, len(s))
	for i, v := range s {
		parts[i] = strconv.Itoa(v)
	}
	return strings.Join(parts, " ")
}

// unreachable is the table value for "no path": far below any real distance
// and exactly representable in JavaScript, which the visualizer reads it in.
const unreachable = -(1 << 31)

// concat joins lists of entries in order.
func concat(lists ...[]Entry) []Entry {
	var out []Entry
	for _, l := range lists {
		out = append(out, l...)
	}
	return out
}

// paths splits "a, b.c d" into paths.
func paths(s string) []string {
	return strings.FieldsFunc(s, func(r rune) bool { return r == ',' || r == ' ' || r == '\n' })
}

func clip(lines []string, n int) []string {
	if len(lines) > n {
		return append(lines[:n:n], fmt.Sprintf("... %d more", len(lines)-n))
	}
	return lines
}

func found(k int, ok bool) string {
	if ok {
		return fmt.Sprintf("found %d", k)
	}
	return fmt.Sprintf("%d not found", k)
}

// flag reads an on/off option: any value but 0 is on.
func flag(in Input, name string, def int) bool { return in.opt(name, def) != 0 }

// spans turns a simulation timeline into the visualizer's lanes.
func spans(tl concurrency.Timeline, lane func(int) string) []Span {
	out := make([]Span, len(tl))
	for i, a := range tl {
		out[i] = Span{Lane: lane(a.Thread), Label: a.What, Start: a.Tick, End: a.Tick + 1, State: a.State}
	}
	return out
}

func thread(i int) string { return "thread " + strconv.Itoa(i) }

func nodeNames(g *graphs.Graph, ids []int) string {
	out := make([]string, len(ids))
	for i, id := range ids {
		out[i] = g.Names[id]
	}
	return strings.Join(out, " ")
}

func distance(d int) string {
	if d == graphs.Infinity {
		return "unreachable"
	}
	return fmt.Sprint(d)
}

// distances lists every node's distance and path from start.
func distances(g *graphs.Graph, start int, dist, prev []int) []string {
	lines := make([]string, g.Len())
	for v := range lines {
		lines[v] = fmt.Sprintf("%s: %s", g.Names[v], distance(dist[v]))
		if p := graphs.PathTo(prev, start, v); p != nil && v != start {
			lines[v] += " via " + nodeNames(g, p)
		}
	}
	return lines
}

// fixed is a call that does not depend on the input.
func fixed(call string) func(Input) string { return func(Input) string { return call } }

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

// seedOption is the option every seeded entry reads.
var seedOption = Option{"seed", "random seed", 1, 0, 1_000_000}

// binary writes x in at least width bits.
func binary(x uint64, width int) string { return fmt.Sprintf("%0*b", width, x) }

// floats writes ints as a Go []float64 literal.
func floats(v []int) string { return "[]float64{" + strings.TrimPrefix(ints(v), "[]int{") }

func toFloats(v []int) []float64 {
	out := make([]float64, len(v))
	for i, x := range v {
		out[i] = float64(x)
	}
	return out
}

// coefficients checks a polynomial entry's coefficients.
func coefficients(in Input) ([]float64, error) {
	if len(in.Values) == 0 || len(in.Values) > 20 {
		return nil, inputError("give 1 to 20 coefficients, the constant term first")
	}
	return toFloats(in.Values), nil
}

// orderedSetTree is what the four self-balancing set structures share.
type orderedSetTree interface {
	Insert(int) bool
	Remove(int) bool
	Contains(int) bool
	Len() int
	Snapshot() trace.Snapshot
}

// rangeP holds the values a range structure is built over, rebuilding it
// when one is pushed.
type rangeP struct{ values []int }

func (p *rangeP) push(op Op, r trace.Recorder) (string, error) {
	v, err := opInt(op)
	if err != nil {
		return "", err
	}
	p.values = append(p.values, v)
	pushed(r, v)
	return fmt.Sprintf("appended %d and rebuilt over %d values", v, len(p.values)), nil
}

func rangeAnswer(what string, l, h, v int, ok bool) string {
	if !ok {
		return fmt.Sprintf("%d..%d is not a range of the values", l, h)
	}
	return fmt.Sprintf("%s of %d..%d is %d", what, l, h, v)
}

// sim describes one simulation entry: the Entry fields, how to run it, and
// the Go call that reproduces a run (its arguments, in order).
type sim struct {
	id, name, file, fn  string
	summary, time, room string
	input               InputKind
	example             Input
	options             []Option
	run                 func(in Input) (Output, error)
	args                func(in Input) []any
}

// entry builds the Entry.
func (s sim) entry() Entry {
	e := entry(s.id, s.name, s.file, s.fn)
	e.Summary, e.Time, e.Space, e.Input = s.summary, s.time, s.room, s.input
	ex := s.example
	e.Example = &ex
	e.Options = s.options
	run := s.run
	e.Run = func(in Input, _ trace.Recorder) (Output, error) {
		out, err := run(in)
		return out, wrap(err)
	}
	if s.args != nil {
		pkg, fn, args := path.Base(path.Dir(s.file)), s.fn, s.args
		e.Call = func(in Input) (string, string) { return "", goCallOf(pkg, fn, args(in)...) }
	}
	return e
}

// goCallOf writes pkg.fn(args...) as Go source.
func goCallOf(pkg, fn string, args ...any) string {
	parts := make([]string, len(args))
	for i, a := range args {
		switch v := a.(type) {
		case string:
			parts[i] = strconv.Quote(v)
		case []int:
			parts[i] = ints(v)
		default:
			parts[i] = fmt.Sprint(v)
		}
	}
	return pkg + "." + fn + "(" + strings.Join(parts, ", ") + ")"
}

// tick turns one-tick events into timeline bars.
func tick(lane, what, state string, at int) Span {
	return Span{Lane: lane, Label: what, Start: at, End: at + 1, State: state}
}

// summary turns counts into a summary.
func summary(counts map[string]int) map[string]float64 {
	out := map[string]float64{}
	for k, v := range counts {
		out[k] = float64(v)
	}
	return out
}

func on(in Input, name string, def int) bool { return in.opt(name, def) == 1 }

// fieldInts reads whole numbers separated by spaces or commas.
func fieldInts(s, what string) ([]int, error) {
	var out []int
	for _, f := range strings.FieldsFunc(s, func(r rune) bool { return r == ' ' || r == ',' || r == '\n' || r == '\t' }) {
		n, err := strconv.Atoi(f)
		if err != nil {
			return nil, inputError(fmt.Sprintf("%s: %q is not a whole number", what, f))
		}
		out = append(out, n)
	}
	return out, nil
}

func goStrings(s []string) string {
	q := make([]string, len(s))
	for i, v := range s {
		q[i] = strconv.Quote(v)
	}
	return "[]string{" + strings.Join(q, ", ") + "}"
}

// callback is a named JavaScript-style callback: its behavior, and the Go
// literal that does the same, for snippets.
type callback struct {
	keep func(v, target int) bool // for filter, find, some, every
	f    func(v int) int          // for map
	src  string
}

// reducers fold from the left. fromFirst ones start from the first element,
// as JS reduce() does without an initial value; the others from init.
var reducers = map[string]struct {
	f         func(acc, v int) int
	init      int
	fromFirst bool
	src       string
}{
	"sum":     {func(a, v int) int { return a + v }, 0, false, "func(acc, v, _ int) int { return acc + v }"},
	"product": {func(a, v int) int { return a * v }, 1, false, "func(acc, v, _ int) int { return acc * v }"},
	"max":     {func(a, v int) int { return max(a, v) }, 0, true, "func(acc, v, _ int) int { return max(acc, v) }"},
	"min":     {func(a, v int) int { return min(a, v) }, 0, true, "func(acc, v, _ int) int { return min(acc, v) }"},
}

func names[V any](m map[string]V) string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	slices.Sort(out)
	return strings.Join(out, ", ")
}

// arrayRow is the array as a row, with roles on the given positions.
func arrayRow(vals []int, roles map[int]string) trace.Snapshot {
	s := trace.Snapshot{Layout: "row", Nodes: make([]trace.Node, len(vals))}
	for i, v := range vals {
		s.Nodes[i] = trace.Node{ID: i, Label: strconv.Itoa(v), Parent: -1, Role: roles[i]}
	}
	return s
}

func spanRoles(from, to int, role string) map[int]string {
	m := map[int]string{}
	for i := from; i < to; i++ {
		m[i] = role
	}
	return m
}

// maxOps is the most operations one structure run applies.
const maxOps = 1000

func unknownOp(op Op, ops ...string) error {
	return inputError(fmt.Sprintf("unknown operation %q (use %s)", op.Name, strings.Join(ops, ", ")))
}

func pushed(r trace.Recorder, v int) { r.Record(trace.Event{Kind: trace.Push, Values: []int{v}}) }

func popped(r trace.Recorder, v int) { r.Record(trace.Event{Kind: trace.Pop, Values: []int{v}}) }

var structuresExamples = map[string]string{
	"stack": "push 1, push 2, push 3, pop, peek", "queue": "enqueue 1, enqueue 2, enqueue 3, dequeue, peek",
	"deque": "push-front 1, push-back 2, push-front 0, pop-back, pop-front", "min-heap": "push 5, push 3, push 8, push 1, pop, pop",
	"set": "add 3, add 1, add 3, remove 1, contains 3", "linked-list": "append 1, append 2, prepend 0, remove 1, reverse",
	"doubly-linked-list": "push-back 1, push-back 2, push-front 0, pop-back, remove 0", "sorted-list": "insert 5, insert 1, insert 4, insert 1, remove 4",
	"binary-search-tree": "insert 8, insert 3, insert 10, insert 1, insert 6, insert 14, remove 3", "avl-tree": "insert 1, insert 2, insert 3, insert 4, insert 5, insert 6, insert 7",
	"trie": "insert car, insert cart, insert care, insert cat, insert dog, prefix car", "hash-table": "put 1 10, put 2 20, put 5 50, put 9 90, put 13 130, get 5",
	"union-find": "add 0, add 1, add 2, add 3, add 4, union 0 1, union 2 3, union 1 3, connected 0 2", "lru-cache": "put 1 1, put 2 2, put 3 3, get 1, put 4 4, get 2",
}

// recorderSwitch lets the AVL tree report rotations to whichever recorder
// the current operation runs with.
type recorderSwitch struct{ r trace.Recorder }

func (s *recorderSwitch) Record(e trace.Event) {
	if s.r != nil {
		s.r.Record(e)
	}
}

// opInts reads an operation's value as exactly n whole numbers.
func opInts(op Op, n int) ([]int, error) {
	f := strings.Fields(op.Value)
	if len(f) != n {
		return nil, inputError(fmt.Sprintf("%s needs %d whole numbers", op.Name, n))
	}
	out := make([]int, n)
	for i, s := range f {
		v, err := strconv.Atoi(s)
		if err != nil {
			return nil, inputError(fmt.Sprintf("%s needs whole numbers, not %q", op.Name, op.Value))
		}
		out[i] = v
	}
	return out, nil
}

// opWord reads an operation's value as one non-empty word.
func opWord(op Op) (string, error) {
	w := strings.TrimSpace(op.Value)
	if w == "" {
		return "", inputError(op.Name + " needs a word")
	}
	return w, nil
}

var disjointOps = []struct {
	name string
	f    func(a, b int) int
}{
	{"sum", func(a, b int) int { return a + b }},
	{"product", func(a, b int) int { return a * b }},
	{"min", func(a, b int) int { return min(a, b) }},
	{"gcd", func(a, b int) int {
		for b != 0 {
			a, b = b, a%b
		}
		return max(a, -a)
	}},
}

// versionsP is what the persistent playgrounds share: every version made so
// far and the one the next update starts from.
type versionsP struct{ current int }

func (p *versionsP) switchTo(op Op, n int) (string, error) {
	v, err := opInt(op)
	if err != nil {
		return "", err
	}
	if v < 0 || v >= n {
		return "", inputError(fmt.Sprintf("there is no version %d (0..%d)", v, n-1))
	}
	p.current = v
	return fmt.Sprintf("updates now start from version %d", v), nil
}

// simEntries builds the entries of a table of simulations.
func simEntries(sims []sim) []Entry {
	out := make([]Entry, len(sims))
	for i, s := range sims {
		out[i] = s.entry()
	}
	return out
}
