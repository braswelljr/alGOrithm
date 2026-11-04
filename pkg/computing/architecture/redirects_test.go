package architecture

import "testing"

// TestRedirects checks which instructions can change the program counter: a
// jump, a call and a return can; a no-op and a halt cannot.
func TestRedirects(t *testing.T) {
	p := mustAssemble(t, "  JMP end\nstart: NOP\nend: CALL f\n HALT\nf: RET")
	want := map[string]bool{"JMP": true, "NOP": false, "CALL": true, "HALT": false, "RET": true}
	for _, in := range p.Code {
		if w, ok := want[in.Op]; ok && in.Redirects() != w {
			t.Errorf("%s.Redirects() = %v, want %v", in.Op, !w, w)
		}
	}
}
