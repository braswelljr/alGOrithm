package automata

// PR is a primitive recursive function of any number of arguments, built
// only from the pieces below. Calls counts every application, to show how
// much work a definition costs.
type PR struct {
	Name  string
	Arity int
	f     func(c *int, args []int) int
}

// Apply calls the function and returns its value and how many piece
// applications it took.
func (p PR) Apply(args ...int) (value, calls int) {
	v := p.f(&calls, args)
	return v, calls
}

// Zero is the function that is always 0.
func Zero(arity int) PR {
	return PR{"zero", arity, func(c *int, _ []int) int { *c++; return 0 }}
}

// Succ is x + 1.
func Succ() PR { return PR{"succ", 1, func(c *int, a []int) int { *c++; return a[0] + 1 }} }

// Proj returns its i-th argument (from 0) of arity.
func Proj(arity, i int) PR {
	return PR{"proj", arity, func(c *int, a []int) int { *c++; return a[i] }}
}

// Compose is f(g1(args), ..., gk(args)).
func Compose(f PR, gs ...PR) PR {
	return PR{f.Name, gs[0].Arity, func(c *int, a []int) int {
		*c++
		inner := make([]int, len(gs))
		for i, g := range gs {
			inner[i] = g.f(c, a)
		}
		return f.f(c, inner)
	}}
}

// Rec is primitive recursion: h(0, args) = base(args), and
// h(n+1, args) = step(n, h(n, args), args). It is the only loop allowed, and
// its count is fixed before it starts: that is why every primitive recursive
// function halts.
func Rec(name string, base, step PR) PR {
	return PR{name, base.Arity + 1, func(c *int, a []int) int {
		*c++
		acc := base.f(c, a[1:])
		for n := 0; n < a[0]; n++ {
			acc = step.f(c, append([]int{n, acc}, a[1:]...))
		}
		return acc
	}}
}

// Add is n + m, built by primitive recursion on n from succ.
func Add() PR { return Rec("add", Proj(1, 0), Compose(Succ(), Proj(3, 1))) }

// Mul is n * m, by recursion on n from Add.
func Mul() PR {
	return Rec("mul", Zero(1), Compose(Add(), Proj(3, 1), Proj(3, 2)))
}

// Pow is m^n, by recursion on n from Mul.
func Pow() PR {
	return Rec("pow", Compose(Succ(), Zero(1)), Compose(Mul(), Proj(3, 1), Proj(3, 2)))
}

// Ackermann is the Ackermann function A(m, n) (Wilhelm Ackermann, 1928),
// written with an explicit stack so it does not overflow Go's: it is
// computable but not primitive recursive, because it grows faster than any
// function built from Rec. It returns A and how many steps it took, or
// ok false when that exceeds limit steps.
//
// Time: about A(m, n) steps. Space: O(A(m, n)) for the stack.
func Ackermann(m, n, limit int) (value, steps int, ok bool) {
	stack := []int{m}
	for len(stack) > 0 {
		if steps++; steps > limit {
			return 0, steps, false
		}
		m = stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		switch {
		case m == 0:
			n++
		case n == 0:
			stack = append(stack, m-1)
			n = 1
		default:
			stack = append(stack, m-1, m)
			n--
		}
	}
	return n, steps, true
}
