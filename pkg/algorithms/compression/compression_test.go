package compression_test

import (
	"bytes"
	"errors"
	"math/rand"
	"reflect"
	"testing"

	"github.com/braswelljr/alGOrithm/pkg/algorithms/compression"
)

// inputs returns texts with and without repetition, including empty ones.
func inputs() [][]byte {
	out := [][]byte{nil, []byte("a"), []byte("aaaa"), []byte("banana"), []byte("abracadabra"),
		[]byte("TOBEORNOTTOBEORTOBEORNOT"), []byte("aaabbbcccaaabbbccc"), {0, 255, 0, 255, 255}}
	r := rand.New(rand.NewSource(3))
	for n := 1; n < 400; n += 37 {
		b := make([]byte, n)
		for i := range b {
			b[i] = byte('a' + r.Intn(1+n%5))
		}
		out = append(out, b)
		noise := make([]byte, n)
		r.Read(noise)
		out = append(out, noise)
	}
	return out
}

func TestRoundTrips(t *testing.T) {
	for _, in := range inputs() {
		if got, err := compression.RunLengthDecode(compression.RunLength(in)); err != nil || !bytes.Equal(got, in) {
			t.Fatalf("run length %q: %q, %v", in, got, err)
		}
		for _, w := range []int{0, 1, 4} {
			if got, err := compression.LZ77Decode(compression.LZ77(in, w)); err != nil || !bytes.Equal(got, in) {
				t.Fatalf("lz77(%d) %q: %q, %v", w, in, got, err)
			}
		}
		codes, added := compression.LZW(in)
		if got, err := compression.LZWDecode(codes); err != nil || !bytes.Equal(got, in) {
			t.Fatalf("lzw %q: %q, %v", in, got, err)
		}
		if len(in) > 0 && len(added) != len(codes)-1 {
			t.Fatalf("lzw %q: %d codes but %d phrases added", in, len(codes), len(added))
		}
		last, idx := compression.BurrowsWheeler(string(in))
		if got, err := compression.InverseBurrowsWheeler(last, idx); err != nil || got != string(in) {
			t.Fatalf("bwt %q: %q, %v", in, got, err)
		}
		if got := compression.MoveToFrontDecode(compression.MoveToFront(in)); !bytes.Equal(got, in) {
			t.Fatalf("mtf %q: %q", in, got)
		}
	}
}

func TestKnownOutputs(t *testing.T) {
	if got := compression.RunLength([]byte("aaabcc")); !reflect.DeepEqual(got, []compression.Run{{'a', 3}, {'b', 1}, {'c', 2}}) {
		t.Errorf("RunLength = %v", got)
	}
	if got := compression.LZ77([]byte("aaaa"), 0); !reflect.DeepEqual(got, []compression.Token{{0, 0, 'a'}, {1, 2, 'a'}}) {
		t.Errorf("LZ77(aaaa) = %v", got)
	}
	codes, added := compression.LZW([]byte("TOBEORNOTTOBEORTOBEORNOT"))
	want := []int{84, 79, 66, 69, 79, 82, 78, 79, 84, 256, 258, 260, 265, 259, 261, 263}
	if !reflect.DeepEqual(codes, want) || added[0] != "TO" || added[9] != "TOB" {
		t.Errorf("LZW = %v, %v", codes, added)
	}
	if last, idx := compression.BurrowsWheeler("banana"); last != "nnbaaa" || idx != 3 {
		t.Errorf("BurrowsWheeler(banana) = %q, %d", last, idx)
	}
	if got := compression.Rotations("abc"); !reflect.DeepEqual(got, []string{"abc", "bca", "cab"}) {
		t.Errorf("Rotations = %v", got)
	}
	if got := compression.MoveToFront([]byte("bbba")); !reflect.DeepEqual(got, []byte{98, 0, 0, 98}) {
		t.Errorf("MoveToFront = %v", got)
	}
	if r := compression.Ratio(100, 25); r != 4 || compression.Ratio(1, 0) != 0 {
		t.Errorf("Ratio = %v", r)
	}
	// BWT then MTF turns repetition into small numbers.
	last, _ := compression.BurrowsWheeler("abababababababab")
	zeros := bytes.Count(compression.MoveToFront([]byte(last)), []byte{0})
	if zeros < 12 {
		t.Errorf("BWT+MTF of a repeated text has %d zeros", zeros)
	}
}

func TestCorrupt(t *testing.T) {
	cases := []error{}
	_, err := compression.RunLengthDecode([]compression.Run{{'a', 0}})
	cases = append(cases, err)
	_, err = compression.LZ77Decode([]compression.Token{{3, 2, 'a'}})
	cases = append(cases, err)
	_, err = compression.LZWDecode([]int{300})
	cases = append(cases, err)
	_, err = compression.LZWDecode([]int{-1})
	cases = append(cases, err)
	_, err = compression.InverseBurrowsWheeler("abc", 3)
	cases = append(cases, err)
	for i, err := range cases {
		if !errors.Is(err, compression.ErrCorrupt) {
			t.Errorf("case %d: error %v, want ErrCorrupt", i, err)
		}
	}
	if s, err := compression.InverseBurrowsWheeler("", -1); s != "" || err != nil {
		t.Errorf("inverse of empty = %q, %v", s, err)
	}
}
