package palindrome

// Palindrome checks if a word is equal to its reversed.
func Palindrome (word string) bool {
  // check length of string
  if len(word) < 2 {
    return true
  }

  // if the first and last characters are the same,
  if word[0] == word[len(word)-1] {
    // Recursively check until results is attained
    return Palindrome(word[1:len(word)-1])
  }

  // return false if the first and last characters are not the same
  return false
}
