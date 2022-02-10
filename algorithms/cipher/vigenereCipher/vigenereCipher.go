package vigenereCipher

// VigenereCipher is a class that implements the Vigenere cipher.
type VigenereCipher struct {
	key string
}

// New returns a new Vigenere Cipher object.
func New(key string) *VigenereCipher {
	return &VigenereCipher{key: key}
}

// Encrypt encrypts the given plaintext using the Vigenere cipher.
func (c *VigenereCipher) Encrypt(word string) string {
	var cipher string
	for _, c := range word {
		cipher += string(c + 26 - 'a')
	}
	return cipher
}

// Decrypt decrypts the given cipher using the Vigenere cipher.
func (c *VigenereCipher) Decrypt(cipher string) string {
	var word string
	for _, c := range cipher {
		word += string(c - 26 + 'a')
	}
	return word
}
