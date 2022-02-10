package vigenereCipher

import "testing"

var (
	key    = "vigenere"
	cipher = New(key)
)

func TestVigenereCipher(t *testing.T) {
	word := "cryptography"
	ciphen := cipher.Encrypt(word)
	plain := cipher.Decrypt(ciphen)
	if word != plain {
		t.Errorf("Expected word(%s) to be equal to plain (%s), got %s", word, plain, plain)
		return
	}
	t.Log("Vigenere Cipher Test Passed with word (", word, ") equal decrypt (", plain, ")")
}
