package hillCipher

import (
	"github.com/braswelljr/alGOrithm/utils"
)

// GenerateKey - generates a key for the hill cipher
// @param key - the key to be used for the hill cipher
// @param n - the size of the key
// @return keyMatrix - the key matrix
func GenerateKey(n int) [][]int {
	// create a slice of slices
	var keyMatrix [][]int
	// get an (n x n) character random key
	key := utils.GenerateRandomKey(n * n)

	// loop through the key
	for i := 0; i < n; i++ {
		var row []int

		// loop through the key and add it to the key matrix
		for j := 0; j < n; j++ {
			// convert the character to an integer and add it to the row
			row = append(row, int(key[i*n+j]))
		}
		// add the row to the key matrix
		keyMatrix = append(keyMatrix, row)
	}

	// return the key matrix
	return keyMatrix
}
