package arrays

import (
	"strconv"
)

type Matrix struct {
	Rows int
	Cols int
	Data [][]int
}

// NewMatrix - creates a new matrix
//
//	@param rows - the number of rows
//	@param cols - the number of columns
//	@return m - the matrix
func NewMatrix(rows, cols int) *Matrix {
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

// MatrixFrom - creates a new matrix from data
//
//	@param data - the data
//	@return m - the matrix
func MatrixFrom(data [][]int) *Matrix {
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
	c := NewMatrix(a.Rows, b.Cols)

	// loop through the rows
	for i := 0; i < a.Rows; i++ {
		// loop through the columns
		for j := 0; j < b.Cols; j++ {
			// loop through a's columns (b's rows)
			for k := 0; k < a.Cols; k++ {
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
	c := NewMatrix(a.Rows, a.Cols)

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
	c := NewMatrix(a.Rows, a.Cols)

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
	b := NewMatrix(a.Cols, a.Rows)

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
	b := NewMatrix(a.Rows, a.Cols)

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

// determinant - calculates the determinant of a square matrix by cofactor
// expansion along the first row
//
//	@param a - the matrix
//	@return det - the determinant of the matrix
func determinant(a *Matrix) int {
	switch a.Rows {
	case 0:
		return 1
	case 1:
		return a.Data[0][0]
	case 2:
		return a.Data[0][0]*a.Data[1][1] - a.Data[0][1]*a.Data[1][0]
	}

	// create a variable to hold the determinant
	det, sign := 0, 1

	// expand along the first row, alternating the cofactor's sign
	for i := 0; i < a.Cols; i++ {
		det += sign * a.Data[0][i] * determinant(minor(a, 0, i))
		sign = -sign
	}

	// return the determinant
	return det
}

// minor - the matrix a without one row and one column
//
//	@param a - the matrix
//	@param row - the row to leave out
//	@param col - the column to leave out
//	@return m - the minor
func minor(a *Matrix, row, col int) *Matrix {
	m := NewMatrix(a.Rows-1, a.Cols-1)
	for i, r := 0, 0; i < a.Rows; i++ {
		if i == row {
			continue
		}
		for j, c := 0, 0; j < a.Cols; j++ {
			if j != col {
				m.Data[r][c] = a.Data[i][j]
				c++
			}
		}
		r++
	}
	return m
}

// Inverse - calculates the inverse of a square matrix as its adjugate
// divided by its determinant. The entries are integers, so the result is
// exact only when the inverse has integer entries (a determinant of 1 or -1,
// for one); otherwise each entry is truncated toward zero. A singular matrix
// (determinant 0) gives the zero matrix.
//
//	@param a - the matrix
//	@return b - the inverse of the matrix
func (a *Matrix) Inverse() *Matrix {
	// create a new matrix
	b := NewMatrix(a.Rows, a.Cols)

	// calculate the determinant
	det := a.Determinant()

	// check if the determinant is 0
	if det == 0 {
		// return the matrix
		return b
	}

	// the adjugate is the transpose of the cofactor matrix
	for i := 0; i < a.Rows; i++ {
		for j := 0; j < a.Cols; j++ {
			cofactor := determinant(minor(a, i, j))
			if (i+j)%2 != 0 {
				cofactor = -cofactor
			}
			b.Data[j][i] = cofactor / det
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
	c := NewMatrix(a.Rows, b.Cols)

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
