package cryptography

import "math/bits"

// gmul multiplies in GF(2^8) modulo x^8 + x^4 + x^3 + x + 1, AES's field.
func gmul(a, b byte) byte {
	var p byte
	for b != 0 {
		if b&1 != 0 {
			p ^= a
		}
		carry := a & 0x80
		a <<= 1
		if carry != 0 {
			a ^= 0x1b
		}
		b >>= 1
	}
	return p
}

// aesSbox and aesInvSbox are computed, not copied: each byte's inverse in
// GF(2^8) (a^254, 0 for 0) run through an affine map that removes any
// algebraic shortcut.
var aesSbox, aesInvSbox = func() (s, inv [256]byte) {
	for x := 0; x < 256; x++ {
		b := byte(x)
		y := byte(1)
		for e := 0; e < 254; e++ {
			y = gmul(y, b)
		}
		if b == 0 {
			y = 0
		}
		v := y ^ bits.RotateLeft8(y, 1) ^ bits.RotateLeft8(y, 2) ^ bits.RotateLeft8(y, 3) ^ bits.RotateLeft8(y, 4) ^ 0x63
		s[x], inv[v] = v, b
	}
	return s, inv
}()

// AESState is the 4x4 grid of bytes AES works on, State[row][column]; a
// block fills it column by column.
type AESState [4][4]byte

func toState(b [16]byte) AESState {
	var s AESState
	for i, v := range b {
		s[i%4][i/4] = v
	}
	return s
}

func (s AESState) bytes() [16]byte {
	var b [16]byte
	for i := range b {
		b[i] = s[i%4][i/4]
	}
	return b
}

// AESStep is the state after one step of one round.
type AESStep struct {
	Round int      `json:"round"`
	Step  string   `json:"step"` // "AddRoundKey", "SubBytes", "ShiftRows", "MixColumns"
	State AESState `json:"state"`
}

// AES128KeyExpansion stretches a 16-byte key into the 11 round keys: each
// new word is the word before XOR the word four back, and every fourth is
// first rotated, run through the S-box and XORed with a round constant.
// For learning: use crypto/aes in real code.
func AES128KeyExpansion(key [16]byte) [11][16]byte {
	var w [44][4]byte
	for i := 0; i < 4; i++ {
		copy(w[i][:], key[4*i:])
	}
	rcon := byte(1)
	for i := 4; i < 44; i++ {
		t := w[i-1]
		if i%4 == 0 {
			t = [4]byte{aesSbox[t[1]] ^ rcon, aesSbox[t[2]], aesSbox[t[3]], aesSbox[t[0]]}
			rcon = gmul(rcon, 2)
		}
		for j := range t {
			w[i][j] = w[i-4][j] ^ t[j]
		}
	}
	var rk [11][16]byte
	for i := range w {
		copy(rk[i/4][4*(i%4):], w[i][:])
	}
	return rk
}

// AES128 encrypts one 16-byte block (FIPS-197): AddRoundKey, then nine
// rounds of SubBytes (confusion), ShiftRows and MixColumns (diffusion) and
// AddRoundKey, and a last round without MixColumns. For learning: use
// crypto/aes with crypto/cipher.NewGCM in real code; a lone block cipher is
// not a way to encrypt messages, and this one's table lookups leak timing.
//
// Time: O(1) per block. Space: O(1).
func AES128(key, block [16]byte) [16]byte {
	out, _ := TraceAES128(key, block)
	return out
}

// TraceAES128 is AES128, also returning the state after every step.
func TraceAES128(key, block [16]byte) ([16]byte, []AESStep) {
	rk := AES128KeyExpansion(key)
	s := toState(block)
	steps := make([]AESStep, 0, 41)
	note := func(round int, step string) { steps = append(steps, AESStep{round, step, s}) }
	s = addRoundKey(s, rk[0])
	note(0, "AddRoundKey")
	for r := 1; r <= 10; r++ {
		s = subBytes(s, &aesSbox)
		note(r, "SubBytes")
		s = shiftRows(s, 1)
		note(r, "ShiftRows")
		if r < 10 {
			s = mixColumns(s, [4]byte{2, 3, 1, 1})
			note(r, "MixColumns")
		}
		s = addRoundKey(s, rk[r])
		note(r, "AddRoundKey")
	}
	return s.bytes(), steps
}

// AES128Decrypt undoes AES128 by running the inverse steps in reverse order.
// For learning: use crypto/aes in real code.
func AES128Decrypt(key, block [16]byte) [16]byte {
	rk := AES128KeyExpansion(key)
	s := addRoundKey(toState(block), rk[10])
	for r := 9; r >= 0; r-- {
		s = subBytes(shiftRows(s, 3), &aesInvSbox)
		s = addRoundKey(s, rk[r])
		if r > 0 {
			s = mixColumns(s, [4]byte{14, 11, 13, 9})
		}
	}
	return s.bytes()
}

func addRoundKey(s AESState, k [16]byte) AESState {
	ks := toState(k)
	for r := range s {
		for c := range s[r] {
			s[r][c] ^= ks[r][c]
		}
	}
	return s
}

func subBytes(s AESState, box *[256]byte) AESState {
	for r := range s {
		for c := range s[r] {
			s[r][c] = box[s[r][c]]
		}
	}
	return s
}

// shiftRows rotates row r left by r*by places (by 3 undoes by 1).
func shiftRows(s AESState, by int) AESState {
	var out AESState
	for r := range s {
		for c := range s[r] {
			out[r][c] = s[r][(c+r*by)%4]
		}
	}
	return out
}

// mixColumns multiplies each column by the circulant matrix whose first row
// is m: {2, 3, 1, 1} encrypts, {14, 11, 13, 9} is its inverse.
func mixColumns(s AESState, m [4]byte) AESState {
	var out AESState
	for c := 0; c < 4; c++ {
		for r := 0; r < 4; r++ {
			var v byte
			for k := 0; k < 4; k++ {
				v ^= gmul(m[(k-r+4)%4], s[k][c])
			}
			out[r][c] = v
		}
	}
	return out
}
