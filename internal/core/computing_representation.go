package core

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/computing/representation"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// floatRows draws a floating-point number as its three fields.
func floatRows(x representation.Float) []BitRow {
	s, e, f := x.Fields()
	return []BitRow{{
		Label: x.Format.Name,
		Fields: []BitField{
			{Name: "sign", Bits: s, Role: "sign"},
			{Name: "exponent", Bits: e, Role: "exponent"},
			{Name: "fraction", Bits: f, Role: "fraction"},
		},
		Note: fmt.Sprintf("%s: exponent %d - %d = %d", x.Class, x.Exponent, x.Format.Bias(), x.Unbiased),
	}}
}

func floatLines(x representation.Float) []string {
	s, e, f := x.Fields()
	lines := []string{
		fmt.Sprintf("bits: %s %s %s (0x%0*X)", s, e, f, x.Format.Width()/4, x.Bits),
		fmt.Sprintf("class: %s", x.Class),
	}
	switch x.Class {
	case "normal":
		lines = append(lines, fmt.Sprintf("value = (-1)^%s * 1.%s (binary) * 2^(%d - %d)", s, strings.TrimRight(f, "0"), x.Exponent, x.Format.Bias()))
	case "subnormal":
		lines = append(lines, fmt.Sprintf("value = (-1)^%s * 0.%s (binary) * 2^(%d), a subnormal: no hidden 1", s, strings.TrimRight(f, "0"), x.Unbiased))
	}
	return append(lines, "exact value stored: "+x.Exact)
}

func parseFloat(text string) (float64, error) {
	v, err := strconv.ParseFloat(strings.TrimSpace(text), 64)
	if err != nil {
		return 0, inputError(fmt.Sprintf("%q is not a number", strings.TrimSpace(text)))
	}
	return v, nil
}

func oneNumber(in Input, what string) (int, error) {
	if len(in.Values) != 1 {
		return 0, inputError("give one " + what)
	}
	return in.Values[0], nil
}

