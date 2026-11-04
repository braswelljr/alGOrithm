package cryptography

import "encoding/binary"

// HMACSHA256 returns the HMAC-SHA256 of msg under key (RFC 2104):
// H((key ^ opad) || H((key ^ ipad) || msg)), with a key longer than a block
// hashed first. The two nested hashes are what stop the length-extension
// attack that breaks H(key || msg). For learning: use crypto/hmac in real
// code.
//
// Time: O(len(key) + len(msg)). Space: O(len(msg)).
func HMACSHA256(key, msg []byte) [32]byte {
	if len(key) > 64 {
		k := SHA256(key)
		key = k[:]
	}
	var ipad, opad [64]byte
	copy(ipad[:], key)
	copy(opad[:], key)
	for i := range ipad {
		ipad[i] ^= 0x36
		opad[i] ^= 0x5c
	}
	inner := SHA256(append(ipad[:], msg...))
	return SHA256(append(opad[:], inner[:]...))
}

// PBKDF2 derives keyLen bytes from a password (RFC 8018 with HMAC-SHA256):
// block i is U1 ^ U2 ^ ... ^ Uc, where U1 = HMAC(password, salt || i) and
// each next U is the HMAC of the one before. The iteration count c makes each
// guess cost c HMACs, which slows down anyone trying passwords from a list.
// For learning: use crypto/pbkdf2 (Go 1.24) or golang.org/x/crypto/pbkdf2,
// or better argon2, in real code.
//
// Time: O(iterations * keyLen). Space: O(keyLen).
func PBKDF2(password, salt []byte, iterations, keyLen int) []byte {
	var out []byte
	for block := uint32(1); len(out) < keyLen; block++ {
		u := HMACSHA256(password, binary.BigEndian.AppendUint32(append([]byte(nil), salt...), block))
		t := u
		for i := 1; i < iterations; i++ {
			u = HMACSHA256(password, u[:])
			for j := range t {
				t[j] ^= u[j]
			}
		}
		out = append(out, t[:]...)
	}
	return out[:keyLen]
}
