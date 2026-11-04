package core

import (
	"fmt"

	"github.com/braswelljr/alGOrithm/pkg/algorithms/recursion"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// recursionEntries is every entry of the recursion category.
func recursionEntries() []Entry {
	return []Entry{
		ackermannEntry(),
		factorialEntry(),
		fibonacciEntry(),
		fibonacciMemoEntry(),
		gcdEntry(),
		lcmEntry(),
		mccarthy91Entry(),
		palindromeEntry(),
		permutationsEntry(),
		towerOfHanoiEntry(),
	}
}

func recursionNum(id, name, file, fn, summary, time, space string, lo, hi int, f func(int) int) Entry {
	e := entry(id, name, "algorithms/recursion/"+file, fn)
	e.Summary, e.Time, e.Space, e.Input, e.Limits = summary, time, space, Number, fmt.Sprintf("%d <= n <= %d", lo, hi)
	e.Example = &Input{Values: []int{min(hi, 10)}}
	e.Run = func(in Input, _ trace.Recorder) (Output, error) {
		n, err := one(in)
		if err != nil {
			return Output{}, err
		}
		if n < lo || n > hi {
			return Output{}, inputError(fmt.Sprintf("%s takes %d <= n <= %d", id, lo, hi))
		}
		return Output{Number: ptr(f(n))}, nil
	}
	return e
}

func recursionPair(id, name, fn, summary string, f func(int, int) int) Entry {
	e := entry(id, name, "algorithms/recursion/gcd.go", fn)
	e.Summary, e.Time, e.Space, e.Input, e.Example = summary, "O(log min(a, b))", "O(log min(a, b))", TwoNumbers, &Input{Values: []int{1071, 462}}
	e.Run = func(in Input, _ trace.Recorder) (Output, error) {
		a, b, err := two(in)
		if err != nil {
			return Output{}, err
		}
		return Output{Number: ptr(f(a, b))}, nil
	}
	return e
}

func factorialEntry() Entry {
	fact := recursionNum("factorial", "Factorial", "factorial.go", "Factorial", "n! = n * (n-1) * ... * 1, computed recursively.", "O(n)", "O(n)", 0, 20, func(n int) int { v, _ := recursion.Factorial(n); return v })
	return fact
}

func ackermannEntry() Entry {
	ack := entry("ackermann", "Ackermann Function", "algorithms/recursion/ackermann.go", "Ackermann")
	ack.Summary, ack.Time, ack.Space, ack.Input, ack.Limits = "A total recursive function that grows faster than any primitive recursive one.", "faster than exponential", "O(A(m, n))", TwoNumbers, "m <= 3, n <= 10"
	ack.Example = &Input{Values: []int{2, 3}}
	ack.Run = func(in Input, _ trace.Recorder) (Output, error) {
		m, n, err := two(in)
		if err != nil {
			return Output{}, err
		}
		if m < 0 || n < 0 || m > 3 || n > 10 {
			return Output{}, inputError("ackermann takes 0 <= m <= 3 and 0 <= n <= 10; larger values do not finish")
		}
		return Output{Number: ptr(recursion.Ackermann(m, n))}, nil
	}
	return ack
}

func towerOfHanoiEntry() Entry {
	hanoi := entry("tower-of-hanoi", "Tower of Hanoi", "algorithms/recursion/hanoi.go", "Hanoi")
	hanoi.Summary, hanoi.Time, hanoi.Space, hanoi.Input, hanoi.Limits = "Move n disks from peg A to C, never a larger disk on a smaller one: 2^n - 1 moves.", "O(2^n)", "O(n)", Number, "1 <= n <= 12"
	hanoi.Example = &Input{Values: []int{4}}
	hanoi.Run = func(in Input, _ trace.Recorder) (Output, error) {
		n, err := one(in)
		if err != nil {
			return Output{}, err
		}
		if n < 1 || n > 12 {
			return Output{}, inputError("tower-of-hanoi takes 1 to 12 disks")
		}
		return Output{Lines: recursion.Hanoi(n), Number: &n}, nil
	}
	return hanoi
}

func palindromeEntry() Entry {
	pal := entry("palindrome", "Palindrome", "algorithms/recursion/palindrome.go", "Palindrome")
	pal.Summary, pal.Time, pal.Space, pal.Input, pal.Example = "Whether a word reads the same backward: its ends match and its middle is a palindrome.", "O(n)", "O(n)", Text, &Input{Text: "racecar"}
	pal.Run = func(in Input, _ trace.Recorder) (Output, error) {
		if in.Text == "" {
			return Output{}, inputError("give a word")
		}
		return Output{Bool: ptr(recursion.Palindrome(in.Text))}, nil
	}
	return pal
}

func permutationsEntry() Entry {
	perm := entry("permutations", "Permutations", "algorithms/recursion/palindrome.go", "Permutations")
	perm.Summary, perm.Time, perm.Space, perm.Input, perm.Limits = "Every ordering of the values, by fixing each first and permuting the rest (backtracking).", "O(n * n!)", "O(n * n!)", Ints, "at most 7 values"
	perm.Example = &Input{Values: []int{1, 2, 3}}
	perm.Run = func(in Input, _ trace.Recorder) (Output, error) {
		if err := values(in, "permute", 7); err != nil {
			return Output{}, err
		}
		var lines []string
		for _, p := range recursion.Permutations(in.Values) {
			lines = append(lines, joinInts(p))
		}
		return Output{Lines: lines, Number: ptr(len(lines))}, nil
	}
	return perm
}

func fibonacciEntry() Entry {
	return recursionNum("fibonacci", "Fibonacci (plain recursion)", "fibonacci.go", "Fibonacci", "F(n) = F(n-1) + F(n-2) by plain recursion: exponential, because the same values are computed again and again.", "O(1.618^n)", "O(n)", 0, 35, recursion.Fibonacci)
}

func fibonacciMemoEntry() Entry {
	return recursionNum("fibonacci-memo", "Fibonacci (memoized)", "fibonacci.go", "FibonacciMemo", "The same recursion, remembering each F(k) once computed: linear.", "O(n)", "O(n)", 0, 90, recursion.FibonacciMemo)
}

func mccarthy91Entry() Entry {
	return recursionNum("mccarthy-91", "McCarthy 91", "ackermann.go", "McCarthy91", "n - 10 above 100, else M(M(n + 11)): 91 for every n <= 101.", "O(101 - n)", "O(101 - n)", -1000, 1000, recursion.McCarthy91)
}

func gcdEntry() Entry {
	return recursionPair("gcd", "Euclidean GCD", "GCD", "The greatest common divisor by Euclid's algorithm: gcd(a, b) = gcd(b, a mod b).", recursion.GCD)
}

func lcmEntry() Entry {
	return recursionPair("lcm", "Least Common Multiple", "LCM", "The smallest number both divide: a / gcd(a, b) * b.", recursion.LCM)
}
