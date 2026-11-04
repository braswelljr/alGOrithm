package ciphers

import "github.com/braswelljr/alGOrithm/pkg/trace"

// Substitution replaces each letter with the letter at the same place in a
// key alphabet of 26 distinct letters (A becomes key[0], B key[1], ...).
// There are 26! (about 4 * 10^26) keys, yet frequency analysis breaks it,
// because each letter always becomes the same letter. A key that is not a
// permutation of the alphabet returns ErrBadKey.
//
// Time: O(n). Space: O(n).
func Substitution(text, key string) (string, error) { return TraceSubstitution(text, key, trace.Nop{}) }

// TraceSubstitution is Substitution, reporting each letter written.
func TraceSubstitution(text, key string, r trace.Recorder) (string, error) {
	perm, _, err := substitutionKey(key)
	if err != nil {
		return "", err
	}
	return mapLetters(text, r, func(idx, _ int) int { return perm[idx] }), nil
}

// SubstitutionDecrypt reverses Substitution with the same key.
func SubstitutionDecrypt(text, key string) (string, error) {
	_, inv, err := substitutionKey(key)
	if err != nil {
		return "", err
	}
	return mapLetters(text, trace.Nop{}, func(idx, _ int) int { return inv[idx] }), nil
}

func substitutionKey(key string) (perm, inv [26]int, err error) {
	k := lettersOnly(key)
	if len(k) != 26 {
		return perm, inv, ErrBadKey
	}
	seen := [26]bool{}
	for i := 0; i < 26; i++ {
		c := int(k[i] - 'A')
		if seen[c] {
			return perm, inv, ErrBadKey
		}
		seen[c] = true
		perm[i], inv[c] = c, i
	}
	return perm, inv, nil
}
