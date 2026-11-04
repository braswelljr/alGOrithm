package cryptography

import (
	"encoding/binary"
	"math/big"
	"math/bits"
)

// sha256K and sha256Init are the first 32 bits of the fractional parts of
// the cube roots of the first 64 primes and of the square roots of the first
// 8: "nothing up my sleeve" numbers anyone can check.
var sha256K, sha256Init = sha256Constants()

func sha256Constants() (k [64]uint32, h [8]uint32) {
	primes := firstPrimes(64)
	for i, p := range primes {
		k[i] = uint32(rootFraction(p, 3))
		if i < 8 {
			h[i] = uint32(rootFraction(p, 2))
		}
	}
	return k, h
}

func firstPrimes(n int) []uint64 {
	var ps []uint64
	for c := uint64(2); len(ps) < n; c++ {
		prime := true
		for _, p := range ps {
			if p*p > c {
				break
			}
			if c%p == 0 {
				prime = false
				break
			}
		}
		if prime {
			ps = append(ps, c)
		}
	}
	return ps
}

// rootFraction returns floor(p^(1/r) * 2^32) mod 2^32, the first 32 bits of
// the fraction of p's r-th root, in whole numbers so no rounding creeps in:
// the largest x with x^r <= p * 2^(32r), by binary search.
func rootFraction(p uint64, r int) uint64 {
	target := new(big.Int).Lsh(new(big.Int).SetUint64(p), uint(32*r))
	lo, hi := uint64(0), uint64(1)<<40
	x, pow := new(big.Int), new(big.Int)
	for lo < hi {
		mid := lo + (hi-lo+1)/2
		pow.Exp(x.SetUint64(mid), big.NewInt(int64(r)), nil)
		if pow.Cmp(target) <= 0 {
			lo = mid
		} else {
			hi = mid - 1
		}
	}
	return lo & 0xffffffff
}

// pad appends the Merkle-Damgård padding shared by MD5 and SHA-256: a 1 bit,
// zeros up to 56 bytes into a 64-byte block, then the length in bits.
func pad(msg []byte, bigEndian bool) []byte {
	n := len(msg)
	out := append(append(make([]byte, 0, n+72), msg...), 0x80)
	for len(out)%64 != 56 {
		out = append(out, 0)
	}
	if bigEndian {
		return binary.BigEndian.AppendUint64(out, uint64(n)*8)
	}
	return binary.LittleEndian.AppendUint64(out, uint64(n)*8)
}

// SHA256 returns the SHA-256 digest of msg (FIPS 180-4): msg is padded to
// 64-byte blocks, and each block goes through 64 rounds that mix eight 32-bit
// words a..h with the block's message schedule. For learning: use
// crypto/sha256 in real code.
//
// Time: O(n). Space: O(n) for the padded copy.
func SHA256(msg []byte) [32]byte {
	h := sha256Init
	p := pad(msg, true)
	for i := 0; i < len(p); i += 64 {
		sha256Block(&h, p[i:i+64], nil)
	}
	var out [32]byte
	for i, v := range h {
		binary.BigEndian.PutUint32(out[4*i:], v)
	}
	return out
}

// SHA256Rounds returns the working words a..h after each of the 64 rounds of
// msg's first block, to watch the avalanche: one changed input bit changes
// about half the bits within a few rounds. For learning: use crypto/sha256 in
// real code.
func SHA256Rounds(msg []byte) [64][8]uint32 {
	var rounds [64][8]uint32
	h := sha256Init
	sha256Block(&h, pad(msg, true)[:64], &rounds)
	return rounds
}

func sha256Block(h *[8]uint32, block []byte, rounds *[64][8]uint32) {
	var w [64]uint32
	for i := 0; i < 16; i++ {
		w[i] = binary.BigEndian.Uint32(block[4*i:])
	}
	for i := 16; i < 64; i++ {
		s0 := bits.RotateLeft32(w[i-15], -7) ^ bits.RotateLeft32(w[i-15], -18) ^ w[i-15]>>3
		s1 := bits.RotateLeft32(w[i-2], -17) ^ bits.RotateLeft32(w[i-2], -19) ^ w[i-2]>>10
		w[i] = w[i-16] + s0 + w[i-7] + s1
	}
	a, b, c, d, e, f, g, hh := h[0], h[1], h[2], h[3], h[4], h[5], h[6], h[7]
	for i := 0; i < 64; i++ {
		s1 := bits.RotateLeft32(e, -6) ^ bits.RotateLeft32(e, -11) ^ bits.RotateLeft32(e, -25)
		ch := e&f ^ ^e&g
		t1 := hh + s1 + ch + sha256K[i] + w[i]
		s0 := bits.RotateLeft32(a, -2) ^ bits.RotateLeft32(a, -13) ^ bits.RotateLeft32(a, -22)
		maj := a&b ^ a&c ^ b&c
		hh, g, f, e, d, c, b, a = g, f, e, d+t1, c, b, a, t1+s0+maj
		if rounds != nil {
			rounds[i] = [8]uint32{a, b, c, d, e, f, g, hh}
		}
	}
	for i, v := range [8]uint32{a, b, c, d, e, f, g, hh} {
		h[i] += v
	}
}
