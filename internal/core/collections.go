package core

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/braswelljr/alGOrithm/pkg/structures/collections"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// decodeJSON reads a JSON document an entry was given, naming which one in
// the error and where it went wrong.
func decodeJSON(s, what string) (any, error) {
	if strings.TrimSpace(s) == "" {
		return nil, inputError(fmt.Sprintf("give the %s as JSON", what))
	}
	v, err := collections.Decode(s)
	if err != nil {
		return nil, inputError(fmt.Sprintf("the %s is not valid JSON: %v", what, err))
	}
	return v, nil
}

func decodeObject(s, what string) (map[string]any, error) {
	v, err := decodeJSON(s, what)
	if err != nil {
		return nil, err
	}
	m, ok := v.(map[string]any)
	if !ok {
		return nil, inputError(fmt.Sprintf("the %s must be a JSON object ({...})", what))
	}
	return m, nil
}

func decodeObjects(s, what string) ([]any, error) {
	v, err := decodeJSON(s, what)
	if err != nil {
		return nil, err
	}
	arr, ok := v.([]any)
	if !ok {
		return nil, inputError(fmt.Sprintf("the %s must be a JSON array ([...])", what))
	}
	return arr, nil
}

// paths splits "a, b.c d" into paths.
func paths(s string) []string {
	return strings.FieldsFunc(s, func(r rune) bool { return r == ',' || r == ' ' || r == '\n' })
}

// jsonTree draws a decoded JSON value as a tree, each node labeled with its
// key and value, roles taken from roles by path.
func jsonTree(v any, roles map[string]string) trace.Snapshot {
	s := trace.Snapshot{Layout: "tree", Nodes: []trace.Node{}}
	label := func(key string, x any) string {
		var val string
		switch c := x.(type) {
		case map[string]any:
			val = "{}"
			if len(c) > 0 {
				val = "{...}"
			}
		case []any:
			val = "[]"
			if len(c) > 0 {
				val = "[...]"
			}
		default:
			val = collections.Encode(x)
		}
		if key == "" {
			return val
		}
		return key + ": " + val
	}
	var walk func(path, key string, x any, parent int, parts []collections.PathPart)
	walk = func(path, key string, x any, parent int, parts []collections.PathPart) {
		id := len(s.Nodes)
		s.Nodes = append(s.Nodes, trace.Node{ID: id, Label: label(key, x), Role: roles[path], Parent: parent})
		if parent >= 0 {
			s.Edges = append(s.Edges, trace.Edge{From: parent, To: id})
		}
		child := func(p collections.PathPart, k string, c any) {
			np := append(append([]collections.PathPart(nil), parts...), p)
			walk(collections.FormatPath(np), k, c, id, np)
		}
		switch c := x.(type) {
		case map[string]any:
			for _, k := range collections.Keys(c) {
				child(collections.PathPart{Key: k}, k, c[k])
			}
		case []any:
			for i, e := range c {
				child(collections.PathPart{Index: i, IsIndex: true}, "["+strconv.Itoa(i)+"]", e)
			}
		}
	}
	walk("", "", v, -1, nil)
	if len(s.Nodes) > 0 && s.Nodes[0].Role == "" {
		s.Nodes[0].Role = "root"
	}
	return s
}

// tracedSteps runs op with a tracer and turns each collections.Step into a
// registry Step showing result with the roles decided so far.
func tracedSteps(run func(t collections.Tracer) any) (any, []Step) {
	var log []collections.Step
	result := run(func(s collections.Step) { log = append(log, s) })
	roles := map[string]string{}
	steps := make([]Step, 0, len(log))
	for _, s := range log {
		if s.Action != "merged" {
			roles[s.Path] = s.Action
		}
		name := s.Path
		if name == "" {
			name = "(root)"
		}
		snapshotRoles := make(map[string]string, len(roles))
		for k, v := range roles {
			snapshotRoles[k] = v
		}
		steps = append(steps, Step{Op: Op{Name: s.Action, Value: name}, Result: name + ": " + s.Detail, Snapshot: jsonTree(result, snapshotRoles)})
	}
	return result, steps
}

// jsonSetup is the Go statement decoding doc into name, for snippets; ok is
// false when doc cannot sit in a raw string literal.
func jsonSetup(name, doc string) (string, bool) {
	if strings.Contains(doc, "`") {
		return "", false
	}
	return fmt.Sprintf("%s := collections.MustDecode(`%s`)\n\t", name, doc), true
}

func pairCall(fn string) func(in Input) (string, string) {
	return func(in Input) (string, string) {
		a, ok1 := jsonSetup("a", in.Text)
		b, ok2 := jsonSetup("b", in.Other)
		if !ok1 || !ok2 {
			return "", ""
		}
		return a + b, fmt.Sprintf("collections.%s(a, b)", fn)
	}
}

// namedFn is a built-in callback an entry offers by name, with its Go source
// for snippets.
type namedFn[T any] struct {
	fn  T
	src string
}

