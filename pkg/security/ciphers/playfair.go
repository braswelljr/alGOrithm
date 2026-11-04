package ciphers

import "strings"

// PlayfairSquare is the 5x5 key square of the Playfair cipher: the keyword's
// letters without repeats, then the rest of the alphabet, with J merged into
// I.
type PlayfairSquare [5][5]byte

// NewPlayfairSquare builds the square for keyword.
func NewPlayfairSquare(keyword string) PlayfairSquare {
	var sq PlayfairSquare
	used := map[byte]bool{}
	i := 0
	for _, c := range []byte(lettersOnly(keyword) + "ABCDEFGHIKLMNOPQRSTUVWXYZ") {
		if c == 'J' {
			c = 'I'
		}
		if used[c] {
			continue
		}
		used[c] = true
		sq[i/5][i%5] = c
		i++
	}
	return sq
}

func (sq PlayfairSquare) find(c byte) (row, col int) {
	for r := range sq {
		for k := range sq[r] {
			if sq[r][k] == c {
				return r, k
			}
		}
	}
	return 0, 0
}

// Playfair encrypts text with the Playfair cipher (Charles Wheatstone,
// 1854): letters taken two at a time, upper-cased, J as I, X between a
// doubled pair and at the end of an odd one; a pair in the same row shifts
// right, in the same column shifts down, otherwise swaps columns (the
// rectangle rule). It returns the ciphertext and the square.
//
// Time: O(n). Space: O(n).
func Playfair(text, keyword string) (string, PlayfairSquare) {
	sq := NewPlayfairSquare(keyword)
	return sq.apply(playfairPairs(text), 1), sq
}

// PlayfairDecrypt reverses Playfair; the X padding stays in the result.
func PlayfairDecrypt(text, keyword string) string {
	sq := NewPlayfairSquare(keyword)
	return sq.apply(lettersOnly(strings.ReplaceAll(strings.ToUpper(text), "J", "I")), 4)
}

// playfairPairs prepares text: letters only, J as I, split into pairs with X
// filling doubled letters and an odd end.
func playfairPairs(text string) string {
	s := strings.ReplaceAll(lettersOnly(text), "J", "I")
	var b strings.Builder
	for i := 0; i < len(s); {
		a := s[i]
		if i+1 < len(s) && s[i+1] != a {
			b.WriteByte(a)
			b.WriteByte(s[i+1])
			i += 2
			continue
		}
		b.WriteByte(a)
		b.WriteByte('X')
		i++
	}
	return b.String()
}

// apply shifts each pair by shift (1 to encrypt, 4 = -1 mod 5 to decrypt).
func (sq PlayfairSquare) apply(pairs string, shift int) string {
	out := make([]byte, 0, len(pairs))
	for i := 0; i+1 < len(pairs); i += 2 {
		r1, c1 := sq.find(pairs[i])
		r2, c2 := sq.find(pairs[i+1])
		switch {
		case r1 == r2:
			out = append(out, sq[r1][(c1+shift)%5], sq[r2][(c2+shift)%5])
		case c1 == c2:
			out = append(out, sq[(r1+shift)%5][c1], sq[(r2+shift)%5][c2])
		default:
			out = append(out, sq[r1][c2], sq[r2][c1])
		}
	}
	return string(out)
}
