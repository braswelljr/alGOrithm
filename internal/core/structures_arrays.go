package core

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/structures/arrays"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// arraysEntries is every entry of the arrays category.
func arraysEntries() []Entry {
	return concat(
		[]Entry{
			dynamicArrayEntry(),
			maxSubarrayEntry(),
			prefixSumsEntry(),
			reverseArrayEntry(),
			rotateArrayEntry(),
			slidingWindowMaxEntry(),
			twoPointersEntry(),
		},
		jsArrayEntries(),
	)
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

var jsArrayEx = []int{3, -1, 4, 1, -5, 9, 2, 6}

func jsArrayVisitEntry(id, name, fn, summary, key, keys string, run func(in Input) (Output, error), call func(in Input) string) Entry {
	e := entry(id, name, "structures/arrays/js.go", fn)
	e.Summary, e.Time, e.Space, e.Input, e.Params = summary+" The callback, by name ("+keys+"), goes in the keyword.", "O(n)", "O(n)", Ints, []string{"key"}
	e.Example = &Input{Values: jsArrayEx, Key: key, Target: 4}
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

func jsArrayPick(m map[string]callback, in Input, def string) (callback, error) {
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

func jsArraySrcOf(c callback, in Input) string {
	if strings.Contains(c.src, "%d") {
		return fmt.Sprintf(c.src, in.Target)
	}
	return c.src
}

func jsArrayStep(vals []int, i int, result string, roles map[int]string) Step {
	r := map[int]string{i: "current"}
	for k, v := range roles {
		if k != i {
			r[k] = v
		}
	}
	return Step{Op: Op{Name: "visit", Value: fmt.Sprintf("index %d: %d", i, vals[i])}, Result: result, Snapshot: arrayRow(vals, r)}
}

func jsArrayPredicateRun(mode string) func(in Input) (Output, error) {
	return func(in Input) (Output, error) {
		c, err := jsArrayPick(predicates, in, "even")
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
			out.Steps = append(out.Steps, jsArrayStep(in.Values, i, fmt.Sprintf("callback(%d) = %v", v, ok), kept))
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

var jsArrayJsMap = jsArrayVisitEntry("js-map", "Array.prototype.map", "Map", "A new array of the callback applied to each element.", "double", names(mappers),
	func(in Input) (Output, error) {
		c, err := jsArrayPick(mappers, in, "double")
		if err != nil {
			return Output{}, err
		}
		out := Output{Initial: slices.Clone(in.Values)}
		var res []int
		for i, v := range in.Values {
			res = append(res, c.f(v))
			out.Steps = append(out.Steps, jsArrayStep(in.Values, i, fmt.Sprintf("callback(%d) = %d", v, c.f(v)), nil))
		}
		out.Values = res
		return out, nil
	},
	func(in Input) string {
		c, _ := jsArrayPick(mappers, in, "double")
		return fmt.Sprintf("arrays.Map(arrays.Of(%s...), %s).Values()", goInts(in.Values), c.src)
	})

func jsArrayPredCall(method string) func(in Input) string {
	return func(in Input) string {
		c, _ := jsArrayPick(predicates, in, "even")
		return fmt.Sprintf("arrays.Of(%s...).%s(%s)", goInts(in.Values), method, jsArraySrcOf(c, in))
	}
}

var jsArrayJsFilter = jsArrayVisitEntry("js-filter", "Array.prototype.filter", "Array.Filter", "A new array of the elements the callback keeps.", "even", names(predicates), jsArrayPredicateRun("filter"),
	func(in Input) string { return jsArrayPredCall("Filter")(in) + ".Values()" })

var jsArrayJsFindIndex = jsArrayVisitEntry("js-find-index", "Array.prototype.findIndex", "Array.FindIndex", "The index of the first element the callback accepts, or -1.", "even", names(predicates), jsArrayPredicateRun("find"), jsArrayPredCall("FindIndex"))

func jsArraySpliceItems(in Input) ([]int, error) { return fieldInts(in.Other, "items to insert") }

func jsArraySpliceArray(in Input) ([]int, error) { return fieldInts(in.Text, "array") }

func jsFindEntry() Entry {
	jsFind := jsArrayVisitEntry("js-find", "Array.prototype.find", "Array.Find", "The first element the callback accepts, stopping there (undefined if none; here ok is false).", "greater", names(predicates), jsArrayPredicateRun("find"), jsArrayPredCall("Find"))
	jsFind.Params = append(jsFind.Params, "target")
	return jsFind
}

func jsSomeEntry() Entry {
	jsSome := jsArrayVisitEntry("js-some", "Array.prototype.some", "Array.Some", "Whether the callback accepts at least one element; it stops at the first (false for []).", "negative", names(predicates), jsArrayPredicateRun("some"), jsArrayPredCall("Some"))
	jsSome.Params = append(jsSome.Params, "target")
	return jsSome
}

func jsEveryEntry() Entry {
	jsEvery := jsArrayVisitEntry("js-every", "Array.prototype.every", "Array.Every", "Whether the callback accepts every element; it stops at the first refusal (true for []).", "positive", names(predicates), jsArrayPredicateRun("every"), jsArrayPredCall("Every"))
	jsEvery.Params = append(jsEvery.Params, "target")
	return jsEvery
}

func jsReduceEntry() Entry {
	jsReduce := jsArrayVisitEntry("js-reduce", "Array.prototype.reduce", "Reduce", "Fold the elements from the left into one value: acc = callback(acc, element).", "sum", names(reducers),
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
				out.Steps = append(out.Steps, jsArrayStep(in.Values, 0, fmt.Sprintf("no initial value: start from %d", acc), nil))
			}
			for i := from; i < len(in.Values); i++ {
				v := in.Values[i]
				next := rd.f(acc, v)
				out.Steps = append(out.Steps, jsArrayStep(in.Values, i, fmt.Sprintf("callback(%d, %d) = %d", acc, v, next), nil))
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
	return jsReduce
}

func jsFlatMapEntry() Entry {
	jsFlatMap := jsArrayVisitEntry("js-flat-map", "Array.prototype.flatMap", "FlatMap", "Map each element to an array, then join them one level deep.", "duplicate", "duplicate, range",
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
				out.Steps = append(out.Steps, jsArrayStep(in.Values, i, fmt.Sprintf("callback(%d) = [%s]", v, strings.ReplaceAll(joinInts(part), " ", ", ")), nil))
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
	return jsFlatMap
}

func jsSortEntry() Entry {
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
	return jsSort
}

func jsSpliceEntry() Entry {
	jsSplice := entry("js-splice", "Array.prototype.splice", "structures/arrays/js.go", "Array.Splice")
	jsSplice.Summary, jsSplice.Time, jsSplice.Space, jsSplice.Input = "Remove count elements from start (negative counts from the end) and insert the new items there, in place; returns what was removed. The array goes in the first field and the items to insert in the second.", "O(n)", "O(n)", TwoTexts
	jsSplice.Options = []Option{{"start", "start (negative from the end)", 1, -1000, 1000}, {"count", "delete count", 2, 0, 1000}}
	jsSplice.Traced = true
	jsSplice.Example = &Input{Text: "10 20 30 40 50", Other: "7 8 9", Options: map[string]int{"start": 1, "count": 2}}
	jsSplice.Run = func(in Input, r trace.Recorder) (Output, error) {
		arr, err := jsArraySpliceArray(in)
		if err != nil {
			return Output{}, err
		}
		if len(arr) == 0 || len(arr) > 10_000 {
			return Output{}, inputError("give 1 to 10000 values to splice")
		}
		items, err := jsArraySpliceItems(in)
		if err != nil {
			return Output{}, err
		}
		a := arrays.From(arr)
		a.SetRecorder(r)
		removed := a.Splice(in.opt("start", 1), in.opt("count", 2), items...)
		return Output{Initial: arr, Values: a.Values(), Lines: []string{"removed: [" + strings.ReplaceAll(joinInts(removed), " ", ", ") + "]"}}, nil
	}
	jsSplice.Call = func(in Input) (string, string) {
		arr, _ := jsArraySpliceArray(in)
		items, _ := jsArraySpliceItems(in)
		args := []string{strconv.Itoa(in.opt("start", 1)), strconv.Itoa(in.opt("count", 2))}
		for _, v := range items {
			args = append(args, strconv.Itoa(v))
		}
		return fmt.Sprintf("a := arrays.Of(%s...)\n\tremoved := a.Splice(%s)\n\t", goInts(arr), strings.Join(args, ", ")), "removed, a.Values()"
	}
	return jsSplice
}

func jsSliceEntry() Entry {
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
	return jsSlice
}

func jsArrayEntry() Entry {
	play := structureEntry("js-array", "JavaScript Array", "structures/arrays/js.go", "Array",
		"Every JavaScript array method on a live array, with JS semantics: negative indexes, clamped ranges, mutating and copying methods.",
		"O(1) push, pop, at; O(n) shift, unshift, splice", "O(n)", "push",
		[]string{"push V...", "pop", "shift", "unshift V...", "splice START COUNT V...", "slice START END", "concat V...", "at I", "with I V", "fill V START END", "copy-within T S E", "reverse", "sort", "sort-default", "index-of V", "last-index-of V", "includes V", "join SEP"},
		func(Input) playground { return &jsArrayP{a: arrays.Of[int]()} })
	play.Example = &Input{Values: []int{1, 2, 3}, Ops: ParseOps("push 4 5, shift, unshift 0, splice 1 2 9 9, at -1, slice 1 -1, fill 7 2 4, copy-within 0 3 5, sort-default, includes 7, join -")}
	return play
}

func jsArrayEntries() []Entry {
	return []Entry{
		jsArrayEntry(),
		jsArrayJsMap,
		jsArrayJsFilter,
		jsReduceEntry(),
		jsFindEntry(),
		jsArrayJsFindIndex,
		jsSomeEntry(),
		jsEveryEntry(),
		jsFlatMapEntry(),
		jsSortEntry(),
		jsSpliceEntry(),
		jsSliceEntry(),
	}
}

func arraysMk(id, name, fn, summary string, run func(arraysIn Input, r trace.Recorder) (Output, error)) Entry {
	file := "structures/arrays/techniques.go"
	e := entry(id, name, file, fn)
	e.Summary, e.Time, e.Space, e.Input, e.Traced, e.Run = summary, "O(n)", "O(1)", Ints, true, run
	e.Example = &Input{Values: []int{-2, 1, -3, 4, -1, 2, 1, -5, 4}, Target: 3}
	return e
}

func arraysIn(arraysIn Input) []int { return append([]int(nil), arraysIn.Values...) }

func dynamicArrayEntry() Entry {
	dyn := structureEntry("dynamic-array", "Dynamic Array", "structures/arrays/dynamic.go", "Dynamic",
		"A growable array: append into spare capacity, double and copy when full, shift to insert or delete.",
		"O(1) amortized append and pop, O(1) get and set, O(n) insert and delete", "O(n)",
		"append", []string{"append VALUE", "pop", "insert INDEX VALUE", "delete INDEX", "get INDEX", "set INDEX VALUE"},
		func(Input) playground { return &dynamicP{d: arrays.NewDynamic(1, nil)} })
	dyn.Example = &Input{Values: []int{4, 8}, Ops: ParseOps("append 15, append 16, append 23, insert 0 42, delete 2, pop")}
	return dyn
}

func rotateArrayEntry() Entry {
	rotate := arraysMk("rotate-array", "Rotate (three reversals)", "Rotate", "Rotate right by k with three in-place reversals: all, then each part.",
		func(i Input, r trace.Recorder) (Output, error) {
			if err := values(i, "rotate", 10_000); err != nil {
				return Output{}, err
			}
			return Output{Initial: arraysIn(i), Values: arrays.TraceRotate(arraysIn(i), i.Target, r)}, nil
		})
	rotate.Params = []string{"target"}
	return rotate
}

func twoPointersEntry() Entry {
	twoSum := arraysMk("two-pointers", "Two Pointers (pair sum)", "TwoSum", "Find two values of a sorted array that sum to a target, moving a pointer in from each end.",
		func(i Input, r trace.Recorder) (Output, error) {
			if err := values(i, "search", 10_000); err != nil {
				return Output{}, err
			}
			s := arraysIn(i)
			slices.Sort(s)
			a, b, ok := arrays.TraceTwoSum(s, i.Target, r)
			out := Output{Initial: s, Bool: &ok}
			if ok {
				out.Lines = []string{fmt.Sprintf("%d + %d = %d (indexes %d and %d)", s[a], s[b], i.Target, a, b)}
			}
			return out, nil
		})
	twoSum.Params, twoSum.Input, twoSum.Example = []string{"target"}, SortedInts, &Input{Values: []int{1, 3, 4, 6, 9, 11, 15}, Target: 13}
	return twoSum
}

func maxSubarrayEntry() Entry {
	kadane := arraysMk("max-subarray", "Maximum Subarray (Kadane)", "MaxSubarray", "The largest sum of consecutive values: extend the run, or start over when that is better.",
		func(i Input, r trace.Recorder) (Output, error) {
			if err := values(i, "scan", 100_000); err != nil {
				return Output{}, err
			}
			sum, s, e := arrays.TraceMaxSubarray(i.Values, r)
			return Output{Initial: arraysIn(i), Number: &sum, Values: i.Values[s:e]}, nil
		})
	return kadane
}

func slidingWindowMaxEntry() Entry {
	window := arraysMk("sliding-window-max", "Sliding Window Maximum", "WindowMax", "The maximum of every window of k values, with a deque of indexes so each value enters and leaves once.",
		func(i Input, r trace.Recorder) (Output, error) {
			if err := values(i, "scan", 100_000); err != nil {
				return Output{}, err
			}
			k := i.opt("window", 3)
			out := arrays.TraceWindowMax(i.Values, k, r)
			if out == nil {
				return Output{}, inputError(fmt.Sprintf("the window must be 1 to %d", len(i.Values)))
			}
			return Output{Initial: arraysIn(i), Values: out}, nil
		})
	window.Options = []Option{{"window", "window size", 3, 1, 1000}}
	return window
}

func prefixSumsEntry() Entry {
	prefix := arraysMk("prefix-sums", "Prefix Sums", "PrefixSums", "p[i] is the sum of the first i values, so any range sum is one subtraction.",
		func(i Input, r trace.Recorder) (Output, error) {
			if err := values(i, "sum", 100_000); err != nil {
				return Output{}, err
			}
			p := arrays.TracePrefixSums(i.Values, r)
			return Output{Initial: make([]int, len(p)), Values: p}, nil
		})
	return prefix
}

func reverseArrayEntry() Entry {
	reverse := arraysMk("reverse-array", "Reverse (two pointers)", "Reverse", "Swap the ends and walk two pointers inward.",
		func(i Input, r trace.Recorder) (Output, error) {
			if err := values(i, "reverse", 100_000); err != nil {
				return Output{}, err
			}
			return Output{Initial: arraysIn(i), Values: arrays.TraceReverse(arraysIn(i), r)}, nil
		})
	return reverse
}
