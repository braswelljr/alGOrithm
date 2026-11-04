package collections_test

import (
	"math"
	"math/rand"
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/braswelljr/alGOrithm/pkg/structures/collections"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

var dec = collections.MustDecode

func TestDeepMerge(t *testing.T) {
	// lodash's documented example.
	object := dec(`{"a": [{"b": 2}, {"d": 4}]}`)
	other := dec(`{"a": [{"c": 3}, {"e": 5}]}`)
	before, beforeOther := collections.Encode(object), collections.Encode(other)
	got := collections.DeepMerge(object, other)
	if want := dec(`{"a": [{"b": 2, "c": 3}, {"d": 4, "e": 5}]}`); !collections.IsEqual(got, want) {
		t.Errorf("merge = %s", collections.Encode(got))
	}
	if collections.Encode(object) != before || collections.Encode(other) != beforeOther {
		t.Error("DeepMerge changed an input")
	}
	cases := []struct{ dst, src, want string }{
		{`{"a": {"b": 1, "c": [1, 2]}}`, `{"a": {"c": [9], "d": true}}`, `{"a": {"b": 1, "c": [9, 2], "d": true}}`},
		{`{"a": 1}`, `{"a": null}`, `{"a": 1}`},
		{`{}`, `{"a": null}`, `{"a": null}`},
		{`{"a": {"b": 1}}`, `{"a": 5}`, `{"a": 5}`},
		{`{"a": 5}`, `{"a": {"b": 1}}`, `{"a": {"b": 1}}`},
	}
	for _, c := range cases {
		if got := collections.DeepMerge(dec(c.dst), dec(c.src)); !collections.IsEqual(got, dec(c.want)) {
			t.Errorf("merge(%s, %s) = %s, want %s", c.dst, c.src, collections.Encode(got), c.want)
		}
	}
	three := collections.DeepMerge(dec(`{"a": 1}`), dec(`{"b": 2}`), dec(`{"a": 3}`))
	if !collections.IsEqual(three, dec(`{"a": 3, "b": 2}`)) {
		t.Errorf("three sources = %s", collections.Encode(three))
	}
	concat := collections.MergeWith(collections.MergeOptions{ConcatArrays: true}, nil, dec(`{"a": [1]}`), dec(`{"a": [2, 3]}`))
	if !collections.IsEqual(concat, dec(`{"a": [1, 2, 3]}`)) {
		t.Errorf("concat = %s", collections.Encode(concat))
	}
	var steps []collections.Step
	collections.TraceDeepMerge(func(s collections.Step) { steps = append(steps, s) }, dec(`{"a": 1, "b": 2}`), dec(`{"a": 9, "c": 3, "b": null}`))
	actions := map[string]string{}
	for _, s := range steps {
		actions[s.Path] = s.Action
	}
	if actions["a"] != "overwritten" || actions["b"] != "kept" || actions["c"] != "new" {
		t.Errorf("trace actions = %v", actions)
	}
}

func TestIsEqualAndClone(t *testing.T) {
	if !collections.IsEqual(math.NaN(), math.NaN()) || collections.IsEqual(nil, map[string]any{}) || collections.IsEqual(dec(`[1,2]`), dec(`[1,2,3]`)) {
		t.Error("tricky IsEqual cases wrong")
	}
	if p, d := collections.FirstDifference(dec(`{"a": {"b": [1, 2]}}`), dec(`{"a": {"b": [1, 3]}}`)); !d || p != "a.b[1]" {
		t.Errorf("first difference %q, %v", p, d)
	}
	orig := dec(`{"a": {"b": [1, {"c": 2}]}}`)
	c := collections.DeepClone(orig)
	c.(map[string]any)["a"].(map[string]any)["b"].([]any)[1].(map[string]any)["c"] = 99
	if collections.Get(orig, "a.b[1].c", nil) != 2.0 {
		t.Error("DeepClone shares a nested map")
	}
}

func TestPaths(t *testing.T) {
	object := dec(`{"a": [{"b": {"c": 3}}]}`)
	if collections.Get(object, "a[0].b.c", nil) != 3.0 || collections.Get(object, "a.b.c", "default") != "default" {
		t.Error("Get wrong (lodash example)")
	}
	if !collections.Has(object, "a[0].b") || collections.Has(object, "a[1]") {
		t.Error("Has wrong")
	}
	set, err := collections.Set(object, "x[0].y.z", 5.0)
	if err != nil || collections.Get(set, "x[0].y.z", nil) != 5.0 || collections.Has(object, "x") {
		t.Errorf("Set = %s, %v (must not change the input)", collections.Encode(set), err)
	}
	if !collections.IsEqual(collections.Get(set, "x", nil), dec(`[{"y": {"z": 5}}]`)) {
		t.Error("Set did not create an array for an index step")
	}
	un, ok := collections.Unset(object, "a[0].b.c")
	if !ok || collections.Has(un, "a[0].b.c") || !collections.Has(object, "a[0].b.c") {
		t.Error("Unset wrong")
	}
	if v := collections.Get(dec(`{"weird.key": {"x": 1}}`), `['weird.key'].x`, nil); v != 1.0 {
		t.Errorf("quoted key = %v", v)
	}
	for _, bad := range []string{"a..b", "a[", "a[x]", "a.", `a['b`} {
		if _, err := collections.ParsePath(bad); err == nil {
			t.Errorf("ParsePath(%q) accepted", bad)
		}
	}
	flat := collections.FlattenObject(dec(`{"a": {"b": [1, {"c": 2}], "e": {}}}`))
	if !reflect.DeepEqual(collections.Keys(flat), []string{"a.b[0]", "a.b[1].c", "a.e"}) {
		t.Errorf("flatten keys %v", collections.Keys(flat))
	}
	back, err := collections.UnflattenObject(flat)
	if err != nil || !collections.IsEqual(back, dec(`{"a": {"b": [1, {"c": 2}], "e": {}}}`)) {
		t.Errorf("unflatten = %s, %v", collections.Encode(back), err)
	}
}

func TestObjects(t *testing.T) {
	object := dec(`{"a": 1, "b": "2", "c": 3}`).(map[string]any)
	if got := collections.Pick(object, "a", "c"); !collections.IsEqual(any(got), dec(`{"a": 1, "c": 3}`)) {
		t.Errorf("Pick = %v", got)
	}
	if got := collections.Omit(object, "a", "c"); !collections.IsEqual(any(got), dec(`{"b": "2"}`)) {
		t.Errorf("Omit = %v", got)
	}
	a := collections.Assign(map[string]any{"a": 0.0}, map[string]any{"b": 2.0}, map[string]any{"a": 1.0})
	if !collections.IsEqual(any(a), dec(`{"a": 1, "b": 2}`)) {
		t.Errorf("Assign = %v", a)
	}
	d := collections.Defaults(map[string]any{"a": 1.0}, map[string]any{"b": 2.0}, map[string]any{"a": 3.0})
	if !collections.IsEqual(any(d), dec(`{"a": 1, "b": 2}`)) {
		t.Errorf("Defaults = %v", d)
	}
	dd := collections.DefaultsDeep(dec(`{"a": {"b": 2}}`).(map[string]any), dec(`{"a": {"b": 1, "c": 3}}`).(map[string]any))
	if !collections.IsEqual(any(dd), dec(`{"a": {"b": 2, "c": 3}}`)) {
		t.Errorf("DefaultsDeep = %v", dd)
	}
	if got := collections.Invert(map[string]int{"a": 1, "b": 2}); got[2] != "b" {
		t.Errorf("Invert = %v", got)
	}
	pairs := collections.Entries(map[string]int{"b": 2, "a": 1})
	if pairs[0].Key != "a" || collections.FromEntries(pairs)["b"] != 2 {
		t.Error("Entries/FromEntries wrong")
	}
	if k, ok := collections.FindKey(map[string]int{"barney": 36, "fred": 40}, func(v int) bool { return v < 40 }); !ok || k != "barney" {
		t.Errorf("FindKey = %q", k)
	}
	up := collections.MapKeys(map[string]int{"a": 1}, func(v int, k string) string { return k + "x" })
	if up["ax"] != 1 || collections.MapValues(map[string]int{"a": 1}, func(v int, _ string) int { return v * 2 })["a"] != 2 {
		t.Error("MapKeys/MapValues wrong")
	}
	if !collections.IsMap(object) || collections.IsMap([]any{}) {
		t.Error("IsMap wrong")
	}
}

func TestSlices(t *testing.T) {
	if got := collections.Chunk([]string{"a", "b", "c", "d"}, 3); !reflect.DeepEqual(got, [][]string{{"a", "b", "c"}, {"d"}}) {
		t.Errorf("Chunk = %v", got)
	}
	if got := collections.Zip([]string{"a", "b"}, []string{"1", "2"}); !reflect.DeepEqual(got, [][]string{{"a", "1"}, {"b", "2"}}) {
		t.Errorf("Zip = %v", got)
	}
	if got := collections.Unzip([][]string{{"a", "1"}, {"b", "2"}}); !reflect.DeepEqual(got, [][]string{{"a", "b"}, {"1", "2"}}) {
		t.Errorf("Unzip = %v", got)
	}
	floor := func(v float64) int { return int(math.Floor(v)) }
	if got := collections.GroupBy([]float64{6.1, 4.2, 6.3}, floor); !reflect.DeepEqual(got, map[int][]float64{4: {4.2}, 6: {6.1, 6.3}}) {
		t.Errorf("GroupBy = %v", got)
	}
	if got := collections.CountBy([]float64{6.1, 4.2, 6.3}, floor); got[6] != 2 || got[4] != 1 {
		t.Errorf("CountBy = %v", got)
	}
	type user struct {
		Name   string
		Age    int
		Active bool
	}
	users := []user{{"barney", 36, false}, {"fred", 40, true}, {"pebbles", 1, false}, {"barney", 34, true}}
	yes, no := collections.Partition(users, func(u user) bool { return u.Active })
	if len(yes) != 2 || len(no) != 2 || yes[0].Name != "fred" {
		t.Errorf("Partition = %v / %v", yes, no)
	}
	if got := collections.KeyBy(users, func(u user) string { return u.Name }); got["barney"].Age != 34 {
		t.Error("KeyBy should keep the last")
	}
	ordered := collections.OrderBy(users, collections.By(func(u user) string { return u.Name }, false), collections.By(func(u user) int { return u.Age }, true))
	var names []string
	for _, u := range ordered {
		names = append(names, u.Name+":"+string(rune('0'+u.Age%10)))
	}
	if !reflect.DeepEqual(names, []string{"barney:6", "barney:4", "fred:0", "pebbles:1"}) {
		t.Errorf("OrderBy = %v", names)
	}
	if got := collections.UniqBy([]float64{2.1, 1.2, 2.3}, floor); !reflect.DeepEqual(got, []float64{2.1, 1.2}) {
		t.Errorf("UniqBy = %v", got)
	}
	if got := collections.Difference([]int{2, 1}, []int{2, 3}); !reflect.DeepEqual(got, []int{1}) {
		t.Errorf("Difference = %v", got)
	}
	if got := collections.Intersection([]int{2, 1}, []int{2, 3}); !reflect.DeepEqual(got, []int{2}) {
		t.Errorf("Intersection = %v", got)
	}
	if got := collections.Union([]int{2}, []int{1, 2}); !reflect.DeepEqual(got, []int{2, 1}) {
		t.Errorf("Union = %v", got)
	}
	if got := collections.Compact([]int{0, 1, 0, 2}); !reflect.DeepEqual(got, []int{1, 2}) {
		t.Errorf("Compact = %v", got)
	}
	if got := collections.FlattenDeep(dec(`[1, [2, [3, [4]], 5]]`).([]any)); len(got) != 5 {
		t.Errorf("FlattenDeep = %v", got)
	}
	for _, c := range []struct {
		start, end, step int
		want             []int
	}{{0, 4, 1, []int{0, 1, 2, 3}}, {0, 20, 5, []int{0, 5, 10, 15}}, {1, 4, 0, []int{1, 2, 3}}, {0, -4, -1, []int{0, -1, -2, -3}}, {0, 4, -1, nil}} {
		if got := collections.Range(c.start, c.end, c.step); !reflect.DeepEqual(got, c.want) {
			t.Errorf("Range(%d, %d, %d) = %v", c.start, c.end, c.step, got)
		}
	}
	if !reflect.DeepEqual(collections.TakeWhile([]int{1, 2, 5, 1}, func(v int) bool { return v < 3 }), []int{1, 2}) ||
		!reflect.DeepEqual(collections.DropWhile([]int{1, 2, 5, 1}, func(v int) bool { return v < 3 }), []int{5, 1}) {
		t.Error("TakeWhile/DropWhile wrong")
	}
	in := []int{1, 2, 3, 4, 5, 6}
	var log trace.Log
	sh := collections.TraceShuffle(in, rand.New(rand.NewSource(1)), &log)
	sorted := append([]int(nil), sh...)
	sort.Ints(sorted)
	if !reflect.DeepEqual(sorted, in) || !reflect.DeepEqual(trace.Replay(append([]int(nil), in...), log.Events), sh) {
		t.Errorf("Shuffle = %v (a permutation whose swaps replay)", sh)
	}
	if v, ok := collections.MaxBy(users, func(u user) int { return u.Age }); !ok || v.Name != "fred" {
		t.Error("MaxBy wrong")
	}
	if collections.Sum([]int{4, 2, 8, 6}) != 20 || collections.Mean([]int{4, 2, 8, 6}) != 5 || collections.Median([]int{3, 1, 2, 10}) != 2.5 {
		t.Error("Sum/Mean/Median wrong")
	}
}

func TestFunctions(t *testing.T) {
	calls := 0
	var fib func(int) int
	fib = collections.Memoize(func(n int) int {
		calls++
		if n < 2 {
			return n
		}
		return fib(n-1) + fib(n-2)
	})
	if fib(40) != 102334155 || calls != 41 {
		t.Errorf("memoized fib(40) took %d calls", calls)
	}
	m := collections.Memoize(func(n int) int { return n * n })
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) { defer wg.Done(); _ = m(i % 5) }(i)
	}
	wg.Wait()
	n := 0
	once := collections.Once(func() int { n++; return n })
	if first, second := once(), once(); first != 1 || second != 1 {
		t.Error("Once ran twice")
	}
	add := func(a, b, c int) int { return a + b + c }
	if collections.Curry3(add)(1)(2)(3) != 6 || collections.Curry2(func(a, b int) int { return a - b })(5)(2) != 3 {
		t.Error("Curry wrong")
	}
	inc := func(v int) int { return v + 1 }
	dbl := func(v int) int { return v * 2 }
	if collections.Pipe(inc, dbl)(3) != 8 || collections.Compose(inc, dbl)(3) != 7 {
		t.Error("Pipe/Compose wrong")
	}
}

