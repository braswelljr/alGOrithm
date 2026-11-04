package core

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/systems/distributed"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// distSpans turns a distributed timeline into one-tick timeline bars.
func distSpans(tl distributed.Timeline) []Span {
	out := make([]Span, len(tl))
	for i, e := range tl {
		out[i] = Span{Lane: e.Lane, Label: e.What, Start: e.Tick, End: e.Tick + 1, State: e.State}
	}
	return out
}

func b2f(b bool) float64 {
	if b {
		return 1
	}
	return 0
}

const clockTraceExample = "A send B\nA local\nB recv A\nB send C\nC local\nC recv B\nC send A\nA recv C"

func distributedEntries() []Entry {
	clock := func(id, name, fn, summary, time, space string, vector bool) Entry {
		e := entry(id, name, "systems/distributed/clocks.go", fn)
		e.Summary, e.Time, e.Space, e.Input = summary, time, space, Text
		e.Example = &Input{Text: clockTraceExample}
		e.Limits = "one event a line: \"P local\", \"P send Q\", \"Q recv P\""
		e.Run = func(in Input, _ trace.Recorder) (Output, error) {
			tr, err := distributed.ParseTrace(in.Text)
			if err != nil {
				return Output{}, wrap(err)
			}
			lam, vec := tr.Lamport(), tr.Vector()
			cols := []string{"Lamport"}
			if vector {
				cols = tr.Procs
			}
			out := Output{Table: &TableOutput{Rows: len(tr.Events), Cols: len(cols), ColLabels: cols}}
			for i, ev := range tr.Events {
				out.Table.RowLabels = append(out.Table.RowLabels, strconv.Itoa(i+1)+". "+ev.Label)
				if vector {
					out.Table.Cells = append(out.Table.Cells, vec[i]...)
				} else {
					out.Table.Cells = append(out.Table.Cells, lam[i])
				}
				out.Timeline = append(out.Timeline, Span{Lane: tr.Procs[ev.Proc], Label: ev.Label, Start: lam[i] - 1, End: lam[i], State: "running"})
			}
			concurrent := 0
			for i := range tr.Events {
				for j := i + 1; j < len(tr.Events); j++ {
					if distributed.Concurrent(vec[i], vec[j]) {
						concurrent++
						if vector && len(out.Lines) < 30 {
							out.Lines = append(out.Lines, fmt.Sprintf("%d (%s) and %d (%s) are concurrent", i+1, tr.Events[i].Label, j+1, tr.Events[j].Label))
						}
					}
				}
			}
			out.Summary = map[string]float64{"events": float64(len(tr.Events)), "concurrent pairs": float64(concurrent)}
			return out, nil
		}
		e.Call = func(in Input) (string, string) {
			return fmt.Sprintf("tr, err := distributed.ParseTrace(%s)\n\t%s", "`"+in.Text+"`", guard), "tr." + strings.TrimPrefix(fn, "Trace.") + "()"
		}
		return e
	}
	lam := clock("lamport-clock", "Lamport Clocks", "Trace.Lamport", "Count events per process and jump past the time a message carries: if a happened before b, a's timestamp is smaller.", "O(events)", "O(events)", false)
	vec := clock("vector-clock", "Vector Clocks", "Trace.Vector", "One counter per process in every timestamp: a happened before b exactly when a's vector is below b's, so concurrent events show.", "O(events * processes)", "O(events * processes)", true)

	bully := entry("bully-election", "Bully Election", "systems/distributed/election.go", "Bully")
	bully.Summary, bully.Time, bully.Space, bully.Input = "A process that notices the leader is gone challenges every higher id; the highest one alive bullies its way to leader.", "O(n^2) messages", "O(n^2)", Simulation
	bully.Options = []Option{{"processes", "processes", 5, 2, 26}, {"down", "highest ids crashed", 1, 0, 25}, {"starter", "process that starts", 1, 0, 25}}
	bully.Example = &Input{Options: map[string]int{"processes": 5, "down": 1, "starter": 1}}
	downOf := func(in Input) []int {
		var down []int
		n := in.opt("processes", 5)
		for i := 0; i < in.opt("down", 1) && i < n; i++ {
			down = append(down, n-1-i)
		}
		return down
	}
	bully.Run = func(in Input, _ trace.Recorder) (Output, error) {
		res, err := distributed.Bully(in.opt("processes", 5), downOf(in), in.opt("starter", 1))
		return electionOutput(res, err)
	}
	bully.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("distributed.Bully(%d, %s, %d)", in.opt("processes", 5), ints(downOf(in)), in.opt("starter", 1))
	}

	ring := entry("ring-election", "Ring Election (Chang-Roberts)", "systems/distributed/election.go", "Ring")
	ring.Summary, ring.Time, ring.Space, ring.Input = "Every process sends its id round a ring; larger ids are passed on, smaller dropped, and the id that comes home is the leader.", "O(n log n) average, O(n^2) worst messages", "O(n)", Ints
	ring.Example = &Input{Values: []int{3, 7, 1, 9, 4, 2}}
	ring.Limits = "distinct ids clockwise, 2 to 64"
	ring.Run = func(in Input, _ trace.Recorder) (Output, error) { return electionOutput(distributed.Ring(in.Values)) }
	ring.Call = func(in Input) (string, string) { return "", fmt.Sprintf("distributed.Ring(%s)", ints(in.Values)) }

	tpc := entry("two-phase-commit", "Two-Phase Commit", "systems/distributed/commit.go", "TwoPhaseCommit")
	tpc.Summary, tpc.Time, tpc.Space, tpc.Input = "A coordinator asks everyone to prepare, then commits only if all vote yes; if it crashes in between, the yes voters are stuck.", "O(n) ticks, 4n messages", "O(n)", Simulation
	tpc.Options = []Option{{"participants", "participants", 3, 1, 26}, {"no", "participant voting no (0 none)", 0, 0, 26}, {"crash", "coordinator crashes after the votes (1 yes)", 0, 0, 1}}
	tpc.Example = &Input{Options: map[string]int{"participants": 3, "no": 0, "crash": 1}}
	tpc.Run = func(in Input, _ trace.Recorder) (Output, error) {
		res, err := distributed.TwoPhaseCommit(in.opt("participants", 3), in.opt("no", 0), in.opt("crash", 0) == 1)
		if err != nil {
			return Output{}, wrap(err)
		}
		return Output{Text: res.Decision, Lines: res.Steps, Timeline: distSpans(res.Timeline),
			Summary: map[string]float64{"messages": float64(res.Messages), "blocked": float64(len(res.Blocked))}}, nil
	}
	tpc.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("distributed.TwoPhaseCommit(%d, %d, %v)", in.opt("participants", 3), in.opt("no", 0), in.opt("crash", 0) == 1)
	}

	raft := entry("raft", "Raft Consensus", "systems/distributed/raft.go", "Raft")
	raft.Summary, raft.Time, raft.Space, raft.Input = "Nodes elect a leader by majority vote in numbered terms, and the leader replicates a log, committing entries a majority stores.", "O(ticks * nodes^2)", "O(ticks * nodes)", Simulation
	raft.Options = []Option{{"nodes", "nodes", 5, 3, 9}, {"ticks", "ticks", 120, 1, 2000}, {"crash", "tick the leader crashes (0 never)", 60, 0, 2000}, {"every", "ticks between client commands (0 none)", 8, 0, 100}, seedOption}
	raft.Example = &Input{Options: map[string]int{"nodes": 5, "ticks": 120, "crash": 60, "every": 8, "seed": 1}}
	raftCfg := func(in Input) distributed.RaftConfig {
		return distributed.RaftConfig{Nodes: in.opt("nodes", 5), Ticks: in.opt("ticks", 120), Seed: int64(in.opt("seed", 1)), CrashAt: in.opt("crash", 60), Every: in.opt("every", 8)}
	}
	raft.Run = func(in Input, _ trace.Recorder) (Output, error) {
		res, err := distributed.Raft(raftCfg(in))
		if err != nil {
			return Output{}, wrap(err)
		}
		out := Output{Lines: res.Steps}
		state := map[string]string{"leader": "running", "candidate": "blocked", "follower": "idle", "crashed": "done"}
		for _, s := range res.Stretches {
			out.Timeline = append(out.Timeline, Span{Lane: "node " + strconv.Itoa(s.Node), Label: fmt.Sprintf("%s t%d", s.State, s.Term), Start: s.Start, End: s.End, State: state[s.State]})
		}
		term := 0
		for _, e := range res.Elections {
			term = max(term, e.Term)
		}
		out.Summary = map[string]float64{"elections": float64(len(res.Elections)), "last term": float64(term), "committed entries": float64(res.Committed), "messages": float64(res.Messages)}
		return out, nil
	}
	raft.Call = func(in Input) (string, string) {
		c := raftCfg(in)
		return "", fmt.Sprintf("distributed.Raft(distributed.RaftConfig{Nodes: %d, Ticks: %d, Seed: %d, CrashAt: %d, Every: %d})", c.Nodes, c.Ticks, c.Seed, c.CrashAt, c.Every)
	}

	paxos := entry("paxos", "Paxos (single decree)", "systems/distributed/paxos.go", "Paxos")
	paxos.Summary, paxos.Time, paxos.Space, paxos.Input = "Proposers race with numbered ballots; acceptors promise and accept, and once a majority accepts a value every later ballot carries it.", "O(attempts * proposers * acceptors) messages", "O(messages)", Simulation
	paxos.Options = []Option{{"proposers", "proposers (values 10, 20, ...)", 2, 1, 5}, {"acceptors", "acceptors", 3, 1, 9}, seedOption}
	paxos.Example = &Input{Options: map[string]int{"proposers": 2, "acceptors": 3, "seed": 4}}
	proposals := func(in Input) []int {
		vs := make([]int, in.opt("proposers", 2))
		for i := range vs {
			vs[i] = 10 * (i + 1)
		}
		return vs
	}
	paxos.Run = func(in Input, _ trace.Recorder) (Output, error) {
		res, err := distributed.Paxos(proposals(in), in.opt("acceptors", 3), int64(in.opt("seed", 1)))
		if err != nil {
			return Output{}, wrap(err)
		}
		out := Output{Lines: res.Steps, Timeline: distSpans(res.Timeline), Bool: ptr(res.Decided),
			Summary: map[string]float64{"ballots": float64(res.Ballots), "messages": float64(res.Messages), "decided": b2f(res.Decided)}}
		if res.Decided {
			out.Number = ptr(res.Chosen)
		}
		return out, nil
	}
	paxos.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("distributed.Paxos(%s, %d, %d)", ints(proposals(in)), in.opt("acceptors", 3), in.opt("seed", 1))
	}

	gossip := entry("gossip-protocol", "Gossip Protocol", "systems/distributed/gossip.go", "Gossip")
	gossip.Summary, gossip.Time, gossip.Space, gossip.Input = "Each node that knows a rumor tells a few random others every round: everyone knows in about log n rounds, and pulling finishes the stragglers.", "O(rounds * n * fanout)", "O(n)", Simulation
	gossip.Options = []Option{{"nodes", "nodes", 1000, 2, 100_000}, {"fanout", "nodes told per round", 1, 1, 50}, {"pull", "uninformed nodes also ask (1 yes)", 0, 0, 1}, seedOption}
	gossip.Example = &Input{Options: map[string]int{"nodes": 1000, "fanout": 1, "pull": 0, "seed": 1}}
	gossip.Run = func(in Input, _ trace.Recorder) (Output, error) {
		counts, err := distributed.Gossip(in.opt("nodes", 1000), in.opt("fanout", 1), in.opt("pull", 0) == 1, int64(in.opt("seed", 1)))
		if err != nil {
			return Output{}, wrap(err)
		}
		out := Output{Values: counts, Table: &TableOutput{Rows: len(counts), Cols: 1, ColLabels: []string{"informed"}, Cells: counts}}
		for i := range counts {
			out.Table.RowLabels = append(out.Table.RowLabels, "round "+strconv.Itoa(i))
		}
		n := in.opt("nodes", 1000)
		logn := 0
		for 1<<logn < n {
			logn++
		}
		out.Summary = map[string]float64{"rounds": float64(len(counts) - 1), "log2(n)": float64(logn)}
		return out, nil
	}
	gossip.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("distributed.Gossip(%d, %d, %v, %d)", in.opt("nodes", 1000), in.opt("fanout", 1), in.opt("pull", 0) == 1, in.opt("seed", 1))
	}

	hrw := entry("rendezvous-hashing", "Rendezvous Hashing", "systems/distributed/rendezvous.go", "Join")
	hrw.Summary, hrw.Time, hrw.Space, hrw.Input = "Every node scores every key and the highest score owns it: a new node takes only its share, while hash mod n reshuffles nearly all.", "O(keys * nodes)", "O(nodes)", Simulation
	hrw.Options = []Option{{"nodes", "nodes before the join", 4, 1, 100}, {"keys", "keys", 10_000, 1, 100_000}}
	hrw.Example = &Input{Options: map[string]int{"nodes": 4, "keys": 10_000}}
	hrw.Run = func(in Input, _ trace.Recorder) (Output, error) {
		res, err := distributed.Join(in.opt("nodes", 4), in.opt("keys", 10_000))
		if err != nil {
			return Output{}, wrap(err)
		}
		out := Output{Values: res.Load, Summary: map[string]float64{
			"keys": float64(res.Keys), "moved (rendezvous)": float64(res.Rendezvous), "moved (mod n)": float64(res.Modulo),
			"ideal share for the new node": float64(res.Keys) / float64(len(res.Load)),
		}}
		for i, l := range res.Load {
			out.Lines = append(out.Lines, fmt.Sprintf("node-%d owns %d keys", i, l))
		}
		return out, nil
	}
	hrw.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("distributed.Join(%d, %d)", in.opt("nodes", 4), in.opt("keys", 10_000))
	}

	quorum := entry("quorum-consistency", "Quorum Reads and Writes", "systems/distributed/quorum.go", "Quorum")
	quorum.Summary, quorum.Time, quorum.Space, quorum.Input = "Write to W of N replicas and read R: when R + W > N every read meets the latest write; below that, some reads come back stale.", "O(trials * n)", "O(n)", Simulation
	quorum.Options = []Option{{"n", "replicas", 5, 1, 50}, {"r", "read quorum R", 2, 1, 50}, {"w", "write quorum W", 2, 1, 50}, {"trials", "trials", 1000, 1, 100_000}, seedOption}
	quorum.Example = &Input{Options: map[string]int{"n": 5, "r": 2, "w": 2, "trials": 1000, "seed": 1}}
	quorum.Run = func(in Input, _ trace.Recorder) (Output, error) {
		res, err := distributed.Quorum(in.opt("n", 5), in.opt("r", 2), in.opt("w", 2), in.opt("trials", 1000), int64(in.opt("seed", 1)))
		if err != nil {
			return Output{}, wrap(err)
		}
		return Output{Lines: res.Steps, Bool: ptr(res.Overlap), Summary: map[string]float64{
			"trials": float64(res.Trials), "stale reads": float64(res.Stale), "R + W > N": b2f(res.Overlap),
		}}, nil
	}
	quorum.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("distributed.Quorum(%d, %d, %d, %d, %d)", in.opt("n", 5), in.opt("r", 2), in.opt("w", 2), in.opt("trials", 1000), in.opt("seed", 1))
	}

	gc := entry("g-counter", "G-Counter (CRDT)", "systems/distributed/crdt.go", "GCounter")
	gc.Summary, gc.Time, gc.Space, gc.Input = "Each replica counts its own increments; merging takes the larger count per replica, so replicas agree in any merge order.", "O(replicas) per merge", "O(replicas)", Text
	gc.Example = &Input{Text: "A inc 3\nB inc 2\nC inc 5\nA merge B\nB merge C\nC merge A\nA merge C\nB merge A"}
	gc.Limits = "one operation a line: \"A inc N\" or \"A merge B\""
	gc.Run = func(in Input, _ trace.Recorder) (Output, error) {
		run, err := distributed.RunGCounter(in.Text)
		if err != nil {
			return Output{}, wrap(err)
		}
		n := len(run.Replicas)
		out := Output{Lines: run.Steps, Table: &TableOutput{Rows: n, Cols: n + 1, RowLabels: run.Replicas, ColLabels: append(append([]string(nil), run.Replicas...), "value")}}
		for _, c := range run.Counters {
			for _, r := range run.Replicas {
				out.Table.Cells = append(out.Table.Cells, c[r])
			}
			out.Table.Cells = append(out.Table.Cells, c.Value())
		}
		return out, nil
	}
	gc.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("distributed.RunGCounter(%s)", "`"+in.Text+"`")
	}

	ors := entry("or-set", "OR-Set (CRDT)", "systems/distributed/crdt.go", "ORSet")
	ors.Summary, ors.Time, ors.Space, ors.Input = "A set whose adds carry unique tags and whose removes delete only the tags they saw, so a concurrent add survives a remove: add wins.", "O(tags) per merge", "O(tags)", Text
	ors.Example = &Input{Text: "A add milk\nA add eggs\nB merge A\nB remove milk\nA add milk\nA merge B\nB merge A"}
	ors.Limits = "one operation a line: \"A add x\", \"A remove x\" or \"A merge B\""
	ors.Run = func(in Input, _ trace.Recorder) (Output, error) {
		run, err := distributed.RunORSet(in.Text)
		if err != nil {
			return Output{}, wrap(err)
		}
		out := Output{Lines: run.Steps}
		for i, r := range run.Replicas {
			out.Lines = append(out.Lines, fmt.Sprintf("%s ends with {%s}", r, strings.Join(run.Sets[i], ", ")))
		}
		return out, nil
	}
	ors.Call = func(in Input) (string, string) { return "", fmt.Sprintf("distributed.RunORSet(%s)", "`"+in.Text+"`") }

	snap := entry("chandy-lamport-snapshot", "Chandy-Lamport Snapshot", "systems/distributed/snapshot.go", "ChandyLamport")
	snap.Summary, snap.Time, snap.Space, snap.Input = "Record a consistent global state while money keeps moving: markers split each channel into before and after the snapshot.", "O(transfers + n^2) messages", "O(n^2)", Simulation
	snap.Options = []Option{{"processes", "processes", 4, 2, 10}, {"transfers", "transfers", 40, 0, 1000}, {"at", "tick the snapshot starts", 15, 0, 10_000}, seedOption}
	snap.Example = &Input{Options: map[string]int{"processes": 4, "transfers": 40, "at": 15, "seed": 2}}
	snap.Run = func(in Input, _ trace.Recorder) (Output, error) {
		res, err := distributed.ChandyLamport(in.opt("processes", 4), in.opt("transfers", 40), in.opt("at", 15), int64(in.opt("seed", 1)))
		if err != nil {
			return Output{}, wrap(err)
		}
		n := len(res.Balances)
		out := Output{Lines: res.Steps, Timeline: distSpans(res.Timeline), Table: &TableOutput{Rows: n, Cols: 1, ColLabels: []string{"recorded balance"}, Cells: res.Balances}}
		for i := 0; i < n; i++ {
			out.Table.RowLabels = append(out.Table.RowLabels, "P"+strconv.Itoa(i))
		}
		chans := make([]string, 0, len(res.Channels))
		for c := range res.Channels {
			chans = append(chans, c)
		}
		sort.Strings(chans)
		for _, c := range chans {
			out.Lines = append(out.Lines, fmt.Sprintf("in flight on %s: %d", c, res.Channels[c]))
		}
		out.Summary = map[string]float64{"recorded total": float64(res.Total), "expected total": float64(res.Expected)}
		return out, nil
	}
	snap.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("distributed.ChandyLamport(%d, %d, %d, %d)", in.opt("processes", 4), in.opt("transfers", 40), in.opt("at", 15), in.opt("seed", 1))
	}

	byz := entry("byzantine-generals", "Byzantine Generals (OM(m))", "systems/distributed/byzantine.go", "Byzantine")
	byz.Summary, byz.Time, byz.Space, byz.Input = "Generals relay orders through rounds of oral messages and take majorities; loyal ones agree despite m traitors only when n >= 3m + 1.", "O(n^(m+1)) messages", "O(n^(m+1))", Simulation
	byz.Options = []Option{{"generals", "generals (commander included)", 4, 2, 10}, {"rounds", "rounds m", 1, 0, 3}, {"traitors", "traitorous lieutenants (the highest numbered)", 1, 0, 9}, {"commander", "the commander is a traitor (1 yes)", 0, 0, 1}, {"attack", "the order: attack (1) or retreat (0)", 1, 0, 1}}
	byz.Example = &Input{Options: map[string]int{"generals": 4, "rounds": 1, "traitors": 1, "commander": 0, "attack": 1}}
	traitorsOf := func(in Input) []int {
		n := in.opt("generals", 4)
		var ts []int
		if in.opt("commander", 0) == 1 {
			ts = append(ts, 0)
		}
		for i := 0; i < in.opt("traitors", 1) && n-1-i > 0; i++ {
			ts = append(ts, n-1-i)
		}
		return ts
	}
	byz.Run = func(in Input, _ trace.Recorder) (Output, error) {
		ts := traitorsOf(in)
		res, err := distributed.Byzantine(in.opt("generals", 4), in.opt("rounds", 1), distributed.Order(in.opt("attack", 1)), ts)
		if err != nil {
			return Output{}, wrap(err)
		}
		out := Output{Bool: ptr(res.Agree), Summary: map[string]float64{"messages": float64(res.Messages), "loyal agree": b2f(res.Agree), "loyal obey": b2f(res.Obey)}}
		for i, d := range res.Decisions {
			who := "loyal"
			for _, t := range ts {
				if t == i+1 {
					who = "traitor"
				}
			}
			out.Lines = append(out.Lines, fmt.Sprintf("lieutenant %d (%s) decides %s", i+1, who, d))
		}
		return out, nil
	}
	byz.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("distributed.Byzantine(%d, %d, distributed.Order(%d), %s)", in.opt("generals", 4), in.opt("rounds", 1), in.opt("attack", 1), ints(traitorsOf(in)))
	}

	return []Entry{lam, vec, bully, ring, tpc, raft, paxos, gossip, hrw, quorum, gc, ors, snap, byz}
}

func electionOutput(res distributed.ElectionResult, err error) (Output, error) {
	if err != nil {
		return Output{}, wrap(err)
	}
	return Output{Number: ptr(res.Leader), Lines: res.Steps, Timeline: distSpans(res.Timeline),
		Summary: map[string]float64{"leader": float64(res.Leader), "messages": float64(res.Messages), "ticks": float64(res.Ticks)}}, nil
}
