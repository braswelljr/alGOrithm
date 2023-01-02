package utils

// ConvertStringToMatrix - converts a string to a matrix
//
//  @param message - the message to be converted
//  @param n - the size of the matrix
//  @return matrix - the matrix
func ConvertStringToMatrix(message string, n int) [][]int {
	// create a slice of slices
	var matrix [][]int
	// get the length of the message
	length := len(message)

	// loop through the message
	for i := 0; i < length; i += n {
		var row []int

		// loop through the message and add it to the matrix
		for j := 0; j < n; j++ {
			// convert the character to an integer and add it to the row
			row = append(row, int(message[i+j]))
		}
		// add the row to the matrix
		matrix = append(matrix, row)
	}

	// return the matrix
	return matrix
}
