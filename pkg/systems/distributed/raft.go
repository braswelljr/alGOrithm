package distributed

import (
	"fmt"
	"math/rand"
)

// RaftConfig is a Raft cluster to simulate.
type RaftConfig struct {
	Nodes int   // 3 to 9
	Ticks int   // how long to run
	Seed  int64 // picks the election timeouts
	// CrashAt is the tick the leader of the moment crashes (0: never), so a
	// new election follows.
	CrashAt int
	// Every is how often, in ticks, a client sends the leader a command
	// (0: never).
	Every int
}

// Stretch is a run of ticks a node spent in one state of one term.
type Stretch struct {
	Node  int    `json:"node"`
	State string `json:"state"` // "follower", "candidate", "leader", "crashed"
	Term  int    `json:"term"`
	Start int    `json:"start"`
	End   int    `json:"end"`
}

// Election is a node becoming leader.
type Election struct {
	Tick   int `json:"tick"`
	Term   int `json:"term"`
	Leader int `json:"leader"`
}

// RaftResult is what a Raft run did.
type RaftResult struct {
	Stretches []Stretch  `json:"stretches"`
	Elections []Election `json:"elections"`
	// Logs is each node's log at the end: the term of each entry.
	Logs [][]int `json:"logs"`
	// Committed is how many entries the leader at the end had committed.
	Committed int      `json:"committed"`
	Messages  int      `json:"messages"`
	Steps     []string `json:"steps"`
}

type raftMsg struct {
	at, from, to int
	kind         string // "vote?", "vote", "append", "appended"
	term         int
	ok           bool
	lastIdx      int   // vote?: candidate's last log index; appended: the follower's match index
	lastTerm     int   // vote?: the term of that entry
	prevIdx      int   // append
	prevTerm     int   // append
	entries      []int // append
	commit       int   // append: leader's commit index (entries 0..commit-1)
}

type raftNode struct {
	state      string
	term, vote int
	log        []int
	commit     int
	deadline   int
	votes      int
	next       []int
	match      []int
}

// Raft simulates Raft (Diego Ongaro and John Ousterhout, 2014): nodes start
// as followers; one whose election timer (random, 10 to 19 ticks) runs out
// becomes a candidate in a new term and asks for votes; a node votes once a
// term, for a candidate whose log is at least as up to date as its own; a
// majority makes a leader, which sends heartbeats (AppendEntries) every 3
// ticks carrying new log entries. An entry is committed once a majority
// stores it and it is from the leader's term. Any message with a higher
// term turns its receiver into a follower of that term. Messages take one
// tick and are not lost; nodes fail only by crashing, and a crashed node
// does not come back. Snapshots, membership changes and persistence are
// left out. At most one leader is elected per term.
//
// Time: O(ticks * nodes^2). Space: O(ticks * nodes + log).
func Raft(c RaftConfig) (RaftResult, error) {
	if c.Nodes < 3 || c.Nodes > 9 || c.Ticks < 1 || c.Ticks > 2000 || c.Every < 0 || c.CrashAt < 0 {
		return RaftResult{}, fmt.Errorf("%w: 3 to 9 nodes and 1 to 2000 ticks", ErrInput)
	}
	sim := newRaftSim(c)
	for t := 0; t < c.Ticks; t++ {
		sim.tick(t)
	}
	return sim.result(), nil
}

// raftHeartbeat is how often, in ticks, a leader sends AppendEntries.
const raftHeartbeat = 3

// raftSim is one Raft run: the nodes, the messages in flight and the record.
type raftSim struct {
	c      RaftConfig
	r      *rand.Rand
	nodes  []*raftNode
	queue  []raftMsg
	leader int
	res    RaftResult
}

func newRaftSim(c RaftConfig) *raftSim {
	s := &raftSim{c: c, r: rand.New(rand.NewSource(c.Seed)), leader: -1}
	s.nodes = make([]*raftNode, c.Nodes)
	for i := range s.nodes {
		s.nodes[i] = &raftNode{state: "follower", vote: -1, deadline: s.timeout(0)}
	}
	return s
}

// timeout is a fresh election deadline: 10 to 19 ticks after t.
func (s *raftSim) timeout(t int) int { return t + 10 + s.r.Intn(10) }

