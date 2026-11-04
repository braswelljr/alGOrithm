// Package representation shows how computers represent data, bit by bit, one
// idea per file: positional bases and conversion between them (bases.go),
// binary addition with its carries (arithmetic.go), signed integers in
// sign-magnitude, ones' complement and two's complement, and fixed point
// (signed.go), IEEE 754 floating point in single and half precision
// (float.go), BCD, parity and Hamming codes (codes.go; Gray code is in
// math/bitwise), and text as UTF-8 bytes, and byte order (text.go).
//
// Every function returns its answer with the steps a student would write
// down, so the visualizer can show the working, not only the result.
package representation

import "errors"

// ErrInput is wrapped by every error about bad input: a digit outside its
// base, a width out of range, a value that does not fit.
var ErrInput = errors.New("invalid input")
