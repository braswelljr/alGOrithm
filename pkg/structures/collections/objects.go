package collections

import (
	"fmt"
	"math"
	"reflect"
	"sort"
)

// DeepClone returns a copy of v that shares no map or slice with it: maps and
// slices are copied all the way down, other values as they are (lodash
// _.cloneDeep, structuredClone in JS).
//
// Time: O(size of v). Space: the same.
func DeepClone(v any) any {
	switch x := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(x))
		for k, e := range x {
			out[k] = DeepClone(e)
		}
		return out
	case []any:
		out := make([]any, len(x))
		for i, e := range x {
			out[i] = DeepClone(e)
		}
		return out
	}
	return v
}

// MergeOptions change how DeepMerge combines values.
type MergeOptions struct {
	// ConcatArrays appends a source array to the destination array instead
	// of merging them index by index.
	ConcatArrays bool
	// Customizer, when it returns ok, decides the value at path itself
	// (lodash _.mergeWith); dst is the value so far, nil when absent.
	Customizer func(path string, dst, src any) (v any, ok bool)
}

// DeepMerge merges srcs, left to right, into a copy of dst and returns it
// (lodash _.merge, but without mutating dst): maps merge key by key, arrays
// merge index by index (so [9] over [1, 2] gives [9, 2]), any other source
// value replaces the destination's, and a source nil (JSON null, standing in
// for JavaScript's undefined) never overwrites an existing value. No input is
// changed and the result shares nothing with them.
//
// Time: O(total size). Space: the same.
func DeepMerge(dst any, srcs ...any) any { return MergeWith(MergeOptions{}, nil, dst, srcs...) }

// TraceDeepMerge is DeepMerge, reporting every decision to t: "new" for a key
// the destination lacked, "overwritten", "kept" when a nil source left the
// destination alone, and "merged" when two maps or arrays were combined.
func TraceDeepMerge(t Tracer, dst any, srcs ...any) any {
	return MergeWith(MergeOptions{}, t, dst, srcs...)
}

// MergeWith is DeepMerge with options, reporting to t (nil for none).
func MergeWith(opts MergeOptions, t Tracer, dst any, srcs ...any) any {
	out := DeepClone(dst)
	for i, src := range srcs {
		out = merge(opts, t, "", out, src, i+1)
	}
	return out
}

func merge(opts MergeOptions, t Tracer, path string, dst, src any, n int) any {
	if opts.Customizer != nil {
		if v, ok := opts.Customizer(path, dst, src); ok {
			t.step(path, "overwritten", fmt.Sprintf("the customizer chose %s", short(v)))
			return v
		}
	}
	switch s := src.(type) {
	case nil:
		t.step(path, "kept", fmt.Sprintf("source %d is null: kept %s", n, short(dst)))
		return dst
	case map[string]any:
		d, ok := dst.(map[string]any)
		if !ok {
			t.step(path, overwriteAction(dst), fmt.Sprintf("took source %d's object", n))
			return DeepClone(s)
		}
		if path != "" {
			t.step(path, "merged", fmt.Sprintf("merged source %d's object into it", n))
		}
		for _, k := range sortedKeys(s) {
			kp := keyPath(path, k)
			if prev, exists := d[k]; exists {
				d[k] = merge(opts, t, kp, prev, s[k], n)
			} else {
				t.step(kp, "new", fmt.Sprintf("took source %d: %s", n, short(s[k])))
				d[k] = DeepClone(s[k])
			}
		}
		return d
	case []any:
		d, ok := dst.([]any)
		if !ok {
			t.step(path, overwriteAction(dst), fmt.Sprintf("took source %d's array", n))
			return DeepClone(s)
		}
		if opts.ConcatArrays {
			t.step(path, "merged", fmt.Sprintf("appended source %d's %d values", n, len(s)))
			return append(d, DeepClone(s).([]any)...)
		}
		t.step(path, "merged", fmt.Sprintf("merged source %d's array index by index", n))
		for i, v := range s {
			ip := indexPath(path, i)
			if i < len(d) {
				d[i] = merge(opts, t, ip, d[i], v, n)
			} else {
				t.step(ip, "new", fmt.Sprintf("took source %d: %s", n, short(v)))
				d = append(d, DeepClone(v))
			}
		}
		return d
	}
	if IsEqual(dst, src) {
		t.step(path, "kept", fmt.Sprintf("source %d has the same value", n))
		return dst
	}
	t.step(path, overwriteAction(dst), fmt.Sprintf("took source %d: %s (was %s)", n, short(src), short(dst)))
	return src
}

