package caesarCipher

// CaesarCipher is a function that takes a string and a number and returns the string encrypted with the caesar cipher

// Encrypt returns the word with shifted values
func Encrypt(word string, shift int) string {
  var cipher, character string

  // loop through the word
  for _, char := range word {
    // Check for validity (in range)
    if char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' {
      // Add shift to character
      code := char + int32(shift)
      character = string(code)
      // check character
    } else {
      character = string(char)
    }
    // Add character to the cipher returned string
    cipher += character
  }

  // Return the cipher
  return cipher
}

// Decrypt returns the word with shifted values
func Decrypt(cipher string, shift int) string {
  var word, character string

  // loop through the word
  for _, char := range cipher {
    // Check for validity (in range)
    if char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' {
      // Add shift to character
      code := char - int32(shift)
      character = string(code)
    } else {
      character = string(char)
    }
    // Add character to the cipher returned string
    word += character
  }

  // Return the word
  return word
}
