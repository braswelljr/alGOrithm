package hillCipher

import (
	"encoding/binary"
	"math/rand"
)

// GenerateKey - generate a random key
func GenerateKey(size int) [][]int {
	// create a byte array of size*size
	var b [8]byte
	// seed the random number generator
	rand.Seed(int64(binary.BigEndian.Uint64(b[:])))
	// create a 2d array of size*size
	key := make([][]int, size)
	// fill the 2d array with random numbers
	for i := 0; i < size; i++ {
		// create a 1d array of size
		key[i] = make([]int, size)

		// fill the 1d array with random numbers (0-25)
		for j := 0; j < size; j++ {
			key[i][j] = int(rand.Int31n(26))
		}

	}
	// return the 2d array
	return key
}

// Encrypt - encrypt a message
func Encrypt(word string, key [][]int) string {
	var cipher string
	for _, char := range word {
		cipher += string(rune(key[0][0]*(int(char)-'a') + key[1][0]))
	}
	return cipher
}

// Decrypt - decrypt a cipher
func Decrypt(cipher string, key [][]int) string {
	var word string
	for _, char := range cipher {
		word += string(rune(key[0][1]*(int(char)-'a') + key[1][1]))
	}
	return word
}