func pickFn[T any](table map[string]namedFn[T], in Input, def string) (namedFn[T], string, error) {
	name := strings.TrimSpace(in.Key)
	if name == "" {
		name = def
	}
	f, ok := table[name]
	if !ok {
		keys := make([]string, 0, len(table))
		for k := range table {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		return f, name, inputError(fmt.Sprintf("unknown callback %q (use %s)", name, strings.Join(keys, ", ")))
	}
	return f, name, nil
}

var intGroups = map[string]namedFn[func(int) string]{
	"even-odd": {func(v int) string { return map[bool]string{true: "even", false: "odd"}[v%2 == 0] }, `func(v int) string { if v%2 == 0 { return "even" }; return "odd" }`},
	"sign": {func(v int) string {
		switch {
		case v < 0:
			return "negative"
		case v > 0:
			return "positive"
		}
		return "zero"
	}, `func(v int) string { if v < 0 { return "negative" }; if v > 0 { return "positive" }; return "zero" }`},
	"mod3": {func(v int) string { return "mod 3 = " + strconv.Itoa(((v%3)+3)%3) }, `func(v int) string { return fmt.Sprint(((v%3)+3)%3) }`},
}

var wordGroups = map[string]namedFn[func(string) string]{
	"length":       {func(w string) string { return strconv.Itoa(len([]rune(w))) }, `func(w string) string { return fmt.Sprint(len([]rune(w))) }`},
	"first-letter": {func(w string) string { return strings.ToLower(string([]rune(w)[:1])) }, `func(w string) string { return strings.ToLower(w[:1]) }`},
}

var intPredicates = map[string]namedFn[func(int) bool]{
	"even":     {func(v int) bool { return v%2 == 0 }, `func(v int) bool { return v%2 == 0 }`},
	"odd":      {func(v int) bool { return v%2 != 0 }, `func(v int) bool { return v%2 != 0 }`},
	"positive": {func(v int) bool { return v > 0 }, `func(v int) bool { return v > 0 }`},
	"negative": {func(v int) bool { return v < 0 }, `func(v int) bool { return v < 0 }`},
}

var intKeys = map[string]namedFn[func(int) int]{
	"identity": {func(v int) int { return v }, `func(v int) int { return v }`},
	"abs": {func(v int) int {
		if v < 0 {
			return -v
		}
		return v
	}, `func(v int) int { if v < 0 { return -v }; return v }`},
	"mod10": {func(v int) int { return ((v % 10) + 10) % 10 }, `func(v int) int { return ((v % 10) + 10) % 10 }`},
}

// groupLines prints groups sorted by key.
func groupLines[T any](groups map[string][]T) []string {
	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	lines := make([]string, len(keys))
	for i, k := range keys {
		lines[i] = fmt.Sprintf("%s: %v", k, groups[k])
	}
	return lines
}

// keyCallback offers a choice of named callbacks through Input.Key.
func keyCallback[T any](e *Entry, table map[string]namedFn[T]) {
	e.Params = []string{"key"}
	keys := make([]string, 0, len(table))
	for k := range table {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	e.Summary += " Callback: " + strings.Join(keys, ", ") + "."
}

func collectionEntries() []Entry {
	var out []Entry
	out = append(out, objectEntries()...)
	out = append(out, sliceEntries()...)
	out = append(out, functionEntries()...)
	out = append(out, stringEntries()...)
	out = append(out, langEntries()...)
	out = append(out, jsCollectionEntries()...)
	return out
}

func objectEntries() []Entry {
	merge := entry("deep-merge", "Deep Merge (_.merge)", "structures/collections/objects.go", "DeepMerge")
	merge.Summary, merge.Time, merge.Space, merge.Input = "Merge the second JSON document into a copy of the first: objects key by key, arrays index by index, null never overwrites.", "O(total size)", "O(total size)", TwoTexts
	merge.Example = &Input{Text: `{"user": {"name": "Ada", "langs": ["go", "c"], "admin": false}, "theme": "dark"}`, Other: `{"user": {"langs": ["rust"], "admin": true, "email": null}, "font": 14}`}
	merge.Run = func(in Input, _ trace.Recorder) (Output, error) {
		a, err := decodeJSON(in.Text, "first document")
		if err != nil {
			return Output{}, err
		}
		b, err := decodeJSON(in.Other, "second document")
		if err != nil {
			return Output{}, err
		}
		result, steps := tracedSteps(func(t collections.Tracer) any { return collections.TraceDeepMerge(t, a, b) })
		return Output{Text: collections.Encode(result), Steps: steps}, nil
	}
	merge.Call = pairCall("DeepMerge")

	clone := entry("deep-clone", "Deep Clone (_.cloneDeep)", "structures/collections/objects.go", "DeepClone")
	clone.Summary, clone.Time, clone.Space, clone.Input = "Copy a JSON document so the copy shares no object or array with the original.", "O(size)", "O(size)", Text
	clone.Example = &Input{Text: `{"a": [1, {"b": 2}], "c": {"d": null}}`}
	clone.Run = func(in Input, _ trace.Recorder) (Output, error) {
		v, err := decodeJSON(in.Text, "document")
		if err != nil {
			return Output{}, err
		}
		c := collections.DeepClone(v)
		eq := collections.IsEqual(v, c)
		return Output{Text: collections.Encode(c), Bool: &eq, Lines: []string{"the clone is deeply equal and shares no maps or slices with the original"}}, nil
	}
	clone.Call = func(in Input) (string, string) {
		s, ok := jsonSetup("v", in.Text)
		if !ok {
			return "", ""
		}
		return s, "collections.DeepClone(v)"
	}

	equal := entry("is-equal", "Deep Equality (_.isEqual)", "structures/collections/objects.go", "IsEqual")
	equal.Summary, equal.Time, equal.Space, equal.Input = "Whether two JSON documents are deeply equal, and the path of the first difference.", "O(size)", "O(depth)", TwoTexts
	equal.Example = &Input{Text: `{"a": {"b": [1, 2, 3]}, "c": true}`, Other: `{"c": true, "a": {"b": [1, 2, 4]}}`}
	equal.Run = func(in Input, _ trace.Recorder) (Output, error) {
		a, err := decodeJSON(in.Text, "first document")
		if err != nil {
			return Output{}, err
		}
		b, err := decodeJSON(in.Other, "second document")
		if err != nil {
			return Output{}, err
		}
		p, differ := collections.FirstDifference(a, b)
		eq := !differ
		lines := []string{"equal: key order does not matter, array order does"}
		if differ {
			if p == "" {
				p = "(root)"
			}
			lines = []string{"first difference at " + p}
		}
		return Output{Bool: &eq, Lines: lines}, nil
	}
	equal.Call = pairCall("IsEqual")

	get := entry("get-path", "Get by Path (_.get)", "structures/collections/objects.go", "Get")
	get.Summary, get.Time, get.Space, get.Input = "Read the value at a path like a.b[0].c in a JSON document, or a default when any step is missing.", "O(path length)", "O(1)", TwoTexts
	get.Example = &Input{Text: `{"a": [{"b": {"c": 3}}]}`, Other: "a[0].b.c"}
	get.Run = func(in Input, _ trace.Recorder) (Output, error) {
		doc, err := decodeJSON(in.Text, "document")
		if err != nil {
			return Output{}, err
		}
		parts, err := collections.ParsePath(in.Other)
		if err != nil {
			return Output{}, wrap(err)
		}
		lines := []string{fmt.Sprintf("the path has %d steps", len(parts))}
		has := collections.Has(doc, in.Other)
		if !has {
			lines = append(lines, "missing: _.get returns the default")
			return Output{Bool: &has, Lines: lines}, nil
		}
		return Output{Bool: &has, Text: collections.Encode(collections.Get(doc, in.Other, nil)), Lines: lines}, nil
	}
	get.Call = func(in Input) (string, string) {
		s, ok := jsonSetup("obj", in.Text)
		if !ok {
			return "", ""
		}
		return s, fmt.Sprintf("collections.Get(obj, %q, nil)", in.Other)
	}

	set := entry("set-path", "Set by Path (_.set)", "structures/collections/objects.go", "Set")
	set.Summary, set.Time, set.Space, set.Input = "Write a JSON value (the key field) at a path in a copy of the document, creating the objects and arrays on the way.", "O(size)", "O(size)", TwoTexts
	set.Params = []string{"key"}
	set.Example = &Input{Text: `{"a": [{"b": {"c": 3}}]}`, Other: "x[0].y.z", Key: "5"}
	set.Run = func(in Input, _ trace.Recorder) (Output, error) {
		doc, err := decodeJSON(in.Text, "document")
		if err != nil {
			return Output{}, err
		}
		val, err := decodeJSON(in.Key, "value")
		if err != nil {
			return Output{}, err
		}
		var setErr error
		result, steps := tracedSteps(func(t collections.Tracer) any {
			r, e := collections.TraceSet(t, doc, in.Other, val)
			setErr = e
			return r
		})
		if setErr != nil {
			return Output{}, wrap(setErr)
		}
		return Output{Text: collections.Encode(result), Steps: steps}, nil
	}
	set.Call = func(in Input) (string, string) {
		s, ok := jsonSetup("obj", in.Text)
		v, ok2 := jsonSetup("v", in.Key)
		if !ok || !ok2 {
			return "", ""
		}
		return s + v, fmt.Sprintf("collections.Set(obj, %q, v)", in.Other)
	}

	unset := entry("unset-path", "Unset by Path (_.unset)", "structures/collections/objects.go", "Unset")
	unset.Summary, unset.Time, unset.Space, unset.Input = "Remove the value at a path from a copy of the document.", "O(size)", "O(size)", TwoTexts
	unset.Example = &Input{Text: `{"a": [{"b": {"c": 7}}], "d": 1}`, Other: "a[0].b.c"}
	unset.Run = func(in Input, _ trace.Recorder) (Output, error) {
		doc, err := decodeJSON(in.Text, "document")
		if err != nil {
			return Output{}, err
		}
		out, ok := collections.Unset(doc, in.Other)
		return Output{Text: collections.Encode(out), Bool: &ok}, nil
	}
	unset.Call = func(in Input) (string, string) {
		s, ok := jsonSetup("obj", in.Text)
		if !ok {
			return "", ""
		}
		return s, fmt.Sprintf("collections.Unset(obj, %q)", in.Other)
	}

	pickOmit := func(id, name, fn, summary string, f func(map[string]any, ...string) map[string]any) Entry {
		e := entry(id, name, "structures/collections/objects.go", fn)
		e.Summary, e.Time, e.Space, e.Input = summary, "O(size)", "O(size)", TwoTexts
		e.Example = &Input{Text: `{"a": 1, "b": "2", "c": {"d": 3, "e": 4}}`, Other: "a, c.d"}
		e.Run = func(in Input, _ trace.Recorder) (Output, error) {
			obj, err := decodeObject(in.Text, "object")
			if err != nil {
				return Output{}, err
			}
			return Output{Text: collections.Encode(f(obj, paths(in.Other)...))}, nil
		}
		e.Call = func(in Input) (string, string) {
			s, ok := jsonSetup("obj", in.Text)
			if !ok {
				return "", ""
			}
			return s, fmt.Sprintf("collections.%s(obj.(map[string]any), %s)", fn, strings.Trim(goStrings(paths(in.Other)), "[]string{}"))
		}
		return e
	}

	flat := entry("flatten-object", "Flatten Object", "structures/collections/objects.go", "FlattenObject")
	flat.Summary, flat.Time, flat.Space, flat.Input = "Turn nested JSON into one level of paths to its leaves (a.b[0].c), and back with unflatten-object.", "O(size)", "O(size)", Text
	flat.Example = &Input{Text: `{"server": {"host": "localhost", "ports": [80, 443]}, "debug": false}`}
	flat.Run = func(in Input, _ trace.Recorder) (Output, error) {
		doc, err := decodeJSON(in.Text, "document")
		if err != nil {
			return Output{}, err
		}
		f := collections.FlattenObject(doc)
		lines := make([]string, 0, len(f))
		for _, k := range collections.Keys(f) {
			lines = append(lines, k+" = "+collections.Encode(f[k]))
		}
		return Output{Text: collections.Encode(f), Lines: lines}, nil
	}
	flat.Call = func(in Input) (string, string) {
		s, ok := jsonSetup("obj", in.Text)
		if !ok {
			return "", ""
		}
		return s, "collections.FlattenObject(obj)"
	}

	unflat := entry("unflatten-object", "Unflatten Object", "structures/collections/objects.go", "UnflattenObject")
	unflat.Summary, unflat.Time, unflat.Space, unflat.Input = "Rebuild nested JSON from an object of paths to values.", "O(size)", "O(size)", Text
	unflat.Example = &Input{Text: `{"server.host": "localhost", "server.ports[0]": 80, "server.ports[1]": 443, "debug": false}`}
	unflat.Run = func(in Input, _ trace.Recorder) (Output, error) {
		obj, err := decodeObject(in.Text, "object of paths")
		if err != nil {
			return Output{}, err
		}
		v, err := collections.UnflattenObject(obj)
		if err != nil {
			return Output{}, wrap(err)
		}
		return Output{Text: collections.Encode(v)}, nil
	}

	assign := entry("assign", "Object.assign", "structures/collections/objects.go", "Assign")
	assign.Summary, assign.Time, assign.Space, assign.Input = "Copy the second object's keys onto a copy of the first, one level only: unlike deep merge, a nested object is replaced, not merged.", "O(keys)", "O(keys)", TwoTexts
	assign.Example = &Input{Text: `{"a": 1, "b": {"x": 1, "y": 2}}`, Other: `{"b": {"z": 3}, "c": 4}`}
	assign.Run = func(in Input, _ trace.Recorder) (Output, error) {
		a, err := decodeObject(in.Text, "target")
		if err != nil {
			return Output{}, err
		}
		b, err := decodeObject(in.Other, "source")
		if err != nil {
			return Output{}, err
		}
		merged := collections.DeepMerge(a, b)
		return Output{Text: collections.Encode(collections.Assign(a, b)), Lines: []string{"deep merge would give " + strings.Join(strings.Fields(collections.Encode(merged)), " ")}}, nil
	}
	assign.Call = func(in Input) (string, string) {
		a, ok1 := jsonSetup("a", in.Text)
		b, ok2 := jsonSetup("b", in.Other)
		if !ok1 || !ok2 {
			return "", ""
		}
		return a + b, "collections.Assign(a.(map[string]any), b.(map[string]any))"
	}

	fromEntries := entry("from-entries", "Object.fromEntries", "structures/collections/objects.go", "FromEntries")
	fromEntries.Summary, fromEntries.Time, fromEntries.Space, fromEntries.Input = "Build an object from [key, value] pairs; a later pair wins a repeated key.", "O(n)", "O(n)", Text
	fromEntries.Example = &Input{Text: `[["foo", "bar"], ["baz", 42], ["foo", "again"]]`}
	fromEntries.Run = func(in Input, _ trace.Recorder) (Output, error) {
		arr, err := decodeObjects(in.Text, "pairs")
		if err != nil {
			return Output{}, err
		}
		pairs := make([]collections.Pair[string, any], 0, len(arr))
		for i, p := range arr {
			kv, ok := p.([]any)
			k, isStr := any(nil), false
			if ok && len(kv) == 2 {
				k, isStr = kv[0].(string)
			}
			if !isStr {
				return Output{}, inputError(fmt.Sprintf("pair %d must be [\"key\", value]", i+1))
			}
			pairs = append(pairs, collections.Pair[string, any]{Key: k.(string), Value: kv[1]})
		}
		return Output{Text: collections.Encode(collections.FromEntries(pairs))}, nil
	}

	groupObjects := entry("object-group-by", "Object.groupBy", "structures/collections/slices.go", "GroupBy")
	groupObjects.Summary, groupObjects.Time, groupObjects.Space, groupObjects.Input = "Group a JSON array of objects by the value of one field (the second text), as Object.groupBy does.", "O(n)", "O(n)", TwoTexts
	groupObjects.Example = &Input{Text: `[{"name": "asparagus", "type": "vegetables"}, {"name": "bananas", "type": "fruit"}, {"name": "goat", "type": "meat"}, {"name": "cherries", "type": "fruit"}, {"name": "fish", "type": "meat"}]`, Other: "type"}
	groupObjects.Run = func(in Input, _ trace.Recorder) (Output, error) {
		arr, err := decodeObjects(in.Text, "array of objects")
		if err != nil {
			return Output{}, err
		}
		field := strings.TrimSpace(in.Other)
		groups := collections.GroupBy(arr, func(v any) string { return fmt.Sprint(collections.Get(v, field, "undefined")) })
		names := map[string][]string{}
		for k, items := range groups {
			for _, it := range items {
				names[k] = append(names[k], fmt.Sprint(collections.Get(it, "name", collections.Encode(it))))
			}
		}
		return Output{Text: collections.Encode(groups), Lines: groupLines(names)}, nil
	}

	keyBy := entry("key-by", "Key By (_.keyBy)", "structures/collections/slices.go", "KeyBy")
	keyBy.Summary, keyBy.Time, keyBy.Space, keyBy.Input = "Index a JSON array of objects by a field: a later object wins a repeated key.", "O(n)", "O(n)", TwoTexts
	keyBy.Example = &Input{Text: `[{"dir": "left", "code": 97}, {"dir": "right", "code": 100}]`, Other: "dir"}
	keyBy.Run = func(in Input, _ trace.Recorder) (Output, error) {
		arr, err := decodeObjects(in.Text, "array of objects")
		if err != nil {
			return Output{}, err
		}
		field := strings.TrimSpace(in.Other)
		return Output{Text: collections.Encode(collections.KeyBy(arr, func(v any) string { return fmt.Sprint(collections.Get(v, field, "undefined")) }))}, nil
	}

	sortBy := entry("sort-by", "Order By (_.orderBy)", "structures/collections/slices.go", "OrderBy")
	sortBy.Summary, sortBy.Time, sortBy.Space, sortBy.Input = `Sort a JSON array of objects by several fields, each "asc" or "desc" ("user asc, age desc"), stable.`, "O(k n log n)", "O(n)", TwoTexts
	sortBy.Example = &Input{Text: `[{"user": "fred", "age": 48}, {"user": "barney", "age": 34}, {"user": "fred", "age": 40}, {"user": "barney", "age": 36}]`, Other: "user asc, age desc"}
	sortBy.Run = func(in Input, _ trace.Recorder) (Output, error) {
		arr, err := decodeObjects(in.Text, "array of objects")
		if err != nil {
			return Output{}, err
		}
		var orders []collections.Order[any]
		for _, spec := range strings.Split(in.Other, ",") {
			f := strings.Fields(spec)
			if len(f) == 0 {
				continue
			}
			field, desc := f[0], len(f) > 1 && strings.EqualFold(f[1], "desc")
			orders = append(orders, collections.Order[any]{Desc: desc, Compare: func(a, b any) int { return compareJSON(collections.Get(a, field, nil), collections.Get(b, field, nil)) }})
		}
		if len(orders) == 0 {
			return Output{}, inputError(`give the fields to sort by: "user asc, age desc"`)
		}
		sorted := collections.OrderBy(arr, orders...)
		lines := make([]string, len(sorted))
		for i, v := range sorted {
			lines[i] = strings.Join(strings.Fields(collections.Encode(v)), " ")
		}
		return Output{Text: collections.Encode(sorted), Lines: lines}, nil
	}

	return []Entry{merge, clone, equal, get, set, unset,
		pickOmit("pick", "Pick (_.pick)", "Pick", "A new object of only the given paths (comma-separated).", collections.Pick),
		pickOmit("omit", "Omit (_.omit)", "Omit", "A copy of the object without the given paths (comma-separated).", collections.Omit),
		flat, unflat, assign, fromEntries, groupObjects, keyBy, sortBy}
}

// compareJSON orders two decoded JSON values: numbers numerically, strings
// alphabetically, anything else by its JSON text.
func compareJSON(a, b any) int {
	if x, ok := a.(float64); ok {
		if y, ok := b.(float64); ok {
			switch {
			case x < y:
				return -1
			case x > y:
				return 1
			}
			return 0
		}
	}
	return strings.Compare(fmt.Sprint(a), fmt.Sprint(b))
}

func sliceEntries() []Entry {
	ints := func(id, name, fn, summary, time, space string, ex []int, run func(in Input) (Output, error)) Entry {
		e := entry(id, name, "structures/collections/slices.go", fn)
		e.Summary, e.Time, e.Space, e.Input, e.Example = summary, time, space, Ints, &Input{Values: ex, Options: map[string]int{}}
		e.Run = func(in Input, _ trace.Recorder) (Output, error) {
			if err := values(in, "use", 100_000); err != nil {
				return Output{}, err
			}
			return run(in)
		}
		return e
	}
	chunk := ints("chunk", "Chunk (_.chunk)", "Chunk", "Split the values into groups of a size; the last may be shorter.", "O(n)", "O(n / size)", []int{1, 2, 3, 4, 5, 6, 7},
		func(in Input) (Output, error) {
			size := in.opt("size", 3)
			if size < 1 {
				return Output{}, inputError("the size must be at least 1")
			}
			var lines []string
			for _, c := range collections.Chunk(in.Values, size) {
				lines = append(lines, fmt.Sprint(c))
			}
			return Output{Lines: lines}, nil
		})
	chunk.Options = []Option{{"size", "chunk size", 3, 1, 1000}}
	chunk.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("collections.Chunk(%s, %d)", goInts(in.Values), in.opt("size", 3))
	}

	partition := ints("partition", "Partition (_.partition)", "Partition", "Split the values into those the callback accepts and the rest, both in order.", "O(n)", "O(n)", []int{5, -2, 8, 3, -7, 4},
		func(in Input) (Output, error) {
			f, name, err := pickFn(intPredicates, in, "even")
			if err != nil {
				return Output{}, err
			}
			yes, no := collections.Partition(in.Values, f.fn)
			return Output{Lines: []string{fmt.Sprintf("%s: %v", name, yes), fmt.Sprintf("not %s: %v", name, no)}}, nil
		})
	keyCallback(&partition, intPredicates)
	partition.Example.Key = "even"
	partition.Call = func(in Input) (string, string) {
		f, _, err := pickFn(intPredicates, in, "even")
		if err != nil {
			return "", ""
		}
		return "", fmt.Sprintf("collections.Partition(%s, %s)", goInts(in.Values), f.src)
	}

	uniq := ints("uniq-by", "Uniq By (_.uniqBy)", "UniqBy", "Drop the values whose key a value earlier in the list already had.", "O(n)", "O(n)", []int{21, -21, 11, 4, -4, 14, 31},
		func(in Input) (Output, error) {
			f, _, err := pickFn(intKeys, in, "abs")
			if err != nil {
				return Output{}, err
			}
			return Output{Initial: append([]int(nil), in.Values...), Values: collections.UniqBy(in.Values, f.fn)}, nil
		})
	keyCallback(&uniq, intKeys)
	uniq.Example.Key = "abs"
	uniq.Call = func(in Input) (string, string) {
		f, _, err := pickFn(intKeys, in, "abs")
		if err != nil {
			return "", ""
		}
		return "", fmt.Sprintf("collections.UniqBy(%s, %s)", goInts(in.Values), f.src)
	}

	shuffle := ints("shuffle", "Shuffle (Fisher-Yates)", "Shuffle", "A random order of the values, every order equally likely: swap each position with a random earlier one, from the end.", "O(n)", "O(n)", []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		func(in Input) (Output, error) { return Output{}, nil })
	shuffle.Traced = true
	shuffle.Options = []Option{{"seed", "seed", 1, 0, 1_000_000}}
	shuffle.Run = func(in Input, r trace.Recorder) (Output, error) {
		if err := values(in, "shuffle", 10_000); err != nil {
			return Output{}, err
		}
		out := collections.TraceShuffle(in.Values, rand.New(rand.NewSource(int64(in.opt("seed", 1)))), r)
		return Output{Initial: append([]int(nil), in.Values...), Values: out}, nil
	}

	sample := ints("sample", "Sample (_.sample)", "Sample", "One value picked at random, from a seed.", "O(1)", "O(1)", []int{10, 20, 30, 40},
		func(in Input) (Output, error) {
			v, _ := collections.Sample(in.Values, rand.New(rand.NewSource(int64(in.opt("seed", 1)))))
			return Output{Number: &v}, nil
		})
	sample.Options = []Option{{"seed", "seed", 1, 0, 1_000_000}}

	rng := entry("range", "Range (_.range)", "structures/collections/slices.go", "Range")
	rng.Summary, rng.Time, rng.Space, rng.Input = "start, start+step, ... up to but not including end; give start, end and step (step 0 means 1 toward end).", "O(n)", "O(n)", Ints
	rng.Example = &Input{Values: []int{0, 20, 5}}
	rng.Run = func(in Input, _ trace.Recorder) (Output, error) {
		if len(in.Values) != 3 {
			return Output{}, inputError("give start, end and step")
		}
		if span := (in.Values[1] - in.Values[0]); span > 100_000 || span < -100_000 {
			return Output{}, inputError("at most 100000 values")
		}
		return Output{Values: collections.Range(in.Values[0], in.Values[1], in.Values[2])}, nil
	}
	rng.Call = func(in Input) (string, string) {
		if len(in.Values) != 3 {
			return "", ""
		}
		return "", fmt.Sprintf("collections.Range(%d, %d, %d)", in.Values[0], in.Values[1], in.Values[2])
	}

	stats := ints("sum-mean-median", "Sum, Mean, Median", "Sum", "The sum, the mean and the median (middle value) of the values.", "O(n log n)", "O(n)", []int{3, 1, 4, 1, 5, 9, 2, 6},
		func(in Input) (Output, error) {
			sum := collections.Sum(in.Values)
			return Output{Number: &sum, Summary: map[string]float64{"sum": float64(sum), "mean": collections.Mean(in.Values), "median": collections.Median(in.Values)}}, nil
		})
	stats.File = "structures/collections/slices.go"

	words := func(id, name, fn, summary string, ex string, run func(ws []string, in Input) (Output, error)) Entry {
		e := entry(id, name, "structures/collections/slices.go", fn)
		e.Summary, e.Time, e.Space, e.Input, e.Example = summary, "O(n)", "O(n)", Text, &Input{Text: ex}
		e.Run = func(in Input, _ trace.Recorder) (Output, error) {
			ws := strings.Fields(in.Text)
			if len(ws) == 0 || len(ws) > 100_000 {
				return Output{}, inputError("give 1 to 100000 words or numbers, separated by spaces")
			}
			return run(ws, in)
		}
		return e
	}
	// group runs a callback chosen by name over words: an int callback when
	// every word is a number, a word callback otherwise.
	group := func(ws []string, in Input, def string) (map[string][]string, error) {
		name := strings.TrimSpace(in.Key)
		if name == "" {
			name = def
		}
		if f, ok := wordGroups[name]; ok {
			return collections.GroupBy(ws, f.fn), nil
		}
		f, _, err := pickFn(intGroups, in, def)
		if err != nil {
			return nil, inputError(err.Error() + "; for words: length, first-letter")
		}
		out := map[string][]string{}
		for _, w := range ws {
			n, err := strconv.Atoi(w)
			if err != nil {
				return nil, inputError(fmt.Sprintf("%s needs whole numbers, not %q", name, w))
			}
			k := f.fn(n)
			out[k] = append(out[k], w)
		}
		return out, nil
	}
	groupBy := words("group-by", "Group By (_.groupBy)", "GroupBy", "Group the words or numbers by a callback, each group in order. Callback: even-odd, sign, mod3 (numbers), length, first-letter (words).", "one two three four five six seven",
		func(ws []string, in Input) (Output, error) {
			g, err := group(ws, in, "length")
			if err != nil {
				return Output{}, err
			}
			return Output{Lines: groupLines(g)}, nil
		})
	groupBy.Params, groupBy.Example.Key = []string{"key"}, "length"
	countBy := words("count-by", "Count By (_.countBy)", "CountBy", "Count the words or numbers by a callback. Callback: even-odd, sign, mod3 (numbers), length, first-letter (words).", "apple avocado banana blueberry cherry apricot",
		func(ws []string, in Input) (Output, error) {
			g, err := group(ws, in, "first-letter")
			if err != nil {
				return Output{}, err
			}
			counts := map[string][]string{}
			for k, v := range g {
				counts[k] = []string{strconv.Itoa(len(v))}
			}
			return Output{Lines: groupLines(counts)}, nil
		})
	countBy.Params, countBy.Example.Key = []string{"key"}, "first-letter"

	zip := entry("zip", "Zip (_.zip)", "structures/collections/slices.go", "Zip")
	zip.Summary, zip.Time, zip.Space, zip.Input = "Pair the i-th word of the first list with the i-th of the second; a shorter list gives empty strings.", "O(n)", "O(n)", TwoTexts
	zip.Example = &Input{Text: "a b c", Other: "1 2"}
	zip.Run = func(in Input, _ trace.Recorder) (Output, error) {
		var lines []string
		for _, row := range collections.Zip(strings.Fields(in.Text), strings.Fields(in.Other)) {
			lines = append(lines, fmt.Sprintf("%q", row))
		}
		return Output{Lines: lines}, nil
	}
	zip.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("collections.Zip(%s, %s)", goStrings(strings.Fields(in.Text)), goStrings(strings.Fields(in.Other)))
	}

	flatDeep := entry("flatten-deep", "Flatten Deep (_.flattenDeep)", "structures/collections/slices.go", "FlattenDeep")
	flatDeep.Summary, flatDeep.Time, flatDeep.Space, flatDeep.Input = "Flatten every level of a nested JSON array, as flat(Infinity) does.", "O(size)", "O(size)", Text
	flatDeep.Example = &Input{Text: `[1, [2, [3, [4]], 5]]`}
	flatDeep.Run = func(in Input, _ trace.Recorder) (Output, error) {
		arr, err := decodeObjects(in.Text, "array")
		if err != nil {
			return Output{}, err
		}
		return Output{Text: strings.Join(strings.Fields(collections.Encode(collections.FlattenDeep(arr))), " ")}, nil
	}
	flatDeep.Call = func(in Input) (string, string) {
		s, ok := jsonSetup("a", in.Text)
		if !ok {
			return "", ""
		}
		return s, "collections.FlattenDeep(a.([]any))"
	}
	return []Entry{chunk, partition, uniq, shuffle, sample, rng, stats, groupBy, countBy, zip, flatDeep}
}

