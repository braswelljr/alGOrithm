package compression

// MoveToFront replaces each byte with its position in a list of all 256
// bytes, then moves it to the front. Recently seen bytes get small numbers,
// so the runs that BurrowsWheeler groups together become runs of zeros,
// which an entropy coder compresses well.
//
// Time: O(256 n) with this list. Space: O(n).
func MoveToFront(data []byte) []byte {
	list := alphabet()
	out := make([]byte, len(data))
	for i, b := range data {
		k := 0
		for list[k] != b {
			k++
		}
		out[i] = byte(k)
		copy(list[1:k+1], list[:k])
		list[0] = b
	}
	return out
}

// MoveToFrontDecode reverses MoveToFront, keeping the same list.
//
// Time: O(256 n). Space: O(n).
func MoveToFrontDecode(ranks []byte) []byte {
	list := alphabet()
	out := make([]byte, len(ranks))
	for i, k := range ranks {
		b := list[k]
		out[i] = b
		copy(list[1:int(k)+1], list[:k])
		list[0] = b
	}
	return out
}

func alphabet() []byte {
	list := make([]byte, 256)
	for i := range list {
		list[i] = byte(i)
	}
	return list
}
