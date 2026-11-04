package logic

import "strconv"

// NetGate is one gate of a netlist: its Name, its Kind ("input", "const0",
// "const1", "not", "and", "or", "xor", "nand", "nor", "xnor" or "output"),
// and the names of the gates feeding it.
type NetGate struct {
	Name string
	Kind string
	In   []string
}

// Netlist turns e into gates, the way a synthesis tool does in miniature: one
// input per variable, one gate per operator, and an output "F". Shared
// variables feed every gate that reads them; equal subexpressions are not
// merged.
//
// Time: O(size of e). Space: O(size of e).
func Netlist(e *Expr) []NetGate {
	var gates []NetGate
	for _, v := range e.Vars() {
		gates = append(gates, NetGate{Name: v, Kind: "input"})
	}
	n := 0
	var build func(*Expr) string
	build = func(x *Expr) string {
		switch x.Op {
		case OpVar:
			return x.Name
		case OpConst:
			n++
			name := "k" + strconv.Itoa(n)
			kind := "const0"
			if x.Value {
				kind = "const1"
			}
			gates = append(gates, NetGate{Name: name, Kind: kind})
			return name
		default: // an operator, below
		}
		var in []string
		for _, a := range x.Args {
			in = append(in, build(a))
		}
		n++
		name := "g" + strconv.Itoa(n)
		gates = append(gates, NetGate{Name: name, Kind: string(x.Op), In: in})
		return name
	}
	out := build(e)
	return append(gates, NetGate{Name: "F", Kind: "output", In: []string{out}})
}
