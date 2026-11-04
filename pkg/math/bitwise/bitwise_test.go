package bitwise_test

import (
	"math/bits"
	"math/rand"
	"testing"

	"github.com/braswelljr/alGOrithm/pkg/math/bitwise"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

func samples() []uint64 {
	out := []uint64{0, 1, 2, 3, 12, 255, 256, 1 << 63, 1<<63 + 1, ^uint64(0)}
	r := rand.New(rand.NewSource(7))
	for range [500]struct{}{} {
		out = append(out, r.Uint64()>>r.Intn(64))
	}
	return out
}

func TestAgainstMathBits(t *testing.T) {
	for _, x := range samples() {
		var log trace.Log
		if got := bitwise.TracePopCount(x, &log); got != bits.OnesCount64(x) || bitwise.PopCount(x) != got || len(log.Events) != got {
			t.Fatalf("PopCount(%b) = %d", x, got)
		}
		y := x
		for _, e := range log.Events {
			y &^= 1 << e.Indices[0]
		}
		if y != 0 {
			t.Fatalf("PopCount(%b) clears left %b", x, y)
		}
		if got := bitwise.CountTrailingZeros(x); got != bits.TrailingZeros64(x) {
			t.Fatalf("CountTrailingZeros(%b) = %d", x, got)
		}
		if got := bitwise.IsPowerOfTwo(x); got != (bits.OnesCount64(x) == 1) {
			t.Fatalf("IsPowerOfTwo(%d) = %v", x, got)
		}
		want := uint64(1)
		for want < x && want != 0 {
			want <<= 1
		}
		if got := bitwise.NextPowerOfTwo(x); got != want {
			t.Fatalf("NextPowerOfTwo(%d) = %d, want %d", x, got, want)
		}
		if got := bitwise.ReverseBits(x, 64); got != bits.Reverse64(x) {
			t.Fatalf("ReverseBits(%b, 64) = %b", x, got)
		}
		if got := bitwise.ReverseBits(x, 8); got != uint64(bits.Reverse8(uint8(x))) {
			t.Fatalf("ReverseBits(%b, 8) = %b", x, got)
		}
		if bitwise.GrayDecode(bitwise.GrayCode(x)) != x {
			t.Fatalf("Gray round trip of %d", x)
		}
	}
	if bitwise.ReverseBits(5, -3) != 0 {
		t.Error("ReverseBits with negative width")
	}
}

func TestGraySequence(t *testing.T) {
	for b := 0; b <= 10; b++ {
		seq := bitwise.GraySequence(b)
		seen := map[uint64]bool{}
		for i, g := range seq {
			seen[g] = true
			if next := seq[(i+1)%len(seq)]; b > 0 && bits.OnesCount64(g^next) != 1 {
				t.Fatalf("Gray %d bits: %b then %b", b, g, next)
			}
		}
		if len(seen) != 1<<b {
			t.Fatalf("Gray %d bits has %d distinct codes", b, len(seen))
		}
	}
	if bitwise.GraySequence(17) != nil || bitwise.GraySequence(-1) != nil {
		t.Error("GraySequence accepts bad widths")
	}
}

func TestTricks(t *testing.T) {
	if got := bitwise.SingleNumber([]int{4, 1, 2, 1, 2}); got != 4 {
		t.Errorf("SingleNumber = %d", got)
	}
	if got := bitwise.SingleNumber([]int{-7, 3, 3}); got != -7 {
		t.Errorf("SingleNumber with negatives = %d", got)
	}
	if a, b := bitwise.SwapWithoutTemp(5, -9); a != -9 || b != 5 {
		t.Errorf("SwapWithoutTemp = %d, %d", a, b)
	}
	subs := bitwise.BitSubsets(0b1011)
	if len(subs) != 8 || subs[0] != 0b1011 || subs[7] != 0 {
		t.Errorf("BitSubsets = %v", subs)
	}
	seen := map[uint64]bool{}
	for _, s := range subs {
		if s&^0b1011 != 0 || seen[s] {
			t.Errorf("bad submask %b", s)
		}
		seen[s] = true
	}
	if len(bitwise.BitSubsets(0)) != 1 || bitwise.BitSubsets(1<<21-1) != nil {
		t.Error("BitSubsets edge cases")
	}
}
