package hillCipher

import "testing"

func TestGenerateKey(t *testing.T) {
	text := "Hello World"
	key := GenerateKey(2)
	encrypted := Encrypt(text, key)
	decrypted := Decrypt(encrypted, key)
	//if text != decrypted {
	//  t.Errorf("Expected %s, got %s", text, decrypted)
	//  return
	//}
	t.Log("Word ", text, " -> Encrypted : ", encrypted, " -> Decrypted : ", decrypted)
}
