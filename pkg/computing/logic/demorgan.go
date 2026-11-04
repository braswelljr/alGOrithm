package logic

// DeMorgan pushes every negation in e down to the variables (negation normal
// form), using De Morgan's laws, !(A & B) = !A | !B and !(A | B) = !A & !B,
// and !!A = A. NAND, NOR and XNOR are first read as the negated AND, OR and
// XOR they are, and a negated XOR becomes an XOR with one side negated. It
// returns the new expression and each law applied, in words.
//
// Time: O(size of e). Space: O(size of e).
func DeMorgan(e *Expr) (*Expr, []string) {
	var steps []string
	return push(e, false, &steps), steps
}

func push(e *Expr, neg bool, steps *[]string) *Expr {
	switch e.Op {
	case OpVar:
		if neg {
			return &Expr{Op: OpNot, Args: []*Expr{e}}
		}
		return e
	case OpConst:
		return &Expr{Op: OpConst, Value: e.Value != neg}
	case OpNot:
		if neg {
			*steps = append(*steps, "!!"+e.Args[0].format(4)+" = "+e.Args[0].String()+" (double negation)")
		}
		return push(e.Args[0], !neg, steps)
	case OpNand, OpNor, OpXnor:
		base := map[Op]Op{OpNand: OpAnd, OpNor: OpOr, OpXnor: OpXor}[e.Op]
		inner := &Expr{Op: base, Args: e.Args}
		*steps = append(*steps, e.String()+" = !("+inner.String()+")")
		return push(inner, !neg, steps)
	case OpXor:
		if neg {
			*steps = append(*steps, "!("+e.String()+") = !"+e.Args[0].format(4)+" ^ "+e.Args[1].format(3))
			return &Expr{Op: OpXor, Args: []*Expr{push(e.Args[0], true, steps), push(e.Args[1], false, steps)}}
		}
		return &Expr{Op: OpXor, Args: []*Expr{push(e.Args[0], false, steps), push(e.Args[1], false, steps)}}
	default: // OpAnd and OpOr, below
	}
	// OpAnd or OpOr.
	op := e.Op
	if neg {
		flip := map[Op]Op{OpAnd: OpOr, OpOr: OpAnd}[op]
		sym := map[Op]string{OpAnd: " & ", OpOr: " | "}
		*steps = append(*steps, "!("+e.String()+") = !"+e.Args[0].format(4)+sym[flip]+"!"+e.Args[1].format(4)+" (De Morgan)")
		op = flip
	}
	return &Expr{Op: op, Args: []*Expr{push(e.Args[0], neg, steps), push(e.Args[1], neg, steps)}}
}
