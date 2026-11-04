package cryptography

// LeakyEqual compares a and b the ordinary way, stopping at the first byte
// that differs, and also returns how many bytes it looked at. That count is
// the leak: an attacker timing a server's check of a secret token learns how
// many leading bytes of a guess are right and can find the token one byte at
// a time, in 256 * n tries instead of 256^n.
func LeakyEqual(a, b []byte) (equal bool, compared int) {
	if len(a) != len(b) {
		return false, 0
	}
	for i := range a {
		if a[i] != b[i] {
			return false, i + 1
		}
	}
	return true, len(a)
}

// ConstantTimeEqual reports whether a and b are equal, always looking at
// every byte: the differences are ORed together and checked once at the end,
// so the time depends only on the length, never on where they differ. For
// learning: use crypto/subtle.ConstantTimeCompare or crypto/hmac.Equal in
// real code, which also keep the compiler from adding a shortcut.
//
// Time: O(n). Space: O(1).
func ConstantTimeEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	var diff byte
	for i := range a {
		diff |= a[i] ^ b[i]
	}
	return diff == 0
}
