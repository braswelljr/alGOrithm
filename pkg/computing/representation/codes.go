package representation

import (
	"fmt"
	"math/bits"
	"strconv"
)

// BCD writes n >= 0 in binary-coded decimal: each decimal digit as its own
// 4 bits, the way calculators and clocks store numbers so they can show them
// without dividing. It returns one 4-bit group per digit.
//
// Time: O(digits). Space: O(digits).
func BCD(n int) ([]string, error) {
	if n < 0 {
		return nil, fmt.Errorf("%w: a number of 0 or more", ErrInput)
	}
	var out []string
	for _, d := range strconv.Itoa(n) {
		out = append(out, Binary(uint64(d-'0'), 4))
	}
	return out, nil
}

// Parity returns the bit that makes the number of ones in v (its low n bits)
// even, the simplest error check: it detects any single flipped bit, and
// misses any two.
//
// Time: O(1). Space: O(1).
func Parity(v uint64, n int) int {
	return bits.OnesCount64(v&(uint64(1)<<n-1)) % 2
}

// Hamming74 encodes 4 data bits d (d1 is bit 3 of d, the first) as 7 bits
// p1 p2 d1 p3 d2 d3 d4 (Richard Hamming, 1950): each parity bit covers the
// positions whose number has its bit set (p1 covers 1, 3, 5, 7; p2 covers
// 2, 3, 6, 7; p3 covers 4, 5, 6, 7), so the parity checks that fail spell the
// position of a single flipped bit. Bit 6 of the result is position 1.
//
// Time: O(1). Space: O(1).
func Hamming74(d int) int {
	d1, d2, d3, d4 := d>>3&1, d>>2&1, d>>1&1, d&1
	p1 := d1 ^ d2 ^ d4
	p2 := d1 ^ d3 ^ d4
	p3 := d2 ^ d3 ^ d4
	pos := []int{p1, p2, d1, p3, d2, d3, d4}
	v := 0
	for _, b := range pos {
		v = v<<1 | b
	}
	return v
}

// HammingDecode checks a 7-bit Hamming(7,4) codeword: it recomputes the three
// parity checks, reads the failing ones as the position of the error (the
// syndrome, 0 for none), flips that bit back, and returns the corrected data
// bits, the syndrome and the corrected codeword. Two flipped bits are
// corrected wrongly; that needs an extra parity bit (SECDED).
//
// Time: O(1). Space: O(1).
func HammingDecode(code int) (data, syndrome, corrected int) {
	bit := func(c, pos int) int { return c >> (7 - pos) & 1 } // pos 1..7
	s1 := bit(code, 1) ^ bit(code, 3) ^ bit(code, 5) ^ bit(code, 7)
	s2 := bit(code, 2) ^ bit(code, 3) ^ bit(code, 6) ^ bit(code, 7)
	s3 := bit(code, 4) ^ bit(code, 5) ^ bit(code, 6) ^ bit(code, 7)
	syndrome = s3<<2 | s2<<1 | s1
	corrected = code & 0x7f
	if syndrome != 0 {
		corrected ^= 1 << (7 - syndrome)
	}
	data = bit(corrected, 3)<<3 | bit(corrected, 5)<<2 | bit(corrected, 6)<<1 | bit(corrected, 7)
	return data, syndrome, corrected
}