func overwriteAction(dst any) string {
	if dst == nil {
		return "new"
	}
	return "overwritten"
}

// short is a value as compact JSON, cut to 40 bytes.
func short(v any) string {
	s := fmt.Sprint(v)
	if b, err := jsonCompact(v); err == nil {
		s = b
	}
	if len(s) > 40 {
		s = s[:37] + "..."
	}
	return s
}

// IsEqual reports whether a and b are deeply equal (lodash _.isEqual): maps
// with the same keys and equal values, arrays with equal values in order, and
// equal scalars, with NaN equal to NaN as lodash has it. nil and an empty map
// are not equal.
//
// Time: O(size). Space: O(depth).
func IsEqual(a, b any) bool {
	_, differ := FirstDifference(a, b)
	return !differ
}

// FirstDifference returns the path of the first place a and b differ, in
// sorted key order, and whether they differ at all ("" is the root).
func FirstDifference(a, b any) (string, bool) { return firstDiff("", a, b) }

func firstDiff(path string, a, b any) (string, bool) {
	switch x := a.(type) {
	case map[string]any:
		y, ok := b.(map[string]any)
		if !ok || len(x) != len(y) {
			if ok {
				for _, k := range sortedKeys(union(x, y)) {
					if _, inX := x[k]; !inX {
						return keyPath(path, k), true
					}
					if _, inY := y[k]; !inY {
						return keyPath(path, k), true
					}
				}
			}
			return path, true
		}
		for _, k := range sortedKeys(x) {
			yv, has := y[k]
			if !has {
				return keyPath(path, k), true
			}
			if p, d := firstDiff(keyPath(path, k), x[k], yv); d {
				return p, true
			}
		}
		return "", false
	case []any:
		y, ok := b.([]any)
		if !ok {
			return path, true
		}
		for i := 0; i < len(x) && i < len(y); i++ {
			if p, d := firstDiff(indexPath(path, i), x[i], y[i]); d {
				return p, true
			}
		}
		if len(x) != len(y) {
			return indexPath(path, min(len(x), len(y))), true
		}
		return "", false
	case float64:
		y, ok := b.(float64)
		if ok && (x == y || (math.IsNaN(x) && math.IsNaN(y))) {
			return "", false
		}
		return path, true
	}
	if reflect.DeepEqual(a, b) {
		return "", false
	}
	return path, true
}