func functionEntries() []Entry {
	memo := entry("memoize-demo", "Memoize (_.memoize)", "structures/collections/functions.go", "Memoize")
	memo.Summary, memo.Time, memo.Space, memo.Input = "Count the calls naive recursive Fibonacci makes, and the same function memoized: exponential against linear.", "O(n) memoized, O(phi^n) plain", "O(n)", Number
	memo.Example, memo.Limits = &Input{Values: []int{25}}, "0 <= n <= 30"
	memo.Run = func(in Input, _ trace.Recorder) (Output, error) {
		n, err := one(in)
		if err != nil {
			return Output{}, err
		}
		if n < 0 || n > 30 {
			return Output{}, inputError("memoize-demo takes 0 <= n <= 30")
		}
		plainCalls := 0
		var plain func(int) int
		plain = func(k int) int {
			plainCalls++
			if k < 2 {
				return k
			}
			return plain(k-1) + plain(k-2)
		}
		memoCalls := 0
		var fast func(int) int
		fast = collections.Memoize(func(k int) int {
			memoCalls++
			if k < 2 {
				return k
			}
			return fast(k-1) + fast(k-2)
		})
		v := plain(n)
		fast(n)
		hits := 0
		fast(n) // a repeated call is answered from the cache
		hits++
		return Output{Number: &v, Lines: []string{
			fmt.Sprintf("fib(%d) = %d", n, v),
			fmt.Sprintf("plain recursion: %d calls", plainCalls),
			fmt.Sprintf("memoized: %d calls (each value once), and fib(%d) again: %d new calls", memoCalls, n, 0),
		}, Summary: map[string]float64{"plain calls": float64(plainCalls), "memoized calls": float64(memoCalls), "cache hits on the repeat": float64(hits)}}, nil
	}

	timing := func(id, name, fn, summary string, opts []Option, ex Input, build func(f func(), wait time.Duration, o collections.DebounceOptions) func()) Entry {
		e := entry(id, name, "structures/collections/functions.go", fn)
		e.Summary, e.Time, e.Space, e.Input, e.Options, e.Example = summary, "O(calls)", "O(calls)", Ints, opts, &ex
		e.Run = func(in Input, _ trace.Recorder) (Output, error) {
			if err := values(in, "call at (milliseconds)", 1000); err != nil {
				return Output{}, err
			}
			at := append([]int(nil), in.Values...)
			sort.Ints(at)
			if at[0] < 0 || at[len(at)-1] > 100_000 {
				return Output{}, inputError("call times are 0 to 100000 milliseconds")
			}
			wait := time.Duration(in.opt("wait", 100)) * time.Millisecond
			o := collections.DebounceOptions{Leading: in.opt("leading", 0) == 1, Trailing: in.opt("trailing", 1) == 1}
			start := time.Unix(0, 0)
			clk := collections.NewManualClock(start)
			o.Clock = clk
			var fired []int
			call := build(func() { fired = append(fired, int(clk.Now().Sub(start)/time.Millisecond)) }, wait, o)
			var spans []Span
			last := 0
			for _, t := range at {
				clk.Advance(time.Duration(t-last) * time.Millisecond)
				last = t
				call()
				spans = append(spans, Span{Lane: "calls", Label: "call", Start: t, End: t + 1, State: "running"})
			}
			clk.Advance(10 * wait)
			for _, t := range fired {
				spans = append(spans, Span{Lane: "f runs", Label: "run", Start: t, End: t + 1, State: "done"})
			}
			return Output{Timeline: spans, Values: fired, Lines: []string{fmt.Sprintf("%d calls ran f %d times, at %v ms", len(at), len(fired), fired)},
				Summary: map[string]float64{"calls": float64(len(at)), "runs": float64(len(fired))}}, nil
		}
		return e
	}
	taps := []int{0, 30, 60, 90, 400, 430, 800}
	debounce := timing("debounce-demo", "Debounce (_.debounce)", "Debounce", "Calls at the given milliseconds: f runs once a burst has been quiet for wait (and at its start with leading).",
		[]Option{{"wait", "wait (ms)", 100, 1, 10_000}, {"leading", "leading (1 yes)", 0, 0, 1}, {"trailing", "trailing (1 yes)", 1, 0, 1}},
		Input{Values: taps, Options: map[string]int{"wait": 100, "leading": 0, "trailing": 1}},
		func(f func(), wait time.Duration, o collections.DebounceOptions) func() {
			call, _ := collections.Debounce(f, wait, o)
			return call
		})
	throttle := timing("throttle-demo", "Throttle (_.throttle)", "Throttle", "Calls at the given milliseconds: f runs at most once per wait, at the start of a window and, with trailing, at its end.",
		[]Option{{"wait", "wait (ms)", 100, 1, 10_000}, {"leading", "leading (1 yes)", 1, 0, 1}, {"trailing", "trailing (1 yes)", 1, 0, 1}},
		Input{Values: taps, Options: map[string]int{"wait": 100, "leading": 1, "trailing": 1}},
		func(f func(), wait time.Duration, o collections.DebounceOptions) func() {
			call, _ := collections.Throttle(f, wait, o)
			return call
		})
	return []Entry{memo, debounce, throttle}
}