// burst calls f at each offset (ms) on a manual clock, then lets time run out,
// and returns the times (ms) f fired.
func burst(at []int, build func(clk collections.Clock, f func()) func()) []int {
	start := time.Unix(0, 0)
	clk := collections.NewManualClock(start)
	var fired []int
	call := build(clk, func() { fired = append(fired, int(clk.Now().Sub(start)/time.Millisecond)) })
	last := 0
	for _, a := range at {
		clk.Advance(time.Duration(a-last) * time.Millisecond)
		last = a
		call()
	}
	clk.Advance(time.Second)
	return fired
}

func TestDebounceThrottle(t *testing.T) {
	at := []int{0, 10, 20, 150, 160}
	deb := func(o collections.DebounceOptions) func(collections.Clock, func()) func() {
		return func(clk collections.Clock, f func()) func() {
			o.Clock = clk
			call, _ := collections.Debounce(f, 50*time.Millisecond, o)
			return call
		}
	}
	if got := burst(at, deb(collections.DebounceOptions{})); !reflect.DeepEqual(got, []int{70, 210}) {
		t.Errorf("trailing debounce fired at %v", got)
	}
	if got := burst(at, deb(collections.DebounceOptions{Leading: true})); !reflect.DeepEqual(got, []int{0, 150}) {
		t.Errorf("leading debounce fired at %v", got)
	}
	if got := burst(at, deb(collections.DebounceOptions{Leading: true, Trailing: true})); !reflect.DeepEqual(got, []int{0, 70, 150, 210}) {
		t.Errorf("leading and trailing debounce fired at %v", got)
	}
	thr := func(o collections.DebounceOptions) func(collections.Clock, func()) func() {
		return func(clk collections.Clock, f func()) func() {
			o.Clock = clk
			call, _ := collections.Throttle(f, 50*time.Millisecond, o)
			return call
		}
	}
	if got := burst(at, thr(collections.DebounceOptions{})); !reflect.DeepEqual(got, []int{0, 50, 150, 200}) {
		t.Errorf("throttle fired at %v", got)
	}
	if got := burst(at, thr(collections.DebounceOptions{Leading: true})); !reflect.DeepEqual(got, []int{0, 150}) {
		t.Errorf("leading-only throttle fired at %v", got)
	}
	clk := collections.NewManualClock(time.Unix(0, 0))
	n := 0
	call, cancel := collections.Debounce(func() { n++ }, time.Millisecond, collections.DebounceOptions{Clock: clk})
	call()
	cancel()
	clk.Advance(time.Second)
	if n != 0 {
		t.Error("cancel did not drop the pending call")
	}
}

