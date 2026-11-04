// Package compilers builds a tiny compiler for a tiny language, one stage per
// file, so each stage can be read and run on its own:
//
//	lexer.go      source text to tokens with their line and column
//	ast.go        the syntax tree, printed as an indented tree or as source
//	parser.go     recursive descent: a program of let, if, while, fn, print
//	pratt.go      Pratt parsing: operator precedence by binding power
//	shunting.go   Dijkstra's shunting-yard: infix to reverse Polish notation
//	rpn.go        evaluating reverse Polish notation with a stack
//	check.go      type checking, with errors at line:column
//	interp.go     a tree-walking interpreter
//	fold.go       constant folding and dead-code elimination
//	tac.go        three-address code
//	bytecode.go   a stack bytecode compiler and its virtual machine
//	regalloc.go   register allocation by coloring an interference graph
//	gc.go         garbage collection: reference counting, mark and sweep,
//	              copying (Cheney) and incremental tri-color marking
//
// The language has integers and booleans:
//
//	fn fact(n) { if n < 2 { return 1; } return n * fact(n - 1); }
//	let i = 1;
//	while i <= 5 { print fact(i); i = i + 1; }
//
// A function sees its parameters, its own variables and every function, not
// the variables of the program around it. Every stage reports what it did,
// step by step, as plain data, so the steps can be printed or animated.
package compilers