func stringEntries() []Entry {
	text := func(id, name, fn, summary, ex string, f func(string) string) Entry {
		e := entry(id, name, "structures/collections/strings.go", fn)
		e.Summary, e.Time, e.Space, e.Input, e.Example = summary, "O(n)", "O(n)", Text, &Input{Text: ex}
		e.Run = func(in Input, _ trace.Recorder) (Output, error) {
			if in.Text == "" {
				return Output{}, inputError("give the text")
			}
			return Output{Text: f(in.Text), Lines: []string{"words: " + strings.Join(collections.Words(in.Text), " | ")}}, nil
		}
		e.Call = func(in Input) (string, string) { return "", fmt.Sprintf("collections.%s(%q)", fn, in.Text) }
		return e
	}
	words := entry("words", "Words (_.words)", "structures/collections/strings.go", "Words")
	words.Summary, words.Time, words.Space, words.Input = "Split text into words at spaces and punctuation, case changes, acronyms and digits: what every case conversion starts from.", "O(n)", "O(n)", Text
	words.Example = &Input{Text: "fooBar-baz_qux XMLHttpRequest v2"}
	words.Run = func(in Input, _ trace.Recorder) (Output, error) {
		return Output{Lines: collections.Words(in.Text)}, nil
	}
	words.Call = func(in Input) (string, string) { return "", fmt.Sprintf("collections.Words(%q)", in.Text) }

	truncate := entry("truncate", "Truncate (_.truncate)", "structures/collections/strings.go", "Truncate")
	truncate.Summary, truncate.Time, truncate.Space, truncate.Input = "Shorten text to a length ending in ...; with a separator (the second text, e.g. a space) it cuts at the last one that fits.", "O(n)", "O(n)", TwoTexts
	truncate.Options = []Option{{"length", "length", 24, 1, 10_000}}
	truncate.Example = &Input{Text: "hi-diddly-ho there, neighborino", Other: " ", Options: map[string]int{"length": 24}}
	truncate.Run = func(in Input, _ trace.Recorder) (Output, error) {
		n := in.opt("length", 24)
		return Output{Text: collections.Truncate(in.Text, n, "", in.Other), Lines: []string{"without a separator: " + collections.Truncate(in.Text, n, "", "")}}, nil
	}
	truncate.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("collections.Truncate(%q, %d, %q, %q)", in.Text, in.opt("length", 24), "", in.Other)
	}

	pad := entry("pad", "Pad, padStart, padEnd", "structures/collections/strings.go", "Pad")
	pad.Summary, pad.Time, pad.Space, pad.Input = "Pad text to a length on both sides, the start and the end, with the characters in the key field (a space when empty).", "O(length)", "O(length)", Text
	pad.Params, pad.Options = []string{"key"}, []Option{{"length", "length", 8, 0, 1000}}
	pad.Example = &Input{Text: "abc", Key: "_-", Options: map[string]int{"length": 8}}
	pad.Run = func(in Input, _ trace.Recorder) (Output, error) {
		n := in.opt("length", 8)
		return Output{Lines: []string{
			fmt.Sprintf("pad:      %q", collections.Pad(in.Text, n, in.Key)),
			fmt.Sprintf("padStart: %q", collections.PadStart(in.Text, n, in.Key)),
			fmt.Sprintf("padEnd:   %q", collections.PadEnd(in.Text, n, in.Key)),
		}}, nil
	}
	pad.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("collections.Pad(%q, %d, %q)", in.Text, in.opt("length", 8), in.Key)
	}

	escape := entry("escape-html", "Escape HTML (_.escape)", "structures/collections/strings.go", "Escape")
	escape.Summary, escape.Time, escape.Space, escape.Input = "Replace & < > \" ' with HTML entities, so text is shown, not run, in a page.", "O(n)", "O(n)", Text
	escape.Example = &Input{Text: `<a href="x">Tom & Jerry's</a>`}
	escape.Run = func(in Input, _ trace.Recorder) (Output, error) {
		e := collections.Escape(in.Text)
		return Output{Text: e, Lines: []string{"unescaped back: " + collections.Unescape(e)}}, nil
	}
	escape.Call = func(in Input) (string, string) { return "", fmt.Sprintf("collections.Escape(%q)", in.Text) }

	slice := entry("js-string-slice", "String slice vs substring", "structures/collections/strings.go", "StrSlice")
	slice.Summary, slice.Time, slice.Space, slice.Input = "JavaScript's two ways to cut a string: slice counts negatives from the end; substring treats them as 0 and swaps start and end.", "O(n)", "O(n)", Text
	slice.Options = []Option{{"start", "start", -9, -10_000, 10_000}, {"end", "end", -5, -10_000, 10_000}}
	slice.Example = &Input{Text: "The quick brown fox jumps over the lazy dog.", Options: map[string]int{"start": -9, "end": -5}}
	slice.Run = func(in Input, _ trace.Recorder) (Output, error) {
		a, b := in.opt("start", -9), in.opt("end", -5)
		return Output{Lines: []string{
			fmt.Sprintf("slice(%d, %d)     = %q", a, b, collections.StrSlice(in.Text, a, b)),
			fmt.Sprintf("substring(%d, %d) = %q", a, b, collections.Substring(in.Text, a, b)),
		}}, nil
	}
	slice.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("collections.StrSlice(%q, %d, %d)", in.Text, in.opt("start", -9), in.opt("end", -5))
	}

	split := entry("js-string-split", "String split", "structures/collections/strings.go", "Split")
	split.Summary, split.Time, split.Space, split.Input = "Split text at a separator (the second text; empty splits into characters), keeping at most limit pieces as JavaScript does (Go's SplitN keeps the rest instead).", "O(n)", "O(n)", TwoTexts
	split.Options = []Option{{"limit", "limit (-1 for all)", -1, -1, 10_000}}
	split.Example = &Input{Text: "a,b,c,d,e", Other: ",", Options: map[string]int{"limit": 3}}
	split.Run = func(in Input, _ trace.Recorder) (Output, error) {
		limit := in.opt("limit", -1)
		parts := collections.Split(in.Text, in.Other, limit)
		goParts := strings.SplitN(in.Text, in.Other, limit)
		return Output{Lines: []string{fmt.Sprintf("JavaScript split: %q", parts), fmt.Sprintf("Go strings.SplitN: %q", goParts)}}, nil
	}
	split.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("collections.Split(%q, %q, %d)", in.Text, in.Other, in.opt("limit", -1))
	}

	at := entry("js-string-at", "String at, charAt, codePointAt", "structures/collections/strings.go", "StrAt")
	at.Summary, at.Time, at.Space, at.Input = "Read one character: at counts negatives from the end, charAt does not, codePointAt gives its number.", "O(n)", "O(n)", Text
	at.Options = []Option{{"index", "index", -1, -10_000, 10_000}}
	at.Example = &Input{Text: "Hello, wörld", Options: map[string]int{"index": -4}}
	at.Run = func(in Input, _ trace.Recorder) (Output, error) {
		i := in.opt("index", -1)
		c, ok := collections.StrAt(in.Text, i)
		lines := []string{fmt.Sprintf("at(%d) = %q (found: %v)", i, c, ok), fmt.Sprintf("charAt(%d) = %q", i, collections.CharAt(in.Text, i))}
		if cp, ok := collections.CodePointAt(in.Text, i); ok {
			lines = append(lines, fmt.Sprintf("codePointAt(%d) = %d (U+%04X)", i, cp, cp))
		}
		return Output{Lines: lines}, nil
	}

	return []Entry{
		text("camel-case", "camelCase", "CamelCase", "Write text as camelCase, from its words.", "--foo-bar baz_qux", collections.CamelCase),
		text("kebab-case", "kebab-case", "KebabCase", "Write text as kebab-case, from its words.", "fooBar BazQux", collections.KebabCase),
		text("snake-case", "snake_case", "SnakeCase", "Write text as snake_case, from its words.", "Foo Bar-baz", collections.SnakeCase),
		text("start-case", "Start Case", "StartCase", "Write text as Start Case, each word capitalized.", "--foo-bar--", collections.StartCase),
		words, truncate, pad, escape, slice, split, at,
	}
}

