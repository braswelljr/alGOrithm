package hashing

// FNV1a32 is the 32-bit FNV-1a hash (Fowler, Noll and Vo): for each byte, XOR
// it in, then multiply by the FNV prime. Simple and well spread for short keys.
//
// Time: O(n). Space: O(1).
func FNV1a32(data []byte) uint32 {
	h := uint32(2166136261)
	for _, b := range data {
		h ^= uint32(b)
		h *= 16777619
	}
	return h
}

// FNV1a64 is the 64-bit FNV-1a hash.
//
// Time: O(n). Space: O(1).
func FNV1a64(data []byte) uint64 {
	h := uint64(14695981039346656037)
	for _, b := range data {
		h ^= uint64(b)
		h *= 1099511628211
	}
	return h
}

// DJB2 is Daniel J. Bernstein's string hash: h = h*33 + byte, starting from
// 5381.
//
// Time: O(n). Space: O(1).
func DJB2(data []byte) uint32 {
	h := uint32(5381)
	for _, b := range data {
		h = h*33 + uint32(b)
	}
	return h
}

// SDBM is the hash from the sdbm database library: h = byte + (h << 6) +
// (h << 16) - h.
//
// Time: O(n). Space: O(1).
func SDBM(data []byte) uint32 {
	var h uint32
	for _, b := range data {
		h = uint32(b) + (h << 6) + (h << 16) - h
	}
	return h
}

// Murmur3_32 is Austin Appleby's MurmurHash3 (x86, 32 bit) with a seed: mix
// four bytes at a time with multiplies and rotations, then the tail, then a
// final avalanche so every input bit affects every output bit.
//
// Time: O(n). Space: O(1).
func Murmur3_32(data []byte, seed uint32) uint32 {
	const c1, c2 = 0xcc9e2d51, 0x1b873593
	rotl := func(x uint32, r uint) uint32 { return x<<r | x>>(32-r) }
	h := seed
	n := len(data)
	for i := 0; i+4 <= n; i += 4 {
		k := uint32(data[i]) | uint32(data[i+1])<<8 | uint32(data[i+2])<<16 | uint32(data[i+3])<<24
		k *= c1
		k = rotl(k, 15)
		k *= c2
		h ^= k
		h = rotl(h, 13)
		h = h*5 + 0xe6546b64
	}
	var k uint32
	tail := data[n&^3:]
	switch len(tail) {
	case 3:
		k ^= uint32(tail[2]) << 16
		fallthrough
	case 2:
		k ^= uint32(tail[1]) << 8
		fallthrough
	case 1:
		k ^= uint32(tail[0])
		k *= c1
		k = rotl(k, 15)
		k *= c2
		h ^= k
	}
	h ^= uint32(n)
	h ^= h >> 16
	h *= 0x85ebca6b
	h ^= h >> 13
	h *= 0xc2b2ae35
	h ^= h >> 16
	return h
}

// crcTable is the CRC-32 (IEEE 802.3) lookup table: the remainder of each byte
// value, reflected polynomial 0xEDB88320.
var crcTable = func() [256]uint32 {
	var t [256]uint32
	for i := range t {
		c := uint32(i)
		for k := 0; k < 8; k++ {
			if c&1 == 1 {
				c = 0xEDB88320 ^ c>>1
			} else {
				c >>= 1
			}
		}
		t[i] = c
	}
	return t
}()

// CRC32 is the IEEE CRC-32 checksum (Ethernet, zip, PNG), one table lookup a
// byte: the remainder of the data as a polynomial divided by a fixed one. It
// detects accidental corruption, not tampering.
//
// Time: O(n). Space: O(1) beyond the 1 KiB table.
func CRC32(data []byte) uint32 {
	c := ^uint32(0)
	for _, b := range data {
		c = crcTable[byte(c)^b] ^ c>>8
	}
	return ^c
}

// Adler32 is Mark Adler's checksum (zlib): two running sums modulo 65521, the
// second of the first. Faster than CRC-32, weaker on short inputs.
//
// Time: O(n). Space: O(1).
func Adler32(data []byte) uint32 {
	const mod = 65521
	a, b := uint32(1), uint32(0)
	for _, x := range data {
		a = (a + uint32(x)) % mod
		b = (b + a) % mod
	}
	return b<<16 | a
}

// Polynomial is the polynomial hash sum(data[i] * base^(n-1-i)) mod mod, the
// hash a rolling window updates in O(1) (Rabin-Karp); base 31 and the prime
// 1e9+7 are the usual choices. mod must be at most 2^31.
//
// Time: O(n). Space: O(1).
func Polynomial(data []byte, base, mod uint64) uint64 {
	var h uint64
	for _, b := range data {
		h = (h*base + uint64(b)) % mod
	}
	return h
}
