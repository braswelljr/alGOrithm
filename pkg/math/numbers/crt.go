package numbers

import "math/big"

// ChineseRemainder solves the system x = residues[i] (mod moduli[i]) with the
// Chinese remainder theorem, combining one congruence at a time: from x = a
// (mod m) and x = b (mod n), with g = gcd(m, n), a solution exists only when
// g divides b - a, and it is unique modulo lcm(m, n). It returns the smallest
// x >= 0 and that modulus, or ErrNumber when the lists differ in length, are
// empty, a modulus is below 1, the system has no solution, or the modulus
// overflows an int64. The moduli need not be coprime.
//
// Time: O(k log M) for k congruences and combined modulus M.
// Space: O(1) beyond the big-integer work.
func ChineseRemainder(residues, moduli []int) (x, m int, err error) {
	if len(residues) != len(moduli) || len(moduli) == 0 {
		return 0, 0, ErrNumber
	}
	a, mod := big.NewInt(0), big.NewInt(1)
	for i, n := range moduli {
		if n < 1 {
			return 0, 0, ErrNumber
		}
		b, nb := big.NewInt(int64(residues[i])), big.NewInt(int64(n))
		b.Mod(b, nb)
		g, p := new(big.Int), new(big.Int)
		g.GCD(p, nil, mod, nb) // p*mod = g (mod n)
		diff := new(big.Int).Sub(b, a)
		if new(big.Int).Mod(diff, g).Sign() != 0 {
			return 0, 0, ErrNumber
		}
		// a + mod * (diff/g * p mod n/g) solves both congruences.
		ng := new(big.Int).Quo(nb, g)
		t := new(big.Int).Quo(diff, g)
		t.Mul(t, p).Mod(t, ng)
		a.Add(a, t.Mul(t, mod))
		mod.Mul(mod, ng)
		a.Mod(a, mod)
		if !mod.IsInt64() {
			return 0, 0, ErrNumber
		}
	}
	return int(a.Int64()), int(mod.Int64()), nil
}