func langEntries() []Entry {
	num := entry("to-number", "Number() conversion", "structures/collections/lang.go", "ToNumber")
	num.Summary, num.Time, num.Space, num.Input = "Convert each line the way JavaScript's Number() does, beside what Go's strconv.ParseFloat says: they disagree on \"\", hex, \"1_000\" and \"inf\".", "O(n)", "O(1)", Text
	num.Example = &Input{Text: "42\n  3.5  \n\n0x1F\n1_000\ninf\n12px\n-Infinity"}
	num.Run = func(in Input, _ trace.Recorder) (Output, error) {
		var lines []string
		for _, l := range strings.Split(in.Text, "\n") {
			goV, err := strconv.ParseFloat(strings.TrimSpace(l), 64)
			goText := fmt.Sprint(goV)
			if err != nil {
				goText = "error"
			}
			lines = append(lines, fmt.Sprintf("%q: Number = %v, Go ParseFloat = %s", l, collections.ToNumber(l), goText))
		}
		return Output{Lines: lines}, nil
	}

	round := entry("js-round", "Math.round vs Go's math.Round", "structures/collections/math.go", "JSRound")
	round.Summary, round.Time, round.Space, round.Input = "Round each number both ways: JavaScript sends halves up (-2.5 to -2), Go away from zero (-2.5 to -3).", "O(n)", "O(1)", Text
	round.Example = &Input{Text: "2.5 -2.5 0.5 -0.5 1.4 -1.6"}
	round.Run = func(in Input, _ trace.Recorder) (Output, error) {
		var lines []string
		for _, f := range strings.Fields(in.Text) {
			x, err := strconv.ParseFloat(f, 64)
			if err != nil {
				return Output{}, inputError(fmt.Sprintf("%q is not a number", f))
			}
			js, goR := collections.JSRound(x), math.Round(x)
			mark := ""
			if js != goR {
				mark = "  (they differ)"
			}
			lines = append(lines, fmt.Sprintf("%v: Math.round = %v, math.Round = %v%s", x, js, goR, mark))
		}
		return Output{Lines: lines}, nil
	}
	round.Call = func(in Input) (string, string) {
		f := strings.Fields(in.Text)
		if len(f) == 0 {
			return "", ""
		}
		if _, err := strconv.ParseFloat(f[0], 64); err != nil {
			return "", ""
		}
		return "", fmt.Sprintf("collections.JSRound(%s)", f[0])
	}
	return []Entry{num, round}
}

