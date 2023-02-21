// CaesarCipher is a function that takes a string and a number and returns the string encrypted with the caesar cipher.
//
//	@function Encrypt - takes up a word with a shift value and returns the word encrypted with the caesar cipher.
//	@function Decrypt - takes up an encrypted word with the corresponding shift value and returns the decrypted word.
package caesarCipher

// Encrypt - takes up a word with a shift value and returns an encryption.
//
//	@param word - word to be encrypted
//	@param shift - shift value
//	@return - string
func Encrypt(word string, shift int) string {
	var cipher string

	// convert shift to int32 type
	x := int32(shift)

	// loop through the word
	for _, char := range word {
		// check if the character is a letter with range of a-z, A-Z
		if char >= 'a'+x && char <= 'z' || char >= 'A'+x && char <= 'Z' {
			// if the character is a letter, shift it
			char -= x
		} else if char >= 'a' && char < 'a'+x || char >= 'A' && char < 'A'+x {
			char = char - x + 26
		}
		// convert char to string and append to cipher
		cipher += string(char)
	}

	// Return the cipher
	return cipher
}

// Decrypt - takes up an encrypted word with the corresponding shift value and returns the decrypted word.
//
//	@param cipher - encrypted word
//	@param shift - shift value
//	@return - string
func Decrypt(cipher string, shift int) string {
	var word string

	// convert shift to int32 type
	x := int32(shift)

	// loop through the word
	for _, char := range cipher {
		// check if char is in the range of a-z or A-Z
		if (char >= 'a' && char <= 'z'-x) || (char >= 'A' && char <= 'Z'-x) {
			// add shift to char
			char += x
		} else if (char > 'z'-x && char <= 'z') || (char > 'Z'-x && char <= 'Z') {
			// subtract 26(the length of characters) from char
			char = char + x - 26
		}
		// convert char to string and append to cipher
		word += string(char)
	}

	// Return the cipher
	return word
}
