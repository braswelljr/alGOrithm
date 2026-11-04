package core

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/structures/arrays"
	"github.com/braswelljr/alGOrithm/pkg/structures/sets"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

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

// asSet sorts s and drops repeats: the sorted-set form the set operations take.
func asSet[T int | string](s []T) []T {
	s = slices.Clone(s)
	slices.Sort(s)
	return slices.Compact(s)
}

func goInts(s []int) string { return "[]int{" + strings.ReplaceAll(joinInts(s), " ", ", ") + "}" }

func goStrings(s []string) string {
	q := make([]string, len(s))
	for i, v := range s {
		q[i] = strconv.Quote(v)
	}
	return "[]string{" + strings.Join(q, ", ") + "}"
}

func setEntries() []Entry {
	setOp := func(id, name, fn, summary string, traced func(a, b []int, r trace.Recorder) []int) Entry {
		e := entry(id, name, "structures/sets/sorted.go", fn)
		e.Summary, e.Time, e.Space, e.Input, e.Traced = summary, "O(n + m)", "O(n + m)", TwoTexts, true
		e.Example = &Input{Text: "1 3 5 7 9 11", Other: "3 4 5 6 11 12"}
		e.Limits = "at most 10000 values each"
		read := func(in Input) ([]int, []int, error) {
			a, err := fieldInts(in.Text, "first set")
			if err != nil {
				return nil, nil, err
			}
			b, err := fieldInts(in.Other, "second set")
			if err != nil {
				return nil, nil, err
			}
			if len(a) > 10_000 || len(b) > 10_000 {
				return nil, nil, inputError("at most 10000 values each")
			}
			return asSet(a), asSet(b), nil
		}
		e.Run = func(in Input, r trace.Recorder) (Output, error) {
			a, b, err := read(in)
			if err != nil {
				return Output{}, err
			}
			got := traced(a, b, r)
			// The two inputs as a table, -1 padding the shorter row, so the
			// comparisons (Indices[0] in a, Indices[1] in b) have something to point at.
			cols := max(len(a), len(b), 1)
			t := &TableOutput{Rows: 2, Cols: cols, Cells: make([]int, 2*cols), RowLabels: []string{"a", "b"}, ColLabels: numberLabels(cols, 0)}
			for i := range t.Cells {
				t.Cells[i] = -1
			}
			copy(t.Cells, a)
			copy(t.Cells[cols:], b)
			return Output{Initial: make([]int, len(got)), Values: got, Table: t, Number: ptr(len(got))}, nil
		}
		e.Call = func(in Input) (string, string) {
			a, b, _ := read(in)
			return "", fmt.Sprintf("sets.%s(%s, %s)", fn, goInts(a), goInts(b))
		}
		return e
	}

	power := entry("power-set", "Power Set", "structures/sets/power.go", "PowerSet")
	power.Summary, power.Time, power.Space, power.Input = "Every subset, by counting a bitmask from 0 to 2^n - 1: bit i says whether value i is in.", "O(n * 2^n)", "O(n * 2^n)", Ints
	power.Example, power.Limits = &Input{Values: []int{1, 2, 3}}, "at most 12 values"
	power.Run = func(in Input, _ trace.Recorder) (Output, error) {
		if err := values(in, "take subsets of", 12); err != nil {
			return Output{}, err
		}
		subs, _ := sets.PowerSet(in.Values)
		lines := make([]string, len(subs))
		for i, s := range subs {
			lines[i] = fmt.Sprintf("%0*b  {%s}", len(in.Values), i, strings.ReplaceAll(joinInts(s), " ", ", "))
		}
		return Output{Lines: lines, Number: ptr(len(subs))}, nil
	}
	power.Call = func(in Input) (string, string) { return "", "sets.PowerSet(" + goInts(in.Values) + ")" }

	product := entry("cartesian-product", "Cartesian Product", "structures/sets/power.go", "CartesianProduct")
	product.Summary, product.Time, product.Space, product.Input = "Every pair (x, y) with x from the first set and y from the second.", "O(n * m)", "O(n * m)", TwoTexts
	product.Example = &Input{Text: "red green", Other: "S M L"}
	product.Run = func(in Input, _ trace.Recorder) (Output, error) {
		a, b := strings.Fields(in.Text), strings.Fields(in.Other)
		if len(a)*len(b) > 10_000 {
			return Output{}, inputError("at most 10000 pairs")
		}
		pairs, _ := sets.CartesianProduct(a, b)
		lines := make([]string, len(pairs))
		for i, p := range pairs {
			lines[i] = "(" + p.First + ", " + p.Second + ")"
		}
		return Output{Lines: lines, Number: ptr(len(pairs))}, nil
	}
	product.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("sets.CartesianProduct(%s, %s)", goStrings(strings.Fields(in.Text)), goStrings(strings.Fields(in.Other)))
	}

	jac := entry("jaccard-similarity", "Jaccard Similarity", "structures/sets/sorted.go", "Jaccard")
	jac.Summary, jac.Time, jac.Space, jac.Input = "How alike two sets are: the size of their intersection over the size of their union.", "O(n + m)", "O(n + m)", TwoTexts
	jac.Example = &Input{Text: "apple banana cherry date", Other: "banana cherry fig"}
	jac.Run = func(in Input, _ trace.Recorder) (Output, error) {
		a, b := asSet(strings.Fields(in.Text)), asSet(strings.Fields(in.Other))
		inter, union := sets.Intersection(a, b), sets.Union(a, b)
		return Output{
			Lines:   []string{"both: " + strings.Join(inter, " "), "either: " + strings.Join(union, " ")},
			Summary: map[string]float64{"jaccard": sets.Jaccard(a, b), "intersection": float64(len(inter)), "union": float64(len(union))},
		}, nil
	}
	jac.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("sets.Jaccard(%s, %s)", goStrings(asSet(strings.Fields(in.Text))), goStrings(asSet(strings.Fields(in.Other))))
	}

	inSets := func(e Entry, example string) Entry {
		e.Category = "sets"
		e.Example = &Input{Ops: ParseOps(example)}
		return e
	}
	return []Entry{
		inSets(structureEntry("bitset", "Bitset", "structures/sets/bitset.go", "Bitset", "A set of small whole numbers, one bit each: membership in O(1), set operations 64 members at a time.",
			"O(1) set, clear, test; O(n/64) count", "O(max / 64)", "set", []string{"set N", "clear N", "test N", "count", "members"},
			func(Input) playground { return &bitsetP{sets.NewBitset()} }), "set 1, set 3, set 5, set 12, clear 3, test 5, count"),
		inSets(structureEntry("multiset", "Multiset (Bag)", "structures/sets/multiset.go", "Multiset", "A set where a value can occur more than once: each value with a count.",
			"O(1) average add, remove, count", "O(distinct)", "add", []string{"add VALUE", "remove VALUE", "count VALUE"},
			func(Input) playground { return &multisetP{sets.NewMultiset[int]()} }), "add 3, add 1, add 3, add 3, remove 3, count 3"),
		inSets(structureEntry("ordered-set", "Ordered Set", "structures/sets/orderedset.go", "OrderedSet", "A sorted set: binary search for membership, rank, floor, ceiling and ranges.",
			"O(log n) search, O(n) insert", "O(n)", "add", []string{"add VALUE", "remove VALUE", "contains VALUE", "rank VALUE", "floor VALUE", "ceiling VALUE", "range LO HI"},
			func(Input) playground { return &orderedP{sets.NewOrderedSet[int]()} }), "add 20, add 5, add 12, add 30, add 5, floor 13, ceiling 13, rank 12, range 6 25"),
		setOp("sorted-union", "Union (two pointers)", "Union", "The values in either set, walking both sorted sets at once like a merge.", sets.TraceUnion),
		setOp("sorted-intersection", "Intersection (two pointers)", "Intersection", "The values in both sets, walking both sorted sets at once.", sets.TraceIntersection),
		setOp("sorted-difference", "Difference (two pointers)", "Difference", "The values in the first set but not the second.", sets.TraceDifference),
		setOp("sorted-symmetric-difference", "Symmetric Difference (two pointers)", "SymmetricDifference", "The values in exactly one of the two sets.", sets.TraceSymmetricDifference),
		power, product, jac,
	}
}