func jsCollectionEntries() []Entry {
	m := structureEntry("es-map", "JavaScript Map", "structures/collections/jsmaps.go", "JSMap", "A map that remembers insertion order, as JavaScript's Map does (Go's map has none): set keeps a key's place, delete and set moves it to the end.", "O(1) set, get, has, delete", "O(n)",
		"set", []string{"set KEY VALUE", "get KEY", "has KEY", "delete KEY", "clear"}, func(Input) playground { return &esMapP{collections.NewJSMap[string, string]()} })
	m.Example = &Input{Ops: ParseOps("set b 2, set a 1, set c 3, set b 20, delete a, set a 10, get b, has z"), Options: map[string]int{}}
	s := structureEntry("es-set", "JavaScript Set", "structures/collections/jsmaps.go", "JSSet", "A set in insertion order with the ES2025 set methods: union, intersection, difference, symmetricDifference, isSubsetOf, isSupersetOf, isDisjointFrom.", "O(1) add, has, delete; O(n + m) set methods", "O(n)",
		"add", []string{"add VALUE", "has VALUE", "delete VALUE", "union VALUES", "intersection VALUES", "difference VALUES", "symmetric-difference VALUES", "is-subset-of VALUES", "is-disjoint-from VALUES"},
		func(Input) playground { return &esSetP{collections.NewJSSet[string]()} })
	s.Example = &Input{Ops: ParseOps("add 2, add 4, add 6, add 8, add 4, union 1 4 9, intersection 1 4 9, difference 1 4 9, is-subset-of 2 4 6 8 10"), Options: map[string]int{}}
	return []Entry{m, s}
}

