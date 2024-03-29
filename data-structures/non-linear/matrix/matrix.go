package matrix

import (
	"strconv"
	"strings"
)

type Matrix struct {
	Rows int
	Cols int
	Data [][]int
}

// New - creates a new matrix
//
//	@param rows - the number of rows
//	@param cols - the number of columns
//	@return m - the matrix
func New(rows, cols int) *Matrix {
	// create a new matrix
	m := &Matrix{}

	// set the rows, columns and data
	m.Rows, m.Cols, m.Data = rows, cols, make([][]int, rows)

	// loop through the rows
	for i := 0; i < rows; i++ {
		// set the columns
		m.Data[i] = make([]int, cols)
	}

	// return the matrix
	return m
}

// NewFromData - creates a new matrix from data
//
//	@param data - the data
//	@return m - the matrix
func NewFromData(data [][]int) *Matrix {
	// create a new matrix
	m := &Matrix{}

	// set the rows, columns and data
	m.Rows, m.Cols, m.Data = len(data), len(data[0]), data

	// return the matrix
	return m
}

// Set - sets the value of a matrix
//
//	@param row - the row
//	@param col - the column
//	@param value - the value
func (m *Matrix) Set(row, col, value int) {
	// set the value
	m.Data[row][col] = value
}

// Multiply - multiplies two matrices
//
//	@param a - the first matrix
//	@param b - the second matrix
//	@return c - the product of the two matrices
func (a *Matrix) Multiply(b *Matrix) *Matrix {
	// create a new matrix
	c := New(a.Rows, b.Cols)

	// loop through the rows
	for i := 0; i < a.Rows; i++ {
		// loop through the columns
		for j := 0; j < b.Cols; j++ {
			// loop through the rows
			for k := 0; k < a.Rows; k++ {
				// multiply the values and add it to the matrix
				c.Data[i][j] += a.Data[i][k] * b.Data[k][j]
			}
		}
	}

	// return the matrix
	return c
}

// Add - adds two matrices
//
//	@param a - the first matrix
//	@param b - the second matrix
//	@return c - the sum of the two matrices
func (a *Matrix) Add(b *Matrix) *Matrix {
	// create a new matrix
	c := New(a.Rows, a.Cols)

	// loop through the rows
	for i := 0; i < a.Rows; i++ {
		// loop through the columns
		for j := 0; j < a.Cols; j++ {
			// add the values and add it to the matrix
			c.Data[i][j] = a.Data[i][j] + b.Data[i][j]
		}
	}

	// return the matrix
	return c
}

// Subtract - subtracts two matrices
//
//	@param a - the first matrix
//	@param b - the second matrix
//	@return c - the difference of the two matrices
func (a *Matrix) Subtract(b *Matrix) *Matrix {
	// create a new matrix
	c := New(a.Rows, a.Cols)

	// loop through the rows
	for i := 0; i < a.Rows; i++ {
		// loop through the columns
		for j := 0; j < a.Cols; j++ {
			// subtract the values and add it to the matrix
			c.Data[i][j] = a.Data[i][j] - b.Data[i][j]
		}
	}

	// return the matrix
	return c
}

// Transpose - transposes a matrix
//
//	@param a - the matrix
//	@return b - the transposed matrix
func (a *Matrix) Transpose() *Matrix {
	// create a new matrix
	b := New(a.Cols, a.Rows)

	// loop through the rows
	for i := 0; i < a.Rows; i++ {
		// loop through the columns
		for j := 0; j < a.Cols; j++ {
			// set the transposed value
			b.Data[j][i] = a.Data[i][j]
		}
	}

	// return the matrix
	return b
}

// Determinant - calculates the determinant of a matrix
//
//	@param a - the matrix
//	@return det - the determinant of the matrix
func (a *Matrix) Determinant() int {
	// create a new matrix
	b := New(a.Rows, a.Cols)

	// loop through the rows
	for i := 0; i < a.Rows; i++ {
		// loop through the columns
		for j := 0; j < a.Cols; j++ {
			// set the value
			b.Data[i][j] = a.Data[i][j]
		}
	}

	// calculate the determinant
	return determinant(b)
}

