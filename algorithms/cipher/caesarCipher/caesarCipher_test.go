package caesarCipher

import (
	"testing"
)

func TestCaesarCipher(t *testing.T) {
	text := "Hello World"
	shift := 2
	x := Encrypt(text, shift)
	y := Decrypt(x, shift)
	t.Log("Text : ", text, "( Shift : ", shift, " )", " -> Encrypted : ", x, " -> Decrypted : ", y)
}

func TestDecrypt(t *testing.T) {
	// `Decrypt` should return the original text if the shift is 0
	text := "Hello World"
	shift := 18
	x := Encrypt(text, shift)
	y := Decrypt(x, shift)
	if y != text {
		t.Error("Expected", text, "but got", y)
		return
	}
	t.Log("Text : ", text, "( Shift : ", shift, " )", " -> Encrypted : ", x, " -> Decrypted : ", y)
}
