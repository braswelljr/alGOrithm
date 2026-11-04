package cryptography

import (
	"encoding/binary"
	"math/bits"
)

// QuarterRound is ChaCha20's core: add, XOR and rotate (ARX) four words, with
// no table lookups, so it runs in constant time on any CPU. For learning: use
// golang.org/x/crypto/chacha20poly1305 in real code.
func QuarterRound(a, b, c, d uint32) (uint32, uint32, uint32, uint32) {
	a += b
	d = bits.RotateLeft32(d^a, 16)
	c += d
	b = bits.RotateLeft32(b^c, 12)
	a += b
	d = bits.RotateLeft32(d^a, 8)
	c += d
	b = bits.RotateLeft32(b^c, 7)
	return a, b, c, d
}

// ChaCha20Rounds returns the 4x4 state of a ChaCha20 block before the rounds
// and after each of the ten double rounds (a column round, then a diagonal
// round), before the input is added back. For learning: use
// golang.org/x/crypto/chacha20 in real code.
func ChaCha20Rounds(key [32]byte, counter uint32, nonce [12]byte) [11][16]uint32 {
	var out [11][16]uint32
	s := [16]uint32{0x61707865, 0x3320646e, 0x79622d32, 0x6b206574} // "expand 32-byte k"
	for i := 0; i < 8; i++ {
		s[4+i] = binary.LittleEndian.Uint32(key[4*i:])
	}
	s[12] = counter
	for i := 0; i < 3; i++ {
		s[13+i] = binary.LittleEndian.Uint32(nonce[4*i:])
	}
	out[0] = s
	qr := func(a, b, c, d int) { s[a], s[b], s[c], s[d] = QuarterRound(s[a], s[b], s[c], s[d]) }
	for i := 1; i <= 10; i++ {
		qr(0, 4, 8, 12)
		qr(1, 5, 9, 13)
		qr(2, 6, 10, 14)
		qr(3, 7, 11, 15)
		qr(0, 5, 10, 15)
		qr(1, 6, 11, 12)
		qr(2, 7, 8, 13)
		qr(3, 4, 9, 14)
		out[i] = s
	}
	return out
}

// ChaCha20Block returns 64 bytes of keystream (RFC 8439): the state after 20
// rounds plus the starting state, so the rounds cannot be run backwards. For
// learning: use golang.org/x/crypto/chacha20 in real code.
func ChaCha20Block(key [32]byte, counter uint32, nonce [12]byte) [64]byte {
	r := ChaCha20Rounds(key, counter, nonce)
	var out [64]byte
	for i := range r[10] {
		binary.LittleEndian.PutUint32(out[4*i:], r[10][i]+r[0][i])
	}
	return out
}

// ChaCha20 encrypts or decrypts msg (the same XOR both ways) with the
// keystream starting at block counter. A nonce must never repeat under one
// key: two messages under the same keystream XOR to the XOR of their
// plaintexts. For learning: use golang.org/x/crypto/chacha20poly1305 in real
// code, which also authenticates.
//
// Time: O(n). Space: O(n).
func ChaCha20(key [32]byte, nonce [12]byte, counter uint32, msg []byte) []byte {
	out := make([]byte, len(msg))
	for off := 0; off < len(msg); off += 64 {
		ks := ChaCha20Block(key, counter, nonce)
		counter++
		for i := off; i < len(msg) && i < off+64; i++ {
			out[i] = msg[i] ^ ks[i-off]
		}
	}
	return out
}
