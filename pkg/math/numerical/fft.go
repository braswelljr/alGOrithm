package numerical

import (
	"math"
	"math/cmplx"
)

// FFT returns the discrete Fourier transform of x, whose length must be a
// power of two, with the radix-2 Cooley-Tukey fast Fourier transform (1965):
// the transform of n values is built from the transforms of the even and the
// odd positions, combined with twiddle factors e^(-2 pi i k / n). That turns
// O(n^2) work into O(n log n), and it underlies signal processing, image
// compression and fast polynomial multiplication. x is not modified.
// ErrShape when the length is not a power of two.
//
// Time: O(n log n). Space: O(n log n) with this recursive version.
func FFT(x []complex128) ([]complex128, error) {
	if len(x) == 0 || len(x)&(len(x)-1) != 0 {
		return nil, ErrShape
	}
	return fft(x, -1), nil
}

// InverseFFT reverses FFT: the same transform with the opposite sign in the
// exponent, divided by n.
func InverseFFT(x []complex128) ([]complex128, error) {
	if len(x) == 0 || len(x)&(len(x)-1) != 0 {
		return nil, ErrShape
	}
	out := fft(x, 1)
	for i := range out {
		out[i] /= complex(float64(len(x)), 0)
	}
	return out, nil
}

func fft(x []complex128, sign float64) []complex128 {
	n := len(x)
	if n == 1 {
		return []complex128{x[0]}
	}
	even, odd := make([]complex128, n/2), make([]complex128, n/2)
	for i := 0; i < n/2; i++ {
		even[i], odd[i] = x[2*i], x[2*i+1]
	}
	e, o := fft(even, sign), fft(odd, sign)
	out := make([]complex128, n)
	for k := 0; k < n/2; k++ {
		t := cmplx.Rect(1, sign*2*math.Pi*float64(k)/float64(n)) * o[k]
		out[k], out[k+n/2] = e[k]+t, e[k]-t
	}
	return out
}

// PolyMultiply returns the product of the polynomials a and b (constant term
// first) through the FFT: evaluate both at the 2^k-th roots of unity,
// multiply pointwise, and interpolate back. The coefficients carry rounding
// error of about 1e-12 relative to their size; round them for integer
// polynomials. nil when either is empty.
//
// Time: O(n log n). Space: O(n).
func PolyMultiply(a, b []float64) []float64 {
	if len(a) == 0 || len(b) == 0 {
		return nil
	}
	size := 1
	for size < len(a)+len(b)-1 {
		size <<= 1
	}
	fa, fb := make([]complex128, size), make([]complex128, size)
	for i, v := range a {
		fa[i] = complex(v, 0)
	}
	for i, v := range b {
		fb[i] = complex(v, 0)
	}
	ta, tb := fft(fa, -1), fft(fb, -1)
	for i := range ta {
		ta[i] *= tb[i]
	}
	prod, _ := InverseFFT(ta)
	out := make([]float64, len(a)+len(b)-1)
	for i := range out {
		out[i] = real(prod[i])
	}
	return out
}
