package distributed

import (
	"fmt"
	"strconv"
)

// CommitResult is the outcome of a two-phase commit.
type CommitResult struct {
	// Decision is "commit", "abort", or "blocked" when the coordinator
	// crashed after the votes and participants that voted yes cannot decide.
	Decision string   `json:"decision"`
	Blocked  []int    `json:"blocked"` // participants left uncertain
	Messages int      `json:"messages"`
	Timeline Timeline `json:"timeline"`
	Steps    []string `json:"steps"`
}

// TwoPhaseCommit runs two-phase commit with a coordinator and n
// participants. Phase one: the coordinator sends PREPARE; each participant
// writes its work to its log and votes yes, except the one listed in no
// (1-based, 0 for none), which votes no. Phase two: the coordinator commits
// only if every vote was yes, logs the decision, and sends it; participants
// apply it and acknowledge. With crash true the coordinator crashes after
// counting the votes and before sending the decision: a participant that
// voted no may abort on its own, but one that voted yes has promised to do
// either and must wait, holding its locks, until the coordinator returns.
// That blocking is two-phase commit's weakness and why three-phase commit
// and consensus-based commit exist. Messages take one tick and are not lost.
//
// Time: O(n) ticks and 4n messages. Space: O(n).
func TwoPhaseCommit(n, no int, crash bool) (CommitResult, error) {
	if n < 1 || n > 26 || no < 0 || no > n {
		return CommitResult{}, fmt.Errorf("%w: 1 to 26 participants, and a no vote from 0 (none) to %d", ErrInput, n)
	}
	var res CommitResult
	p := func(i int) string { return "participant " + strconv.Itoa(i+1) }
	step := func(t int, lane, what, state string) {
		res.Timeline.add(t, lane, what, state)
		res.Steps = append(res.Steps, fmt.Sprintf("t%d %s: %s", t, lane, what))
	}
	step(0, "coordinator", "send PREPARE", "running")
	allYes := true
	for i := 0; i < n; i++ {
		res.Messages += 2 // PREPARE and the vote
		if i+1 == no {
			allYes = false
			step(1, p(i), "vote NO", "running")
		} else {
			step(1, p(i), "log prepared, vote YES", "running")
		}
	}
	decision := "abort"
	if allYes {
		decision = "commit"
	}
	if crash {
		step(2, "coordinator", "crashes before sending the decision", "blocked")
		res.Decision = "blocked"
		for i := 0; i < n; i++ {
			if i+1 == no {
				step(3, p(i), "voted no: aborts on its own", "done")
				continue
			}
			res.Blocked = append(res.Blocked, i+1)
			step(3, p(i), "uncertain: waits, holding its locks", "blocked")
		}
		if len(res.Blocked) == 0 {
			res.Decision = "abort"
		}
		return res, nil
	}
	step(2, "coordinator", "log "+decision+", send "+decision, "running")
	for i := 0; i < n; i++ {
		res.Messages += 2 // the decision and the ack
		step(3, p(i), decision+", ack", "done")
	}
	step(4, "coordinator", "all acked: done", "done")
	res.Decision = decision
	return res, nil
}
