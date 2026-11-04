// Package ciphers holds classical ciphers, one file each: Caesar, ROT13,
// Vigenère and Hill. They teach how encryption works and how it is broken;
// none of them is secure, so never use one to protect real data (use
// crypto/aes or golang.org/x/crypto instead).
//
// Every cipher keeps letters' case where it can, leaves other characters
// alone, and decrypts what it encrypts:
//
//	secret := ciphers.Caesar("Attack at dawn", 3) // "Dwwdfn dw gdzq"
//	plain := ciphers.Caesar(secret, -3)
package ciphers

// shiftLetter shifts a letter by k places (any int), keeping its case; any
// other byte is returned unchanged.
func shiftLetter(c byte, k int) byte {
	k = ((k % 26) + 26) % 26
	switch {
	case c >= 'a' && c <= 'z':
		return 'a' + byte((int(c-'a')+k)%26)
	case c >= 'A' && c <= 'Z':
		return 'A' + byte((int(c-'A')+k)%26)
	}
	return c
}

func isLetter(c byte) bool { return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') }
