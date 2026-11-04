package core

import (
	"fmt"
	"slices"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/structures/sets"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// setsEntries is every entry of the sets category.
func setsEntries() []Entry {
	return concat(
		setEntries(),
		structureSpecEntries("sets"),
	)
}

// asSet sorts s and drops repeats: the sorted-set form the set operations take.
func asSet[T int | string](s []T) []T {
	s = slices.Clone(s)
	slices.Sort(s)
	return slices.Compact(s)
}

func setsEntrySetOp(id, name, fn, summary string, traced func(a, b []int, r trace.Recorder) []int) Entry {
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

func setsEntryInSets(e Entry, example string) Entry {
	e.Category = "sets"
	e.Example = &Input{Ops: ParseOps(example)}
	return e
}

func setEntries() []Entry {
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
	return []Entry{
		setsEntryInSets(structureEntry("bitset", "Bitset", "structures/sets/bitset.go", "Bitset", "A set of small whole numbers, one bit each: membership in O(1), set operations 64 members at a time.",
			"O(1) set, clear, test; O(n/64) count", "O(max / 64)", "set", []string{"set N", "clear N", "test N", "count", "members"},
			func(Input) playground { return &bitsetP{sets.NewBitset()} }), "set 1, set 3, set 5, set 12, clear 3, test 5, count"),
		setsEntryInSets(structureEntry("multiset", "Multiset (Bag)", "structures/sets/multiset.go", "Multiset", "A set where a value can occur more than once: each value with a count.",
			"O(1) average add, remove, count", "O(distinct)", "add", []string{"add VALUE", "remove VALUE", "count VALUE"},
			func(Input) playground { return &multisetP{sets.NewMultiset[int]()} }), "add 3, add 1, add 3, add 3, remove 3, count 3"),
		setsEntryInSets(structureEntry("ordered-set", "Ordered Set", "structures/sets/orderedset.go", "OrderedSet", "A sorted set: binary search for membership, rank, floor, ceiling and ranges.",
			"O(log n) search, O(n) insert", "O(n)", "add", []string{"add VALUE", "remove VALUE", "contains VALUE", "rank VALUE", "floor VALUE", "ceiling VALUE", "range LO HI"},
			func(Input) playground { return &orderedP{sets.NewOrderedSet[int]()} }), "add 20, add 5, add 12, add 30, add 5, floor 13, ceiling 13, rank 12, range 6 25"),
		setsEntrySetOp("sorted-union", "Union (two pointers)", "Union", "The values in either set, walking both sorted sets at once like a merge.", sets.TraceUnion),
		setsEntrySetOp("sorted-intersection", "Intersection (two pointers)", "Intersection", "The values in both sets, walking both sorted sets at once.", sets.TraceIntersection),
		setsEntrySetOp("sorted-difference", "Difference (two pointers)", "Difference", "The values in the first set but not the second.", sets.TraceDifference),
		setsEntrySetOp("sorted-symmetric-difference", "Symmetric Difference (two pointers)", "SymmetricDifference", "The values in exactly one of the two sets.", sets.TraceSymmetricDifference),
		power, product, jac,
	}
}
