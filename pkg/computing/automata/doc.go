// Package automata simulates the machines of the theory of computation, one
// file each, so every step can be watched:
//
//	finite automata   DFA and NFA runs, the subset construction, minimization,
//	                  the product of two DFAs
//	regular           Thompson's construction from a regular expression to an
//	expressions       NFA, matching with it, and backtracking's blowup
//	transducers       Mealy and Moore machines (a vending machine)
//	stack machines    pushdown automata (balanced parentheses, a^n b^n)
//	grammars          derivations, CYK parsing, ambiguity, the pumping lemma
//	Turing machines   a simulator, binary increment, palindromes, unary
//	                  addition, the 3-state busy beaver
//	cellular          elementary rules (30, 110) and Conway's Game of Life
//	computability     a diagonal-argument toy for the halting problem,
//	                  primitive recursion, and the lambda calculus
//
// Machines are read from a small text format, one transition a line, so a
// learner can type their own:
//
//	start p
//	accept r
//	p a q
//	p b p
//	q b r
//
// A run returns its Steps: the states the machine is in, the transitions it
// just took, how much input it has read, and, for the machines that have
// them, the tape, the stack or the output so far.
package automata
