package anagram

// Anagram - Anagrams are words or phrases that contain the same number of characters.
//
//	@word - string
//	@anagram - string
//	@return - bool
func Anagram(word string, anagram string) bool {
	// compare the length of the two strings
	if len(word) != len(anagram) {
		return false
	}

	// create a map to store the characters of the word
	wordMap := make(map[rune]int)

	// loop through the word and store the characters in the map
	for _, char := range word {
		wordMap[char]++
	}

	// loop through the anagram and check if the characters are in the map
	for _, char := range anagram {
		if _, ok := wordMap[char]; !ok {
			return false
		}

		wordMap[char]--
	}

	// loop through the map and check if the values are 0
	for _, value := range wordMap {
		if value != 0 {
			return false
		}
	}

	return true
}