// determinant - calculates the determinant of a matrix
//
//	@param a - the matrix
//	@return det - the determinant of the matrix
func determinant(a *Matrix) int {
	// check if the matrix is 2x2
	if a.Rows == 2 && a.Cols == 2 {
		// return the determinant
		return a.Data[0][0]*a.Data[1][1] - a.Data[0][1]*a.Data[1][0]
	}

	// create a new matrix
	b := New(a.Rows-1, a.Cols-1)

	// create a variable to hold the determinant
	det := 0

	// loop through the columns
	for i := 0; i < a.Cols; i++ {
		// loop through the rows
		for j := 1; j < a.Rows; j++ {
			// loop through the columns
			for k := 0; k < a.Cols; k++ {
				// check if the column is not the current column
				if k != i {
					// set the value
					b.Data[j-1][k] = a.Data[j][k]
				}
			}
		}

		// calculate the determinant
		det += a.Data[0][i] * determinant(b)
	}

	// return the determinant
	return det
}

// Inverse - calculates the inverse of a matrix
//
//	@param a - the matrix
//	@return b - the inverse of the matrix
func (a *Matrix) Inverse() *Matrix {
	// create a new matrix
	b := New(a.Rows, a.Cols)

	// calculate the determinant
	det := b.Determinant()

	// check if the determinant is 0
	if det == 0 {
		// return the matrix
		return b
	}

	// create a new matrix
	c := New(a.Rows-1, a.Cols-1)

	// loop through the rows
	for i := 0; i < a.Rows; i++ {
		// loop through the columns
		for j := 0; j < a.Cols; j++ {
			// loop through the rows
			for k := 0; k < a.Rows; k++ {
				// check if the row is not the current row
				if k != i {
					// loop through the columns
					for l := 0; l < a.Cols; l++ {
						// check if the column is not the current column
						if l != j {
							// set the value
							c.Data[k][l] = a.Data[k][l]
						}
					}
				}
			}

			// set the value
			b.Data[i][j] = determinant(c)
		}
	}

	// loop through the rows
	for i := 0; i < a.Rows; i++ {
		// loop through the columns
		for j := 0; j < a.Cols; j++ {
			// check if the row and column are even
			if (i+j)%2 != 0 {
				// set the value
				b.Data[i][j] = -b.Data[i][j]
			}
		}
	}

	// return the matrix
	return b
}

// Transform - transforms a matrix
//
//	@param a - the matrix
//	@param b - the transformation matrix
//	@return c - the transformed matrix
func (a *Matrix) Transform(b *Matrix) *Matrix {
	// create a new matrix
	c := New(a.Rows, b.Cols)

	// loop through the rows
	for i := 0; i < a.Rows; i++ {
		// loop through the columns
		for j := 0; j < b.Cols; j++ {
			// loop through the columns
			for k := 0; k < a.Cols; k++ {
				// add the values and add it to the matrix
				c.Data[i][j] += a.Data[i][k] * b.Data[k][j]
			}
		}
	}

	// return the matrix
	return c
}

// ConvertStringToMatrix - converts a string to a matrix
//
//	@param s - the s to be converted
//	@param n - the size of the matrix
//	@return matrix - the matrix
func ConvertStringToMatrix(s string, n int) *Matrix {
	// create a new matrix
	a := New(n, n)

	// split the string by new lines
	lines := strings.Split(s, " ")

	// loop through the lines
	for i := 0; i < len(lines); i++ {
		// split the line by spaces
		values := strings.Split(lines[i], " ")

		// loop through the values
		for j := 0; j < len(values); j++ {
			// convert the value to an integer
			value, _ := strconv.Atoi(values[j])

			// check if the matrix is empty
			if a.Rows == 0 && a.Cols == 0 {
				// set the number of rows and columns
				a.Rows = len(lines)
				a.Cols = len(values)
			}

			// set the value
			a.Data[i][j] = value
		}
	}

	// return the matrix
	return a
}

// ConvertToString - converts a matrix to a string
//
//	@param a - the matrix
//	@return s - the string
func (a *Matrix) ConvertToString() string {
	// create a new string
	s := ""

	// loop through the rows
	for i := 0; i < a.Rows; i++ {
		// loop through the columns
		for j := 0; j < a.Cols; j++ {
			// add the value to the string
			s += strconv.Itoa(a.Data[i][j])

			// check if the value is not the last value
			if j != a.Cols-1 {
				// add a space
				s += " "
			}
		}

		// check if the value is not the last value
		if i != a.Rows-1 {
			// add a new line
			s += "\n"
		}
	}

	// return the string
	return s
}
