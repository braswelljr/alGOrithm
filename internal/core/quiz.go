package core

import (
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// Question is one "what happens next?" question about a run: the learner
// watches the run up to Step, then picks one of Choices. Answer is the index
// of the right choice, which comes from the run's real events and output, not
// from a second implementation.
type Question struct {
	// Kind names what is asked: "compare-next", "swap-next", "probe-next",
	// "visit-next", "runs-at", "head-next", "page-reference", "op-result" or
	// "result".
	Kind   string `json:"kind"`
	Prompt string `json:"prompt"`
	// Step is where the player pauses, in Unit: the number of events applied
	// ("step"), the time ("time"), the seek ("seek") or the operation
	// ("operation") shown when the question is asked.
	Step int    `json:"step"`
	Unit string `json:"unit"`
	// Shown is the array as it stands at Step, when the run works on one.
	Shown []int `json:"shown,omitempty"`
	// Context is what the learner knows at Step, in words.
	Context     string   `json:"context,omitempty"`
	Choices     []string `json:"choices"`
	Answer      int      `json:"answer"`
	Explanation string   `json:"explanation"`
}

// The units a Question's Step is counted in, matching the visualizer's
// player for each kind of run.
const (
	UnitStep      = "step"
	UnitTime      = "time"
	UnitSeek      = "seek"
	UnitOperation = "operation"
)

// maxQuestions bounds one quiz, so a large count cannot make a huge reply.
const maxQuestions = 50

// Quiz runs the entry id on in with tracing and asks up to n questions about
// what it does, in the order they come up. The same seed gives the same
// questions and the same order of choices. An entry with nothing to ask
// about gives no questions and no error.
//
// Time: O(n * E) for E recorded events (each question replays up to its step).
// Space: O(E).
func Quiz(id string, in Input, n int, seed int64) ([]Question, error) {
	e, ok := Lookup(id)
	if !ok {
		return nil, fmt.Errorf("%w: no entry %q (see `algorithm list`)", ErrInput, id)
	}
	if n <= 0 {
		n = 5
	}
	n = min(n, maxQuestions)
	res, err := execute(e, in, 0)
	if err != nil {
		return nil, err
	}
	q := &quiz{e: e, in: in, res: res, o: res.Output, ev: res.Events, rng: rand.New(rand.NewSource(seed))}
	return q.ask(n), nil
}

type quiz struct {
	e   Entry
	in  Input
	res Result
	o   Output
	ev  []trace.Event
	rng *rand.Rand
}

// candidate builds one question; false means it could not find enough
// distinct choices.
type candidate func() (Question, bool)

func (q *quiz) ask(n int) []Question {
	cands := q.candidates()
	if len(cands) == 0 {
		cands = q.resultCandidates()
	}
	var out []Question
	for _, i := range q.rng.Perm(len(cands)) {
		if len(out) == n {
			break
		}
		if qu, ok := cands[i](); ok {
			out = append(out, qu)
		}
	}
	sort.SliceStable(out, func(a, b int) bool { return out[a].Step < out[b].Step })
	return out
}

// frameKind mirrors runKind in the visualizer (run-frames.tsx): which view,
// and so which player unit, a run gets.
func (q *quiz) frameKind() string {
	o := q.o
	matrix := o.Graph != nil && o.Table != nil && len(o.Table.Cells) == len(o.Graph.Nodes)*len(o.Graph.Nodes)
	switch {
	case len(o.Timeline) > 0:
		return "timeline"
	case o.Track != nil:
		return "track"
	case o.Grid != nil:
		return "grid"
	case o.Graph != nil && !matrix:
		return "graph"
	case len(o.Points) > 0:
		return "points"
	case o.Table != nil && len(q.ev) > 0:
		return "table"
	case len(o.Matches) > 0:
		return "text"
	case len(o.Steps) > 0:
		return "structure"
	case len(o.Initial) > 0 && len(q.ev) > 0:
		if o.Index != nil {
			return "cells"
		}
		return "bars"
	}
	return "other"
}

func (q *quiz) candidates() []candidate {
	switch q.frameKind() {
	case "bars":
		return q.pairCandidates()
	case "cells":
		return q.probeCandidates()
	case "grid":
		return q.visitCandidates(len(q.o.Grid.Cost), q.cellName, q.gridNeighbors)
	case "graph":
		return q.visitCandidates(len(q.o.Graph.Nodes), func(i int) string { return q.o.Graph.Nodes[i] }, q.graphNeighbors)
	case "timeline":
		if q.e.Input == Processes && len(q.ev) == 0 {
			return q.timelineCandidates()
		}
	case "track":
		return q.trackCandidates()
	case "table":
		if q.e.Input == Pages {
			return q.pageCandidates()
		}
	case "structure":
		return q.opCandidates()
	}
	return nil
}

// end is the player's last index for this run, where a result question
// pauses.
func (q *quiz) end() (int, string) {
	switch q.frameKind() {
	case "timeline":
		if len(q.ev) == 0 {
			end := 0
			for _, s := range q.o.Timeline {
				end = max(end, s.End)
			}
			return end, UnitTime
		}
	case "track":
		return max(0, len(q.o.Track.Path)-1), UnitSeek
	case "structure":
		return max(0, len(q.o.Steps)-1), UnitOperation
	case "other":
		return 0, UnitStep
	}
	return len(q.ev), UnitStep
}

// choose returns answer and up to 3 distractors from pool (in pool
// order first, then shuffled), shuffled together, and the answer's index.
// It fails when fewer than 3 distinct choices can be made.
func (q *quiz) choose(answer string, pool []string) ([]string, int, bool) {
	const want = 4
	seen := map[string]bool{answer: true}
	choices := []string{answer}
	for _, p := range pool {
		if len(choices) == want {
			break
		}
		if !seen[p] {
			seen[p] = true
			choices = append(choices, p)
		}
	}
	if len(choices) < 3 {
		return nil, 0, false
	}
	q.rng.Shuffle(len(choices), func(i, j int) { choices[i], choices[j] = choices[j], choices[i] })
	for i, c := range choices {
		if c == answer {
			return choices, i, true
		}
	}
	return nil, 0, false
}

// arrayAt replays Swap and Set events up to step on a copy of initial.
func arrayAt(initial []int, ev []trace.Event, step int) []int {
	a := append([]int(nil), initial...)
	for _, e := range ev[:step] {
		switch e.Kind {
		case trace.Swap:
			if len(e.Indices) == 2 && inRange(e.Indices[0], len(a)) && inRange(e.Indices[1], len(a)) {
				a[e.Indices[0]], a[e.Indices[1]] = a[e.Indices[1]], a[e.Indices[0]]
			}
		case trace.Set:
			if len(e.Indices) == 1 && len(e.Values) == 1 && inRange(e.Indices[0], len(a)) {
				a[e.Indices[0]] = e.Values[0]
			}
		default:
		}
	}
	return a
}

func inRange(i, n int) bool { return i >= 0 && i < n }

// pairName names two positions the same way whichever comes first.
func pairName(i, j int) string {
	if i > j {
		i, j = j, i
	}
	return fmt.Sprintf("positions %d and %d", i, j)
}

// pairCandidates asks which two positions a sort compares or swaps next.
func (q *quiz) pairCandidates() []candidate {
	n := len(q.o.Initial)
	if n < 3 {
		return nil
	}
	var out []candidate
	for k, e := range q.ev {
		if (e.Kind != trace.Compare && e.Kind != trace.Swap) || len(e.Indices) != 2 ||
			!inRange(e.Indices[0], n) || !inRange(e.Indices[1], n) || e.Indices[0] == e.Indices[1] {
			continue
		}
		out = append(out, func() (Question, bool) {
			i, j := e.Indices[0], e.Indices[1]
			var pool []string
			for _, d := range [][2]int{{i - 1, j - 1}, {i + 1, j + 1}, {i, j + 1}, {i - 1, j}} {
				if inRange(d[0], n) && inRange(d[1], n) && d[0] != d[1] {
					pool = append(pool, pairName(d[0], d[1]))
				}
			}
			q.rng.Shuffle(len(pool), func(a, b int) { pool[a], pool[b] = pool[b], pool[a] })
			for len(pool) < 12 {
				a, b := q.rng.Intn(n), q.rng.Intn(n)
				if a != b {
					pool = append(pool, pairName(a, b))
				}
			}
			choices, ans, ok := q.choose(pairName(i, j), pool)
			if !ok {
				return Question{}, false
			}
			shown := arrayAt(q.o.Initial, q.ev, k)
			verb, kind := "compare", "compare-next"
			if e.Kind == trace.Swap {
				verb, kind = "swap", "swap-next"
			}
			return Question{
				Kind: kind, Step: k, Unit: UnitStep, Shown: shown,
				Prompt:      fmt.Sprintf("Which two positions does %s %s next?", q.e.Name, verb),
				Context:     fmt.Sprintf("The array after %d steps: %s.", k, joinInts(shown)),
				Choices:     choices,
				Answer:      ans,
				Explanation: fmt.Sprintf("It %ss %s, holding %d and %d.", verb, pairName(i, j), shown[min(i, j)], shown[max(i, j)]),
			}, true
		})
	}
	return out
}

// probeCandidates asks which position a search looks at next.
func (q *quiz) probeCandidates() []candidate {
	n := len(q.o.Initial)
	if n < 3 {
		return nil
	}
	var out []candidate
	lo, hi := 0, n-1
	var probed []int
	for k, e := range q.ev {
		k := k
		if e.Kind == trace.Mark && e.Note == "range" && len(e.Indices) == 2 {
			lo, hi = e.Indices[0], e.Indices[1]
			continue
		}
		if (e.Kind != trace.Compare && e.Kind != trace.Visit) || len(e.Indices) != 1 || !inRange(e.Indices[0], n) {
			continue
		}
		at, l, h := e.Indices[0], lo, hi
		before := append([]int(nil), probed...)
		probed = append(probed, at)
		out = append(out, func() (Question, bool) {
			var pool []string
			for i := l; i <= h; i++ {
				if inRange(i, n) && i != at {
					pool = append(pool, posName(i))
				}
			}
			q.rng.Shuffle(len(pool), func(a, b int) { pool[a], pool[b] = pool[b], pool[a] })
			for _, d := range []int{at - 1, at + 1, (l + h) / 2, l, h} {
				if inRange(d, n) {
					pool = append(pool, posName(d))
				}
			}
			for i := 0; i < n; i++ {
				pool = append(pool, posName(q.rng.Intn(n)))
			}
			choices, ans, ok := q.choose(posName(at), pool)
			if !ok {
				return Question{}, false
			}
			ctx := fmt.Sprintf("Looking for %d in %s.", q.in.Target, joinInts(q.o.Initial))
			if len(before) > 0 {
				ctx += " Already looked at " + joinInts(before) + "."
			}
			return Question{
				Kind: "probe-next", Step: k, Unit: UnitStep, Shown: q.o.Initial,
				Prompt:      fmt.Sprintf("Which position does %s look at next?", q.e.Name),
				Context:     ctx,
				Choices:     choices,
				Answer:      ans,
				Explanation: fmt.Sprintf("It looks at position %d, which holds %d.", at, q.o.Initial[at]),
			}, true
		})
	}
	return out
}

func posName(i int) string { return "position " + strconv.Itoa(i) }

// cellName names a grid cell by row and column.
func (q *quiz) cellName(i int) string {
	return fmt.Sprintf("row %d, column %d", i/q.o.Grid.Cols, i%q.o.Grid.Cols)
}

func (q *quiz) gridNeighbors(i int) []int {
	g := q.o.Grid
	r, c := i/g.Cols, i%g.Cols
	var out []int
	for _, d := range [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
		rr, cc := r+d[0], c+d[1]
		if rr >= 0 && rr < g.Rows && cc >= 0 && cc < g.Cols && g.Cost[rr*g.Cols+cc] > 0 {
			out = append(out, rr*g.Cols+cc)
		}
	}
	return out
}

func (q *quiz) graphNeighbors(i int) []int {
	var out []int
	for _, e := range q.o.Graph.Edges {
		switch {
		case e.From == i:
			out = append(out, e.To)
		case e.To == i && !q.o.Graph.Directed:
			out = append(out, e.From)
		}
	}
	return out
}

// visitCandidates asks which cell or node a search explores next. The
// distractors are what it could have explored instead: the frontier first,
// then the neighbors of what it has seen.
func (q *quiz) visitCandidates(size int, name func(int) string, neighbors func(int) []int) []candidate {
	if size < 3 {
		return nil
	}
	var out []candidate
	var pushedOrder, visitedOrder []int
	pushed := map[int]bool{}
	for k, e := range q.ev {
		k := k
		if len(e.Indices) == 0 || !inRange(e.Indices[0], size) {
			continue
		}
		at := e.Indices[0]
		switch e.Kind {
		case trace.Push:
			if !pushed[at] {
				pushed[at] = true
				pushedOrder = append(pushedOrder, at)
			}
			continue
		case trace.Visit:
		default:
			continue
		}
		// Prefixes of slices only ever appended to: what was known at step k.
		pushedNow, visitedNow := pushedOrder[:len(pushedOrder):len(pushedOrder)], visitedOrder[:len(visitedOrder):len(visitedOrder)]
		visitedOrder = append(visitedOrder, at)
		out = append(out, func() (Question, bool) {
			visited := map[int]bool{}
			for _, v := range visitedNow {
				visited[v] = true
			}
			var frontier, near []int
			for _, p := range pushedNow {
				if !visited[p] && p != at {
					frontier = append(frontier, p)
				}
			}
			for _, v := range visitedNow {
				for _, w := range neighbors(v) {
					if !visited[w] && w != at {
						near = append(near, w)
					}
				}
			}
			var pool []string
			q.rng.Shuffle(len(frontier), func(a, b int) { frontier[a], frontier[b] = frontier[b], frontier[a] })
			for _, f := range frontier {
				pool = append(pool, name(f))
			}
			q.rng.Shuffle(len(near), func(a, b int) { near[a], near[b] = near[b], near[a] })
			for _, v := range near {
				pool = append(pool, name(v))
			}
			for _, v := range neighbors(at) {
				pool = append(pool, name(v))
			}
			for i := 0; i < 8; i++ {
				pool = append(pool, name(q.rng.Intn(size)))
			}
			seen := len(visitedNow)
			choices, ans, ok := q.choose(name(at), pool)
			if !ok {
				return Question{}, false
			}
			return Question{
				Kind: "visit-next", Step: k, Unit: UnitStep,
				Prompt:      fmt.Sprintf("Which one does %s explore next?", q.e.Name),
				Context:     fmt.Sprintf("%d explored so far; %d waiting to be explored.", seen, len(frontier)+1),
				Choices:     choices,
				Answer:      ans,
				Explanation: fmt.Sprintf("It explores %s next.", name(at)),
			}, true
		})
	}
	return out
}

// timelineCandidates asks which process a CPU scheduler runs at the start
// of each slice.
func (q *quiz) timelineCandidates() []candidate {
	var lanes []string
	seen := map[string]bool{}
	for _, s := range q.o.Timeline {
		if !seen[s.Lane] {
			seen[s.Lane] = true
			lanes = append(lanes, s.Lane)
		}
	}
	var out []candidate
	for i, s := range q.o.Timeline {
		s := s
		if s.State != "" && s.State != "running" {
			continue
		}
		if i > 0 && q.o.Timeline[i-1].Lane == s.Lane && q.o.Timeline[i-1].End == s.Start {
			continue // the same process carrying on is not a decision
		}
		out = append(out, func() (Question, bool) {
			pool := append([]string(nil), lanes...)
			q.rng.Shuffle(len(pool), func(a, b int) { pool[a], pool[b] = pool[b], pool[a] })
			choices, ans, ok := q.choose(s.Lane, pool)
			if !ok {
				return Question{}, false
			}
			return Question{
				Kind: "runs-at", Step: s.Start, Unit: UnitTime,
				Prompt:      fmt.Sprintf("At time %d, which process does %s run?", s.Start, q.e.Name),
				Context:     "The processes (name, arrival, burst):\n" + strings.TrimSpace(q.in.Text),
				Choices:     choices,
				Answer:      ans,
				Explanation: fmt.Sprintf("%s runs from time %d to %d.", s.Lane, s.Start, s.End),
			}, true
		})
	}
	return out
}

// trackCandidates asks where a disk head goes next.
func (q *quiz) trackCandidates() []candidate {
	t := q.o.Track
	var out []candidate
	for k := 0; k+1 < len(t.Path); k++ {
		k := k
		at, next := t.Path[k], t.Path[k+1]
		pending := map[int]int{}
		for _, v := range q.in.Values {
			pending[v]++
		}
		for _, v := range t.Path[1 : k+1] {
			pending[v]--
		}
		out = append(out, func() (Question, bool) {
			var left []int
			for v, c := range pending {
				if c > 0 && v != next {
					left = append(left, v)
				}
			}
			// Closest first: the requests a head could plausibly go to.
			sort.Slice(left, func(a, b int) bool {
				da, db := abs(left[a]-at), abs(left[b]-at)
				return da < db || (da == db && left[a] < left[b])
			})
			var pool []string
			for _, v := range left {
				pool = append(pool, trackName(v))
			}
			pool = append(pool, trackName(0), trackName(max(0, t.Size-1)))
			choices, ans, ok := q.choose(trackName(next), pool)
			if !ok {
				return Question{}, false
			}
			var waiting []int
			for _, v := range q.in.Values {
				if pending[v] > 0 {
					waiting = append(waiting, v)
				}
			}
			sort.Ints(waiting)
			return Question{
				Kind: "head-next", Step: k, Unit: UnitSeek,
				Prompt:      fmt.Sprintf("The head is at track %d. Where does %s send it next?", at, q.e.Name),
				Context:     "Requests still waiting: " + joinInts(waiting) + ".",
				Choices:     choices,
				Answer:      ans,
				Explanation: fmt.Sprintf("It moves the head from track %d to track %d, %d tracks.", at, next, abs(next-at)),
			}, true
		})
	}
	return out
}

func trackName(v int) string { return "track " + strconv.Itoa(v) }

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// pageCandidates asks, for each page reference, whether it hits or faults
// and which page it evicts.
func (q *quiz) pageCandidates() []candidate {
	t := q.o.Table
	refs := q.in.Values
	if t.Cols != len(refs) || t.Rows == 0 {
		return nil
	}
	blank := -1
	if t.Blank != nil {
		blank = *t.Blank
	}
	column := func(c int) []int {
		var out []int
		if c < 0 {
			return out
		}
		for r := 0; r < t.Rows; r++ {
			if v := t.Cells[r*t.Cols+c]; v != blank {
				out = append(out, v)
			}
		}
		return out
	}
	first := map[int]int{} // column -> its first event
	for k, e := range q.ev {
		if len(e.Indices) > 0 && e.Indices[0] >= 0 && e.Indices[0] < len(t.Cells) {
			if _, ok := first[e.Indices[0]%t.Cols]; !ok {
				first[e.Indices[0]%t.Cols] = k
			}
		}
	}
	var out []candidate
	for c, page := range refs {
		page := page
		step, ok := first[c]
		if !ok {
			continue
		}
		before, after := column(c-1), column(c)
		answer, why, ok := pageOutcome(page, before, after, t.Rows)
		if !ok {
			continue
		}
		out = append(out, func() (Question, bool) {
			pool := []string{"hit", "fault, into an empty frame"}
			ev := append([]int(nil), before...)
			q.rng.Shuffle(len(ev), func(a, b int) { ev[a], ev[b] = ev[b], ev[a] })
			for _, v := range ev {
				pool = append(pool, fmt.Sprintf("fault, evicting page %d", v))
			}
			if len(before) == t.Rows { // full: an empty frame is not plausible
				pool = pool[:1:1]
				for _, v := range ev {
					pool = append(pool, fmt.Sprintf("fault, evicting page %d", v))
				}
			}
			choices, ans, ok := q.choose(answer, pool)
			if !ok {
				return Question{}, false
			}
			held := "nothing"
			if len(before) > 0 {
				held = joinInts(before)
			}
			return Question{
				Kind: "page-reference", Step: step, Unit: UnitStep,
				Prompt:      fmt.Sprintf("Page %d is referenced. What does %s do?", page, q.e.Name),
				Context:     fmt.Sprintf("The %d frames hold: %s.", t.Rows, held),
				Choices:     choices,
				Answer:      ans,
				Explanation: why,
			}, true
		})
	}
	return out
}

// pageOutcome says what a reference did, from the frames before and after
// it; false when the two do not fit any single outcome.
func pageOutcome(page int, before, after []int, frames int) (string, string, bool) {
	has := func(s []int, v int) bool {
		for _, x := range s {
			if x == v {
				return true
			}
		}
		return false
	}
	if has(before, page) {
		return "hit", fmt.Sprintf("Page %d is already in a frame, so it is a hit.", page), true
	}
	if !has(after, page) {
		return "", "", false
	}
	var gone []int
	for _, v := range before {
		if !has(after, v) {
			gone = append(gone, v)
		}
	}
	switch {
	case len(gone) == 0 && len(before) < frames:
		return "fault, into an empty frame", fmt.Sprintf("Page %d is not in memory and a frame is free, so it faults into the empty frame.", page), true
	case len(gone) == 1:
		return fmt.Sprintf("fault, evicting page %d", gone[0]), fmt.Sprintf("Page %d is not in memory and the frames are full, so it faults and evicts page %d.", page, gone[0]), true
	}
	return "", "", false
}

// readOps are operations that answer with a value from the structure.
var readOps = map[string]bool{
	"get": true, "find": true, "search": true, "lookup": true, "query": true, "sum": true,
	"rank": true, "select": true, "min": true, "max": true, "range": true, "count": true,
}

// opCandidates asks what a reading operation on a structure returns.
func (q *quiz) opCandidates() []candidate {
	steps := q.o.Steps
	var out []candidate
	for k := 1; k < len(steps); k++ {
		k := k
		st := steps[k]
		if st.Error != "" || (st.Op.Value != "" && !readOps[st.Op.Name]) {
			continue
		}
		answer, ok := opAnswer(st)
		if !ok {
			continue
		}
		out = append(out, func() (Question, bool) {
			var pool []string
			for _, nd := range steps[k-1].Snapshot.Nodes {
				for _, f := range strings.FieldsFunc(nd.Label, func(r rune) bool { return r != '-' && (r < '0' || r > '9') }) {
					if _, err := strconv.Atoi(f); err == nil {
						pool = append(pool, f)
					}
				}
			}
			q.rng.Shuffle(len(pool), func(a, b int) { pool[a], pool[b] = pool[b], pool[a] })
			var others []string
			for _, s := range steps[1:] {
				if s.Op.Value != "" {
					others = append(others, strings.Fields(s.Op.Value)...)
				}
			}
			for _, v := range q.in.Values {
				others = append(others, strconv.Itoa(v))
			}
			q.rng.Shuffle(len(others), func(a, b int) { others[a], others[b] = others[b], others[a] })
			pool = append(pool, others...)
			if n, err := strconv.Atoi(answer); err == nil {
				pool = append(pool, strconv.Itoa(n+1), strconv.Itoa(n-1))
			}
			choices, ans, ok := q.choose(answer, pool)
			if !ok {
				return Question{}, false
			}
			var done []string
			for _, s := range steps[1:k] {
				done = append(done, opName(s.Op))
			}
			ctx := "No operations yet."
			if len(done) > 0 {
				ctx = "Operations so far: " + strings.Join(done, ", ") + "."
			}
			if len(q.in.Values) > 0 {
				ctx = "It starts with " + joinInts(q.in.Values) + ". " + ctx
			}
			return Question{
				Kind: "op-result", Step: k - 1, Unit: UnitOperation,
				Prompt:      fmt.Sprintf("What does `%s` give back?", opName(st.Op)),
				Context:     ctx,
				Choices:     choices,
				Answer:      ans,
				Explanation: fmt.Sprintf("The %s says: %s.", strings.ToLower(q.e.Name), st.Result),
			}, true
		})
	}
	return out
}

func opName(o Op) string {
	if o.Value == "" {
		return o.Name
	}
	return o.Name + " " + o.Value
}

// opAnswer is the first number in an operation's result, after the key it
// was asked about ("5 is 50" answers 50).
func opAnswer(st Step) (string, bool) {
	r := st.Result
	if st.Op.Value != "" {
		r = strings.TrimPrefix(r, st.Op.Value)
	}
	for _, f := range strings.FieldsFunc(r, func(c rune) bool { return c != '-' && (c < '0' || c > '9') }) {
		if _, err := strconv.Atoi(f); err == nil {
			return f, true
		}
	}
	return "", false
}

// resultCandidates asks for the run's result when nothing step by step can
// be asked: its number, its index, or one of its summary figures.
func (q *quiz) resultCandidates() []candidate {
	end, unit := q.end()
	ask := func(prompt, answer string, near []string, why string) candidate {
		return func() (Question, bool) {
			pool := append([]string(nil), near...)
			q.rng.Shuffle(len(pool), func(a, b int) { pool[a], pool[b] = pool[b], pool[a] })
			choices, ans, ok := q.choose(answer, pool)
			if !ok {
				return Question{}, false
			}
			return Question{
				Kind: "result", Step: end, Unit: unit, Prompt: prompt,
				Context:     "The input: " + describeInput(q.e, q.in) + ".",
				Choices:     choices,
				Answer:      ans,
				Explanation: why,
			}, true
		}
	}
	switch {
	case q.o.Number != nil:
		n := *q.o.Number
		return []candidate{ask(fmt.Sprintf("What number does %s give?", q.e.Name), strconv.Itoa(n), nearInts(n),
			fmt.Sprintf("%s gives %d.", q.e.Name, n))}
	case q.o.Index != nil:
		i := *q.o.Index
		near := []string{"not found"}
		for _, d := range []int{-2, -1, 1, 2} {
			if inRange(i+d, len(q.o.Initial)) {
				near = append(near, posName(i+d))
			}
		}
		answer, why := "not found", "The value is not there."
		if i >= 0 {
			answer, why = posName(i), fmt.Sprintf("It is at position %d.", i)
		}
		return []candidate{ask(fmt.Sprintf("Where does %s find it?", q.e.Name), answer, near, why)}
	case len(q.o.Summary) > 0:
		keys := make([]string, 0, len(q.o.Summary))
		for k := range q.o.Summary {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var out []candidate
		for _, k := range keys {
			v := q.o.Summary[k]
			var near []string
			for _, f := range []float64{0.5, 0.75, 1.25, 1.5, 2} {
				near = append(near, fmtFloat(v*f))
			}
			near = append(near, fmtFloat(v+1), fmtFloat(v-1))
			out = append(out, ask(fmt.Sprintf("What is the %s?", k), fmtFloat(v), near, fmt.Sprintf("The %s is %s.", k, fmtFloat(v))))
		}
		return out
	}
	return nil
}

// nearInts are plausible wrong answers near n.
func nearInts(n int) []string {
	step := max(1, abs(n)/10)
	return []string{
		strconv.Itoa(n + 1), strconv.Itoa(n - 1), strconv.Itoa(n + step), strconv.Itoa(n - step),
		strconv.Itoa(n * 2), strconv.Itoa(n / 2), strconv.Itoa(n + 2*step),
	}
}

func fmtFloat(v float64) string { return strconv.FormatFloat(v, 'f', -1, 64) }

// describeInput says in a few words what a run was given.
func describeInput(e Entry, in Input) string {
	var parts []string
	if len(in.Values) > 0 {
		parts = append(parts, "values "+joinInts(in.Values))
	}
	if in.Target != 0 {
		parts = append(parts, "target "+strconv.Itoa(in.Target))
	}
	if in.Text != "" {
		t := strings.TrimSpace(in.Text)
		if len(t) > 60 {
			t = t[:60] + "..."
		}
		parts = append(parts, fmt.Sprintf("text %q", t))
	}
	if in.Other != "" {
		parts = append(parts, fmt.Sprintf("and %q", in.Other))
	}
	keys := make([]string, 0, len(in.Options))
	for k := range in.Options {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%d", k, in.Options[k]))
	}
	if len(parts) == 0 {
		return e.Name + "'s defaults"
	}
	return strings.Join(parts, ", ")
}