func (s *raftSim) send(m raftMsg) {
	s.res.Messages++
	s.queue = append(s.queue, m)
}

func (s *raftSim) step(format string, args ...any) {
	s.res.Steps = append(s.res.Steps, fmt.Sprintf(format, args...))
}

// tick runs one tick: a scheduled crash, this tick's messages, a client
// command, the timers, and the timeline.
func (s *raftSim) tick(t int) {
	if s.c.CrashAt > 0 && t == s.c.CrashAt && s.isLeader(s.leader) {
		s.nodes[s.leader].state = "crashed"
		s.step("t%d: leader %d crashes", t, s.leader)
		s.leader = -1
	}
	for _, m := range s.due(t) {
		s.deliver(t, m)
	}
	if s.c.Every > 0 && t%s.c.Every == 0 && s.isLeader(s.leader) {
		ld := s.nodes[s.leader]
		ld.log = append(ld.log, ld.term)
		ld.match[s.leader] = len(ld.log) - 1
	}
	for i, nd := range s.nodes {
		if nd.state != "crashed" && t >= nd.deadline {
			s.expire(t, i)
		}
	}
	if s.leader >= 0 && !s.isLeader(s.leader) {
		s.leader = -1
	}
	s.record(t)
}

func (s *raftSim) isLeader(i int) bool { return i >= 0 && s.nodes[i].state == "leader" }

// due takes this tick's messages off the queue, in the order they were sent.
func (s *raftSim) due(t int) []raftMsg {
	var due, later []raftMsg
	for _, m := range s.queue {
		if m.at == t {
			due = append(due, m)
		} else {
			later = append(later, m)
		}
	}
	s.queue = later
	return due
}

// deliver hands m to its receiver; a higher term first turns it into a follower.
func (s *raftSim) deliver(t int, m raftMsg) {
	nd := s.nodes[m.to]
	if nd.state == "crashed" {
		return
	}
	if m.term > nd.term {
		nd.term, nd.vote, nd.state = m.term, -1, "follower"
	}
	switch m.kind {
	case "vote?":
		s.onVoteRequest(t, nd, m)
	case "vote":
		s.onVote(t, nd, m)
	case "append":
		s.onAppend(t, nd, m)
	case "appended":
		s.onAppended(nd, m)
	}
}

// onVoteRequest grants a vote once a term, to a candidate whose log is at
// least as up to date.
func (s *raftSim) onVoteRequest(t int, nd *raftNode, m raftMsg) {
	upToDate := m.lastTerm > lastLogTerm(nd) || (m.lastTerm == lastLogTerm(nd) && m.lastIdx >= len(nd.log)-1)
	grant := m.term == nd.term && (nd.vote == -1 || nd.vote == m.from) && upToDate
	if grant {
		nd.vote, nd.deadline = m.from, s.timeout(t)
	}
	s.send(raftMsg{at: t + 1, from: m.to, to: m.from, kind: "vote", term: nd.term, ok: grant})
}

// onVote counts a granted vote; a majority makes the candidate leader.
func (s *raftSim) onVote(t int, nd *raftNode, m raftMsg) {
	if nd.state != "candidate" || m.term != nd.term || !m.ok {
		return
	}
	nd.votes++
	if nd.votes < majority(s.c.Nodes) {
		return
	}
	nd.state, s.leader = "leader", m.to
	nd.next, nd.match = make([]int, s.c.Nodes), make([]int, s.c.Nodes)
	for p := range nd.next {
		nd.next[p], nd.match[p] = len(nd.log), -1
	}
	nd.match[m.to] = len(nd.log) - 1
	s.res.Elections = append(s.res.Elections, Election{Tick: t, Term: nd.term, Leader: m.to})
	s.step("t%d: node %d is leader of term %d with %d votes", t, m.to, nd.term, nd.votes)
	s.broadcastAppend(t, m.to)
}

