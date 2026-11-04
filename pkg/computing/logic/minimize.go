package logic

import (
	"fmt"
	"math/bits"
	"slices"
	"strings"
)

// Implicant is a product term over n variables: the bits of Value where Mask
// is 0 are fixed, and the bits where Mask is 1 are free (the dashes in
// "1-0"). Bit n-1 is the first variable.
type Implicant struct {
	Value, Mask int
}

// Covers reports whether minterm m satisfies the implicant.
func (p Implicant) Covers(m int) bool { return m&^p.Mask == p.Value&^p.Mask }

// Literals returns the number of variables the term fixes.
func (p Implicant) Literals(n int) int { return n - bits.OnesCount(uint(p.Mask)) }

// Pattern writes the term as bits and dashes, first variable first ("1-0").
func (p Implicant) Pattern(n int) string {
	var b strings.Builder
	for k := n - 1; k >= 0; k-- {
		switch {
		case p.Mask>>k&1 == 1:
			b.WriteByte('-')
		case p.Value>>k&1 == 1:
			b.WriteByte('1')
		default:
			b.WriteByte('0')
		}
	}
	return b.String()
}

// Term writes the product with the variables' names ("A & !C"), or "1" when
// every variable is free.
func (p Implicant) Term(vars []string) string {
	n := len(vars)
	var parts []string
	for k, v := range vars {
		bit := n - 1 - k
		if p.Mask>>bit&1 == 1 {
			continue
		}
		if p.Value>>bit&1 == 1 {
			parts = append(parts, v)
		} else {
			parts = append(parts, "!"+v)
		}
	}
	if len(parts) == 0 {
		return "1"
	}
	return strings.Join(parts, " & ")
}

// Minimized is the result of minimizing a function: its prime implicants,
// the essential ones, the cover chosen (a minimal sum of products), and the
// steps in words.
type Minimized struct {
	Primes    []Implicant
	Essential []Implicant
	Cover     []Implicant
	Steps     []string
}

// SOP writes the cover as a sum of products with the variables' names, "0"
// for a function that is never true.
func (m Minimized) SOP(vars []string) string {
	if len(m.Cover) == 0 {
		return "0"
	}
	terms := make([]string, len(m.Cover))
	for i, p := range m.Cover {
		terms[i] = p.Term(vars)
	}
	return strings.Join(terms, " | ")
}

// maxPetrick bounds the products Petrick's method keeps while it multiplies
// out; past it the smallest products found so far are kept.
const maxPetrick = 4096

