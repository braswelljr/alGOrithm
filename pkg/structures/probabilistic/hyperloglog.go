package probabilistic

import (
	"fmt"
	"math"
	"math/bits"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// HyperLogLog estimates how many distinct items a stream holds, in a few
// kilobytes for billions of items (Philippe Flajolet et al., 2007). Each item's
// hash picks one of m = 2^p registers with its first p bits, and the register
// keeps the longest run of leading zeros seen in the remaining bits: a run of
// k zeros turns up about once in 2^k distinct hashes. The harmonic mean of
// 2^register over all registers, scaled, is the estimate; adding an item twice
// changes nothing. The standard error is about 1.04 / sqrt(m).
//
// For small counts, while registers are still zero, the estimate switches to
// linear counting, which is more accurate there.
//
// Time: O(1) per Add, O(m) per Count.
// Space: O(m) bytes.
type HyperLogLog struct {
	p         uint
	registers []uint8
}

// NewHyperLogLog returns an estimator with 2^p registers; p is held to 4..16.
func NewHyperLogLog(p int) *HyperLogLog {
	p = min(max(p, 4), 16)
	return &HyperLogLog{p: uint(p), registers: make([]uint8, 1<<p)}
}

// Add records item and returns the register it fell in and that register's
// value afterwards.
func (h *HyperLogLog) Add(item string) (register int, rank int) {
	x := hash64(item, 0)
	register = int(x >> (64 - h.p))
	rest := x<<h.p | 1<<(h.p-1) // a stop bit, so rank stays at most 64-p+1
	rank = bits.LeadingZeros64(rest) + 1
	if uint8(rank) > h.registers[register] {
		h.registers[register] = uint8(rank)
	}
	return register, int(h.registers[register])
}

// Count returns the estimated number of distinct items added.
func (h *HyperLogLog) Count() int {
	m := float64(len(h.registers))
	sum, zeros := 0.0, 0
	for _, r := range h.registers {
		sum += math.Ldexp(1, -int(r))
		if r == 0 {
			zeros++
		}
	}
	alpha := 0.7213 / (1 + 1.079/m)
	switch m {
	case 16:
		alpha = 0.673
	case 32:
		alpha = 0.697
	case 64:
		alpha = 0.709
	}
	est := alpha * m * m / sum
	if est <= 2.5*m && zeros > 0 {
		est = m * math.Log(m/float64(zeros)) // linear counting
	}
	return int(math.Round(est))
}

// StandardError returns the relative standard error, 1.04 / sqrt(m).
func (h *HyperLogLog) StandardError() float64 { return 1.04 / math.Sqrt(float64(len(h.registers))) }

// Registers returns a copy of the registers.
func (h *HyperLogLog) Registers() []int {
	out := make([]int, len(h.registers))
	for i, r := range h.registers {
		out[i] = int(r)
	}
	return out
}

// Snapshot describes the registers, at most 64 of them, as a row.
func (h *HyperLogLog) Snapshot() trace.Snapshot {
	n := min(len(h.registers), 64)
	labels := make([]string, n)
	for i := range labels {
		labels[i] = fmt.Sprintf("r%d=%d", i, h.registers[i])
	}
	return trace.Row(labels, "", "", "")
}
