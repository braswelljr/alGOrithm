package recursion_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/braswelljr/alGOrithm/pkg/algorithms/recursion"
)

func TestFactorial(t *testing.T) {
	for n, want := range map[int]int{0: 1, 1: 1, 5: 120, 20: 2432902008176640000} {
		if got, err := recursion.Factorial(n); err != nil || got != want {
			t.Errorf("Factorial(%d) = %d, %v", n, got, err)
		}
	}
	for _, n := range []int{-1, 21} {
		if _, err := recursion.Factorial(n); !errors.Is(err, recursion.ErrRange) {
			t.Errorf("Factorial(%d) error %v", n, err)
		}
	}
}

func TestFibonacciVariantsAgree(t *testing.T) {
	for n := -10; n <= 30; n++ {
		a, b, c := recursion.Fibonacci(n), recursion.FibonacciMemo(n), recursion.FibonacciIterative(n)
		if a != b || b != c {
			t.Fatalf("F(%d): %d %d %d", n, a, b, c)
		}
	}
	if recursion.FibonacciIterative(90) != 2880067194370816120 || recursion.Fibonacci(9) != 34 {
		t.Error("Fibonacci values wrong")
	}
}

func TestGCDAndLCM(t *testing.T) {
	for _, c := range [][3]int{{48, 18, 6}, {17, 5, 1}, {0, 7, 7}, {-12, 8, 4}, {0, 0, 0}} {
		if got := recursion.GCD(c[0], c[1]); got != c[2] {
			t.Errorf("GCD(%d, %d) = %d", c[0], c[1], got)
		}
	}
	if recursion.LCM(4, 6) != 12 || recursion.LCM(0, 5) != 0 || recursion.LCM(-3, 4) != 12 {
		t.Error("LCM wrong")
	}
}

func TestHanoiMovesAreLegal(t *testing.T) {
	for n := 0; n <= 8; n++ {
		moves := recursion.Hanoi(n)
		if len(moves) != 1<<n-1 {
			t.Fatalf("Hanoi(%d) = %d moves", n, len(moves))
		}
		pegs := map[string][]int{"A": nil, "B": nil, "C": nil}
		for d := n; d >= 1; d-- {
			pegs["A"] = append(pegs["A"], d)
		}
		for _, m := range moves {
			var from, to string
			if _, err := fmt.Sscanf(m, "%s -> %s", &from, &to); err != nil {
				t.Fatalf("move %q: %v", m, err)
			}
			src, dst := pegs[from], pegs[to]
			disk := src[len(src)-1]
			if len(dst) > 0 && dst[len(dst)-1] < disk {
				t.Fatalf("%q puts %d on %d", m, disk, dst[len(dst)-1])
			}
			pegs[from], pegs[to] = src[:len(src)-1], append(dst, disk)
		}
		if len(pegs["C"]) != n {
			t.Errorf("n=%d: %d disks on C", n, len(pegs["C"]))
		}
	}
}

func TestAckermannAndMcCarthy(t *testing.T) {
	if recursion.Ackermann(2, 3) != 9 || recursion.Ackermann(3, 3) != 61 || recursion.Ackermann(0, 5) != 6 {
		t.Error("Ackermann wrong")
	}
	for n := -50; n <= 101; n++ {
		if recursion.McCarthy91(n) != 91 {
			t.Fatalf("M(%d) = %d", n, recursion.McCarthy91(n))
		}
	}
	if recursion.McCarthy91(150) != 140 {
		t.Error("M(150) wrong")
	}
}

func TestPalindromeAndPermutations(t *testing.T) {
	for s, want := range map[string]bool{"": true, "a": true, "racecar": true, "abca": false, "abba": true} {
		if recursion.Palindrome(s) != want {
			t.Errorf("Palindrome(%q)", s)
		}
	}
	p := recursion.Permutations([]int{1, 2, 3})
	seen := map[string]bool{}
	for _, x := range p {
		seen[fmt.Sprint(x)] = true
	}
	if len(p) != 6 || len(seen) != 6 || len(recursion.Permutations([]int{})) != 1 {
		t.Errorf("permutations %v", p)
	}
}
