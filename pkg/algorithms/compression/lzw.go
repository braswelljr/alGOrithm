package compression

// LZW encodes data with LZW (Lempel, Ziv and Terry Welch, 1984), the
// compressor of GIF and Unix compress: the dictionary starts with the 256
// single bytes (codes 0-255); read the longest phrase already in it, emit
// its code, and add that phrase plus the next byte as a new code. Both sides
// grow the same dictionary, so it is never sent. added lists the phrases in
// the order they got codes 256, 257, ..., the dictionary's growth.
//
// Time: O(n) with a hash map. Space: O(n).
func LZW(data []byte) (codes []int, added []string) {
	dict := make(map[string]int, 256)
	for i := 0; i < 256; i++ {
		dict[string([]byte{byte(i)})] = i
	}
	phrase := ""
	for _, b := range data {
		next := phrase + string([]byte{b})
		if _, ok := dict[next]; ok {
			phrase = next
			continue
		}
		codes = append(codes, dict[phrase])
		dict[next] = 256 + len(added)
		added = append(added, next)
		phrase = string([]byte{b})
	}
	if phrase != "" {
		codes = append(codes, dict[phrase])
	}
	return codes, added
}

// LZWDecode rebuilds the bytes from codes, growing the same dictionary as
// LZW. The one tricky case is a code not yet in the dictionary: it can only
// be the previous phrase plus its own first byte. It returns ErrCorrupt for
// any other unknown code.
//
// Time: O(n). Space: O(n).
func LZWDecode(codes []int) ([]byte, error) {
	dict := make([]string, 256, 256+len(codes))
	for i := range dict {
		dict[i] = string([]byte{byte(i)})
	}
	var out []byte
	prev := ""
	for _, c := range codes {
		var phrase string
		switch {
		case c >= 0 && c < len(dict):
			phrase = dict[c]
		case c == len(dict) && prev != "":
			phrase = prev + prev[:1]
		default:
			return nil, ErrCorrupt
		}
		out = append(out, phrase...)
		if prev != "" {
			dict = append(dict, prev+phrase[:1])
		}
		prev = phrase
	}
	return out, nil
}
