// Package logic is digital logic from the ground up, one idea per file: the
// seven gates and their truth tables (gates.go), Boolean expressions parsed
// from text (expr.go), truth tables of any expression and the canonical sum
// of products and product of sums (table.go), De Morgan's laws (demorgan.go),
// Quine-McCluskey minimization with Petrick's method (minimize.go), Karnaugh
// maps (kmap.go), and satisfiability, first by brute force over the truth
// table and then by DPLL (sat.go).
//
// Expressions are written the way textbooks and programmers both write them:
//
//	A & (B | !C)        symbols: ! & | ^, and ~ for not
//	A AND (B OR NOT C)  words: NOT AND OR XOR NAND NOR XNOR, any case
//	A * (B + C')        algebra: * product, + sum, ' complement
//
// Precedence, from tightest: not, and (and nand), xor (and xnor), or (and nor).
package logic

import "errors"

// ErrInput is wrapped by every error about bad input: an expression that does
// not parse, too many variables, a malformed clause.
var ErrInput = errors.New("invalid input")
