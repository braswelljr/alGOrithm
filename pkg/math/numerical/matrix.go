package numerical

// strassenCutoff is the size at or below which Strassen multiplies directly.
const strassenCutoff = 2

// Multiply returns the product of a (n x m) and b (m x p) the direct way,
// each entry a row of a dotted with a column of b. ErrShape when the inner
// sizes differ or a row is ragged. Integer overflow wraps.
//
// Time: O(n m p). Space: O(n p).
func Multiply(a, b [][]int) ([][]int, error) {
	if len(a) == 0 || len(b) == 0 || !rectangular(a) || !rectangular(b) || len(a[0]) != len(b) {
		return nil, ErrShape
	}
	out := make([][]int, len(a))
	for i := range a {
		out[i] = make([]int, len(b[0]))
		for k, aik := range a[i] {
			for j, bkj := range b[k] {
				out[i][j] += aik * bkj
			}
		}
	}
	return out, nil
}

// MatrixPower returns the square matrix m raised to the kth power with
// repeated squaring: O(log k) products instead of k. m^0 is the identity.
// It computes Fibonacci numbers ([[1 1] [1 0]]^n) and counts walks of length
// k in a graph (its adjacency matrix to the kth). ErrShape when m is not
// square or k is negative. Integer overflow wraps.
//
// Time: O(n^3 log k). Space: O(n^2).
func MatrixPower(m [][]int, k int) ([][]int, error) {
	if len(m) == 0 || !rectangular(m) || len(m[0]) != len(m) || k < 0 {
		return nil, ErrShape
	}
	result := identity(len(m))
	base := m
	for ; k > 0; k >>= 1 {
		if k&1 == 1 {
			result, _ = Multiply(result, base)
		}
		if k > 1 {
			base, _ = Multiply(base, base)
		}
	}
	return result, nil
}

// Strassen returns the product of the square matrices a and b with Volker
// Strassen's algorithm (1969): split each into four blocks and form the
// product from seven block products instead of eight, recursively. Sizes
// that are not a power of two are padded with zeros. It was the first proof
// that matrix multiplication needs fewer than n^3 steps; in practice it wins
// only for large matrices, so it switches to Multiply on small blocks.
// ErrShape unless a and b are square and the same size.
//
// Time: O(n^log2(7)), about O(n^2.81). Space: O(n^2).
func Strassen(a, b [][]int) ([][]int, error) {
	n := len(a)
	if n == 0 || len(b) != n || !rectangular(a) || !rectangular(b) || len(a[0]) != n || len(b[0]) != n {
		return nil, ErrShape
	}
	size := 1
	for size < n {
		size <<= 1
	}
	c := strassen(pad(a, size), pad(b, size))
	out := make([][]int, n)
	for i := range out {
		out[i] = c[i][:n]
	}
	return out, nil
}

func strassen(a, b [][]int) [][]int {
	n := len(a)
	if n <= strassenCutoff {
		c, _ := Multiply(a, b)
		return c
	}
	h := n / 2
	a11, a12, a21, a22 := quarter(a, h)
	b11, b12, b21, b22 := quarter(b, h)
	m1 := strassen(addM(a11, a22, 1), addM(b11, b22, 1))
	m2 := strassen(addM(a21, a22, 1), b11)
	m3 := strassen(a11, addM(b12, b22, -1))
	m4 := strassen(a22, addM(b21, b11, -1))
	m5 := strassen(addM(a11, a12, 1), b22)
	m6 := strassen(addM(a21, a11, -1), addM(b11, b12, 1))
	m7 := strassen(addM(a12, a22, -1), addM(b21, b22, 1))
	c11 := addM(addM(m1, m4, 1), addM(m7, m5, -1), 1)
	c12 := addM(m3, m5, 1)
	c21 := addM(m2, m4, 1)
	c22 := addM(addM(m1, m2, -1), addM(m3, m6, 1), 1)
	out := make([][]int, n)
	for i := 0; i < h; i++ {
		out[i] = append(append([]int(nil), c11[i]...), c12[i]...)
		out[i+h] = append(append([]int(nil), c21[i]...), c22[i]...)
	}
	return out
}

// quarter splits a 2h x 2h matrix into its four h x h blocks.
func quarter(m [][]int, h int) (a, b, c, d [][]int) {
	for i := 0; i < h; i++ {
		a = append(a, m[i][:h])
		b = append(b, m[i][h:])
		c = append(c, m[i+h][:h])
		d = append(d, m[i+h][h:])
	}
	return a, b, c, d
}

// addM returns a + sign*b.
func addM(a, b [][]int, sign int) [][]int {
	out := make([][]int, len(a))
	for i := range a {
		out[i] = make([]int, len(a[i]))
		for j := range a[i] {
			out[i][j] = a[i][j] + sign*b[i][j]
		}
	}
	return out
}

func pad(m [][]int, size int) [][]int {
	out := make([][]int, size)
	for i := range out {
		out[i] = make([]int, size)
		if i < len(m) {
			copy(out[i], m[i])
		}
	}
	return out
}

func identity(n int) [][]int {
	out := make([][]int, n)
	for i := range out {
		out[i] = make([]int, n)
		out[i][i] = 1
	}
	return out
}

func rectangular(m [][]int) bool {
	for _, row := range m {
		if len(row) != len(m[0]) || len(row) == 0 {
			return false
		}
	}
	return true
}
