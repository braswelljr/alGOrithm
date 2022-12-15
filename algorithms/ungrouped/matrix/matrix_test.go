package matrix

import (
	"fmt"
	"testing"
)

// TestMultiply - tests the multiply function
func TestMultiply(t *testing.T) {
	// create a matrix
	a := New(2, 2)
	a.Data = [][]int{{1, 2}, {3, 4}}

	// create a matrix
	b := New(2, 2)
	b.Data = [][]int{{5, 6}, {7, 8}}

	// multiply the matrices
	c := a.Multiply(b)

	// print the matrix
	t.Log("Matrix A:", a.Data)
	t.Log("Matrix B:", b.Data)
	t.Log("Answer of A * B:", c.Data)
}

// TestAdd - tests the add function
func TestAdd(t *testing.T) {
	// create a matrix
	a := New(2, 2)
	a.Data = [][]int{{1, 2}, {3, 4}}

	// create a matrix
	b := New(2, 2)
	b.Data = [][]int{{5, 6}, {7, 8}}

	// add the matrices
	c := a.Add(b)

	// print the matrix
	t.Log("Matrix A:", a.Data)
	t.Log("Matrix B:", b.Data)
	t.Log("Answer of A * B:", c)
}

// TestNew - tests the new function
func TestNew(t *testing.T) {
	// create a matrix
	m := New(2, 2)

	// print the matrix
	fmt.Println(m.Data)
}

// TestTransform - transforms a matrix
func TestTransform(t *testing.T) {
	// create a matrix
	m := New(2, 2)
	m.Data = [][]int{{1, 2}, {3, 4}}

	x := New(2, 2)
	x.Data = [][]int{{5, 6}, {7, 8}}

	// transform the matrix
	res := m.Transform(x)

	// print the matrix
	t.Log("Matrix A:", m.Data)
	t.Log("Matrix B:", x.Data)
	t.Log("Answer of A * B:", res.Data)
}
