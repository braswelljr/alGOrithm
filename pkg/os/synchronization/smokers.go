package synchronization

import (
	"fmt"
	"math/rand"
)

// The three ingredients; smoker i holds an endless supply of ingredient i.
var ingredients = [3]string{"tobacco", "paper", "matches"}

// CigaretteSmokers simulates Patil's cigarette smokers problem (1971). An
// agent puts two of the three ingredients on the table, chosen with a seed;
// the smoker who holds the third should take both and smoke. Naive: each
// smoker grabs the ingredients it needs one at a time as they appear, so two
// smokers can each grab one of the pair and wait forever for the other
// (Deadlocked). Fixed: a "pusher" looks at the pair as a whole and wakes only
// the smoker it completes.
//
// Time: O(rounds). Space: O(rounds).
func CigaretteSmokers(rounds int, naive bool, seed int64) (Result, error) {
	if rounds < 1 || rounds > 30 {
		return Result{}, inputErr("1-30 rounds")
	}
	rng := rand.New(rand.NewSource(seed))
	res := Result{Counts: map[string]int{}}
	tick := 0
	for r := 0; r < rounds; r++ {
		missing := rng.Intn(3) // the smoker who should smoke
		a, b := (missing+1)%3, (missing+2)%3
		res.Timeline = append(res.Timeline, Event{Tick: tick, Lane: "agent", What: fmt.Sprintf("puts %s and %s", ingredients[a], ingredients[b]), State: "running"})
		res.Log = append(res.Log, fmt.Sprintf("round %d: the agent puts %s and %s on the table", r+1, ingredients[a], ingredients[b]))
		tick++
		if naive {
			// The first ingredient is grabbed by a seeded smoker that needs it
			// (not necessarily the one it completes).
			needers := []int{}
			for s := 0; s < 3; s++ {
				if s != a {
					needers = append(needers, s)
				}
			}
			grabber := needers[rng.Intn(len(needers))]
			res.Timeline = append(res.Timeline, Event{Tick: tick, Lane: "smoker with " + ingredients[grabber], What: "grabs the " + ingredients[a], State: "running"})
			if grabber != missing {
				res.Log = append(res.Log, fmt.Sprintf("  the smoker with %s grabs the %s and waits for %s, which the agent will never put down while the %s sits there taken: deadlock", ingredients[grabber], ingredients[a], ingredients[3-grabber-a], ingredients[b]))
				res.Timeline = append(res.Timeline, Event{Tick: tick + 1, Lane: "smoker with " + ingredients[grabber], What: "waits forever", State: "blocked"},
					Event{Tick: tick + 1, Lane: "smoker with " + ingredients[missing], What: "waits forever", State: "blocked"},
					Event{Tick: tick + 1, Lane: "agent", What: "waits for a smoker to finish", State: "blocked"})
				res.Deadlocked = true
				res.Ticks = tick + 2
				return res, nil
			}
		}
		lane := "smoker with " + ingredients[missing]
		res.Timeline = append(res.Timeline, Event{Tick: tick, Lane: lane, What: "takes both and rolls", State: "running"},
			Event{Tick: tick + 1, Lane: lane, What: "smokes", State: "running"},
			Event{Tick: tick + 1, Lane: "agent", What: "waits for the smoker", State: "blocked"})
		res.Counts[ingredients[missing]]++
		res.Log = append(res.Log, fmt.Sprintf("  the smoker with %s smokes", ingredients[missing]))
		tick += 2
	}
	res.Ticks = tick
	return res, nil
}
