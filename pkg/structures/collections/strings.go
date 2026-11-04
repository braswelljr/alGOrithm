package collections

import (
	"strings"
	"unicode"
)

// These work on runes (Unicode code points). JavaScript strings are UTF-16,
// so an index or a length differs from JavaScript's for characters outside
// the Basic Multilingual Plane (most emoji count 2 in JavaScript, 1 here).

// Words splits s into words the way lodash _.words does: at spaces and
// punctuation, at a lower-to-upper case change ("fooBar"), before the last
// capital of an acronym ("XMLHttp" -> XML Http), and between letters and
// digits.
//
// Time: O(n). Space: O(n).
func Words(s string) []string {
	r := []rune(s)
	var words []string
	start := -1
	flush := func(end int) {
		if start >= 0 && end > start {
			words = append(words, string(r[start:end]))
		}
		start = -1
	}
	kind := func(c rune) int {
		switch {
		case unicode.IsUpper(c):
			return 1
		case unicode.IsLower(c):
			return 2
		case unicode.IsDigit(c):
			return 3
		case unicode.IsLetter(c):
			return 2
		}
		return 0
	}
	for i, c := range r {
		k := kind(c)
		if k == 0 {
			flush(i)
			continue
		}
		if start < 0 {
			start = i
			continue
		}
		prev := kind(r[i-1])
		switch {
		case prev == 2 && k == 1, prev == 3 && k != 3, prev != 3 && k == 3:
			flush(i)
			start = i
		case prev == 1 && k == 1 && i+1 < len(r) && kind(r[i+1]) == 2:
			flush(i) // "XMLHttp": the H starts a word
			start = i
		}
	}
	flush(len(r))
	return words
}

func mapWords(s string, f func(i int, w string) string, sep string) string {
	ws := Words(s)
	for i, w := range ws {
		ws[i] = f(i, w)
	}
	return strings.Join(ws, sep)
}

// CamelCase writes s as camelCase ("Foo Bar" -> "fooBar", lodash _.camelCase).
func CamelCase(s string) string {
	return mapWords(s, func(i int, w string) string {
		if i == 0 {
			return strings.ToLower(w)
		}
		return Capitalize(w)
	}, "")
}

// KebabCase writes s as kebab-case (lodash _.kebabCase).
func KebabCase(s string) string { return mapWords(s, lowerWord, "-") }

// SnakeCase writes s as snake_case (lodash _.snakeCase).
func SnakeCase(s string) string { return mapWords(s, lowerWord, "_") }

// StartCase writes s as Start Case, each word's first letter upper case
// (lodash _.startCase).
func StartCase(s string) string {
	return mapWords(s, func(_ int, w string) string { return UpperFirst(w) }, " ")
}

func lowerWord(_ int, w string) string { return strings.ToLower(w) }

// Capitalize upper-cases the first letter and lower-cases the rest (lodash
// _.capitalize).
func Capitalize(s string) string { return UpperFirst(strings.ToLower(s)) }

// UpperFirst upper-cases the first letter (lodash _.upperFirst).
func UpperFirst(s string) string {
	r := []rune(s)
	if len(r) > 0 {
		r[0] = unicode.ToUpper(r[0])
	}
	return string(r)
}

// LowerFirst lower-cases the first letter (lodash _.lowerFirst).
func LowerFirst(s string) string {
	r := []rune(s)
	if len(r) > 0 {
		r[0] = unicode.ToLower(r[0])
	}
	return string(r)
}

// PadStart pads s on the left with chars, repeated and cut, to length runes
// (String.prototype.padStart, lodash _.padStart); chars "" means a space.
func PadStart(s string, length int, chars string) string {
	return padding(len([]rune(s)), length, chars) + s
}

// PadEnd pads s on the right (String.prototype.padEnd, lodash _.padEnd).
func PadEnd(s string, length int, chars string) string {
	return s + padding(len([]rune(s)), length, chars)
}

// Pad pads s on both sides, the extra character on the right (lodash _.pad).
func Pad(s string, length int, chars string) string {
	n := len([]rune(s))
	if length <= n {
		return s
	}
	left := (length - n) / 2
	return padding(0, left, chars) + s + padding(0, length-n-left, chars)
}

func padding(have, want int, chars string) string {
	if chars == "" {
		chars = " "
	}
	if want <= have {
		return ""
	}
	c := []rune(chars)
	out := make([]rune, want-have)
	for i := range out {
		out[i] = c[i%len(c)]
	}
	return string(out)
}