type bitsetP struct{ b *sets.Bitset }

func (p *bitsetP) Snapshot() trace.Snapshot { return p.b.Snapshot() }
func (p *bitsetP) apply(op Op, r trace.Recorder) (string, error) {
	switch op.Name {
	case "count":
		return fmt.Sprintf("%d members (popcount)", p.b.Count()), nil
	case "members":
		return "members: " + joinInts(p.b.Members()), nil
	case "set", "clear", "test":
		n, err := opInt(op)
		if err != nil {
			return "", err
		}
		if n < 0 || n > 4095 {
			return "", inputError("members are 0 to 4095 here")
		}
		switch op.Name {
		case "set":
			if !p.b.Set(n) {
				return fmt.Sprintf("bit %d was already set", n), nil
			}
			pushed(r, n)
			return fmt.Sprintf("set bit %d (word %d, bit %d)", n, n/64, n%64), nil
		case "clear":
			if !p.b.Clear(n) {
				return fmt.Sprintf("bit %d was not set", n), nil
			}
			popped(r, n)
			return fmt.Sprintf("cleared bit %d", n), nil
		}
		return fmt.Sprintf("bit %d is set: %v", n, p.b.Test(n)), nil
	}
	return "", unknownOp(op, "set", "clear", "test", "count", "members")
}