func TestStrings(t *testing.T) {
	for in, want := range map[string]string{"Foo Bar": "fooBar", "--foo-bar--": "fooBar", "__FOO_BAR__": "fooBar"} {
		if got := collections.CamelCase(in); got != want {
			t.Errorf("CamelCase(%q) = %q, want %q", in, got, want)
		}
	}
	if collections.KebabCase("fooBar") != "foo-bar" || collections.KebabCase("__FOO_BAR__") != "foo-bar" || collections.SnakeCase("Foo Bar") != "foo_bar" {
		t.Error("KebabCase/SnakeCase wrong (lodash examples)")
	}
	if collections.StartCase("--foo-bar--") != "Foo Bar" || collections.StartCase("fooBar") != "Foo Bar" || collections.StartCase("__FOO_BAR__") != "FOO BAR" {
		t.Errorf("StartCase wrong: %q", collections.StartCase("__FOO_BAR__"))
	}
	if got := collections.Words("fooBar-baz_qux XMLHttpRequest v2"); !reflect.DeepEqual(got, []string{"foo", "Bar", "baz", "qux", "XML", "Http", "Request", "v", "2"}) {
		t.Errorf("Words = %v", got)
	}
	if collections.Capitalize("FRED") != "Fred" || collections.UpperFirst("fred") != "Fred" || collections.LowerFirst("FRED") != "fRED" {
		t.Error("Capitalize/UpperFirst/LowerFirst wrong")
	}
	if collections.Pad("abc", 8, "_-") != "_-abc_-_" || collections.PadStart("abc", 6, "_-") != "_-_abc" || collections.PadEnd("abc", 6, "_-") != "abc_-_" || collections.PadStart("abc", 2, "") != "abc" {
		t.Error("Pad wrong (lodash examples)")
	}
	text := "hi-diddly-ho there, neighborino"
	if got := collections.Truncate(text, 24, "", ""); got != "hi-diddly-ho there, n..." {
		t.Errorf("Truncate = %q", got)
	}
	if got := collections.Truncate(text, 24, "", " "); got != "hi-diddly-ho there,..." {
		t.Errorf("Truncate with separator = %q", got)
	}
	if got := collections.Truncate(text, 24, " [...]", ""); got != "hi-diddly-ho there [...]" {
		t.Errorf("Truncate with omission = %q", got)
	}
	if e := collections.Escape(`fred, barney, & <pebbles>`); e != "fred, barney, &amp; &lt;pebbles&gt;" || collections.Unescape(e) != `fred, barney, & <pebbles>` {
		t.Error("Escape wrong")
	}
	if collections.Trim("-_-abc-_-", "_-") != "abc" || collections.TrimStart("  abc  ", "") != "abc  " || collections.TrimEnd("-_-abc-_-", "_-") != "-_-abc" {
		t.Error("Trim wrong")
	}
	// MDN examples.
	if c, _ := collections.StrAt("The quick brown fox", -4); c != " " {
		t.Errorf("at(-4) = %q", c)
	}
	if collections.StrSlice("The quick brown fox jumps over the lazy dog.", -9, -5) != "lazy" || collections.StrSlice("Mozilla", 4, 1) != "" || collections.Substring("Mozilla", 4, 1) != "ozi" || collections.Substring("Mozilla", -5, 2) != "Mo" {
		t.Error("slice/substring wrong")
	}
	if got := collections.Split("a b c d", " ", 2); !reflect.DeepEqual(got, []string{"a", "b"}) || len(collections.Split("héllo", "", -1)) != 5 {
		t.Errorf("Split = %v", got)
	}
	if !collections.StartsWith("To be, or not", "or", 7) || !collections.EndsWith("To be, or not", "be", 5) || collections.CharAt("abc", -1) != "" {
		t.Error("StartsWith/EndsWith/CharAt wrong")
	}
	if r, ok := collections.CodePointAt("☃★♲", 1); !ok || r != '★' {
		t.Error("CodePointAt wrong")
	}
	_ = strings.ToLower
}

