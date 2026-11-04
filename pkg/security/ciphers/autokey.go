package ciphers

import "github.com/braswelljr/alGOrithm/pkg/trace"

// Autokey is Vigenère whose key, after the keyword, continues with the
// plaintext itself (Blaise de Vigenère's own design), so the key never
// repeats and Kasiski's method does not apply. Only letters use up the key;
// a keyword with no letters returns ErrBadKey.
//
// Time: O(n). Space: O(n).
func Autokey(text, keyword string) (string, error) { return TraceAutokey(text, keyword, trace.Nop{}) }

// TraceAutokey is Autokey, reporting each letter written.
func TraceAutokey(text, keyword string, r trace.Recorder) (string, error) {
	key := keyShifts(keyword)
	if len(key) == 0 {
		return "", ErrBadKey
	}
	// key[n] is the keyword's nth letter, then the plaintext's, appended as
	// each letter goes by.
	return mapLetters(text, r, func(idx, n int) int {
		k := key[n]
		key = append(key, idx)
		return idx + k
	}), nil
}

// AutokeyDecrypt reverses Autokey: each recovered letter extends the key.
func AutokeyDecrypt(text, keyword string) (string, error) {
	key := keyShifts(keyword)
	if len(key) == 0 {
		return "", ErrBadKey
	}
	return mapLetters(text, trace.Nop{}, func(idx, n int) int {
		p := ((idx-key[n])%26 + 26) % 26
		key = append(key, p)
		return p
	}), nil
}
