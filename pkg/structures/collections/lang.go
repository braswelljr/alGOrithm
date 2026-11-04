package collections

import (
	"math"
	"reflect"
	"strconv"
	"strings"
)

// IsEmpty reports whether v is nil, "", or an empty map, slice or array
// (lodash _.isEmpty for the values Go has). Numbers and booleans are not
// collections, so like lodash they count as empty.
func IsEmpty(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.String, reflect.Map, reflect.Slice, reflect.Array, reflect.Chan:
		return rv.Len() == 0
	case reflect.Pointer, reflect.Interface:
		return rv.IsNil()
	default:
		return true // not a collection
	}
}

// IsNil reports whether v is nil or a nil pointer, map, slice, channel or
// function (lodash _.isNil, without JavaScript's undefined).
func IsNil(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Pointer, reflect.Map, reflect.Slice, reflect.Chan, reflect.Func, reflect.Interface:
		return rv.IsNil()
	default:
		return false
	}
}

// CastArray returns v unchanged when it is a []any, and []any{v} otherwise
// (lodash _.castArray).
func CastArray(v any) []any {
	if s, ok := v.([]any); ok {
		return s
	}
	return []any{v}
}

// Clamp limits v to [lo, hi] (lodash _.clamp).
func Clamp[T Number](v, lo, hi T) T { return max(lo, min(v, hi)) }

// InRange reports whether lo <= v < hi, swapping lo and hi when lo is larger
// (lodash _.inRange).
func InRange[T Number](v, lo, hi T) bool {
	if lo > hi {
		lo, hi = hi, lo
	}
	return v >= lo && v < hi
}

// DefaultTo returns v unless it is the zero value, in which case def (lodash
// _.defaultTo, where JavaScript also replaces NaN).
func DefaultTo[T comparable](v, def T) T {
	var zero T
	if v == zero {
		return def
	}
	return v
}

// ToNumber converts s the way JavaScript's Number(s) does: surrounding
// whitespace is ignored, "" is 0, "0x", "0o" and "0b" prefixes are read,
// "Infinity" and "-Infinity" are accepted, and anything else that is not a
// decimal number is NaN (where Go's strconv would report an error).
func ToNumber(s string) float64 {
	s = strings.TrimSpace(s)
	switch s {
	case "":
		return 0
	case "Infinity", "+Infinity":
		return math.Inf(1)
	case "-Infinity":
		return math.Inf(-1)
	}
	for prefix, base := range map[string]int{"0x": 16, "0X": 16, "0o": 8, "0O": 8, "0b": 2, "0B": 2} {
		if strings.HasPrefix(s, prefix) {
			n, err := strconv.ParseUint(s[2:], base, 64)
			if err != nil {
				return math.NaN()
			}
			return float64(n)
		}
	}
	if strings.ContainsAny(s, "_") || strings.EqualFold(s, "nan") || strings.EqualFold(strings.TrimLeft(s, "+-"), "inf") {
		return math.NaN() // Go accepts these; JavaScript does not
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return math.NaN()
	}
	return f
}