// QuineMcCluskey minimizes the function of n variables (n <= MaxVars) that is
// true on minterms and does not matter on dontCares (Willard Quine 1952,
// Edward McCluskey 1956). It groups the terms by their number of ones,
// combines terms in neighboring groups that differ in one bit until nothing
// combines (the terms never combined are the prime implicants), takes the
// essential primes (the only cover of some minterm), and chooses the rest by
// Petrick's method: multiply out "this minterm is covered by P1 or P2" for
// every remaining minterm, and keep the product with the fewest terms, then
// the fewest literals.
//
// Time: O(3^n) implicants in the worst case, and Petrick's method is
// exponential in the minterms left after the essential primes; this is why
// tools use heuristics (Espresso) for large n. Space: the same.
func QuineMcCluskey(n int, minterms, dontCares []int) (Minimized, error) {
	if n < 1 || n > MaxVars {
		return Minimized{}, fmt.Errorf("%w: give 1 to %d variables", ErrInput, MaxVars)
	}
	var res Minimized
	terms := map[Implicant]bool{}
	for _, list := range [][]int{minterms, dontCares} {
		for _, m := range list {
			if m < 0 || m >= 1<<n {
				return Minimized{}, fmt.Errorf("%w: minterm %d is outside 0 to %d", ErrInput, m, 1<<n-1)
			}
			terms[Implicant{Value: m}] = true
		}
	}
	need := slices.Clone(minterms)
	slices.Sort(need)
	need = slices.Compact(need)
	if len(need) == 0 {
		res.Steps = append(res.Steps, "no minterms: the function is 0")
		return res, nil
	}

	// Combine until nothing combines.
	for round := 1; len(terms) > 0; round++ {
		used := map[Implicant]bool{}
		next := map[Implicant]bool{}
		list := sortedTerms(terms)
		for i, a := range list {
			for _, b := range list[i+1:] {
				diff := a.Value ^ b.Value
				if a.Mask == b.Mask && bits.OnesCount(uint(diff&^a.Mask)) == 1 {
					c := Implicant{Value: a.Value &^ diff, Mask: a.Mask | diff}
					next[c] = true
					used[a], used[b] = true, true
				}
			}
		}
		var primes []string
		for _, t := range list {
			if !used[t] {
				res.Primes = append(res.Primes, t)
				primes = append(primes, t.Pattern(n))
			}
		}
		res.Steps = append(res.Steps, fmt.Sprintf("round %d: %d terms, %d combined into %d; prime: %s",
			round, len(list), len(used), len(next), strings.Join(orNone(primes), " ")))
		terms = next
	}
	sortImplicants(res.Primes)

	// Essential primes.
	covered := map[int]bool{}
	for _, m := range need {
		var only []Implicant
		for _, p := range res.Primes {
			if p.Covers(m) {
				only = append(only, p)
			}
		}
		if len(only) == 1 && !slices.Contains(res.Essential, only[0]) {
			res.Essential = append(res.Essential, only[0])
			res.Steps = append(res.Steps, fmt.Sprintf("%s is essential: the only prime covering minterm %d", only[0].Pattern(n), m))
		}
	}
	for _, p := range res.Essential {
		for _, m := range need {
			if p.Covers(m) {
				covered[m] = true
			}
		}
	}
	res.Cover = slices.Clone(res.Essential)

	// Petrick's method for the rest.
	var left []int
	for _, m := range need {
		if !covered[m] {
			left = append(left, m)
		}
	}
	if len(left) > 0 {
		res.Steps = append(res.Steps, fmt.Sprintf("minterms left after the essential primes: %v; Petrick's method", left))
		best := petrick(n, res.Primes, left)
		res.Cover = append(res.Cover, best...)
	}
	sortImplicants(res.Cover)
	return res, nil
}

func orNone(s []string) []string {
	if len(s) == 0 {
		return []string{"none"}
	}
	return s
}

// petrick returns the cheapest set of primes covering every minterm in left.
func petrick(n int, primes []Implicant, left []int) []Implicant {
	// A product is a set of prime indexes, as a bit set over len(primes) <= 3^8.
	type product map[int]bool
	products := []product{{}}
	for _, m := range left {
		var next []product
		for _, p := range products {
			for i, pr := range primes {
				if !pr.Covers(m) {
					continue
				}
				q := product{i: true}
				for k := range p {
					q[k] = true
				}
				next = append(next, q)
			}
		}
		// Absorb duplicates and supersets cheaply: keep the smallest.
		slices.SortFunc(next, func(a, b product) int { return len(a) - len(b) })
		if len(next) > maxPetrick {
			next = next[:maxPetrick]
		}
		products = next
	}
	cost := func(p product) (int, int) {
		lits := 0
		for i := range p {
			lits += primes[i].Literals(n)
		}
		return len(p), lits
	}
	best := products[0]
	for _, p := range products[1:] {
		a, b := cost(p)
		c, d := cost(best)
		if a < c || (a == c && b < d) {
			best = p
		}
	}
	out := make([]Implicant, 0, len(best))
	for i := range best {
		out = append(out, primes[i])
	}
	return out
}

func sortedTerms(set map[Implicant]bool) []Implicant {
	list := make([]Implicant, 0, len(set))
	for t := range set {
		list = append(list, t)
	}
	sortImplicants(list)
	return list
}

// sortImplicants orders by number of ones, then value, then mask, so every
// result is deterministic.
func sortImplicants(list []Implicant) {
	slices.SortFunc(list, func(a, b Implicant) int {
		if d := bits.OnesCount(uint(a.Value&^a.Mask)) - bits.OnesCount(uint(b.Value&^b.Mask)); d != 0 {
			return d
		}
		if a.Mask != b.Mask {
			return a.Mask - b.Mask
		}
		return a.Value - b.Value
	})
}