type multisetP struct{ m *sets.Multiset[int] }

func (p *multisetP) Snapshot() trace.Snapshot { return p.m.Snapshot() }
func (p *multisetP) apply(op Op, r trace.Recorder) (string, error) {
	if op.Name != "add" && op.Name != "remove" && op.Name != "count" {
		return "", unknownOp(op, "add", "remove", "count")
	}
	v, err := opInt(op)
	if err != nil {
		return "", err
	}
	switch op.Name {
	case "add":
		pushed(r, v)
		return fmt.Sprintf("added %d (now %d of it)", v, p.m.Add(v)), nil
	case "remove":
		if !p.m.Remove(v) {
			return fmt.Sprintf("%d is not in the bag", v), nil
		}
		popped(r, v)
		return fmt.Sprintf("removed one %d (%d left)", v, p.m.Count(v)), nil
	}
	return fmt.Sprintf("%d occurs %d times", v, p.m.Count(v)), nil
}

type orderedP struct{ s *sets.OrderedSet[int] }

func (p *orderedP) Snapshot() trace.Snapshot { return p.s.Snapshot() }
func (p *orderedP) apply(op Op, r trace.Recorder) (string, error) {
	if op.Name == "range" {
		lh, err := opPair(op)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("members in [%d, %d]: %s", lh[0], lh[1], joinInts(p.s.Range(lh[0], lh[1]))), nil
	}
	switch op.Name {
	case "add", "remove", "contains", "rank", "floor", "ceiling":
	default:
		return "", unknownOp(op, "add", "remove", "contains", "rank", "floor", "ceiling", "range")
	}
	v, err := opInt(op)
	if err != nil {
		return "", err
	}
	near := func(name string, x int, ok bool) string {
		if !ok {
			return fmt.Sprintf("no %s of %d", name, v)
		}
		return fmt.Sprintf("%s of %d is %d", name, v, x)
	}
	switch op.Name {
	case "add":
		if !p.s.Add(v) {
			return fmt.Sprintf("%d is already in the set", v), nil
		}
		pushed(r, v)
		return fmt.Sprintf("added %d at position %d", v, p.s.Rank(v)), nil
	case "remove":
		if !p.s.Remove(v) {
			return fmt.Sprintf("%d is not in the set", v), nil
		}
		popped(r, v)
		return fmt.Sprintf("removed %d", v), nil
	case "contains":
		return fmt.Sprintf("contains %d: %v", v, p.s.Contains(v)), nil
	case "rank":
		return fmt.Sprintf("%d members are smaller than %d", p.s.Rank(v), v), nil
	case "floor":
		x, ok := p.s.Floor(v)
		return near("floor", x, ok), nil
	}
	x, ok := p.s.Ceiling(v)
	return near("ceiling", x, ok), nil
}

// callback is a named JavaScript-style callback: its behavior, and the Go
// literal that does the same, for snippets.
type callback struct {
	keep func(v, target int) bool // for filter, find, some, every
	f    func(v int) int          // for map
	src  string
}

var predicates = map[string]callback{
	"even":     {keep: func(v, _ int) bool { return v%2 == 0 }, src: "func(v, _ int) bool { return v%2 == 0 }"},
	"odd":      {keep: func(v, _ int) bool { return v%2 != 0 }, src: "func(v, _ int) bool { return v%2 != 0 }"},
	"positive": {keep: func(v, _ int) bool { return v > 0 }, src: "func(v, _ int) bool { return v > 0 }"},
	"negative": {keep: func(v, _ int) bool { return v < 0 }, src: "func(v, _ int) bool { return v < 0 }"},
	"greater":  {keep: func(v, t int) bool { return v > t }, src: "func(v, _ int) bool { return v > %d }"},
}

