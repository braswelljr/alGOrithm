package cryptography

import (
	"fmt"
	"strings"
)

const (
	base64Alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	base32Alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"
	hexAlphabet    = "0123456789abcdef"
)

// encodeBits writes data in groups of bitsPer bits (6 for Base64, 5 for
// Base32, 4 for hex) as letters of alphabet, padding with '=' to a multiple
// of block characters (block 0: no padding).
func encodeBits(data []byte, alphabet string, bitsPer, block int) string {
	var b strings.Builder
	var acc, n uint
	for _, c := range data {
		acc, n = acc<<8|uint(c), n+8
		for n >= uint(bitsPer) {
			n -= uint(bitsPer)
			b.WriteByte(alphabet[acc>>n&(1<<bitsPer-1)])
		}
	}
	if n > 0 {
		b.WriteByte(alphabet[acc<<(uint(bitsPer)-n)&(1<<bitsPer-1)])
	}
	for block > 0 && b.Len()%block != 0 {
		b.WriteByte('=')
	}
	return b.String()
}

func decodeBits(s, alphabet string, bitsPer int) ([]byte, error) {
	s = strings.TrimRight(s, "=")
	var out []byte
	var acc, n uint
	for i := 0; i < len(s); i++ {
		v := strings.IndexByte(alphabet, s[i])
		if v < 0 {
			return nil, fmt.Errorf("%w: %q is not in the alphabet", ErrInvalid, s[i])
		}
		acc, n = acc<<uint(bitsPer)|uint(v), n+uint(bitsPer)
		if n >= 8 {
			n -= 8
			out = append(out, byte(acc>>n))
		}
	}
	if acc&(1<<n-1) != 0 {
		return nil, fmt.Errorf("%w: stray bits at the end", ErrInvalid)
	}
	return out, nil
}

// Base64 encodes data in RFC 4648 Base64: every 3 bytes become 4 of 64
// printable characters, padded with '='. An encoding, not encryption: anyone
// can decode it. For learning: use encoding/base64 in real code.
//
// Time: O(n). Space: O(n).
func Base64(data []byte) string { return encodeBits(data, base64Alphabet, 6, 4) }

// Base64Decode undoes Base64.
func Base64Decode(s string) ([]byte, error) { return decodeBits(s, base64Alphabet, 6) }

// Base32 encodes data in RFC 4648 Base32: every 5 bytes become 8 characters
// of A-Z and 2-7, case-insensitive and free of look-alikes like 0 and O. For
// learning: use encoding/base32 in real code.
func Base32(data []byte) string { return encodeBits(data, base32Alphabet, 5, 8) }

// Base32Decode undoes Base32.
func Base32Decode(s string) ([]byte, error) {
	return decodeBits(strings.ToUpper(s), base32Alphabet, 5)
}

// Hex encodes each byte as two lower-case hex digits. For learning: use
// encoding/hex in real code.
func Hex(data []byte) string { return encodeBits(data, hexAlphabet, 4, 0) }

// HexDecode undoes Hex, taking either case.
func HexDecode(s string) ([]byte, error) {
	if len(s)%2 != 0 {
		return nil, fmt.Errorf("%w: hex needs an even number of digits", ErrInvalid)
	}
	return decodeBits(strings.ToLower(s), hexAlphabet, 4)
}