func TestLangAndMath(t *testing.T) {
	if !collections.IsEmpty(nil) || !collections.IsEmpty("") || !collections.IsEmpty(map[string]any{}) || collections.IsEmpty([]int{1}) || !collections.IsEmpty(5) {
		t.Error("IsEmpty wrong")
	}
	var p *int
	if !collections.IsNil(p) || collections.IsNil(0) {
		t.Error("IsNil wrong")
	}
	if collections.Clamp(-10, -5, 5) != -5 || !collections.InRange(3, 2, 4) || collections.InRange(5.2, 0, 4) || !collections.InRange(-3, -2, -6) || collections.DefaultTo("", "x") != "x" {
		t.Error("Clamp/InRange/DefaultTo wrong")
	}
	for s, want := range map[string]float64{"": 0, " 12 ": 12, "0x11": 17, "0b11": 3, "1e3": 1000, "-Infinity": math.Inf(-1)} {
		if got := collections.ToNumber(s); got != want {
			t.Errorf("Number(%q) = %v, want %v", s, got, want)
		}
	}
	for _, s := range []string{"12px", "1_000", "NaN", "inf"} {
		if !math.IsNaN(collections.ToNumber(s)) {
			t.Errorf("Number(%q) should be NaN", s)
		}
	}
	if collections.JSRound(-2.5) != -2 || math.Round(-2.5) != -3 || collections.JSRound(2.5) != 3 || collections.JSRound(-0.4) != 0 {
		t.Error("JSRound differs from Math.round")
	}
	if collections.Sign(-3) != -1 || collections.Hypot(3, 4) != 5 || collections.Clz32(1) != 31 || collections.Clz32(0) != 32 || collections.Fround(5.5) != 5.5 {
		t.Error("JS math wrong")
	}
	if got := collections.CastArray(1); len(got) != 1 {
		t.Error("CastArray wrong")
	}
}

