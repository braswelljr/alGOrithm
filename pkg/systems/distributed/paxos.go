package distributed

import (
	"fmt"
	"math/rand"
	"strconv"
)

// PaxosResult is what a run of single-decree Paxos did.
type PaxosResult struct {
	// Chosen is the value a majority of acceptors accepted, and Decided
	// whether any was; Conflict is true if two different values were ever
	// chosen, which Paxos never allows (tests check it stays false).
	Chosen   int      `json:"chosen"`
	Decided  bool     `json:"decided"`
	Conflict bool     `json:"conflict"`
	Ballots  int      `json:"ballots"` // proposal numbers tried
	Messages int      `json:"messages"`
	Timeline Timeline `json:"timeline"`
	Steps    []string `json:"steps"`
}

type paxosMsg struct {
	from, to   int // proposers are 0..P-1, acceptors P..P+A-1
	kind       string
	ballot     int
	accBallot  int // promise: the highest ballot the acceptor accepted, 0 if none
	value      int // accept: the value; promise: the accepted value
	acceptedOK bool
}

// Paxos runs single-decree Paxos (Leslie Lamport, 1998) with one proposer per
// value in values and the given number of acceptors. A proposer picks a
// ballot (unique to it: i+1, i+1+P, ...) and sends PREPARE; an acceptor
// promises to ignore lower ballots and reports the value it last accepted.
// With promises from a majority the proposer sends ACCEPT, with the value of
// the highest ballot reported, or its own if none was: so once a value is
// chosen every later ballot carries it. A refusal makes the proposer retry
// with a higher ballot, up to ten times. The network delivers in-flight
// messages in a random order (seeded), so proposers race and interleave;
// no message is lost. Dueling proposers can delay a decision forever (the
// FLP result), which the retry limit cuts short.
//
// Time: O(attempts * proposers * acceptors) messages. Space: the same.
func Paxos(values []int, acceptors int, seed int64) (PaxosResult, error) {
	if len(values) < 1 || len(values) > 5 || acceptors < 1 || acceptors > 9 {
		return PaxosResult{}, fmt.Errorf("%w: 1 to 5 proposers and 1 to 9 acceptors", ErrInput)
	}
	px := newPaxosSim(values, acceptors)
	for i := range px.props {
		px.prepare(i)
	}
	r := rand.New(rand.NewSource(seed))
	for step := 0; len(px.flight) > 0 && step < MaxTicks; step++ {
		k := r.Intn(len(px.flight))
		m := px.flight[k]
		px.flight = append(px.flight[:k], px.flight[k+1:]...)
		px.res.Steps = append(px.res.Steps, fmt.Sprintf("%d: %s gets %s %d from %s", step, px.name(m.to), m.kind, m.ballot, px.name(m.from)))
		if m.to >= px.np {
			px.toAcceptor(step, m)
		} else {
			px.toProposer(step, m)
		}
	}
	px.res.Conflict = len(px.chosen) > 1
	return px.res, nil
}

// paxosAttempts is how many ballots a proposer tries before giving up.
const paxosAttempts = 10

type paxosProposer struct {
	ballot, tries, phase, promises, accepts, bestBallot, value int
}

// paxosSim is one Paxos run: proposers, acceptors and messages in flight.
type paxosSim struct {
	values        []int
	np, acceptors int
	res           PaxosResult
	flight        []paxosMsg
	props         []paxosProposer
	promised      []int
	accB, accV    []int
	acceptedAt    map[int]map[int]bool // ballot -> acceptors that accepted it
	chosen        map[int]bool
}

func newPaxosSim(values []int, acceptors int) *paxosSim {
	return &paxosSim{values: values, np: len(values), acceptors: acceptors, res: PaxosResult{Chosen: -1},
		props: make([]paxosProposer, len(values)), promised: make([]int, acceptors),
		accB: make([]int, acceptors), accV: make([]int, acceptors), acceptedAt: map[int]map[int]bool{}, chosen: map[int]bool{}}
}

func (px *paxosSim) name(i int) string {
	if i < px.np {
		return "proposer " + strconv.Itoa(i+1)
	}
	return "acceptor " + strconv.Itoa(i-px.np+1)
}

func (px *paxosSim) send(m paxosMsg) {
	px.res.Messages++
	px.flight = append(px.flight, m)
}

