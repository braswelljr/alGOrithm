package ciphers

import "github.com/braswelljr/alGOrithm/pkg/trace"

// Atbash mirrors the alphabet (A and Z swap, B and Y, ...), the Hebrew
// cipher from about 500 BC. It is its own inverse: Atbash(Atbash(s)) == s.
// Case is kept; other characters are copied.
//
// Time: O(n). Space: O(n).
func Atbash(text string) string { return TraceAtbash(text, trace.Nop{}) }

// TraceAtbash is Atbash, reporting each letter written (Set at its position,
// Values[0] the new byte).
func TraceAtbash(text string, r trace.Recorder) string {
	return mapLetters(text, r, func(idx, _ int) int { return 25 - idx })
}
