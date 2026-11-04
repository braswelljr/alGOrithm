package text

import "strings"

// Soundex returns the American Soundex code of word: its first letter and
// three digits for the consonants that follow, so names that sound alike
// share a code (Robert and Rupert are both R163). Vowels separate repeated
// digits, H and W do not, and non-letters are ignored; "" when word has no
// letter.
//
// Time: O(n). Space: O(1).
func Soundex(word string) string {
	const codes = "01230120022455012623010202" // A to Z
	var b strings.Builder
	last := byte(0)
	for i := 0; i < len(word) && b.Len() < 4; i++ {
		c := word[i] | 0x20 // lower case for letters
		if c < 'a' || c > 'z' {
			continue
		}
		code := codes[c-'a']
		if b.Len() == 0 {
			b.WriteByte(c - 0x20)
			last = code
			continue
		}
		switch {
		case c == 'h' || c == 'w':
			// skip, keeping last: the same digit on both sides is written once
		case code == '0':
			last = 0
		case code != last:
			b.WriteByte(code)
			last = code
		}
	}
	if b.Len() == 0 {
		return ""
	}
	for b.Len() < 4 {
		b.WriteByte('0')
	}
	return b.String()
}
