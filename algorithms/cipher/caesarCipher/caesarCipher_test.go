package caesarCipher

import (
	"strings"
	"testing"
)

var (
	messages = []string{
		"Hello World!",
		"Go is awesome!\nBut we gotta do it right!",
		"Go is awesome!\nBut we gotta do it right!\nAnd we gotta do it fast!",
	}
	shifts = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
)

func TestCaesarCipher(t *testing.T) {
	for _, message := range messages {
		for _, shift := range shifts {
			x := Encrypt(message, shift)
			y := Decrypt(x, shift)
			if y != message {
				t.Error("Expected", message, "but got", y)
				return
			}
			t.Log("Text : ", message, "( Shift : ", shift, " )", " -> Encrypted : ", x, " -> Decrypted : ", y)
		}
	}
}

// TestEncrypt - tests the encrypt function
func TestEncrypt(t *testing.T) {
	// uses the messages and shifts variables
	for _, message := range messages {
		for _, shift := range shifts {
			x := Encrypt(message, shift)
			// check if the encrypted message is not empty
			if strings.TrimSpace(x) == "" || len(x) == 0 {
				t.Error("Expected encrypted message to be not empty")
				return
			}

			// check if the encrypted message is not the same as the original message
			if x == message {
				t.Error("Expected encrypted message to be different from the original message")
				//
				return
			}

			// check for correct number of character shifts
			t.Log("Text : ", message, "( Shift : ", shift, " )", " -> Encrypted : ", x)
		}
	}

}

// TestDecrypt - tests the decrypt function
func TestDecrypt(t *testing.T) {
	// uses the messages and shifts variables
	for _, message := range messages {
		for _, shift := range shifts {
			x := Encrypt(message, shift)
			y := Decrypt(x, shift)
			// check if the decrypted message is not empty
			if strings.TrimSpace(y) == "" || len(y) == 0 {
				t.Error("Expected decrypted message to be not empty")
				return
			}

			// check if the decrypted message is the same as the original message
			if y != message {
				t.Error("Expected decrypted message to be the same as the original message")
				return
			}

			// check for correct number of character shifts
			t.Log("Text : ", message, "( Shift : ", shift, " )", " -> Encrypted : ", x, " -> Decrypted : ", y)
		}
	}
}
