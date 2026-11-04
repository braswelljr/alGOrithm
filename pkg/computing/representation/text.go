package representation

import (
	"fmt"
	"unicode/utf8"
)

// Char is one character of a text as UTF-8: its code point, its bytes, and
// the bit pattern of the first byte that says how many bytes follow (0xxxxxxx
// for one, 110xxxxx 10xxxxxx for two, 1110xxxx for three, 11110xxx for four).
type Char struct {
	Rune      rune
	CodePoint string // "U+00E9"
	Bytes     []byte
	Pattern   string
}

// UTF8 encodes text as UTF-8 (Ken Thompson and Rob Pike, 1992), the encoding
// of the web: ASCII characters stay one byte, so ASCII text is already UTF-8,
// and every other code point takes 2 to 4 bytes whose leading bits say which
// byte is which, so a reader can resynchronize anywhere.
//
// Time: O(n). Space: O(n).
func UTF8(text string) ([]Char, error) {
	if !utf8.ValidString(text) {
		return nil, fmt.Errorf("%w: the text is not valid UTF-8", ErrInput)
	}
	var out []Char
	for _, r := range text {
		b := []byte(string(r))
		pattern := map[int]string{1: "0xxxxxxx", 2: "110xxxxx 10xxxxxx", 3: "1110xxxx 10xxxxxx 10xxxxxx", 4: "11110xxx 10xxxxxx 10xxxxxx 10xxxxxx"}[len(b)]
		out = append(out, Char{Rune: r, CodePoint: fmt.Sprintf("U+%04X", r), Bytes: b, Pattern: pattern})
	}
	return out, nil
}

// Endian returns the bytes of a 32-bit value in memory order, big-endian
// (most significant byte first: network order, and how people write numbers)
// and little-endian (least significant first: x86 and most ARM systems).
//
// Time: O(1). Space: O(1).
func Endian(v uint32) (big, little [4]byte) {
	for i := 0; i < 4; i++ {
		big[i] = byte(v >> (24 - 8*i))
		little[i] = byte(v >> (8 * i))
	}
	return big, little
}