func TestJSMapAndSet(t *testing.T) {
	m := collections.NewJSMap[string, int]()
	m.Set("b", 1).Set("a", 2).Set("b", 3)
	if !reflect.DeepEqual(m.Keys(), []string{"b", "a"}) || m.Size() != 2 {
		t.Errorf("keys %v: insertion order, update keeps place", m.Keys())
	}
	m.Delete("b")
	m.Set("b", 4)
	if !reflect.DeepEqual(m.Keys(), []string{"a", "b"}) {
		t.Errorf("delete then set should move to the end: %v", m.Keys())
	}
	for i := 0; i < 100; i++ {
		m.Set(string(rune('c'+i%20)), i)
		if i%3 == 0 {
			m.Delete(string(rune('c' + i%20)))
		}
	}
	if v, ok := m.Get("a"); !ok || v != 2 || len(m.Keys()) != m.Size() {
		t.Error("JSMap lost track after compaction")
	}
	m.Clear()
	if m.Size() != 0 || m.Has("a") {
		t.Error("Clear wrong")
	}
	a, b := collections.NewJSSet(2, 4, 6, 8), collections.NewJSSet(1, 4, 9)
	if !reflect.DeepEqual(a.Union(b).Values(), []int{2, 4, 6, 8, 1, 9}) || !reflect.DeepEqual(a.Intersection(b).Values(), []int{4}) ||
		!reflect.DeepEqual(a.Difference(b).Values(), []int{2, 6, 8}) || !reflect.DeepEqual(a.SymmetricDifference(b).Values(), []int{2, 6, 8, 1, 9}) {
		t.Error("ES2025 set methods wrong (MDN examples)")
	}
	if !collections.NewJSSet(4).IsSubsetOf(a) || !a.IsSupersetOf(collections.NewJSSet(2, 8)) || a.IsDisjointFrom(b) || !a.IsDisjointFrom(collections.NewJSSet(3)) {
		t.Error("subset/superset/disjoint wrong")
	}
	if len(a.Snapshot().Nodes) != 4 || len(m.Snapshot().Nodes) != 0 {
		t.Error("snapshots wrong")
	}
}

func TestFormatPath(t *testing.T) {
	for _, p := range []string{"a.b[0].c", `a["weird.key"][2]`, "x"} {
		parts, err := collections.ParsePath(p)
		if err != nil {
			t.Fatal(err)
		}
		if got := collections.FormatPath(parts); got != p {
			t.Errorf("FormatPath(ParsePath(%q)) = %q", p, got)
		}
	}
}
