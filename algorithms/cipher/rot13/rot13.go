package rot13

// Rot13 is a function that takes a string and returns the string encrypted with shifts of 13.
// Rot13 is Caesar Cipher with a shift of 13.

// Encrypt returns the word with shifted values
func Encrypt(word string) string {
  var cipher string

  // shift 13 of type int32
  var shift int32 = 13

  // loop through the word
  for _, char := range word {
    // check if the character is a letter with range of a-z, A-Z
    if char >= 'a'+shift && char <= 'z' || char >= 'A'+shift && char <= 'Z' {
      // if the character is a letter, shift it
      char -= shift
    } else if char >= 'a' && char < 'a'+shift || char >= 'A' && char < 'A'+shift {
      char = char - shift + 26
    }
    // convert char to string and append to cipher
    cipher += string(char)
  }

  // Return the cipher
  return cipher
}

// Decrypt returns the word with shifted values
func Decrypt(cipher string) string {
  var word string

  // convert shift to int32 type
  var shift int32 = 13

  // loop through the word
  for _, char := range cipher {
    // check if char is in the range of a-z or A-Z
    if (char >= 'a' && char <= 'z'-shift) || (char >= 'A' && char <= 'Z'-shift) {
      // add shift to char
      char += shift
    } else if (char > 'z'-shift && char <= 'z') || (char > 'Z'-shift && char <= 'Z') {
      // subtract 26(the length of characters) from char
      char = char + shift - 26
    }
    // convert char to string and append to cipher
    word += string(char)
  }

  // Return the cipher
  return word
}
