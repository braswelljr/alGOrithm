package representation

import (
	"encoding/binary"
	"errors"
	"math"
	"strconv"
	"testing"
)

func TestBases(t *testing.T) {
	for _, n := range []int{0, 1, 7, 255, 1000, 123456789} {
		for _, base := range []int{2, 8, 10, 16, 36} {
			s, steps, err := ToBase(n, base)
			if err != nil || s != stringsUpper(strconv.FormatInt(int64(n), base)) || len(steps) == 0 {
				t.Errorf("ToBase(%d, %d) = %q, %v", n, base, s, err)
			}
			v, _, err := FromBase(s, base)
			if err != nil || v != n {
				t.Errorf("FromBase(%q, %d) = %d, %v", s, base, v, err)
			}
		}
	}
	_, steps, _ := ToBase(13, 2)
	if steps[0] != (Division{13, 6, 1}) || len(steps) != 4 {
		t.Errorf("13 in binary: %v", steps)
	}
	for _, bad := range []struct {
		s string
		b int
	}{{"12", 2}, {"G", 16}, {"", 10}, {"1", 1}} {
		if _, _, err := FromBase(bad.s, bad.b); !errors.Is(err, ErrInput) {
			t.Errorf("FromBase(%q, %d) should fail", bad.s, bad.b)
		}
	}
}

func stringsUpper(s string) string {
	b := []byte(s)
	for i, c := range b {
		if c >= 'a' && c <= 'z' {
			b[i] = c - 32
		}
	}
	return string(b)
}

func TestAddBinary(t *testing.T) {
	s, err := AddBinary("1011", "111")
	if err != nil || s.Result != "10010" || !s.Carries[4] {
		t.Errorf("1011 + 111 = %+v, %v", s, err)
	}
	for a := 0; a < 64; a++ {
		for b := 0; b < 64; b++ {
			s, _ := AddBinary(strconv.FormatInt(int64(a), 2), strconv.FormatInt(int64(b), 2))
			if v, _ := strconv.ParseInt(s.Result, 2, 64); int(v) != a+b {
				t.Fatalf("%d + %d = %s", a, b, s.Result)
			}
		}
	}
}

func TestSignedEncodings(t *testing.T) {
	s, _ := Encodings(-5, 8)
	if s.SignMagnitude != "10000101" || s.OnesComp != "11111010" || s.TwosComp != "11111011" {
		t.Errorf("-5: %+v", s)
	}
	s, _ = Encodings(-128, 8)
	if s.TwosComp != "10000000" || s.SignMagnitude != "" {
		t.Errorf("-128 fits only in two's complement: %+v", s)
	}
	o, _ := TwosAdd(100, 50, 8)
	if !o.Overflowed || o.SumValue != -106 {
		t.Errorf("100 + 50 on 8 bits: %+v", o)
	}
	o, _ = TwosAdd(-1, 1, 8)
	if o.Overflowed || !o.Carry || o.SumValue != 0 {
		t.Errorf("-1 + 1 carries without overflow: %+v", o)
	}
	f, _ := FixedPoint(3.14159, 4, 4)
	if f.Raw != 50 || f.Bits != "0011.0010" || math.Abs(f.Err) > 1.0/32 {
		t.Errorf("fixed: %+v", f)
	}
	if _, err := FixedPoint(100, 4, 4); !errors.Is(err, ErrInput) {
		t.Error("100 does not fit in Q4.4")
	}
}

func TestIEEE754(t *testing.T) {
	for _, v := range []float64{0, 1, -2.5, 0.1, 3.14159, 1e-40, 1e38, math.Inf(1), math.Inf(-1)} {
		x := EncodeSingle(v)
		if x.Bits != uint64(math.Float32bits(float32(v))) {
			t.Errorf("%g: bits %x", v, x.Bits)
		}
		if x.Value != float64(float32(v)) {
			t.Errorf("%g: value %g", v, x.Value)
		}
	}
	one := EncodeSingle(1)
	if s, e, f := one.Fields(); s != "0" || e != "01111111" || f != "00000000000000000000000" || one.Class != "normal" {
		t.Errorf("1.0: %s %s %s %s", s, e, f, one.Class)
	}
	if EncodeSingle(1e-40).Class != "subnormal" || !math.IsNaN(Decode(Single, 0x7fc00000).Value) {
		t.Error("classes")
	}
	if EncodeSingle(0.1).Exact != "0.100000001490116119384765625" {
		t.Errorf("0.1 in single is %s", EncodeSingle(0.1).Exact)
	}
	// Half precision: known patterns.
	for v, want := range map[float64]uint64{1: 0x3c00, -2: 0xc000, 65504: 0x7bff, 1e6: 0x7c00, 0.1: 0x2e66, 5.960464477539063e-08: 0x0001} {
		if got := EncodeHalf(v).Bits; got != want {
			t.Errorf("half(%g) = %04x, want %04x", v, got, want)
		}
	}
	if EncodeHalf(2049).Value != 2048 { // a tie rounds to even
		t.Errorf("half(2049) = %g", EncodeHalf(2049).Value)
	}
	_, _, sum, c, eq := Rounding(0.1, 0.2, 0.3)
	if eq || sum == c {
		t.Error("0.1 + 0.2 != 0.3 in binary floating point")
	}
	if _, err := ParseBits(Half, "101"); !errors.Is(err, ErrInput) {
		t.Error("short bit string")
	}
}

func TestCodes(t *testing.T) {
	b, _ := BCD(409)
	if len(b) != 3 || b[0] != "0100" || b[1] != "0000" || b[2] != "1001" {
		t.Errorf("BCD(409) = %v", b)
	}
	if Parity(0b1011, 4) != 1 || Parity(0b1001, 4) != 0 {
		t.Error("parity")
	}
	for d := 0; d < 16; d++ {
		code := Hamming74(d)
		if got, s, _ := HammingDecode(code); got != d || s != 0 {
			t.Fatalf("clean %d: %d syndrome %d", d, got, s)
		}
		for pos := 1; pos <= 7; pos++ {
			bad := code ^ 1<<(7-pos)
			got, s, fixed := HammingDecode(bad)
			if got != d || s != pos || fixed != code {
				t.Fatalf("data %d, flipped position %d: got %d syndrome %d", d, pos, got, s)
			}
		}
	}
}

func TestText(t *testing.T) {
	cs, err := UTF8("Aé€😀")
	if err != nil || len(cs) != 4 {
		t.Fatal(err)
	}
	for i, n := range []int{1, 2, 3, 4} {
		if len(cs[i].Bytes) != n {
			t.Errorf("%q takes %d bytes", cs[i].Rune, len(cs[i].Bytes))
		}
	}
	if cs[1].CodePoint != "U+00E9" || cs[1].Bytes[0] != 0xC3 {
		t.Errorf("é: %+v", cs[1])
	}
	if _, err := UTF8("\xff"); !errors.Is(err, ErrInput) {
		t.Error("invalid UTF-8")
	}
	big, little := Endian(0x12345678)
	var want [4]byte
	binary.BigEndian.PutUint32(want[:], 0x12345678)
	if big != want || little[0] != 0x78 {
		t.Errorf("endian %x %x", big, little)
	}
}
