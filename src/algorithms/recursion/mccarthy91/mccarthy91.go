package mccarthy91

// McCarthy91 returns the number of words in the input string that are less than
// or equal to the given limit.
// The input string is assumed to be a single line of text.

func McCarthy91(input string, limit int) int {
  // if the input is empty, return 0
  if len(input) == 0 {
    return 0
  }

  // If the first character is less than or equal to the limit,
  if len(input) <= limit {
    return 1
  }

  // Recurse on the first half of the string.
  return McCarthy91(input[1:], limit)
}
