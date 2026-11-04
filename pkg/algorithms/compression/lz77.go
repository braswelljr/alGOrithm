package compression

// Token is one step of LZ77: copy Length bytes starting Offset bytes back in
// the output, then append Next. A token with Length 0 is just a literal.
type Token struct {
	Offset, Length int
	Next           byte
}

// LZ77 encodes data with LZ77 (Abraham Lempel and Jacob Ziv, 1977): at each
// position, find the longest match for the bytes ahead within the last window
// bytes already seen, and emit (offset, length, next byte). A match may run
// into the bytes it is copying, so "aaaa" is a literal then one token that
// copies from one back. It is the idea under gzip, PNG and zip (DEFLATE adds
// Huffman coding). window below 1 means 4096. This version searches the
// window directly; real encoders use hash chains.
//
// Time: O(n * window * match length). Space: O(tokens).
func LZ77(data []byte, window int) []Token {
	if window < 1 {
		window = 4096
	}
	var tokens []Token
	for i := 0; i < len(data); {
		best, bestLen := 0, 0
		limit := len(data) - i - 1 // keep one byte for Next
		for start := max(0, i-window); start < i; start++ {
			n := 0
			for n < limit && data[start+n] == data[i+n] {
				n++
			}
			if n > bestLen {
				best, bestLen = i-start, n
			}
		}
		tokens = append(tokens, Token{best, bestLen, data[i+bestLen]})
		i += bestLen + 1
	}
	return tokens
}

// LZ77Decode rebuilds the bytes from tokens, copying byte by byte so that
// overlapping matches work, or returns ErrCorrupt when a token reaches back
// before the start.
//
// Time: O(n). Space: O(n).
func LZ77Decode(tokens []Token) ([]byte, error) {
	var out []byte
	for _, t := range tokens {
		if t.Length < 0 || t.Offset < 0 || (t.Length > 0 && (t.Offset == 0 || t.Offset > len(out))) {
			return nil, ErrCorrupt
		}
		start := len(out) - t.Offset
		for k := 0; k < t.Length; k++ {
			out = append(out, out[start+k])
		}
		out = append(out, t.Next)
	}
	return out, nil
}
