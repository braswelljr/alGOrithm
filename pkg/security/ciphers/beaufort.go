package ciphers

import "github.com/braswelljr/alGOrithm/pkg/trace"

// Beaufort encrypts each letter as (key - plain) mod 26 with a repeating
// keyword (Francis Beaufort, 1857). Unlike Vigenère it is its own inverse:
// Beaufort(Beaufort(s, k), k) == s. A keyword with no letters returns
// ErrBadKey.
//
// Time: O(n). Space: O(n).
func Beaufort(text, keyword string) (string, error) { return TraceBeaufort(text, keyword, trace.Nop{}) }

// TraceBeaufort is Beaufort, reporting each letter written.
func TraceBeaufort(text, keyword string, r trace.Recorder) (string, error) {
	key := keyShifts(keyword)
	if len(key) == 0 {
		return "", ErrBadKey
	}
	return mapLetters(text, r, func(idx, n int) int { return key[n%len(key)] - idx }), nil
}
