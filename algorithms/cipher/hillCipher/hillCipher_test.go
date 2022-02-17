package hillCipher

import "testing"

func TestGenerateKey(t *testing.T) {
	text := "Hello World"
	key := GenerateKey(2)
	encrypted := Encrypt(text, key)
	decrypted := Decrypt(encrypted, key)
	if text != decrypted {
		t.Errorf("Expected %s, got %s", text, decrypted)
		return
	}
	t.Log("Word ", text, " -> Encrypted : ", encrypted, " -> Decrypted : ", decrypted)
}

func TestDecrypt(t *testing.T) {
	// `Decrypt` should return the original text if the shift is 0
	text := "Hello World"
	key := GenerateKey(2)
	encrypted := Encrypt(text, key)
	decrypted := Decrypt(encrypted, key)
	if decrypted != text {
		t.Error("Expected", text, "but got", decrypted)
		return
	}
	t.Log("Text : ", text, " -> Encrypted : ", encrypted, " -> Decrypted : ", decrypted)
}