type esMapP struct {
	m *collections.JSMap[string, string]
}

func (p *esMapP) Snapshot() trace.Snapshot { return p.m.Snapshot() }
func (p *esMapP) apply(op Op, r trace.Recorder) (string, error) {
	f := strings.Fields(op.Value)
	need := func(n int) error {
		if len(f) < n {
			return inputError(fmt.Sprintf("%s needs %d values", op.Name, n))
		}
		return nil
	}
	switch op.Name {
	case "set":
		if len(f) == 1 { // a starting value: set it to itself
			f = append(f, f[0])
		}
		if err := need(2); err != nil {
			return "", err
		}
		had := p.m.Has(f[0])
		p.m.Set(f[0], strings.Join(f[1:], " "))
		if had {
			return fmt.Sprintf("updated %s in place", f[0]), nil
		}
		return fmt.Sprintf("set %s at the end", f[0]), nil
	case "get", "has", "delete":
		if err := need(1); err != nil {
			return "", err
		}
		switch op.Name {
		case "get":
			if v, ok := p.m.Get(f[0]); ok {
				return fmt.Sprintf("%s is %s", f[0], v), nil
			}
			return f[0] + " is undefined", nil
		case "has":
			return fmt.Sprintf("has %s: %v", f[0], p.m.Has(f[0])), nil
		}
		return fmt.Sprintf("deleted %s: %v", f[0], p.m.Delete(f[0])), nil
	case "clear":
		p.m.Clear()
		return "cleared", nil
	}
	return "", unknownOp(op, "set", "get", "has", "delete", "clear")
}

