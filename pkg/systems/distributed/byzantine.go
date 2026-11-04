package distributed

import (
	"fmt"
	"slices"
)

// Order is a general's order: Retreat or Attack.
type Order int

// The two orders.
const (
	Retreat Order = iota
	Attack
)

func (o Order) String() string {
	if o == Attack {
		return "attack"
	}
	return "retreat"
}

// ByzantineResult is what the lieutenants decided.
type ByzantineResult struct {
	// Decisions is each lieutenant's decision, for lieutenants 1..n-1.
	Decisions []Order `json:"decisions"`
	// Agree is whether every loyal lieutenant decided the same; Obey whether
	// they all decided the order the commander was given (asked for only
	// when the commander is loyal).
	Agree    bool `json:"agree"`
	Obey     bool `json:"obey"`
	Messages int  `json:"messages"`
}

// Byzantine runs the oral-messages algorithm OM(m) (Leslie Lamport, Robert
// Shostak and Marshall Pease, 1982) for general 0, the commander, giving
// order to lieutenants 1..n-1; the generals in traitors lie. In OM(0) the
// commander sends its order and each lieutenant uses what it got. In OM(m)
// each lieutenant then acts as the commander of OM(m-1), passing on what it
// got to the others, and decides the majority of the values it holds
// (retreat on a tie). A traitorous commander orders odd-numbered
// lieutenants to attack and even ones to retreat; a traitorous lieutenant
// relays the opposite of what it was told. With n >= 3m + 1 and at most m
// traitors the loyal lieutenants agree, and obey a loyal commander; with
// three generals and one traitor no algorithm can guarantee it.
//
// Time: O(n^(m+1)) messages. Space: O(n^(m+1)) for the recursion.
func Byzantine(n, m int, order Order, traitors []int) (ByzantineResult, error) {
	if n < 2 || n > 10 || m < 0 || m > 3 || m >= n-1 {
		return ByzantineResult{}, fmt.Errorf("%w: 2 to 10 generals and 0 to 3 rounds, fewer than n-1", ErrInput)
	}
	for _, t := range traitors {
		if t < 0 || t >= n {
			return ByzantineResult{}, fmt.Errorf("%w: traitors are generals 0 to %d", ErrInput, n-1)
		}
	}
	traitor := func(g int) bool { return slices.Contains(traitors, g) }
	var res ByzantineResult
	sendTo := func(from, to int, v Order) Order {
		res.Messages++
		switch {
		case traitor(from) && from == 0:
			return Order(to % 2) // a traitorous commander splits the army
		case traitor(from):
			return 1 - v // a traitorous lieutenant relays the opposite
		}
		return v
	}
	// om returns, for each lieutenant in lts, the value it settles on for
	// the order cmdr sent.
	var om func(m, cmdr int, v Order, lts []int) map[int]Order
	om = func(m, cmdr int, v Order, lts []int) map[int]Order {
		got := map[int]Order{}
		for _, l := range lts {
			got[l] = sendTo(cmdr, l, v)
		}
		if m == 0 {
			return got
		}
		relayed := map[int]map[int]Order{} // relayed[l][j]: what j settled on for l's relay
		for _, l := range lts {
			rest := slices.DeleteFunc(slices.Clone(lts), func(x int) bool { return x == l })
			relayed[l] = om(m-1, l, got[l], rest)
		}
		out := map[int]Order{}
		for _, j := range lts {
			votes := []Order{got[j]}
			for _, l := range lts {
				if l != j {
					votes = append(votes, relayed[l][j])
				}
			}
			out[j] = majorityOrder(votes)
		}
		return out
	}
	lts := make([]int, 0, n-1)
	for i := 1; i < n; i++ {
		lts = append(lts, i)
	}
	dec := om(m, 0, order, lts)
	res.Agree, res.Obey = true, true
	first := -1
	for _, l := range lts {
		res.Decisions = append(res.Decisions, dec[l])
		if traitor(l) {
			continue
		}
		if first < 0 {
			first = int(dec[l])
		} else if int(dec[l]) != first {
			res.Agree = false
		}
		if dec[l] != order {
			res.Obey = false
		}
	}
	return res, nil
}

func majorityOrder(votes []Order) Order {
	attack := 0
	for _, v := range votes {
		if v == Attack {
			attack++
		}
	}
	if 2*attack > len(votes) {
		return Attack
	}
	return Retreat
}
