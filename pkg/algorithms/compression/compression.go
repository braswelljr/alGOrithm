// Package compression holds lossless compression algorithms, one file each,
// every one with its inverse: run-length encoding, LZ77 (sliding-window
// matches), LZW (a dictionary that grows as it reads), the Burrows-Wheeler
// transform and move-to-front (which compress nothing alone but make data
// easier to compress, the front of bzip2). Ratio compares sizes.
//
// Each encoder returns its tokens rather than packed bits, so the output can
// be read and checked; the decoder turns the tokens back into the input
// exactly:
//
//	runs := compression.RunLength([]byte("aaabcc")) // [{a 3} {b 1} {c 2}]
//	compression.RunLengthDecode(runs)               // "aaabcc"
package compression

import "errors"

// ErrCorrupt is returned by a decoder given tokens no encoder would produce.
var ErrCorrupt = errors.New("compression: corrupt input")

// Ratio returns originalSize / compressedSize, how many times smaller the
// data became: above 1 it shrank, below 1 it grew. 0 when compressedSize is
// not positive.
func Ratio(originalSize, compressedSize int) float64 {
	if compressedSize <= 0 {
		return 0
	}
	return float64(originalSize) / float64(compressedSize)
}
