package architecture

import (
	"errors"
	"math"
	"slices"
	"strings"
	"testing"
)

const sumProgram = `; add 1..5 and print the total
      LI r1, 0        ; total
      LI r2, 5        ; counter
      LI r3, 0
loop: ADD r1, r1, r2
      ADDI r2, r2, -1
      BNE r2, r3, loop
      OUT r1
      ST r1, 0(r0)
      HALT`

func mustAssemble(t *testing.T, src string) Program {
	t.Helper()
	p, err := Assemble(src)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestAssembleResolvesLabels(t *testing.T) {
	p := mustAssemble(t, "  JMP end\nstart: NOP\nend: CALL f\n HALT\nf: RET")
	if p.Labels["end"] != 2 || p.Labels["f"] != 4 || p.Code[0].Imm != 2 || p.Code[2].Imm != 4 {
		t.Errorf("labels %v, code %v", p.Labels, p.Code)
	}
	if !strings.Contains(p.Listing()[2], "end: CALL f") {
		t.Errorf("listing %v", p.Listing())
	}
	for _, bad := range []string{"JMP nowhere", "ADD r1, r2", "LI r9, 1", "x: NOP\nx: NOP", "FOO r1"} {
		if _, err := Assemble(bad); !errors.Is(err, ErrInput) {
			t.Errorf("%q: %v", bad, err)
		}
	}
}

func TestSimulate(t *testing.T) {
	r, err := Simulate(mustAssemble(t, sumProgram), 16, 1000)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(r.Out, []int{15}) || r.Mem[0] != 15 || r.Regs[2] != 0 {
		t.Errorf("out %v, mem[0] %d, r2 %d", r.Out, r.Mem[0], r.Regs[2])
	}
	if len(r.Cycles) != 4*len(r.Trace) || r.Cycles[0].Phase != Fetch || r.Cycles[3].Phase != Writeback {
		t.Errorf("four phases per instruction: %d cycles, %d instructions", len(r.Cycles), len(r.Trace))
	}
	calls := mustAssemble(t, "LI r1, 3\nCALL double\nOUT r1\nHALT\ndouble: ADD r1, r1, r1\nRET")
	r, err = Simulate(calls, 16, 100)
	if err != nil || !slices.Equal(r.Out, []int{6}) || r.Regs[SP] != 15 {
		t.Errorf("call and return: %v, sp %d, %v", r.Out, r.Regs[SP], err)
	}
	if _, err := Simulate(mustAssemble(t, "x: JMP x"), 16, 50); !errors.Is(err, ErrRuntime) {
		t.Errorf("an endless loop stops: %v", err)
	}
}

func TestPipelineHazards(t *testing.T) {
	dep := mustAssemble(t, "LI r1, 5\nADD r2, r1, r1\nSUB r3, r2, r1\nHALT")
	slow, err := RunPipeline(dep, false, 100)
	if err != nil {
		t.Fatal(err)
	}
	if slow.Stalls != 4 || slow.Cycles != 12 || slow.CPI() != 3 {
		t.Errorf("without forwarding: %d stalls, %d cycles, CPI %.2f; want 4, 12, 3", slow.Stalls, slow.Cycles, slow.CPI())
	}
	fast, _ := RunPipeline(dep, true, 100)
	if fast.Stalls != 0 || fast.Cycles != 8 || len(fast.Forwards) != 3 {
		t.Errorf("with forwarding: %d stalls, %d cycles, %d forwards; want 0, 8, 3", fast.Stalls, fast.Cycles, len(fast.Forwards))
	}
	load, _ := RunPipeline(mustAssemble(t, "LD r1, 0(r0)\nADD r2, r1, r1\nHALT"), true, 100)
	if load.Stalls != 1 {
		t.Errorf("load-use costs one stall even with forwarding: %d", load.Stalls)
	}
	if got := strings.Join(slow.Cells[1], " "); !strings.HasPrefix(strings.TrimSpace(got), "IF ID stall stall EX MEM WB") {
		t.Errorf("row 2: %q", got)
	}
	branchy, _ := RunPipeline(mustAssemble(t, "LI r1, 1\nJMP over\nLI r1, 2\nLI r1, 3\nover: OUT r1\nHALT"), true, 100)
	flushed := 0
	for _, f := range branchy.Flushed {
		if f {
			flushed++
		}
	}
	if branchy.Flushes != 1 || flushed != 2 || branchy.Instructions != 4 {
		t.Errorf("a jump flushes the two instructions behind it: %d flushes, %d rows, %d instructions", branchy.Flushes, flushed, branchy.Instructions)
	}
}

func TestPredictors(t *testing.T) {
	trace, err := ParseBranches(strings.Repeat("TTTN ", 4))
	if err != nil {
		t.Fatal(err)
	}
	if _, c := OneBit(trace); c != 8 {
		t.Errorf("1-bit: %d of 16, want 8 (two misses per loop)", c)
	}
	if _, c := TwoBit(trace); c != 11 {
		t.Errorf("2-bit: %d of 16, want 11 (one miss per loop after the first)", c)
	}
	long, _ := ParseBranches(strings.Repeat("TTTN", 20))
	preds, _, err := Gshare(long, 4)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range preds[40:] {
		if p.Predicted != p.Taken {
			t.Fatal("gshare with 4 bits of history learns a loop of four")
		}
	}
	if _, err := ParseBranches("T X"); err == nil {
		t.Error("bad trace")
	}
}

func TestPerformance(t *testing.T) {
	mix, err := ParseMix("alu 50 1\nload 20 5\nstore 10 3\nbranch 20 2")
	if err != nil {
		t.Fatal(err)
	}
	if got := CPI(mix); math.Abs(got-2.2) > 1e-9 {
		t.Errorf("CPI %.2f, want 2.2", got)
	}
	if _, err := ParseMix("alu 50 1"); err == nil {
		t.Error("percents must add to 100")
	}
	if got := Amdahl(90, 10); math.Abs(got-1/(0.1+0.09)) > 1e-9 {
		t.Errorf("Amdahl 90%% on 10: %.3f", got)
	}
	if Amdahl(90, 1_000_000) >= 10 {
		t.Error("90% parallel never beats 10x")
	}
	vn, h, err := Buses(mustAssemble(t, sumProgram), 1000)
	if err != nil || vn != h+1 {
		t.Errorf("one store: von Neumann %d, Harvard %d (%v)", vn, h, err)
	}
}

func TestLayout(t *testing.T) {
	fields, err := ParseFields("a bool; b int64; c int16; d *int")
	if err != nil {
		t.Fatal(err)
	}
	s := Layout(fields)
	offsets := []int{s.Fields[0].Offset, s.Fields[1].Offset, s.Fields[2].Offset, s.Fields[3].Offset}
	if !slices.Equal(offsets, []int{0, 8, 16, 24}) || s.Size != 32 || s.Padding != 13 {
		t.Errorf("offsets %v, size %d, padding %d", offsets, s.Size, s.Padding)
	}
	packed := Layout(Packed(fields))
	if packed.Size != 24 || packed.Padding != 5 {
		t.Errorf("packed: size %d, padding %d", packed.Size, packed.Padding)
	}
}

func TestSystem(t *testing.T) {
	spans, steps, err := Interrupt(10, 3, 4)
	if err != nil || spans[len(spans)-1].End != 19 || len(steps) == 0 {
		t.Errorf("interrupt: %v, %v", spans, err)
	}
	_, _, overhead, err := ContextSwitch(2, 4, 1, 20)
	if err != nil || overhead <= 0 || overhead >= 0.5 {
		t.Errorf("switching overhead %.2f (%v)", overhead, err)
	}
	r, st, err := CallStack(4)
	if err != nil || r != 24 {
		t.Fatalf("fact(4) = %d (%v)", r, err)
	}
	deepest := 0
	for _, s := range st {
		deepest = max(deepest, len(s.Frames))
	}
	if deepest != 5 || st[len(st)-1].SP != 1000 {
		t.Errorf("fact(4) has main plus 4 frames at the deepest: %d; sp back to %d", deepest, st[len(st)-1].SP)
	}
}