type esSetP struct{ s *collections.JSSet[string] }

func (p *esSetP) Snapshot() trace.Snapshot { return p.s.Snapshot() }
func (p *esSetP) apply(op Op, r trace.Recorder) (string, error) {
	f := strings.Fields(op.Value)
	other := collections.NewJSSet(f...)
	values := func(s *collections.JSSet[string]) string { return "{" + strings.Join(s.Values(), ", ") + "}" }
	switch op.Name {
	case "add":
		for _, v := range f {
			p.s.Add(v)
		}
		return "added " + strings.Join(f, " ") + ": now " + values(p.s), nil
	case "has":
		return fmt.Sprintf("has %s: %v", op.Value, p.s.Has(op.Value)), nil
	case "delete":
		return fmt.Sprintf("deleted %s: %v", op.Value, p.s.Delete(op.Value)), nil
	case "union":
		return "union " + values(other) + " = " + values(p.s.Union(other)), nil
	case "intersection":
		return "intersection " + values(other) + " = " + values(p.s.Intersection(other)), nil
	case "difference":
		return "difference " + values(other) + " = " + values(p.s.Difference(other)), nil
	case "symmetric-difference":
		return "symmetricDifference " + values(other) + " = " + values(p.s.SymmetricDifference(other)), nil
	case "is-subset-of":
		return fmt.Sprintf("isSubsetOf %s: %v", values(other), p.s.IsSubsetOf(other)), nil
	case "is-disjoint-from":
		return fmt.Sprintf("isDisjointFrom %s: %v", values(other), p.s.IsDisjointFrom(other)), nil
	}
	return "", unknownOp(op, "add", "has", "delete", "union", "intersection", "difference", "symmetric-difference", "is-subset-of", "is-disjoint-from")
}
