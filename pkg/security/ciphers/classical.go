package ciphers

import (
	"errors"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// ErrBadKey is returned for a key that cannot work: an affine a sharing a
// factor with 26, a substitution alphabet that is not a permutation, a
// keyword with no letters. (Hill reports its own ErrKey.)
var ErrBadKey = errors.New("ciphers: invalid key")

// letterIndex returns c's position in the alphabet (0 for A or a) and
// whether c is a letter, and base is 'A' or 'a' to rebuild the case.
func letterIndex(c byte) (idx int, base byte, ok bool) {
	switch {
	case c >= 'a' && c <= 'z':
		return int(c - 'a'), 'a', true
	case c >= 'A' && c <= 'Z':
		return int(c - 'A'), 'A', true
	}
	return 0, 0, false
}

// mapLetters applies f to every letter of text (its alphabet index and its
// position among the letters), keeping case and copying everything else, and
// reports each changed position as a Set event (Values[0] the new byte).
func mapLetters(text string, r trace.Recorder, f func(idx, n int) int) string {
	out := []byte(text)
	n := 0
	for i, c := range out {
		idx, base, ok := letterIndex(c)
		if !ok {
			continue
		}
		out[i] = base + byte(((f(idx, n)%26)+26)%26)
		r.Record(trace.Event{Kind: trace.Set, Indices: []int{i}, Values: []int{int(out[i])}})
		n++
	}
	return string(out)
}

// keyShifts returns the alphabet indexes of key's letters.
func keyShifts(key string) []int {
	var s []int
	for i := 0; i < len(key); i++ {
		if idx, _, ok := letterIndex(key[i]); ok {
			s = append(s, idx)
		}
	}
	return s
}

// lettersOnly returns text's letters upper-cased, everything else dropped.
func lettersOnly(text string) string {
	var b strings.Builder
	for i := 0; i < len(text); i++ {
		if idx, _, ok := letterIndex(text[i]); ok {
			b.WriteByte(byte('A' + idx))
		}
	}
	return b.String()
}