// onAppend accepts entries that extend a matching log, else refuses.
func (s *raftSim) onAppend(t int, nd *raftNode, m raftMsg) {
	refuse := raftMsg{at: t + 1, from: m.to, to: m.from, kind: "appended", term: nd.term}
	if m.term < nd.term {
		s.send(refuse)
		return
	}
	nd.state, nd.deadline = "follower", s.timeout(t)
	if m.prevIdx >= len(nd.log) || (m.prevIdx >= 0 && nd.log[m.prevIdx] != m.prevTerm) {
		s.send(refuse)
		return
	}
	nd.log = append(nd.log[:m.prevIdx+1], m.entries...)
	nd.commit = max(nd.commit, min(m.commit, len(nd.log)))
	s.send(raftMsg{at: t + 1, from: m.to, to: m.from, kind: "appended", term: nd.term, ok: true, lastIdx: len(nd.log) - 1})
}

// onAppended moves the follower's indexes and advances the commit point.
func (s *raftSim) onAppended(nd *raftNode, m raftMsg) {
	if nd.state != "leader" || m.term != nd.term {
		return
	}
	if m.ok {
		nd.match[m.from] = max(nd.match[m.from], m.lastIdx)
		nd.next[m.from] = nd.match[m.from] + 1
	} else if nd.next[m.from] > 0 {
		nd.next[m.from]--
	}
	s.advanceCommit(nd)
}

// advanceCommit commits the highest entry of the leader's term that a
// majority stores.
func (s *raftSim) advanceCommit(nd *raftNode) {
	for idx := len(nd.log) - 1; idx >= nd.commit; idx-- {
		count := 0
		for _, mi := range nd.match {
			if mi >= idx {
				count++
			}
		}
		if nd.log[idx] == nd.term && count >= majority(s.c.Nodes) {
			nd.commit = idx + 1
			return
		}
	}
}

// expire runs node i's timer: a leader sends heartbeats, anyone else stands
// for election in a new term.
func (s *raftSim) expire(t, i int) {
	nd := s.nodes[i]
	if nd.state == "leader" {
		s.broadcastAppend(t, i)
		return
	}
	nd.state, nd.term, nd.vote, nd.votes = "candidate", nd.term+1, i, 1
	nd.deadline = s.timeout(t)
	s.step("t%d: node %d times out and stands in term %d", t, i, nd.term)
	for p := range s.nodes {
		if p != i {
			s.send(raftMsg{at: t + 1, from: i, to: p, kind: "vote?", term: nd.term, lastIdx: len(nd.log) - 1, lastTerm: lastLogTerm(nd)})
		}
	}
}

// broadcastAppend sends AppendEntries from leader l to every other node and
// resets its heartbeat.
func (s *raftSim) broadcastAppend(t, l int) {
	ld := s.nodes[l]
	for p := range s.nodes {
		if p == l {
			continue
		}
		prev := ld.next[p] - 1
		pt := 0
		if prev >= 0 {
			pt = ld.log[prev]
		}
		s.send(raftMsg{at: t + 1, from: l, to: p, kind: "append", term: ld.term, prevIdx: prev, prevTerm: pt,
			entries: append([]int(nil), ld.log[ld.next[p]:]...), commit: ld.commit})
	}
	ld.deadline = t + raftHeartbeat
}

// record extends each node's current stretch on the timeline, or starts one.
func (s *raftSim) record(t int) {
	for i, nd := range s.nodes {
		k := len(s.res.Stretches) - 1
		for k >= 0 && s.res.Stretches[k].Node != i {
			k--
		}
		if k >= 0 && s.res.Stretches[k].State == nd.state && s.res.Stretches[k].Term == nd.term && s.res.Stretches[k].End == t {
			s.res.Stretches[k].End = t + 1
			continue
		}
		s.res.Stretches = append(s.res.Stretches, Stretch{Node: i, State: nd.state, Term: nd.term, Start: t, End: t + 1})
	}
}

func (s *raftSim) result() RaftResult {
	for _, nd := range s.nodes {
		s.res.Logs = append(s.res.Logs, append([]int(nil), nd.log...))
		if nd.state == "leader" {
			s.res.Committed = nd.commit
		}
	}
	return s.res
}

// lastLogTerm is the term of nd's last entry, or 0 for an empty log.
func lastLogTerm(nd *raftNode) int {
	if len(nd.log) == 0 {
		return 0
	}
	return nd.log[len(nd.log)-1]
}
