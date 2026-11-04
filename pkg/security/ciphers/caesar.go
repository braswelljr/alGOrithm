package ciphers

// Caesar shifts every letter of text forward by shift places, wrapping Z to A
// ("A" becomes "D" for 3); a negative shift goes back, so Caesar(s, -k)
// decrypts Caesar(s, k). Case is kept; everything else is copied.
//
// Time: O(n). Space: O(n).
func Caesar(text string, shift int) string {
	out := []byte(text)
	for i := range out {
		out[i] = shiftLetter(out[i], shift)
	}
	return string(out)
}

// ROT13 is Caesar with a shift of 13: applying it twice gives the text back.
//
// Time: O(n). Space: O(n).
func ROT13(text string) string { return Caesar(text, 13) }
