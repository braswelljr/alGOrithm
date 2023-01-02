package hillCipher

import (
	"github.com/braswelljr/alGOrithm/data-structures/non-linear/matrix"
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

// Encrypt - encrypts a message using the hill cipher
//
//	@param message - the message to be encrypted
//	@param keyMatrix - the key matrix
//	@return encryptedMessage - the encrypted message
func Encrypt(message string, keyMatrix [][]int) string {
	// convert the message to an integer matrix
	// multiply the key matrix and the message matrix
	// import from the matrix package
	// messageMatrix := utils.ConvertStringToMatrix(message, len(keyMatrix))
	messageMatrix := matrix.ConvertStringToMatrix(message, len(keyMatrix))

	// set keyMatrix as the base matrix
	k := matrix.NewFromData(keyMatrix)

	// multiply the key matrix and the message matrix
	product := k.Multiply(messageMatrix)

	// convert the product matrix to a string
	encryptedMessage := product.ConvertToString()

	// return the encrypted message
	return encryptedMessage
}
