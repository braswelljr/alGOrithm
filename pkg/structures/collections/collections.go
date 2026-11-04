// Package collections holds the utilities every codebase writes again, in the
// shape lodash and the JavaScript standard library gave them, for Go: deep
// merge, deep clone, deep equality and path access on decoded JSON
// (map[string]any and []any); slice helpers (chunk, zip, group by, order by,
// uniq by, shuffle); function helpers (memoize, once, debounce, throttle,
// curry, compose); string helpers (camelCase and friends, words, truncate,
// pad, escape) and JS String methods with JS semantics; an insertion-ordered
// Map and Set like JavaScript's; and the JS Math functions Go lacks.
//
// Each doc comment names the lodash or JS function it mirrors, says what is
// copied and what is shared, and states the cost. Nothing here mutates its
// arguments unless its name says so.
//
//	merged := collections.DeepMerge(
//		collections.MustDecode(`{"a": {"b": 1, "c": [1, 2]}}`),
//		collections.MustDecode(`{"a": {"c": [9], "d": true}}`),
//	) // {"a": {"b": 1, "c": [9, 2], "d": true}}
package collections

import (
	"encoding/json"
	"strings"
)

// Decode parses a JSON document into the values this package works on:
// map[string]any, []any, string, float64, bool and nil. A syntax error names
// the byte offset where it happened.
func Decode(s string) (any, error) {
	dec := json.NewDecoder(strings.NewReader(s))
	var v any
	if err := dec.Decode(&v); err != nil {
		return nil, err
	}
	if dec.More() {
		return nil, &json.SyntaxError{Offset: dec.InputOffset()}
	}
	return v, nil
}

// MustDecode is Decode for literals in examples and tests: it panics on
// invalid JSON.
func MustDecode(s string) any {
	v, err := Decode(s)
	if err != nil {
		panic(err)
	}
	return v
}

// Encode returns v as indented JSON with sorted keys.
func Encode(v any) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "<" + err.Error() + ">"
	}
	return string(b)
}

// Step is one decision a traced operation made at a path, for a step by step
// view: Action is "new", "overwritten", "kept", "merged", "created", "set",
// "removed" or "flattened".
type Step struct {
	Path   string
	Action string
	Detail string
}

// Tracer receives the Steps of a traced operation; nil discards them.
type Tracer func(Step)

func (t Tracer) step(path, action, detail string) {
	if t != nil {
		t(Step{Path: path, Action: action, Detail: detail})
	}
}

func jsonCompact(v any) (string, error) {
	b, err := json.Marshal(v)
	return string(b), err
}
