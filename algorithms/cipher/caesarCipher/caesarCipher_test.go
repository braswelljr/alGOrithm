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
