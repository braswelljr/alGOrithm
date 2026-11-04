package cryptography

import (
	"fmt"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// LFSR returns n output bits of a Fibonacci linear-feedback shift register of
// width bits (2 to 32) started at seed: each step outputs the lowest bit,
// shifts right and puts the XOR of the tapped bits (0 is the lowest) in at
// the top. With the right taps the states run through all 2^width - 1
// nonzero values before repeating. LFSRs are fast in hardware, but the
// output is linear: 2*width bits reveal the taps (Berlekamp-Massey), so an
// LFSR alone is no cipher. For learning: use crypto/rand in real code.
//
// Time: O(n * taps). Space: O(n).
func LFSR(seed uint32, taps []int, width, n int) ([]int, error) {
	return TraceLFSR(seed, taps, width, n, trace.Nop{})
}

// TraceLFSR is LFSR, recording a Set for each output bit (Indices[0] the
// step, Values[0] the bit, Values[1] the state after).
func TraceLFSR(seed uint32, taps []int, width, n int, r trace.Recorder) ([]int, error) {
	if err := checkLFSR(seed, taps, width); err != nil {
		return nil, err
	}
	out := make([]int, n)
	s := seed
	for i := range out {
		out[i] = int(s & 1)
		s = lfsrStep(s, taps, width)
		r.Record(trace.Event{Kind: trace.Set, Indices: []int{i}, Values: []int{out[i], int(s)}})
	}
	return out, nil
}

// LFSRPeriod returns how many steps the register takes to come back to seed.
//
// Time: O(2^width * taps). Space: O(1).
func LFSRPeriod(seed uint32, taps []int, width int) (int, error) {
	if err := checkLFSR(seed, taps, width); err != nil {
		return 0, err
	}
	s, n := lfsrStep(seed, taps, width), 1
	for s != seed {
		s, n = lfsrStep(s, taps, width), n+1
	}
	return n, nil
}

func lfsrStep(s uint32, taps []int, width int) uint32 {
	var fb uint32
	for _, t := range taps {
		fb ^= s >> t & 1
	}
	return s>>1 | fb<<(width-1)
}

func checkLFSR(seed uint32, taps []int, width int) error {
	if width < 2 || width > 32 || len(taps) == 0 {
		return fmt.Errorf("%w: an LFSR needs a width of 2 to 32 and some taps", ErrInvalid)
	}
	if seed == 0 || width < 32 && seed>>width != 0 {
		return fmt.Errorf("%w: the seed must be nonzero and fit in %d bits", ErrInvalid, width)
	}
	for _, t := range taps {
		if t < 0 || t >= width {
			return fmt.Errorf("%w: tap %d is outside the register", ErrInvalid, t)
		}
	}
	return nil
}
