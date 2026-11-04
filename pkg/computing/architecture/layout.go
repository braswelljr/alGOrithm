package architecture

import (
	"fmt"
	"slices"
	"strings"
)

// Field is a struct field: its name and its Go type.
type Field struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// Placed is a field where the compiler puts it: its offset from the start,
// size and alignment, and the padding inserted before it.
type Placed struct {
	Field
	Offset int `json:"offset"`
	Size   int `json:"size"`
	Align  int `json:"align"`
	Pad    int `json:"pad"`
}

// Struct is a laid-out struct: its fields, its total size (a multiple of
// its alignment, which is its largest field's), and the padding in all.
type Struct struct {
	Fields  []Placed `json:"fields"`
	Size    int      `json:"size"`
	Align   int      `json:"align"`
	Padding int      `json:"padding"`
}

// sizes are size and alignment of Go types on a 64-bit machine.
var sizes = map[string][2]int{
	"bool": {1, 1}, "int8": {1, 1}, "uint8": {1, 1}, "byte": {1, 1},
	"int16": {2, 2}, "uint16": {2, 2},
	"int32": {4, 4}, "uint32": {4, 4}, "float32": {4, 4}, "rune": {4, 4},
	"int64": {8, 8}, "uint64": {8, 8}, "float64": {8, 8}, "int": {8, 8}, "uint": {8, 8}, "uintptr": {8, 8},
	"complex64": {8, 4}, "complex128": {16, 8},
	"string": {16, 8}, "slice": {24, 8}, "pointer": {8, 8}, "map": {8, 8}, "chan": {8, 8}, "func": {8, 8}, "interface": {16, 8},
}

func sizeOf(t string) ([2]int, bool) {
	switch {
	case strings.HasPrefix(t, "*"):
		t = "pointer"
	case strings.HasPrefix(t, "[]"):
		t = "slice"
	case strings.HasPrefix(t, "map["):
		t = "map"
	case t == "any" || strings.HasPrefix(t, "interface"):
		t = "interface"
	}
	s, ok := sizes[t]
	return s, ok
}

// ParseFields reads fields as "name type", separated by ";" or new lines
// ("a bool; b int64; c int16").
func ParseFields(text string) ([]Field, error) {
	var out []Field
	for _, part := range strings.FieldsFunc(text, func(r rune) bool { return r == ';' || r == '\n' }) {
		f := strings.Fields(part)
		if len(f) == 0 {
			continue
		}
		if len(f) != 2 {
			return nil, fmt.Errorf("%w: %q: want name type", ErrInput, strings.TrimSpace(part))
		}
		if _, ok := sizeOf(f[1]); !ok {
			return nil, fmt.Errorf("%w: unknown type %s", ErrInput, f[1])
		}
		out = append(out, Field{f[0], f[1]})
	}
	if len(out) == 0 || len(out) > 32 {
		return nil, fmt.Errorf("%w: 1 to 32 fields", ErrInput)
	}
	return out, nil
}

// Layout places fields in order the way Go (and C) compilers do on a 64-bit
// machine: each at the next offset that is a multiple of its alignment,
// padding the gap, and the whole struct padded to a multiple of its largest
// alignment, so that in an array every element's fields stay aligned.
// unsafe.Sizeof and unsafe.Offsetof report the same numbers.
//
// Time: O(n) for n fields.
func Layout(fields []Field) Struct {
	var s Struct
	s.Align = 1
	off := 0
	for _, f := range fields {
		sz, _ := sizeOf(f.Type)
		pad := (sz[1] - off%sz[1]) % sz[1]
		off += pad
		s.Fields = append(s.Fields, Placed{Field: f, Offset: off, Size: sz[0], Align: sz[1], Pad: pad})
		s.Padding += pad
		off += sz[0]
		s.Align = max(s.Align, sz[1])
	}
	tail := (s.Align - off%s.Align) % s.Align
	s.Size, s.Padding = off+tail, s.Padding+tail
	return s
}

// Packed orders the fields by alignment, largest first (keeping the given
// order among equals), which never needs more padding and often needs
// none.
//
// Time: O(n log n).
func Packed(fields []Field) []Field {
	out := slices.Clone(fields)
	slices.SortStableFunc(out, func(a, b Field) int {
		sa, _ := sizeOf(a.Type)
		sb, _ := sizeOf(b.Type)
		return sb[1] - sa[1]
	})
	return out
}
