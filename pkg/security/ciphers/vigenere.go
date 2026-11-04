package ciphers

// Vigenere encrypts text with the Vigenère cipher: each letter is shifted by
// the letter of key under it (A by 0, B by 1, ... Z by 25), the key repeating
// along the message. Only the key's letters count; other characters of the
// text are copied and do not use up a key letter. A key with no letters
// leaves the text unchanged.
//
// Time: O(n). Space: O(n).
func Vigenere(text, key string) string { return vigenere(text, key, 1) }

// VigenereDecrypt reverses Vigenere for the same key.
func VigenereDecrypt(text, key string) string { return vigenere(text, key, -1) }

func vigenere(text, key string, sign int) string {
	var shifts []int
	for i := 0; i < len(key); i++ {
		if c := key[i]; isLetter(c) {
			shifts = append(shifts, int((c|0x20)-'a'))
		}
	}
	if len(shifts) == 0 {
		return text
	}
	out := []byte(text)
	k := 0
	for i, c := range out {
		if isLetter(c) {
			out[i] = shiftLetter(c, sign*shifts[k%len(shifts)])
			k++
		}
	}
	return string(out)
}
