package cryptography

import (
	"encoding/binary"
	"math"
	"math/bits"
)

// md5K[i] is floor(|sin(i+1)| * 2^32), and md5S the rotation of each round.
var md5K = func() (k [64]uint32) {
	for i := range k {
		k[i] = uint32(math.Floor(math.Abs(math.Sin(float64(i+1))) * (1 << 32)))
	}
	return k
}()

var md5S = [4][4]int{{7, 12, 17, 22}, {5, 9, 14, 20}, {4, 11, 16, 23}, {6, 10, 15, 21}}

// MD5 returns the MD5 digest of msg (RFC 1321): four rounds of 16 steps over
// four 32-bit words for each 64-byte block. MD5 is broken: collisions take
// seconds on a laptop (Wang and Yu, 2005), so it must not sign or protect
// anything. For learning: use crypto/sha256 in real code.
//
// Time: O(n). Space: O(n) for the padded copy.
func MD5(msg []byte) [16]byte {
	h := [4]uint32{0x67452301, 0xefcdab89, 0x98badcfe, 0x10325476}
	p := pad(msg, false)
	for off := 0; off < len(p); off += 64 {
		var m [16]uint32
		for i := range m {
			m[i] = binary.LittleEndian.Uint32(p[off+4*i:])
		}
		a, b, c, d := h[0], h[1], h[2], h[3]
		for i := 0; i < 64; i++ {
			var f uint32
			var g int
			switch i / 16 {
			case 0:
				f, g = b&c|^b&d, i
			case 1:
				f, g = d&b|^d&c, (5*i+1)%16
			case 2:
				f, g = b^c^d, (3*i+5)%16
			default:
				f, g = c^(b|^d), 7*i%16
			}
			f += a + md5K[i] + m[g]
			a, d, c, b = d, c, b, b+bits.RotateLeft32(f, md5S[i/16][i%4])
		}
		h[0], h[1], h[2], h[3] = h[0]+a, h[1]+b, h[2]+c, h[3]+d
	}
	var out [16]byte
	for i, v := range h {
		binary.LittleEndian.PutUint32(out[4*i:], v)
	}
	return out
}