// Truncate shortens s to at most length runes, ending in omission ("..." when
// empty), cutting at the last separator that fits when separator is set
// (lodash _.truncate).
func Truncate(s string, length int, omission, separator string) string {
	if omission == "" {
		omission = "..."
	}
	r := []rune(s)
	if len(r) <= length {
		return s
	}
	keep := length - len([]rune(omission))
	if keep < 0 {
		keep = 0
	}
	cut := string(r[:keep])
	if separator != "" {
		if i := strings.LastIndex(cut, separator); i > 0 {
			cut = cut[:i]
		}
	}
	return cut + omission
}

var htmlEscapes = strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;", "'", "&#39;")
var htmlUnescapes = strings.NewReplacer("&amp;", "&", "&lt;", "<", "&gt;", ">", "&quot;", `"`, "&#39;", "'")

// Escape replaces & < > " ' with HTML entities (lodash _.escape).
func Escape(s string) string { return htmlEscapes.Replace(s) }

// Unescape reverses Escape (lodash _.unescape).
func Unescape(s string) string { return htmlUnescapes.Replace(s) }

// Trim removes chars (whitespace when empty) from both ends (lodash _.trim).
func Trim(s, chars string) string {
	if chars == "" {
		return strings.TrimSpace(s)
	}
	return strings.Trim(s, chars)
}

// TrimStart removes chars from the start (lodash _.trimStart).
func TrimStart(s, chars string) string {
	if chars == "" {
		return strings.TrimLeftFunc(s, unicode.IsSpace)
	}
	return strings.TrimLeft(s, chars)
}

// TrimEnd removes chars from the end (lodash _.trimEnd).
func TrimEnd(s, chars string) string {
	if chars == "" {
		return strings.TrimRightFunc(s, unicode.IsSpace)
	}
	return strings.TrimRight(s, chars)
}

// StrAt returns the rune at i, counting from the end when i is negative
// (String.prototype.at); ok is false out of range, where JavaScript gives
// undefined.
func StrAt(s string, i int) (string, bool) {
	r := []rune(s)
	if i < 0 {
		i += len(r)
	}
	if i < 0 || i >= len(r) {
		return "", false
	}
	return string(r[i]), true
}

// CharAt returns the rune at i, or "" out of range; unlike At, a negative i
// is out of range (String.prototype.charAt).
func CharAt(s string, i int) string {
	if i < 0 {
		return ""
	}
	c, _ := StrAt(s, i)
	return c
}

// CodePointAt returns the code point at i (String.prototype.codePointAt);
// ok is false out of range.
func CodePointAt(s string, i int) (rune, bool) {
	r := []rune(s)
	if i < 0 || i >= len(r) {
		return 0, false
	}
	return r[i], true
}

// StrSlice returns the runes from start up to end like String.prototype.slice:
// negative positions count from the end, positions are clamped, and an end
// before start gives "".
func StrSlice(s string, start, end int) string {
	r := []rune(s)
	n := len(r)
	rel := func(i int) int {
		if i < 0 {
			return max(n+i, 0)
		}
		return min(i, n)
	}
	a, b := rel(start), rel(end)
	if a >= b {
		return ""
	}
	return string(r[a:b])
}

// Substring returns the runes between start and end like
// String.prototype.substring: negatives count as 0, positions are clamped,
// and start and end are swapped when start is larger.
func Substring(s string, start, end int) string {
	r := []rune(s)
	a, b := clampIndex(start, len(r)), clampIndex(end, len(r))
	if a > b {
		a, b = b, a
	}
	return string(r[a:b])
}

// Split splits s at sep like String.prototype.split: an empty sep splits into
// runes, and limit >= 0 keeps at most limit pieces (dropping the rest, where
// Go's strings.SplitN keeps the remainder in the last piece); limit < 0 keeps
// all.
func Split(s, sep string, limit int) []string {
	var parts []string
	if sep == "" {
		for _, c := range s {
			parts = append(parts, string(c))
		}
	} else {
		parts = strings.Split(s, sep)
	}
	if limit >= 0 && limit < len(parts) {
		parts = parts[:limit]
	}
	return parts
}

// StartsWith reports whether s starts with prefix at rune position pos
// (String.prototype.startsWith).
func StartsWith(s, prefix string, pos int) bool {
	return strings.HasPrefix(StrSlice(s, max(pos, 0), len([]rune(s))), prefix)
}

// EndsWith reports whether the first end runes of s end with suffix
// (String.prototype.endsWith); end < 0 means the whole string.
func EndsWith(s, suffix string, end int) bool {
	if end < 0 {
		end = len([]rune(s))
	}
	return strings.HasSuffix(StrSlice(s, 0, end), suffix)
}