var mappers = map[string]callback{
	"double":    {f: func(v int) int { return 2 * v }, src: "func(v, _ int) int { return 2 * v }"},
	"square":    {f: func(v int) int { return v * v }, src: "func(v, _ int) int { return v * v }"},
	"negate":    {f: func(v int) int { return -v }, src: "func(v, _ int) int { return -v }"},
	"increment": {f: func(v int) int { return v + 1 }, src: "func(v, _ int) int { return v + 1 }"},
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

func jsArrayEntries() []Entry {
	ex := []int{3, -1, 4, 1, -5, 9, 2, 6}
	// visitEntry is a higher-order method that calls a callback once per
	// element: Steps show each call with the array, the current element, and
	// the callback's answer.
	visitEntry := func(id, name, fn, summary, key, keys string, run func(in Input) (Output, error), call func(in Input) string) Entry {
		e := entry(id, name, "structures/arrays/js.go", fn)
		e.Summary, e.Time, e.Space, e.Input, e.Params = summary+" The callback, by name ("+keys+"), goes in the keyword.", "O(n)", "O(n)", Ints, []string{"key"}
		e.Example = &Input{Values: ex, Key: key, Target: 4}
		e.Choices = strings.Split(keys, ", ")
		e.Limits = "at most 1000 values"
		e.Run = func(in Input, _ trace.Recorder) (Output, error) {
			if err := values(in, "go through", 1000); err != nil {
				return Output{}, err
			}
			return run(in)
		}
		e.Call = func(in Input) (string, string) { return "", call(in) }
		return e
	}
	pick := func(m map[string]callback, in Input, def string) (callback, error) {
		k := strings.ToLower(strings.TrimSpace(in.Key))
		if k == "" {
			k = def
		}
		c, ok := m[k]
		if !ok {
			return callback{}, inputError(fmt.Sprintf("unknown callback %q (use %s)", in.Key, names(m)))
		}
		return c, nil
	}
	srcOf := func(c callback, in Input) string {
		if strings.Contains(c.src, "%d") {
			return fmt.Sprintf(c.src, in.Target)
		}
		return c.src
	}
	//nolint:unparam // predicateRun passes the kept positions; unparam misreads the closure call
	step := func(vals []int, i int, result string, roles map[int]string) Step {
		r := map[int]string{i: "current"}
		for k, v := range roles {
			if k != i {
				r[k] = v
			}
		}
		return Step{Op: Op{Name: "visit", Value: fmt.Sprintf("index %d: %d", i, vals[i])}, Result: result, Snapshot: arrayRow(vals, r)}
	}
	// predicateRun covers filter, find, findIndex, some and every: walk while
	// stop says to go on, mark the elements kept.
	predicateRun := func(mode string) func(in Input) (Output, error) {
		return func(in Input) (Output, error) {
			c, err := pick(predicates, in, "even")
			if err != nil {
				return Output{}, err
			}
			kept := map[int]string{}
			out := Output{Initial: slices.Clone(in.Values)}
			var result []int
			found := -1
			for i, v := range in.Values {
				ok := c.keep(v, in.Target)
				if ok {
					kept[i] = "kept"
					result = append(result, v)
				}
				out.Steps = append(out.Steps, step(in.Values, i, fmt.Sprintf("callback(%d) = %v", v, ok), kept))
				if (mode == "find" || mode == "some") && ok || mode == "every" && !ok {
					found = i
					break
				}
			}
			switch mode {
			case "filter":
				out.Values = result
				if out.Values == nil {
					out.Values = []int{}
				}
			case "find":
				out.Index = &found
				if found >= 0 {
					out.Number = ptr(in.Values[found])
				}
			case "some":
				out.Bool = ptr(found >= 0)
			case "every":
				out.Bool = ptr(found < 0)
			}
			return out, nil
		}
	}

	jsMap := visitEntry("js-map", "Array.prototype.map", "Map", "A new array of the callback applied to each element.", "double", names(mappers),
		func(in Input) (Output, error) {
			c, err := pick(mappers, in, "double")
			if err != nil {
				return Output{}, err
			}
			out := Output{Initial: slices.Clone(in.Values)}
			var res []int
			for i, v := range in.Values {
				res = append(res, c.f(v))
				out.Steps = append(out.Steps, step(in.Values, i, fmt.Sprintf("callback(%d) = %d", v, c.f(v)), nil))
			}
			out.Values = res
			return out, nil
		},
		func(in Input) string {
			c, _ := pick(mappers, in, "double")
			return fmt.Sprintf("arrays.Map(arrays.Of(%s...), %s).Values()", goInts(in.Values), c.src)
		})

	predCall := func(method string) func(in Input) string {
		return func(in Input) string {
			c, _ := pick(predicates, in, "even")
			return fmt.Sprintf("arrays.Of(%s...).%s(%s)", goInts(in.Values), method, srcOf(c, in))
		}
	}
	jsFilter := visitEntry("js-filter", "Array.prototype.filter", "Array.Filter", "A new array of the elements the callback keeps.", "even", names(predicates), predicateRun("filter"),
		func(in Input) string { return predCall("Filter")(in) + ".Values()" })
	jsFind := visitEntry("js-find", "Array.prototype.find", "Array.Find", "The first element the callback accepts, stopping there (undefined if none; here ok is false).", "greater", names(predicates), predicateRun("find"), predCall("Find"))
	jsFindIndex := visitEntry("js-find-index", "Array.prototype.findIndex", "Array.FindIndex", "The index of the first element the callback accepts, or -1.", "even", names(predicates), predicateRun("find"), predCall("FindIndex"))
	jsSome := visitEntry("js-some", "Array.prototype.some", "Array.Some", "Whether the callback accepts at least one element; it stops at the first (false for []).", "negative", names(predicates), predicateRun("some"), predCall("Some"))
	jsEvery := visitEntry("js-every", "Array.prototype.every", "Array.Every", "Whether the callback accepts every element; it stops at the first refusal (true for []).", "positive", names(predicates), predicateRun("every"), predCall("Every"))
	for _, e := range []*Entry{&jsFind, &jsSome, &jsEvery} {
		e.Params = append(e.Params, "target")
	}

	jsReduce := visitEntry("js-reduce", "Array.prototype.reduce", "Reduce", "Fold the elements from the left into one value: acc = callback(acc, element).", "sum", names(reducers),
		func(in Input) (Output, error) {
			k := strings.ToLower(strings.TrimSpace(in.Key))
			if k == "" {
				k = "sum"
			}
			rd, ok := reducers[k]
			if !ok {
				return Output{}, inputError(fmt.Sprintf("unknown reducer %q (use %s)", in.Key, names(reducers)))
			}
			out := Output{Initial: slices.Clone(in.Values)}
			acc, from := rd.init, 0
			if rd.fromFirst {
				acc, from = in.Values[0], 1
				out.Steps = append(out.Steps, step(in.Values, 0, fmt.Sprintf("no initial value: start from %d", acc), nil))
			}
			for i := from; i < len(in.Values); i++ {
				v := in.Values[i]
				next := rd.f(acc, v)
				out.Steps = append(out.Steps, step(in.Values, i, fmt.Sprintf("callback(%d, %d) = %d", acc, v, next), nil))
				acc = next
			}
			out.Number = &acc
			return out, nil
		},
		nil)
	jsReduce.Call = func(in Input) (string, string) {
		rd, ok := reducers[strings.ToLower(strings.TrimSpace(in.Key))]
		if !ok {
			rd = reducers["sum"]
		}
		if rd.fromFirst && len(in.Values) > 0 {
			return "", fmt.Sprintf("arrays.Reduce(arrays.Of(%s...), %s, %d)", goInts(in.Values[1:]), rd.src, in.Values[0])
		}
		return "", fmt.Sprintf("arrays.Reduce(arrays.Of(%s...), %s, %d)", goInts(in.Values), rd.src, rd.init)
	}

	jsFlatMap := visitEntry("js-flat-map", "Array.prototype.flatMap", "FlatMap", "Map each element to an array, then join them one level deep.", "duplicate", "duplicate, range",
		func(in Input) (Output, error) {
			k := strings.ToLower(strings.TrimSpace(in.Key))
			if k == "" {
				k = "duplicate"
			}
			if k != "duplicate" && k != "range" {
				return Output{}, inputError(fmt.Sprintf("unknown callback %q (use duplicate, range)", in.Key))
			}
			out := Output{Initial: slices.Clone(in.Values)}
			var res []int
			for i, v := range in.Values {
				part := []int{v, v}
				if k == "range" {
					part = nil
					for j := 1; j <= min(max(v, 0), 20); j++ {
						part = append(part, j)
					}
				}
				res = append(res, part...)
				out.Steps = append(out.Steps, step(in.Values, i, fmt.Sprintf("callback(%d) = [%s]", v, strings.ReplaceAll(joinInts(part), " ", ", ")), nil))
			}
			if res == nil {
				res = []int{}
			}
			out.Values = res
			return out, nil
		}, nil)
	jsFlatMap.Example = &Input{Values: []int{1, 2, 3}, Key: "duplicate"}
	jsFlatMap.Call = func(in Input) (string, string) {
		f := "func(v, _ int) []int { return []int{v, v} }"
		if strings.EqualFold(strings.TrimSpace(in.Key), "range") {
			f = "func(v, _ int) []int {\n\t\tvar r []int\n\t\tfor j := 1; j <= min(max(v, 0), 20); j++ {\n\t\t\tr = append(r, j)\n\t\t}\n\t\treturn r\n\t}"
		}
		return "", fmt.Sprintf("arrays.FlatMap(arrays.Of(%s...), %s).Values()", goInts(in.Values), f)
	}

	jsSort := entry("js-sort", "Array.prototype.sort", "structures/arrays/js.go", "Array.Sort")
	jsSort.Summary, jsSort.Time, jsSort.Space, jsSort.Input, jsSort.Params = "JavaScript sorts by string form unless given a compare function: [1, 10, 2].sort() stays [1, 10, 2]. Keyword \"numeric\" sorts as numbers; \"default\" (or empty) as JS does.", "O(n log n)", "O(n)", Ints, []string{"key"}
	jsSort.Example, jsSort.Traced = &Input{Values: []int{1, 10, 2, 21, 100, 3}}, true
	jsSort.Run = func(in Input, r trace.Recorder) (Output, error) {
		if err := values(in, "sort", 10_000); err != nil {
			return Output{}, err
		}
		a := arrays.From(in.Values)
		a.SetRecorder(r)
		mode := "default (by string form)"
		if strings.EqualFold(strings.TrimSpace(in.Key), "numeric") {
			a.Sort(func(x, y int) bool { return x < y })
			mode = "numeric (a - b)"
		} else {
			a.SortDefault()
		}
		return Output{Initial: slices.Clone(in.Values), Values: a.Values(), Lines: []string{"compare: " + mode}}, nil
	}
	jsSort.Call = func(in Input) (string, string) {
		if strings.EqualFold(strings.TrimSpace(in.Key), "numeric") {
			return "", fmt.Sprintf("arrays.Of(%s...).Sort(func(x, y int) bool { return x < y }).Values()", goInts(in.Values))
		}
		return "", fmt.Sprintf("arrays.Of(%s...).SortDefault().Values()", goInts(in.Values))
	}

	jsSplice := entry("js-splice", "Array.prototype.splice", "structures/arrays/js.go", "Array.Splice")
	jsSplice.Summary, jsSplice.Time, jsSplice.Space, jsSplice.Input = "Remove count elements from start (negative counts from the end) and insert the new items there, in place; returns what was removed. The array goes in the first field and the items to insert in the second.", "O(n)", "O(n)", TwoTexts
	jsSplice.Options = []Option{{"start", "start (negative from the end)", 1, -1000, 1000}, {"count", "delete count", 2, 0, 1000}}
	jsSplice.Traced = true
	jsSplice.Example = &Input{Text: "10 20 30 40 50", Other: "7 8 9", Options: map[string]int{"start": 1, "count": 2}}
	spliceItems := func(in Input) ([]int, error) { return fieldInts(in.Other, "items to insert") }
	spliceArray := func(in Input) ([]int, error) { return fieldInts(in.Text, "array") }
	jsSplice.Run = func(in Input, r trace.Recorder) (Output, error) {
		arr, err := spliceArray(in)
		if err != nil {
			return Output{}, err
		}
		if len(arr) == 0 || len(arr) > 10_000 {
			return Output{}, inputError("give 1 to 10000 values to splice")
		}
		items, err := spliceItems(in)
		if err != nil {
			return Output{}, err
		}
		a := arrays.From(arr)
		a.SetRecorder(r)
		removed := a.Splice(in.opt("start", 1), in.opt("count", 2), items...)
		return Output{Initial: arr, Values: a.Values(), Lines: []string{"removed: [" + strings.ReplaceAll(joinInts(removed), " ", ", ") + "]"}}, nil
	}
	jsSplice.Call = func(in Input) (string, string) {
		arr, _ := spliceArray(in)
		items, _ := spliceItems(in)
		args := []string{strconv.Itoa(in.opt("start", 1)), strconv.Itoa(in.opt("count", 2))}
		for _, v := range items {
			args = append(args, strconv.Itoa(v))
		}
		return fmt.Sprintf("a := arrays.Of(%s...)\n\tremoved := a.Splice(%s)\n\t", goInts(arr), strings.Join(args, ", ")), "removed, a.Values()"
	}

	jsSlice := entry("js-slice", "Array.prototype.slice", "structures/arrays/js.go", "Array.Slice")
	jsSlice.Summary, jsSlice.Time, jsSlice.Space, jsSlice.Input = "A copy of the elements from start up to (not including) end; negative indexes count from the end, and both are clamped. The array is not changed.", "O(end - start)", "O(end - start)", Ints
	jsSlice.Options = []Option{{"start", "start (negative from the end)", 1, -1000, 1000}, {"end", "end (negative from the end)", -1, -1000, 1000}}
	jsSlice.Example = &Input{Values: []int{10, 20, 30, 40, 50}, Options: map[string]int{"start": 1, "end": -1}}
	jsSlice.Run = func(in Input, _ trace.Recorder) (Output, error) {
		if err := values(in, "slice", 10_000); err != nil {
			return Output{}, err
		}
		a := arrays.From(in.Values)
		got := a.Slice(in.opt("start", 1), in.opt("end", -1)).Values()
		if got == nil {
			got = []int{}
		}
		return Output{Initial: slices.Clone(in.Values), Values: got}, nil
	}
	jsSlice.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("arrays.Of(%s...).Slice(%d, %d).Values()", goInts(in.Values), in.opt("start", 1), in.opt("end", -1))
	}

	play := structureEntry("js-array", "JavaScript Array", "structures/arrays/js.go", "Array",
		"Every JavaScript array method on a live array, with JS semantics: negative indexes, clamped ranges, mutating and copying methods.",
		"O(1) push, pop, at; O(n) shift, unshift, splice", "O(n)", "push",
		[]string{"push V...", "pop", "shift", "unshift V...", "splice START COUNT V...", "slice START END", "concat V...", "at I", "with I V", "fill V START END", "copy-within T S E", "reverse", "sort", "sort-default", "index-of V", "last-index-of V", "includes V", "join SEP"},
		func(Input) playground { return &jsArrayP{a: arrays.Of[int]()} })
	play.Example = &Input{Values: []int{1, 2, 3}, Ops: ParseOps("push 4 5, shift, unshift 0, splice 1 2 9 9, at -1, slice 1 -1, fill 7 2 4, copy-within 0 3 5, sort-default, includes 7, join -")}

	return []Entry{play, jsMap, jsFilter, jsReduce, jsFind, jsFindIndex, jsSome, jsEvery, jsFlatMap, jsSort, jsSplice, jsSlice}
}