func union(a, b map[string]any) map[string]any {
	out := make(map[string]any, len(a)+len(b))
	for k := range a {
		out[k] = true
	}
	for k := range b {
		out[k] = true
	}
	return out
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// Get returns the value at path in obj, or def when any step is missing
// (lodash _.get). A path that does not parse also gives def.
//
// Time: O(len(path)).
func Get(obj any, path string, def any) any {
	v, ok := lookup(obj, path)
	if !ok {
		return def
	}
	return v
}

// Has reports whether path leads to a value in obj, even a null one (lodash
// _.has).
func Has(obj any, path string) bool {
	_, ok := lookup(obj, path)
	return ok
}

func lookup(obj any, path string) (any, bool) {
	parts, err := ParsePath(path)
	if err != nil {
		return nil, false
	}
	cur := obj
	for _, p := range parts {
		switch c := cur.(type) {
		case map[string]any:
			v, ok := c[p.Key]
			if !ok {
				return nil, false
			}
			cur = v
		case []any:
			if !p.IsIndex || p.Index >= len(c) {
				return nil, false
			}
			cur = c[p.Index]
		default:
			return nil, false
		}
	}
	return cur, true
}

// Set returns a copy of obj with value at path, creating the maps and arrays
// on the way like lodash _.set: an index step creates an array (padded with
// nil), any other step a map. obj is not changed. A step through a scalar
// replaces it.
//
// Time: O(size of obj) for the copy. Space: the same.
func Set(obj any, path string, value any) (any, error) { return TraceSet(nil, obj, path, value) }

// TraceSet is Set, reporting each container created ("created") and the final
// write ("set" or "overwritten") to t.
func TraceSet(t Tracer, obj any, path string, value any) (any, error) {
	parts, err := ParsePath(path)
	if err != nil {
		return obj, err
	}
	if len(parts) == 0 {
		t.step("", "set", "replaced the whole value")
		return DeepClone(value), nil
	}
	return setAt(t, "", DeepClone(obj), parts, DeepClone(value)), nil
}

func setAt(t Tracer, prefix string, cur any, parts []PathPart, value any) any {
	p := parts[0]
	here := joinPath(prefix, p)
	last := len(parts) == 1
	if p.IsIndex {
		arr, ok := cur.([]any)
		if !ok {
			t.step(prefix, "created", "made an array")
			arr = nil
		}
		for len(arr) <= p.Index {
			arr = append(arr, nil)
		}
		if last {
			t.step(here, overwriteAction(arr[p.Index]), "set to "+short(value))
			arr[p.Index] = value
		} else {
			arr[p.Index] = setAt(t, here, arr[p.Index], parts[1:], value)
		}
		return arr
	}
	m, ok := cur.(map[string]any)
	if !ok {
		t.step(prefix, "created", "made an object")
		m = map[string]any{}
	}
	if last {
		t.step(here, overwriteAction(m[p.Key]), "set to "+short(value))
		m[p.Key] = value
	} else {
		m[p.Key] = setAt(t, here, m[p.Key], parts[1:], value)
	}
	return m
}

// Unset returns a copy of obj without the value at path, and whether there
// was one (lodash _.unset, without mutating). Removing an array element
// leaves nil in its place, as JavaScript's delete leaves a hole.
func Unset(obj any, path string) (any, bool) {
	parts, err := ParsePath(path)
	if err != nil || len(parts) == 0 || !Has(obj, path) {
		return obj, false
	}
	out := DeepClone(obj)
	cur := out
	for _, p := range parts[:len(parts)-1] {
		switch c := cur.(type) {
		case map[string]any:
			cur = c[p.Key]
		case []any:
			cur = c[p.Index]
		}
	}
	switch c := cur.(type) {
	case map[string]any:
		delete(c, parts[len(parts)-1].Key)
	case []any:
		c[parts[len(parts)-1].Index] = nil
	}
	return out, true
}

// Pick returns a new object of only the given paths of obj (lodash _.pick);
// paths that are missing are left out.
func Pick(obj map[string]any, paths ...string) map[string]any {
	out := any(map[string]any{})
	for _, p := range paths {
		if v, ok := lookup(obj, p); ok {
			out, _ = Set(out, p, v)
		}
	}
	return out.(map[string]any)
}

// Omit returns a copy of obj without the given paths (lodash _.omit).
func Omit(obj map[string]any, paths ...string) map[string]any {
	out := DeepClone(obj)
	for _, p := range paths {
		out, _ = Unset(out, p)
	}
	return out.(map[string]any)
}

// Assign copies the own properties of each source onto a copy of dst, left to
// right, without recursing (Object.assign, lodash _.assign). Values are
// shared, not cloned, as in JavaScript.
//
// Time: O(total keys).
func Assign(dst map[string]any, srcs ...map[string]any) map[string]any {
	out := make(map[string]any, len(dst))
	for k, v := range dst {
		out[k] = v
	}
	for _, s := range srcs {
		for k, v := range s {
			out[k] = v
		}
	}
	return out
}

// Defaults returns a copy of obj with each missing (or nil) key taken from
// the first source that has it (lodash _.defaults). It does not recurse.
func Defaults(obj map[string]any, srcs ...map[string]any) map[string]any {
	out := Assign(obj)
	for _, s := range srcs {
		for k, v := range s {
			if cur, ok := out[k]; !ok || cur == nil {
				out[k] = v
			}
		}
	}
	return out
}

// DefaultsDeep is Defaults that recurses into nested objects (lodash
// _.defaultsDeep): obj's values always win.
func DefaultsDeep(obj map[string]any, srcs ...map[string]any) map[string]any {
	out := DeepClone(obj).(map[string]any)
	for _, s := range srcs {
		fillDefaults(out, s)
	}
	return out
}

func fillDefaults(dst, src map[string]any) {
	for k, v := range src {
		cur, ok := dst[k]
		switch {
		case !ok || cur == nil:
			dst[k] = DeepClone(v)
		default:
			d, dm := cur.(map[string]any)
			s, sm := v.(map[string]any)
			if dm && sm {
				fillDefaults(d, s)
			}
		}
	}
}

// Pair is one key and value of a map.
type Pair[K comparable, V any] struct {
	Key   K
	Value V
}

// Keys returns m's keys, sorted (Object.keys, lodash _.keys). Go maps keep no
// insertion order, so where JavaScript returns keys in the order they were
// added this returns them sorted; use Map for insertion order.
func Keys[V any](m map[string]V) []string { return sortedKeys(m) }

// Values returns m's values in the order of Keys (Object.values).
func Values[V any](m map[string]V) []V {
	out := make([]V, 0, len(m))
	for _, k := range sortedKeys(m) {
		out = append(out, m[k])
	}
	return out
}

// Entries returns m's key and value pairs in the order of Keys
// (Object.entries, lodash _.toPairs).
func Entries[V any](m map[string]V) []Pair[string, V] {
	out := make([]Pair[string, V], 0, len(m))
	for _, k := range sortedKeys(m) {
		out = append(out, Pair[string, V]{k, m[k]})
	}
	return out
}

// FromEntries builds a map from pairs, a later pair winning a repeated key
// (Object.fromEntries, lodash _.fromPairs).
func FromEntries[K comparable, V any](pairs []Pair[K, V]) map[K]V {
	out := make(map[K]V, len(pairs))
	for _, p := range pairs {
		out[p.Key] = p.Value
	}
	return out
}

// Invert swaps keys and values; with repeated values a later key (in sorted
// key order) wins (lodash _.invert).
func Invert[V comparable](m map[string]V) map[V]string {
	out := make(map[V]string, len(m))
	for _, k := range sortedKeys(m) {
		out[m[k]] = k
	}
	return out
}

// MapKeys returns a map with each key replaced by f(value, key); on a clash
// the later key in sorted order wins (lodash _.mapKeys).
func MapKeys[V any](m map[string]V, f func(v V, k string) string) map[string]V {
	out := make(map[string]V, len(m))
	for _, k := range sortedKeys(m) {
		out[f(m[k], k)] = m[k]
	}
	return out
}

// MapValues returns a map with each value replaced by f(value, key) (lodash
// _.mapValues).
func MapValues[V, W any](m map[string]V, f func(v V, k string) W) map[string]W {
	out := make(map[string]W, len(m))
	for k, v := range m {
		out[k] = f(v, k)
	}
	return out
}

// MapEntries returns a map built from f applied to each pair (the
// Object.fromEntries(Object.entries(o).map(f)) idiom).
func MapEntries[V, W any](m map[string]V, f func(k string, v V) (string, W)) map[string]W {
	out := make(map[string]W, len(m))
	for _, k := range sortedKeys(m) {
		nk, nv := f(k, m[k])
		out[nk] = nv
	}
	return out
}

// ForOwn calls f with each key and value in sorted key order, stopping when f
// returns false (lodash _.forOwn).
func ForOwn[V any](m map[string]V, f func(v V, k string) bool) {
	for _, k := range sortedKeys(m) {
		if !f(m[k], k) {
			return
		}
	}
}

// FindKey returns the first key, in sorted order, whose value pred accepts
// (lodash _.findKey); ok is false when none does.
func FindKey[V any](m map[string]V, pred func(v V) bool) (string, bool) {
	for _, k := range sortedKeys(m) {
		if pred(m[k]) {
			return k, true
		}
	}
	return "", false
}

// IsMap reports whether v is a decoded JSON object, a map[string]any (lodash
// _.isPlainObject for the values this package works on).
func IsMap(v any) bool {
	_, ok := v.(map[string]any)
	return ok
}

// FlattenObject returns obj as one level of paths to its leaves
// ("a.b[0].c": v). Empty maps and arrays are leaves.
//
// Time: O(size). Space: O(size).
func FlattenObject(obj any) map[string]any { return TraceFlattenObject(nil, obj) }

// TraceFlattenObject is FlattenObject, reporting each leaf ("flattened").
func TraceFlattenObject(t Tracer, obj any) map[string]any {
	out := map[string]any{}
	var walk func(path string, v any)
	walk = func(path string, v any) {
		switch x := v.(type) {
		case map[string]any:
			if len(x) > 0 {
				for _, k := range sortedKeys(x) {
					walk(keyPath(path, k), x[k])
				}
				return
			}
		case []any:
			if len(x) > 0 {
				for i, e := range x {
					walk(indexPath(path, i), e)
				}
				return
			}
		}
		t.step(path, "flattened", short(v))
		out[path] = DeepClone(v)
	}
	walk("", obj)
	return out
}

// UnflattenObject rebuilds the nested value FlattenObject made from its
// paths.
func UnflattenObject(flat map[string]any) (any, error) {
	var out any = map[string]any{}
	for _, k := range sortedKeys(flat) {
		var err error
		if out, err = Set(out, k, flat[k]); err != nil {
			return nil, err
		}
	}
	return out, nil
}
