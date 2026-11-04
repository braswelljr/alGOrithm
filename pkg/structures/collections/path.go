package collections

import (
	"fmt"
	"strconv"
	"strings"
)

// PathPart is one step of a property path: a map key, or an index into an
// array.
type PathPart struct {
	Key     string
	Index   int
	IsIndex bool
}

// PathError is a path that could not be read, with the 1-based column of the
// problem.
type PathError struct {
	Path   string
	Column int
	Reason string
}

func (e *PathError) Error() string {
	return fmt.Sprintf("path %q, column %d: %s", e.Path, e.Column, e.Reason)
}

// ParsePath reads a lodash property path: keys separated by dots, array
// indexes in brackets, and any key in quoted brackets ("a.b[0].c",
// "a['weird.key'][\"x\"]"). The empty path is the value itself.
//
// Time: O(len(path)). Space: O(len(path)).
func ParsePath(path string) ([]PathPart, error) {
	var parts []PathPart
	fail := func(i int, why string) ([]PathPart, error) {
		return nil, &PathError{Path: path, Column: i + 1, Reason: why}
	}
	i, expectKey := 0, true
	for i < len(path) {
		switch path[i] {
		case '.':
			if expectKey {
				return fail(i, "empty key")
			}
			expectKey = true
			i++
		case '[':
			end := i + 1
			if end < len(path) && (path[end] == '\'' || path[end] == '"') {
				q := path[end]
				closing := strings.IndexByte(path[end+1:], q)
				if closing < 0 {
					return fail(end, "unclosed quote")
				}
				key := path[end+1 : end+1+closing]
				end = end + 1 + closing + 1
				if end >= len(path) || path[end] != ']' {
					return fail(end, "expected ] after the quoted key")
				}
				parts = append(parts, PathPart{Key: key})
				i = end + 1
			} else {
				closing := strings.IndexByte(path[i:], ']')
				if closing < 0 {
					return fail(i, "unclosed [")
				}
				text := path[i+1 : i+closing]
				n, err := strconv.Atoi(text)
				if err != nil || n < 0 {
					return fail(i+1, fmt.Sprintf("%q is not an array index (quote a key: ['%s'])", text, text))
				}
				parts = append(parts, PathPart{Index: n, IsIndex: true, Key: text})
				i += closing + 1
			}
			expectKey = false
		default:
			if !expectKey {
				return fail(i, "expected . or [ between keys")
			}
			end := i
			for end < len(path) && path[end] != '.' && path[end] != '[' {
				end++
			}
			parts = append(parts, PathPart{Key: path[i:end]})
			i, expectKey = end, false
		}
	}
	if expectKey && len(path) > 0 {
		return fail(len(path)-1, "the path ends with a dot")
	}
	return parts, nil
}

// joinPath writes parts back as a path ("a.b[0]").
func joinPath(prefix string, p PathPart) string {
	if p.IsIndex {
		return prefix + "[" + strconv.Itoa(p.Index) + "]"
	}
	if strings.ContainsAny(p.Key, ".[]'\"") || p.Key == "" {
		return prefix + "[" + strconv.Quote(p.Key) + "]"
	}
	if prefix == "" {
		return p.Key
	}
	return prefix + "." + p.Key
}

func keyPath(prefix, key string) string { return joinPath(prefix, PathPart{Key: key}) }

func indexPath(prefix string, i int) string {
	return joinPath(prefix, PathPart{Index: i, IsIndex: true})
}

// FormatPath writes parts as a path ParsePath reads back: dots between plain
// keys, brackets for indexes, and quoted brackets for keys that need them.
func FormatPath(parts []PathPart) string {
	out := ""
	for _, p := range parts {
		out = joinPath(out, p)
	}
	return out
}
