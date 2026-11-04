// Package recursion holds algorithms that are defined in terms of themselves,
// one file each: factorial, Fibonacci (plain, memoized and iterative, to
// compare), the Euclidean GCD and the LCM, the Tower of Hanoi, the Ackermann
// function, McCarthy's 91 function, palindromes, and permutations. Each is
// written recursively on purpose, with its cost stated, so the cost of
// recursion itself can be seen.
//
//	recursion.Factorial(5) // 120
//	recursion.Hanoi(3)     // [A -> C A -> B C -> B A -> C B -> A B -> C A -> C]
package recursion
