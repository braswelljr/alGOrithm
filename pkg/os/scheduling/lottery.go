package scheduling

import "math/rand"

// Lottery schedules by lottery (Waldspurger and Weihl, 1994): each tick, draw
// one of the ready processes' tickets at random and run its holder. A process
// holds Priority tickets (at least 1), so over time it gets the CPU in
// proportion to its tickets, and none starves. seed makes the draws, and so
// the schedule, reproducible.
//
// Time: O(T * n). Space: O(T).
func Lottery(ps []Process, seed int64) (Schedule, error) {
	if err := validate(ps); err != nil {
		return Schedule{}, err
	}
	r := rand.New(rand.NewSource(seed))
	tickets := func(i int) int { return max(1, ps[i].Priority) }
	return simulate(ps, func(ready []int, _ int, _ []int) int {
		total := 0
		for _, i := range ready {
			total += tickets(i)
		}
		draw := r.Intn(total)
		for _, i := range ready {
			if draw -= tickets(i); draw < 0 {
				return i
			}
		}
		return ready[len(ready)-1]
	}), nil
}
