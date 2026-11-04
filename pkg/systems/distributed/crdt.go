package distributed

import (
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"
)

// GCounter is a grow-only counter, a state-based CRDT: each replica counts
// its own increments in its own slot, and merging takes the larger count in
// every slot. Merge is commutative, associative and idempotent, so replicas
// that have seen the same updates agree whatever the order they merged in.
type GCounter map[string]int

// Inc adds n (not negative) to replica's slot.
//
// Time: O(1). Space: O(1).
func (g GCounter) Inc(replica string, n int) { g[replica] += n }

// Merge folds other into g, slot by slot.
//
// Time: O(replicas). Space: O(replicas).
func (g GCounter) Merge(other GCounter) {
	for k, v := range other {
		g[k] = max(g[k], v)
	}
}

// Value is the counter's total.
//
// Time: O(replicas). Space: O(1).
func (g GCounter) Value() int {
	total := 0
	for _, v := range g {
		total += v
	}
	return total
}

// ORSet is an observed-remove set, a CRDT: every add gets a unique tag, and a
// remove deletes only the tags its replica has seen. An add that a remove has
// not seen survives the merge, so concurrent add and remove resolve as "add
// wins" without coordination. Tags of removed adds are kept (tombstones);
// collecting them is left out.
type ORSet struct {
	replica string
	next    int
	adds    map[string]map[string]bool // element -> its tags
	removed map[string]bool            // tags removed
}

// NewORSet returns an empty set held by replica.
func NewORSet(replica string) *ORSet {
	return &ORSet{replica: replica, adds: map[string]map[string]bool{}, removed: map[string]bool{}}
}

// Add puts e in the set under a new tag.
//
// Time: O(1). Space: O(1).
func (s *ORSet) Add(e string) {
	s.next++
	if s.adds[e] == nil {
		s.adds[e] = map[string]bool{}
	}
	s.adds[e][s.replica+strconv.Itoa(s.next)] = true
}

// Remove takes e out by removing every tag of it this replica has seen.
//
// Time: O(tags of e). Space: O(tags of e).
func (s *ORSet) Remove(e string) {
	for tag := range s.adds[e] {
		s.removed[tag] = true
	}
}

// Contains reports whether e has a tag that was not removed.
//
// Time: O(tags of e). Space: O(1).
func (s *ORSet) Contains(e string) bool {
	for tag := range s.adds[e] {
		if !s.removed[tag] {
			return true
		}
	}
	return false
}

// Merge folds other's adds and removes into s.
//
// Time: O(tags). Space: O(tags).
func (s *ORSet) Merge(other *ORSet) {
	for e, tags := range other.adds {
		if s.adds[e] == nil {
			s.adds[e] = map[string]bool{}
		}
		maps.Copy(s.adds[e], tags)
	}
	maps.Copy(s.removed, other.removed)
}

// Elements returns what the set holds, sorted.
//
// Time: O(tags + k log k). Space: O(k).
func (s *ORSet) Elements() []string {
	var out []string
	for e := range s.adds {
		if s.Contains(e) {
			out = append(out, e)
		}
	}
	slices.Sort(out)
	return out
}

// CRDTRun is the replicas of a scripted CRDT run and what each line did.
type CRDTRun struct {
	Replicas []string `json:"replicas"`
	Steps    []string `json:"steps"`
	// Counters or Sets hold each replica's final state, in Replicas order.
	Counters []GCounter `json:"counters,omitempty"`
	Sets     [][]string `json:"sets,omitempty"`
}

// RunGCounter plays a script of one operation a line on G-counter replicas:
// "A inc 3" or "A merge B" (A takes in B's state).
//
// Time: O(lines * replicas). Space: O(replicas^2).
func RunGCounter(script string) (CRDTRun, error) {
	var run CRDTRun
	counters := map[string]GCounter{}
	get := func(r string) GCounter {
		if counters[r] == nil {
			counters[r] = GCounter{}
			run.Replicas = append(run.Replicas, r)
		}
		return counters[r]
	}
	err := script2(script, func(i int, a, op, b string) error {
		g := get(a)
		switch op {
		case "inc":
			n, err := strconv.Atoi(b)
			if err != nil || n < 0 {
				return fmt.Errorf("%w: line %d: inc takes a whole number, not negative", ErrInput, i)
			}
			g.Inc(a, n)
		case "merge":
			g.Merge(get(b))
		default:
			return fmt.Errorf("%w: line %d: want \"A inc N\" or \"A merge B\"", ErrInput, i)
		}
		run.Steps = append(run.Steps, fmt.Sprintf("%s %s %s: %s = %d", a, op, b, slots(g), g.Value()))
		return nil
	})
	for _, r := range run.Replicas {
		run.Counters = append(run.Counters, counters[r])
	}
	return run, err
}

// RunORSet plays a script of one operation a line on OR-set replicas:
// "A add x", "A remove x" or "A merge B".
//
// Time: O(lines * tags). Space: O(tags).
func RunORSet(script string) (CRDTRun, error) {
	var run CRDTRun
	sets := map[string]*ORSet{}
	get := func(r string) *ORSet {
		if sets[r] == nil {
			sets[r] = NewORSet(r)
			run.Replicas = append(run.Replicas, r)
		}
		return sets[r]
	}
	err := script2(script, func(i int, a, op, b string) error {
		s := get(a)
		switch op {
		case "add":
			s.Add(b)
		case "remove":
			s.Remove(b)
		case "merge":
			s.Merge(get(b))
		default:
			return fmt.Errorf("%w: line %d: want \"A add x\", \"A remove x\" or \"A merge B\"", ErrInput, i)
		}
		run.Steps = append(run.Steps, fmt.Sprintf("%s %s %s: {%s}", a, op, b, strings.Join(s.Elements(), ", ")))
		return nil
	})
	for _, r := range run.Replicas {
		run.Sets = append(run.Sets, sets[r].Elements())
	}
	return run, err
}

func script2(script string, do func(line int, a, op, b string) error) error {
	n := 0
	for i, line := range strings.Split(script, "\n") {
		f := strings.Fields(line)
		if len(f) == 0 || strings.HasPrefix(f[0], "#") {
			continue
		}
		if len(f) != 3 {
			return fmt.Errorf("%w: line %d: want three words, \"A op x\"", ErrInput, i+1)
		}
		if n++; n > 1000 {
			return fmt.Errorf("%w: at most 1000 lines", ErrInput)
		}
		if err := do(i+1, f[0], strings.ToLower(f[1]), f[2]); err != nil {
			return err
		}
	}
	if n == 0 {
		return fmt.Errorf("%w: the script is empty", ErrInput)
	}
	return nil
}

func slots(g GCounter) string {
	keys := make([]string, 0, len(g))
	for k := range g {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	parts := make([]string, len(keys))
	for i, k := range keys {
		parts[i] = k + ":" + strconv.Itoa(g[k])
	}
	return "{" + strings.Join(parts, " ") + "}"
}