// jsArrayP is the js-array playground: the array, and the roles to draw
// after the last operation.
type jsArrayP struct {
	a     *arrays.Array[int]
	roles map[int]string
}

func (p *jsArrayP) Snapshot() trace.Snapshot { return arrayRow(p.a.Values(), p.roles) }

func spanRoles(from, to int, role string) map[int]string {
	m := map[int]string{}
	for i := from; i < to; i++ {
		m[i] = role
	}
	return m
}

func (p *jsArrayP) apply(op Op, r trace.Recorder) (string, error) {
	p.a.SetRecorder(r)
	defer p.a.SetRecorder(nil)
	p.roles = nil
	n := p.a.Len()
	args, err := fieldInts(op.Value, op.Name)
	need := func(k int) error {
		if err != nil {
			return err
		}
		if len(args) < k {
			return inputError(fmt.Sprintf("%s needs %d number(s)", op.Name, k))
		}
		return nil
	}
	list := func(s []int) string { return "[" + strings.ReplaceAll(joinInts(s), " ", ", ") + "]" }
	switch op.Name {
	case "push":
		if err := need(1); err != nil {
			return "", err
		}
		l := p.a.Push(args...)
		p.roles = spanRoles(n, l, "new")
		return fmt.Sprintf("push returned the new length %d", l), nil
	case "pop", "shift":
		pop := p.a.Pop
		if op.Name == "shift" {
			pop = p.a.Shift
		}
		v, ok := pop()
		if !ok {
			return op.Name + " on an empty array returns undefined", nil
		}
		return fmt.Sprintf("%s returned %d", op.Name, v), nil
	case "unshift":
		if err := need(1); err != nil {
			return "", err
		}
		l := p.a.Unshift(args...)
		p.roles = spanRoles(0, len(args), "new")
		return fmt.Sprintf("unshift returned the new length %d (every element moved)", l), nil
	case "splice":
		if err := need(2); err != nil {
			return "", err
		}
		start := args[0]
		removed := p.a.Splice(start, args[1], args[2:]...)
		s := start
		if s < 0 {
			s = max(n+s, 0)
		}
		s = min(s, n)
		p.roles = spanRoles(s, s+len(args)-2, "new")
		return fmt.Sprintf("splice removed %s", list(removed)), nil
	case "slice":
		if err := need(2); err != nil {
			return "", err
		}
		got := p.a.Slice(args[0], args[1]).Values()
		s, e := args[0], args[1]
		if s < 0 {
			s = max(n+s, 0)
		}
		if e < 0 {
			e = max(n+e, 0)
		}
		p.roles = spanRoles(min(s, n), min(e, n), "range")
		return fmt.Sprintf("slice returned the copy %s (the array is unchanged)", list(got)), nil
	case "concat":
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("concat returned the new array %s (this one is unchanged)", list(p.a.Concat(args).Values())), nil
	case "at":
		if err := need(1); err != nil {
			return "", err
		}
		v, ok := p.a.At(args[0])
		if !ok {
			return fmt.Sprintf("at(%d) returns undefined", args[0]), nil
		}
		i := args[0]
		if i < 0 {
			i += n
		}
		p.roles = map[int]string{i: "range"}
		return fmt.Sprintf("at(%d) is %d", args[0], v), nil
	case "with":
		if err := need(2); err != nil {
			return "", err
		}
		w, ok := p.a.With(args[0], args[1])
		if !ok {
			return fmt.Sprintf("with(%d, ...) throws a RangeError", args[0]), nil
		}
		return fmt.Sprintf("with returned the copy %s (this one is unchanged)", list(w.Values())), nil
	case "fill":
		if err := need(1); err != nil {
			return "", err
		}
		s, e := 0, n
		if len(args) > 1 {
			s = args[1]
		}
		if len(args) > 2 {
			e = args[2]
		}
		p.a.Fill(args[0], s, e)
		if s < 0 {
			s = max(n+s, 0)
		}
		if e < 0 {
			e = max(n+e, 0)
		}
		p.roles = spanRoles(min(s, n), min(e, n), "new")
		return fmt.Sprintf("filled with %d", args[0]), nil
	case "copy-within":
		if err := need(2); err != nil {
			return "", err
		}
		e := n
		if len(args) > 2 {
			e = args[2]
		}
		p.a.CopyWithin(args[0], args[1], e)
		return "copied within, the length unchanged", nil
	case "reverse":
		p.a.Reverse()
		return "reversed in place", nil
	case "sort":
		p.a.Sort(func(x, y int) bool { return x < y })
		return "sorted numerically, with sort((a, b) => a - b)", nil
	case "sort-default":
		p.a.SortDefault()
		return "sorted the JavaScript default way: by string form", nil
	case "index-of", "last-index-of", "includes":
		if err := need(1); err != nil {
			return "", err
		}
		switch op.Name {
		case "index-of":
			return fmt.Sprintf("indexOf(%d) = %d", args[0], arrays.IndexOf(p.a, args[0])), nil
		case "last-index-of":
			return fmt.Sprintf("lastIndexOf(%d) = %d", args[0], arrays.LastIndexOf(p.a, args[0])), nil
		}
		return fmt.Sprintf("includes(%d) = %v", args[0], arrays.Includes(p.a, args[0])), nil
	case "join":
		sep := op.Value
		if sep == "" {
			sep = ","
		}
		return fmt.Sprintf("join(%q) = %q", sep, p.a.Join(sep)), nil
	}
	return "", unknownOp(op, "push", "pop", "shift", "unshift", "splice", "slice", "concat", "at", "with", "fill", "copy-within", "reverse", "sort", "sort-default", "index-of", "last-index-of", "includes", "join")
}
