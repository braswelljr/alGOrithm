package ciphers

import "github.com/braswelljr/alGOrithm/pkg/trace"

// Affine encrypts each letter x as (a*x + b) mod 26. a must share no factor
// with 26 (1, 3, 5, 7, 9, 11, 15, 17, 19, 21, 23 or 25) so that the mapping
// can be undone; otherwise it returns ErrBadKey. Caesar is a = 1.
//
// Time: O(n). Space: O(n).
func Affine(text string, a, b int) (string, error) { return TraceAffine(text, a, b, trace.Nop{}) }

// TraceAffine is Affine, reporting each letter written.
func TraceAffine(text string, a, b int, r trace.Recorder) (string, error) {
	if inverse26(a) < 0 {
		return "", ErrBadKey
	}
	return mapLetters(text, r, func(idx, _ int) int { return a*idx + b }), nil
}

// AffineDecrypt reverses Affine: x = a^-1 * (y - b) mod 26.
func AffineDecrypt(text string, a, b int) (string, error) {
	inv := inverse26(a)
	if inv < 0 {
		return "", ErrBadKey
	}
	return mapLetters(text, trace.Nop{}, func(idx, _ int) int { return inv * (idx - b) }), nil
}

// inverse26 returns a's inverse modulo 26, or -1 when there is none.
func inverse26(a int) int {
	a = ((a % 26) + 26) % 26
	for i := 1; i < 26; i++ {
		if a*i%26 == 1 {
			return i
		}
	}
	return -1
}
