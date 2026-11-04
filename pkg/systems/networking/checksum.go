package networking

// Checksum is the Internet checksum of RFC 1071, used by IPv4, ICMP, UDP and
// TCP headers: add the data as 16-bit big-endian words in ones' complement
// arithmetic (a carry out of the top bit is added back at the bottom), and
// invert the sum. An odd last byte is padded with a zero. A receiver sums the
// data with the checksum included and gets 0xFFFF when nothing changed, so
// Checksum of the data followed by its checksum is 0. It catches every
// one-bit error but not two words swapped.
//
// Time: O(n). Space: O(1).
func Checksum(data []byte) uint16 {
	return ^uint16(onesSum(data))
}

// ChecksumSteps returns the running ones'-complement sum after each 16-bit
// word, for showing the folding of carries.
//
// Time: O(n). Space: O(n).
func ChecksumSteps(data []byte) (words, sums []uint16) {
	var sum uint32
	for i := 0; i < len(data); i += 2 {
		w := uint32(data[i]) << 8
		if i+1 < len(data) {
			w |= uint32(data[i+1])
		}
		sum += w
		sum = sum&0xFFFF + sum>>16 // fold the carry back in
		words, sums = append(words, uint16(w)), append(sums, uint16(sum))
	}
	return words, sums
}

func onesSum(data []byte) uint32 {
	var sum uint32
	for i := 0; i < len(data); i += 2 {
		w := uint32(data[i]) << 8
		if i+1 < len(data) {
			w |= uint32(data[i+1])
		}
		sum += w
	}
	for sum>>16 != 0 {
		sum = sum&0xFFFF + sum>>16
	}
	return sum
}