func (px *paxosSim) toAll(from int, m paxosMsg) {
	for a := 0; a < px.acceptors; a++ {
		m.from, m.to = from, px.np+a
		px.send(m)
	}
}

// prepare starts proposer i's next ballot, unique to it: i+1, i+1+P, ...
func (px *paxosSim) prepare(i int) {
	p := &px.props[i]
	p.tries++
	p.ballot = i + 1 + (p.tries-1)*px.np
	p.phase, p.promises, p.accepts, p.bestBallot, p.value = 1, 0, 0, 0, px.values[i]
	px.res.Ballots++
	px.toAll(i, paxosMsg{kind: "PREPARE", ballot: p.ballot})
}

// toAcceptor promises or accepts a ballot no lower than its promise, and
// refuses the rest.
func (px *paxosSim) toAcceptor(step int, m paxosMsg) {
	a := m.to - px.np
	switch {
	case m.kind == "PREPARE" && m.ballot > px.promised[a]:
		px.promised[a] = m.ballot
		px.send(paxosMsg{from: m.to, to: m.from, kind: "PROMISE", ballot: m.ballot, accBallot: px.accB[a], value: px.accV[a]})
		px.res.Timeline.add(step, px.name(m.to), "promise "+strconv.Itoa(m.ballot), "running")
	case m.kind == "ACCEPT" && m.ballot >= px.promised[a]:
		px.accept(step, a, m)
	case m.kind == "PREPARE" || m.kind == "ACCEPT":
		px.send(paxosMsg{from: m.to, to: m.from, kind: "NACK", ballot: m.ballot})
		px.res.Timeline.add(step, px.name(m.to), "refuse "+strconv.Itoa(m.ballot), "blocked")
	}
}

// accept records acceptor a accepting m, and the value as chosen once a
// majority has accepted its ballot.
func (px *paxosSim) accept(step, a int, m paxosMsg) {
	px.promised[a], px.accB[a], px.accV[a] = m.ballot, m.ballot, m.value
	if px.acceptedAt[m.ballot] == nil {
		px.acceptedAt[m.ballot] = map[int]bool{}
	}
	px.acceptedAt[m.ballot][a] = true
	if len(px.acceptedAt[m.ballot]) >= majority(px.acceptors) {
		px.chosen[m.value] = true
		if !px.res.Decided {
			px.res.Decided, px.res.Chosen = true, m.value
			px.res.Steps = append(px.res.Steps, fmt.Sprintf("%d: value %d is chosen (ballot %d)", step, m.value, m.ballot))
		}
	}
	px.send(paxosMsg{from: m.to, to: m.from, kind: "ACCEPTED", ballot: m.ballot, acceptedOK: true})
	px.res.Timeline.add(step, px.name(m.to), fmt.Sprintf("accept %d=%d", m.ballot, m.value), "done")
}

// toProposer counts promises and accepts for the current ballot, and
// retries higher on a refusal.
func (px *paxosSim) toProposer(step int, m paxosMsg) {
	p := &px.props[m.to]
	if m.ballot != p.ballot || p.phase == 3 {
		return // an answer to an old ballot, or already done
	}
	switch m.kind {
	case "PROMISE":
		if p.phase != 1 {
			return
		}
		p.promises++
		if m.accBallot > p.bestBallot {
			p.bestBallot, p.value = m.accBallot, m.value
		}
		if p.promises >= majority(px.acceptors) {
			p.phase = 2
			px.res.Timeline.add(step, px.name(m.to), fmt.Sprintf("propose %d=%d", p.ballot, p.value), "running")
			px.toAll(m.to, paxosMsg{kind: "ACCEPT", ballot: p.ballot, value: p.value})
		}
	case "ACCEPTED":
		p.accepts++
		if p.accepts >= majority(px.acceptors) {
			p.phase = 3
			px.res.Timeline.add(step, px.name(m.to), fmt.Sprintf("learns %d", p.value), "done")
		}
	case "NACK":
		if p.tries < paxosAttempts {
			px.res.Timeline.add(step, px.name(m.to), "refused: retry higher", "idle")
			px.prepare(m.to)
		} else {
			p.phase = 3
			px.res.Timeline.add(step, px.name(m.to), "gives up", "blocked")
		}
	}
}
