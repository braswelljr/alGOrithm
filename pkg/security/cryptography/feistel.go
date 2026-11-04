package cryptography

import "math/bits"

// FeistelRound is the two 16-bit halves after one round.
type FeistelRound struct {
	Round int    `json:"round"`
	Left  uint16 `json:"left"`
	Right uint16 `json:"right"`
}

// feistelF is the toy round function: any function works, invertible or
// not, which is the point of the Feistel design.
func feistelF(r, k uint16) uint16 { return bits.RotateLeft16(r*0x9e37^k, 5) + k }

// Feistel encrypts a 32-bit block with a toy Feistel network, the structure
// of DES and Blowfish: split the block into halves L and R, and each round
// sets L, R = R, L ^ F(R, key). Decryption is the same network with the
// round keys reversed, whatever F is. A toy with 16-bit halves and a weak F:
// for learning only; use crypto/aes in real code.
//
// Time: O(rounds). Space: O(rounds) for the trace.
func Feistel(block uint32, keys []uint16) (uint32, []FeistelRound) {
	return feistel(block, keys, false)
}

// FeistelDecrypt undoes Feistel by running the rounds with the keys reversed.
func FeistelDecrypt(block uint32, keys []uint16) (uint32, []FeistelRound) {
	return feistel(block, keys, true)
}

func feistel(block uint32, keys []uint16, reverse bool) (uint32, []FeistelRound) {
	l, r := uint16(block>>16), uint16(block)
	rounds := make([]FeistelRound, 0, len(keys))
	for i := range keys {
		k := keys[i]
		if reverse {
			k = keys[len(keys)-1-i]
		}
		l, r = r, l^feistelF(r, k)
		rounds = append(rounds, FeistelRound{Round: i + 1, Left: l, Right: r})
	}
	// Undo the last swap, so decrypting is the same network run backwards.
	return uint32(r)<<16 | uint32(l), rounds
}
