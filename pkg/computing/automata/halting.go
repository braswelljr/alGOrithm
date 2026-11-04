package automata

// Program is a toy program for the halting-problem demonstration: given an
// input and a budget of steps, it reports whether it halted within them.
// Programs never really run forever here; a program that "loops" simply
// never halts however large the budget.
type Program struct {
	Name string
	Run  func(x, budget int) bool
}

// Programs is the toy list the diagonal argument is played against: some
// halt at once, some never do, some only on some inputs, and one halts but
// only after more steps than a small budget allows.
var Programs = []Program{
	{"halt at once", func(_, _ int) bool { return true }},
	{"loop forever", func(_, _ int) bool { return false }},
	{"halt on even inputs, loop on odd", func(x, _ int) bool { return x%2 == 0 }},
	{"count down from x, then halt", func(x, budget int) bool { return x <= budget }},
	{"Collatz from x+1 until it reaches 1", func(x, budget int) bool {
		n := x + 1
		for steps := 0; steps <= budget; steps++ {
			if n == 1 {
				return true
			}
			if n%2 == 0 {
				n /= 2
			} else {
				n = 3*n + 1
			}
		}
		return false
	}},
	{"count to 1000, then halt", func(_, budget int) bool { return budget >= 1000 }},
	{"loop when x is 3, else halt", func(x, _ int) bool { return x != 3 }},
}

// Diagonal is the result of the diagonal argument against a would-be halting
// decider.
type Diagonal struct {
	// Says[p][x] is what the decider says program p does on input x: true
	// "halts".
	Says [][]bool
	// Truth[p][x] is what p really does (run with a budget far larger than
	// the decider's; every toy program here is understood).
	Truth [][]bool
	// D[x] is what the diagonal program does on x: the opposite of what the
	// decider says program x does on input x.
	D []bool
	// Wrong lists the (program, input) pairs where the decider is wrong.
	Wrong [][2]int
}

// Halting plays the diagonal argument behind the halting problem (Alan
// Turing, 1936) against one honest attempt at a decider: "run the program
// for budget steps, and say it halts if it did". Number the programs
// P0, P1, ...; build D, which on input x asks the decider about Px on x and
// does the opposite. D differs from every Px on input x, so D is none of
// them, and if D were on the list the decider would be wrong about D on
// itself. That holds for any decider, not just this one; this one is also
// simply wrong about slow programs, which Wrong shows.
//
// Time: O(p^2 * budget) for p programs. Space: O(p^2).
func Halting(programs []Program, budget int) Diagonal {
	const truthBudget = 1 << 20
	n := len(programs)
	d := Diagonal{D: make([]bool, n)}
	for p, prog := range programs {
		says, truth := make([]bool, n), make([]bool, n)
		for x := 0; x < n; x++ {
			says[x] = prog.Run(x, budget)
			truth[x] = prog.Run(x, truthBudget)
			if says[x] != truth[x] {
				d.Wrong = append(d.Wrong, [2]int{p, x})
			}
		}
		d.Says, d.Truth = append(d.Says, says), append(d.Truth, truth)
		d.D[p] = !says[p]
	}
	return d
}