func representationEntries() []Entry {
	base := Option{"base", "base", 2, 2, 36}
	bitsOpt := func(def int) Option { return Option{"bits", "bits", def, 2, 32} }

	toBase := computingEntry("representation", "base-conversion", "Base Conversion", "bases.go", "ToBase",
		"Write a number in binary, octal, hex or any base by repeated division: the remainders, read from last to first, are the digits.",
		"O(log n)", "O(log n)", Number, &Input{Values: []int{156}, Options: map[string]int{"base": 2}})
	toBase.Options = []Option{base}
	toBase.Limits = "a number of 0 or more and a base from 2 to 36"
	toBase.Run = func(in Input, _ trace.Recorder) (Output, error) {
		n, err := oneNumber(in, "number of 0 or more")
		if err != nil {
			return Output{}, err
		}
		b := in.opt("base", 2)
		s, steps, err := representation.ToBase(n, b)
		if err != nil {
			return Output{}, wrap(err)
		}
		t := &TableOutput{Cols: 3, ColLabels: []string{"divide", "quotient", "remainder"}}
		for _, d := range steps {
			t.Rows++
			t.RowLabels = append(t.RowLabels, fmt.Sprintf("%d / %d", d.N, b))
			t.Cells = append(t.Cells, d.N, d.Quotient, d.Remainder)
		}
		return Output{Text: s, Table: t, Lines: []string{fmt.Sprintf("%d in base %d is %s: the remainders read from the bottom up", n, b, s)}}, nil
	}
	toBase.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("representation.ToBase(%d, %d)", in.Values[0], in.opt("base", 2))
	}

	fromBase := computingEntry("representation", "base-to-decimal", "Base to Decimal (Horner)", "bases.go", "FromBase",
		"Read a number written in any base: start at zero and, for each digit from the left, multiply by the base and add the digit.",
		"O(digits)", "O(digits)", Text, &Input{Text: "9C", Options: map[string]int{"base": 16}})
	fromBase.Options = []Option{{"base", "base", 16, 2, 36}}
	fromBase.Limits = "1 to 40 digits of the chosen base"
	fromBase.Run = func(in Input, _ trace.Recorder) (Output, error) {
		b := in.opt("base", 16)
		v, steps, err := representation.FromBase(in.Text, b)
		if err != nil {
			return Output{}, wrap(err)
		}
		lines := make([]string, len(steps))
		for i, h := range steps {
			lines[i] = fmt.Sprintf("%d * %d + %d = %d", h.Value, b, h.Digit, h.Result)
		}
		return Output{Number: &v, Lines: lines}, nil
	}
	fromBase.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("representation.FromBase(%q, %d)", in.Text, in.opt("base", 16))
	}

	add := computingEntry("representation", "binary-addition", "Binary Addition", "arithmetic.go", "AddBinary",
		"Add two binary numbers the way you add on paper, column by column from the right: 1 + 1 is 0, carry 1.",
		"O(n)", "O(n)", TwoTexts, &Input{Text: "1011", Other: "111"})
	add.Traced = true
	add.Limits = "two binary numbers of 1 to 64 digits"
	add.Run = func(in Input, r trace.Recorder) (Output, error) {
		a, b := strings.TrimSpace(in.Text), strings.TrimSpace(in.Other)
		n := max(len(a), len(b))
		cols := n + 1
		// The table has the carries, A, B and the sum in rows, columns from
		// the left; column i from the right is table column cols-1-i.
		sum, err := representation.TraceAddBinary(a, b, columnRecorder{r: r, cols: cols})
		if err != nil {
			return Output{}, wrap(err)
		}
		t := &TableOutput{Rows: 4, Cols: cols, RowLabels: []string{"carry", "A", "B", "sum"}}
		blank := -1
		t.Blank = &blank
		t.Cells = make([]int, 4*cols)
		for i := range t.Cells {
			t.Cells[i] = blank
		}
		pad := func(s string) string { return strings.Repeat(" ", cols-len(s)) + s }
		for row, s := range []string{"", pad(sum.A), pad(sum.B)} {
			for c, ch := range s {
				if ch == '0' || ch == '1' {
					t.Cells[row*cols+c] = int(ch - '0')
				}
			}
		}
		for i, c := range sum.Carries {
			if c {
				t.Cells[cols-1-i] = 1
			}
		}
		res := pad(sum.Result)
		for c, ch := range res {
			if ch == '0' || ch == '1' {
				t.Cells[3*cols+c] = int(ch - '0')
			}
		}
		va, _ := strconv.ParseUint(a, 2, 64)
		vb, _ := strconv.ParseUint(b, 2, 64)
		return Output{Text: sum.Result, Table: t, Lines: []string{fmt.Sprintf("%s + %s = %s (%d + %d = %d)", a, b, sum.Result, va, vb, va+vb)}}, nil
	}
	add.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("representation.AddBinary(%q, %q)", in.Text, in.Other)
	}

	signed := computingEntry("representation", "signed-numbers", "Signed Number Encodings", "signed.go", "Encodings",
		"Write a negative number three ways: sign-magnitude, ones' complement and two's complement, and see why two's complement won (one zero, and the same adder for signed and unsigned).",
		"O(n)", "O(n)", Number, &Input{Values: []int{-5}, Options: map[string]int{"bits": 8}})
	signed.Options = []Option{bitsOpt(8)}
	signed.Limits = "a whole number, positive or negative, and a width of 2 to 32 bits"
	signed.Run = func(in Input, _ trace.Recorder) (Output, error) {
		v, err := oneNumber(in, "whole number")
		if err != nil {
			return Output{}, err
		}
		n := in.opt("bits", 8)
		s, err := representation.Encodings(v, n)
		if err != nil {
			return Output{}, wrap(err)
		}
		row := func(label, bits, note string) BitRow {
			if bits == "" {
				return BitRow{Label: label, Note: "does not fit"}
			}
			return BitRow{Label: label, Fields: []BitField{{Name: "sign", Bits: bits[:1], Role: "sign"}, {Name: "rest", Bits: bits[1:], Role: "value"}}, Note: note}
		}
		return Output{Bits: []BitRow{
			row("sign-magnitude", s.SignMagnitude, "sign bit, then the magnitude"),
			row("ones' complement", s.OnesComp, "negative: every bit of the magnitude flipped"),
			row("two's complement", s.TwosComp, "negative: flip every bit, then add 1"),
		}, Lines: []string{
			fmt.Sprintf("sign-magnitude and ones' complement hold %d to %d on %d bits, with two zeros", s.MinSM, s.MaxSM, n),
			fmt.Sprintf("two's complement holds %d to %d, with one zero", s.MinTwo, s.MaxTwo),
		}}, nil
	}
	signed.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("representation.Encodings(%d, %d)", in.Values[0], in.opt("bits", 8))
	}

	overflow := computingEntry("representation", "twos-complement-overflow", "Two's Complement Overflow", "signed.go", "TwosAdd",
		"Add two signed numbers on a fixed number of bits: past the largest value the result wraps around to a negative, and the overflow flag says so. A carry out is not the same thing.",
		"O(n)", "O(n)", TwoNumbers, &Input{Values: []int{100, 50}, Options: map[string]int{"bits": 8}})
	overflow.Options = []Option{bitsOpt(8)}
	overflow.Limits = "two numbers that fit in the chosen width"
	overflow.Run = func(in Input, _ trace.Recorder) (Output, error) {
		if len(in.Values) != 2 {
			return Output{}, inputError("give two numbers")
		}
		o, err := representation.TwosAdd(in.Values[0], in.Values[1], in.opt("bits", 8))
		if err != nil {
			return Output{}, wrap(err)
		}
		row := func(label, bits string) BitRow {
			return BitRow{Label: label, Fields: []BitField{{Name: "sign", Bits: bits[:1], Role: "sign"}, {Name: "rest", Bits: bits[1:], Role: "value"}}}
		}
		verdict := "no overflow"
		if o.Overflowed {
			verdict = fmt.Sprintf("overflow: the true sum %d does not fit, so it wrapped to %d", o.TrueSum, o.SumValue)
		}
		return Output{Number: &o.SumValue, Bool: &o.Overflowed,
			Bits:  []BitRow{row(strconv.Itoa(in.Values[0]), o.A), row(strconv.Itoa(in.Values[1]), o.B), row("sum", o.Sum)},
			Lines: []string{verdict, fmt.Sprintf("carry out of the top bit: %d", b2int(o.Carry))}}, nil
	}
	overflow.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("representation.TwosAdd(%d, %d, %d)", in.Values[0], in.Values[1], in.opt("bits", 8))
	}

	fixed := computingEntry("representation", "fixed-point", "Fixed Point", "signed.go", "FixedPoint",
		"Store a fraction as an integer with an imaginary binary point: fast and exact in steps of 2^-f, the way DSPs, audio and old games do arithmetic.",
		"O(bits)", "O(bits)", Text, &Input{Text: "3.14159", Options: map[string]int{"int": 4, "frac": 4}})
	fixed.Options = []Option{{"int", "integer bits (with the sign)", 4, 1, 31}, {"frac", "fraction bits", 4, 0, 31}}
	fixed.Limits = "a number, and at most 32 bits in all"
	fixed.Run = func(in Input, _ trace.Recorder) (Output, error) {
		v, err := parseFloat(in.Text)
		if err != nil {
			return Output{}, err
		}
		ib, fb := in.opt("int", 4), in.opt("frac", 4)
		f, err := representation.FixedPoint(v, ib, fb)
		if err != nil {
			return Output{}, wrap(err)
		}
		ip, fp, _ := strings.Cut(f.Bits, ".")
		fields := []BitField{{Name: "integer", Bits: ip, Role: "value"}}
		if fp != "" {
			fields = append(fields, BitField{Name: "fraction", Bits: fp, Role: "fraction"})
		}
		return Output{Text: f.Bits, Bits: []BitRow{{Label: fmt.Sprintf("Q%d.%d", ib, fb), Fields: fields}},
			Lines: []string{
				fmt.Sprintf("stored integer: %d (that is %s / 2^%d)", f.Raw, strings.TrimSpace(in.Text), fb),
				fmt.Sprintf("value held: %g, error %.3g", f.Value, f.Err),
			}}, nil
	}
	fixed.Call = func(in Input) (string, string) {
		v, _ := parseFloat(in.Text)
		return "", fmt.Sprintf("representation.FixedPoint(%v, %d, %d)", v, in.opt("int", 4), in.opt("frac", 4))
	}

	float := func(id, name, fn, summary string, ex string, encode func(float64) representation.Float) Entry {
		e := computingEntry("representation", id, name, "float.go", fn, summary, "O(1)", "O(1)", Text, &Input{Text: ex})
		e.Limits = "a number, like 0.1, -2.5, 1e-40 or inf"
		e.Run = func(in Input, _ trace.Recorder) (Output, error) {
			v, err := parseFloat(in.Text)
			if err != nil {
				return Output{}, err
			}
			x := encode(v)
			return Output{Bits: floatRows(x), Lines: floatLines(x), Text: x.Exact}, nil
		}
		e.Call = func(in Input) (string, string) {
			v, _ := parseFloat(in.Text)
			return "", fmt.Sprintf("representation.%s(%v).Exact", fn, v)
		}
		return e
	}
	single := float("ieee754-single", "IEEE 754 Single Precision", "EncodeSingle",
		"How a float is stored in 32 bits: a sign, an 8-bit exponent with a bias of 127, and 23 fraction bits after a hidden leading 1, and the exact value that really holds.",
		"0.1", representation.EncodeSingle)
	half := float("ieee754-half", "IEEE 754 Half Precision", "EncodeHalf",
		"The 16-bit float of GPUs and machine learning: 5 exponent bits, 10 fraction bits, a range up to 65504, and rounding you can see.",
		"3.14159", representation.EncodeHalf)

	decode := computingEntry("representation", "float-decode", "Decode Float Bits", "float.go", "Decode",
		"Go the other way: give 32 bits and read off the sign, exponent and fraction, including zeros, subnormals, infinities and NaN.",
		"O(1)", "O(1)", Text, &Input{Text: "01000000010010010000111111011011"})
	decode.Limits = "32 bits of single precision, or 16 bits of half precision"
	decode.Run = func(in Input, _ trace.Recorder) (Output, error) {
		s := strings.Join(strings.Fields(in.Text), "")
		f := representation.Single
		if len(s) == representation.Half.Width() {
			f = representation.Half
		}
		bits, err := representation.ParseBits(f, s)
		if err != nil {
			return Output{}, wrap(err)
		}
		x := representation.Decode(f, bits)
		return Output{Bits: floatRows(x), Lines: floatLines(x), Text: x.Exact}, nil
	}
	decode.Call = func(in Input) (string, string) {
		s := strings.Join(strings.Fields(in.Text), "")
		return fmt.Sprintf("bits, err := representation.ParseBits(representation.Single, %q)\n\tif err != nil {\n\t\tpanic(err)\n\t}\n\t", s),
			"representation.Decode(representation.Single, bits).Exact"
	}

	rounding := computingEntry("representation", "float-rounding", "Why 0.1 + 0.2 != 0.3", "float.go", "Rounding",
		"Decimal fractions like 0.1 have no exact binary form, so each is stored as the nearest double; the small errors add up and 0.1 + 0.2 lands one step away from 0.3.",
		"O(1)", "O(1)", Text, &Input{Text: "0.1 0.2 0.3"})
	rounding.Limits = "three numbers a, b and c, to compare a + b with c"
	parse3 := func(text string) ([3]float64, error) {
		f := strings.Fields(text)
		var v [3]float64
		if len(f) != 3 {
			return v, inputError("give three numbers: a, b and c")
		}
		for i := range v {
			x, err := parseFloat(f[i])
			if err != nil {
				return v, err
			}
			v[i] = x
		}
		return v, nil
	}
	rounding.Run = func(in Input, _ trace.Recorder) (Output, error) {
		v, err := parse3(in.Text)
		if err != nil {
			return Output{}, err
		}
		a, b, s, c, eq := representation.Rounding(v[0], v[1], v[2])
		return Output{Bool: &eq, Lines: []string{
			"a is stored as       " + a,
			"b is stored as       " + b,
			"a + b is stored as   " + s,
			"c is stored as       " + c,
			fmt.Sprintf("a + b == c: %v", eq),
		}, Bits: append(floatRows(representation.EncodeDouble(v[0]+v[1])), floatRows(representation.EncodeDouble(v[2]))...)}, nil
	}
	rounding.Call = func(in Input) (string, string) {
		v, _ := parse3(in.Text)
		return "", fmt.Sprintf("representation.Rounding(%v, %v, %v)", v[0], v[1], v[2])
	}

	bcd := computingEntry("representation", "bcd", "Binary-Coded Decimal", "codes.go", "BCD",
		"Store each decimal digit in its own 4 bits, so a clock or calculator can show a number without dividing it: 409 is 0100 0000 1001.",
		"O(digits)", "O(digits)", Number, &Input{Values: []int{409}})
	bcd.Limits = "a number of 0 or more"
	bcd.Run = func(in Input, _ trace.Recorder) (Output, error) {
		v, err := oneNumber(in, "number of 0 or more")
		if err != nil {
			return Output{}, err
		}
		groups, err := representation.BCD(v)
		if err != nil {
			return Output{}, wrap(err)
		}
		var fields []BitField
		for i, g := range groups {
			fields = append(fields, BitField{Name: string(strconv.Itoa(v)[i]), Bits: g, Role: "data"})
		}
		bin := strconv.FormatInt(int64(v), 2)
		return Output{Text: strings.Join(groups, " "), Bits: []BitRow{
			{Label: "BCD", Fields: fields, Note: fmt.Sprintf("%d bits", 4*len(groups))},
			{Label: "binary", Fields: []BitField{{Name: "value", Bits: bin, Role: "value"}}, Note: fmt.Sprintf("%d bits", len(bin))},
		}}, nil
	}
	bcd.Call = func(in Input) (string, string) { return "", fmt.Sprintf("representation.BCD(%d)", in.Values[0]) }

	parity := computingEntry("representation", "parity-bit", "Parity Bit", "codes.go", "Parity",
		"Add one bit that makes the count of 1s even: the receiver recounts, and any single flipped bit shows up (two flipped bits do not).",
		"O(n)", "O(1)", Text, &Input{Text: "1011001"})
	parity.Limits = "1 to 64 bits"
	parity.Run = func(in Input, _ trace.Recorder) (Output, error) {
		s := strings.Join(strings.Fields(in.Text), "")
		if s == "" || len(s) > 64 || strings.Trim(s, "01") != "" {
			return Output{}, inputError("give 1 to 64 bits")
		}
		v, _ := strconv.ParseUint(s, 2, 64)
		p := representation.Parity(v, len(s))
		return Output{Number: &p, Bits: []BitRow{{Label: "sent", Fields: []BitField{{Name: "data", Bits: s, Role: "data"}, {Name: "parity", Bits: strconv.Itoa(p), Role: "parity"}}}},
			Lines: []string{fmt.Sprintf("%d ones, so the even parity bit is %d", strings.Count(s, "1"), p)}}, nil
	}
	parity.Call = func(in Input) (string, string) {
		s := strings.Join(strings.Fields(in.Text), "")
		v, _ := strconv.ParseUint(s, 2, 64)
		return "", fmt.Sprintf("representation.Parity(%d, %d)", v, len(s))
	}

	hamming := computingEntry("representation", "hamming-code", "Hamming(7,4) Code", "codes.go", "Hamming74",
		"Protect 4 data bits with 3 parity bits placed at positions 1, 2 and 4: when one bit flips, the failing checks spell out its position, so the receiver fixes it.",
		"O(1)", "O(1)", Number, &Input{Values: []int{11}, Options: map[string]int{"flip": 6}})
	hamming.Options = []Option{{"flip", "position to flip in transit (0 for none)", 6, 0, 7}}
	hamming.Limits = "4 data bits as a number from 0 to 15"
	hamming.Run = func(in Input, _ trace.Recorder) (Output, error) {
		d, err := oneNumber(in, "value from 0 to 15")
		if err != nil {
			return Output{}, err
		}
		if d < 0 || d > 15 {
			return Output{}, inputError("give 4 data bits: a value from 0 to 15")
		}
		flip := in.opt("flip", 0)
		code := representation.Hamming74(d)
		received := code
		if flip > 0 {
			received ^= 1 << (7 - flip)
		}
		data, syndrome, corrected := representation.HammingDecode(received)
		positions := func(c int, mark int) []BitField {
			var f []BitField
			for pos := 1; pos <= 7; pos++ {
				role := "data"
				if pos == 1 || pos == 2 || pos == 4 {
					role = "parity"
				}
				if pos == mark {
					role = "error"
				}
				name := map[int]string{1: "p1", 2: "p2", 3: "d1", 4: "p3", 5: "d2", 6: "d3", 7: "d4"}[pos]
				f = append(f, BitField{Name: name, Bits: strconv.Itoa(c >> (7 - pos) & 1), Role: role})
			}
			return f
		}
		lines := []string{fmt.Sprintf("data %04b is sent as %07b", d, code)}
		if flip > 0 {
			lines = append(lines, fmt.Sprintf("position %d flipped in transit: received %07b", flip, received))
		}
		lines = append(lines, fmt.Sprintf("syndrome %03b = %d: %s", syndrome, syndrome, map[bool]string{true: "no error", false: "the error is at that position"}[syndrome == 0]),
			fmt.Sprintf("corrected %07b, data %04b", corrected, data))
		return Output{Number: &data, Bits: []BitRow{
			{Label: "data", Fields: []BitField{{Name: "d1-d4", Bits: fmt.Sprintf("%04b", d), Role: "data"}}},
			{Label: "sent", Fields: positions(code, 0)},
			{Label: "received", Fields: positions(received, flip)},
			{Label: "corrected", Fields: positions(corrected, 0)},
		}, Lines: lines, Summary: map[string]float64{"syndrome": float64(syndrome)}}, nil
	}
	hamming.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("representation.HammingDecode(representation.Hamming74(%d) ^ %d)", in.Values[0], hammingFlipMask(in.opt("flip", 0)))
	}

	utf := computingEntry("representation", "utf8-encoding", "UTF-8 Encoding", "text.go", "UTF8",
		"How text becomes bytes: ASCII letters stay one byte, and every other character takes 2 to 4, with leading bits that say how many.",
		"O(n)", "O(n)", Text, &Input{Text: "Aé€😀"})
	utf.Limits = "a text of at most 64 characters"
	utf.Run = func(in Input, _ trace.Recorder) (Output, error) {
		if in.Text == "" || len([]rune(in.Text)) > 64 {
			return Output{}, inputError("give a text of 1 to 64 characters")
		}
		cs, err := representation.UTF8(in.Text)
		if err != nil {
			return Output{}, wrap(err)
		}
		var rows []BitRow
		var lines []string
		total := 0
		for _, c := range cs {
			var f []BitField
			var hex []string
			for i, b := range c.Bytes {
				name := "lead"
				if i > 0 {
					name = "cont"
				}
				if len(c.Bytes) == 1 {
					name = "ascii"
				}
				f = append(f, BitField{Name: name, Bits: fmt.Sprintf("%08b", b), Role: "byte"})
				hex = append(hex, fmt.Sprintf("%02X", b))
			}
			total += len(c.Bytes)
			rows = append(rows, BitRow{Label: fmt.Sprintf("%q %s", c.Rune, c.CodePoint), Fields: f, Note: c.Pattern})
			lines = append(lines, fmt.Sprintf("%q is %s: %d byte(s), %s", c.Rune, c.CodePoint, len(c.Bytes), strings.Join(hex, " ")))
		}
		return Output{Bits: rows, Lines: lines, Number: &total}, nil
	}
	utf.Call = func(in Input) (string, string) { return "", fmt.Sprintf("representation.UTF8(%q)", in.Text) }

	endian := computingEntry("representation", "endianness", "Endianness", "text.go", "Endian",
		"The same 32-bit number laid out in memory two ways: big-endian puts the most significant byte first (network order), little-endian the least (x86, ARM).",
		"O(1)", "O(1)", Number, &Input{Values: []int{305419896}})
	endian.Limits = "a number from 0 to 4294967295"
	endian.Run = func(in Input, _ trace.Recorder) (Output, error) {
		v, err := oneNumber(in, "number from 0 to 4294967295")
		if err != nil {
			return Output{}, err
		}
		if v < 0 || v > 1<<32-1 {
			return Output{}, inputError("give a number from 0 to 4294967295")
		}
		big, little := representation.Endian(uint32(v))
		row := func(label string, b [4]byte) BitRow {
			var f []BitField
			for i, x := range b {
				f = append(f, BitField{Name: fmt.Sprintf("+%d: %02X", i, x), Bits: fmt.Sprintf("%08b", x), Role: "byte"})
			}
			return BitRow{Label: label, Fields: f}
		}
		return Output{Bits: []BitRow{row("big-endian", big), row("little-endian", little)},
			Lines: []string{fmt.Sprintf("0x%08X in memory, address +0 first:", v),
				fmt.Sprintf("big-endian    % X", big[:]), fmt.Sprintf("little-endian % X", little[:])}}, nil
	}
	endian.Call = func(in Input) (string, string) { return "", fmt.Sprintf("representation.Endian(%d)", in.Values[0]) }

	return []Entry{toBase, fromBase, add, signed, overflow, fixed, single, half, decode, rounding, bcd, parity, hamming, utf, endian}
}

func hammingFlipMask(pos int) int {
	if pos <= 0 {
		return 0
	}
	return 1 << (7 - pos)
}

// columnRecorder turns binary addition's per-column events into table
// events: column i from the right of an n+1 wide table with rows carry, A,
// B and sum.
type columnRecorder struct {
	r    trace.Recorder
	cols int
}

func (c columnRecorder) Record(e trace.Event) {
	if len(e.Indices) == 0 {
		return
	}
	col := c.cols - 1 - e.Indices[0]
	switch e.Kind {
	case trace.Visit:
		// Reading the column: A, B and the carry into it.
		c.r.Record(trace.Event{Kind: trace.Visit, Indices: []int{1*c.cols + col}})
		c.r.Record(trace.Event{Kind: trace.Visit, Indices: []int{2*c.cols + col}})
		if len(e.Values) == 3 && e.Values[2] == 1 {
			c.r.Record(trace.Event{Kind: trace.Set, Indices: []int{col}, Values: []int{1}})
		}
	case trace.Set:
		c.r.Record(trace.Event{Kind: trace.Set, Indices: []int{3*c.cols + col}, Values: e.Values})
	default:
	}
}
