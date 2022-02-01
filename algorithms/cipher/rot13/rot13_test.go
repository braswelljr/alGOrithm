package rot13

import (
"testing"
)

func TestCaesarCipher(t *testing.T) {
  text := "Hello World"
  x := Encrypt(text)
  y := Decrypt(x)
  t.Log("Text : ", text, " -> Encrypted : ", x, " -> Decrypted : ", y)
}

func TestDecrypt(t *testing.T) {
  // `Decrypt` should return the original text if the shift is 0
  text := "Hello World"
  x := Encrypt(text)
  y := Decrypt(x)
  if y != text {
    t.Error("Expected", text, "but got", y)
    return
  }
  t.Log("Text : ", text, " -> Encrypted : ", x, " -> Decrypted : ", y)
}
