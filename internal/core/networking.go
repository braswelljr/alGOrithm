package core

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/systems/networking"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// netSpans turns a networking timeline into one-tick timeline bars.
func netSpans(tl networking.Timeline) []Span {
	out := make([]Span, len(tl))
	for i, e := range tl {
		out[i] = Span{Lane: e.Lane, Label: e.What, Start: e.Tick, End: e.Tick + 1, State: e.State}
	}
	return out
}

// guard is the setup line that stops a snippet on an error.
const guard = "if err != nil {\n\t\tpanic(err)\n\t}\n\t"

func networkingEntries() []Entry {
	link := []Option{
		{"frames", "frames to send", 8, 1, 200},
		{"loss", "loss percent", 20, 0, 90},
		{"delay", "one-way delay (ticks)", 1, 1, 10},
		{"timeout", "timeout (ticks)", 4, 3, 100},
		seedOption,
	}
	channel := func(in Input) networking.Channel {
		return networking.Channel{Loss: in.opt("loss", 20), Delay: in.opt("delay", 1), Timeout: in.opt("timeout", 4), Seed: int64(in.opt("seed", 1))}
	}
	channelCall := func(in Input) string {
		return fmt.Sprintf("networking.Channel{Loss: %d, Delay: %d, Timeout: %d, Seed: %d}", in.opt("loss", 20), in.opt("delay", 1), in.opt("timeout", 4), in.opt("seed", 1))
	}
	arqOutput := func(res networking.ARQResult, err error) (Output, error) {
		if err != nil {
			return Output{}, wrap(err)
		}
		return Output{Timeline: netSpans(res.Timeline), Values: res.Delivered, Summary: map[string]float64{
			"transmissions": float64(res.Sent), "retransmissions": float64(res.Retransmitted),
			"lost": float64(res.Lost), "ticks": float64(res.Ticks),
		}}, nil
	}
	arqEntry := func(id, name, fn, summary, time string, windowed bool, run func(Input) (networking.ARQResult, error)) Entry {
		e := entry(id, name, "systems/networking/arq.go", fn)
		e.Summary, e.Time, e.Space, e.Input = summary, time, "O(frames + ticks)", Simulation
		e.Options = append([]Option(nil), link...)
		ex := map[string]int{"frames": 8, "loss": 20, "delay": 1, "timeout": 5, "seed": 3}
		if windowed {
			e.Options = append(e.Options, Option{"window", "window size", 4, 1, 64})
			ex["window"] = 4
		}
		e.Example = &Input{Options: ex}
		e.Run = func(in Input, _ trace.Recorder) (Output, error) { return arqOutput(run(in)) }
		e.Call = func(in Input) (string, string) {
			if windowed {
				return "", fmt.Sprintf("networking.%s(%d, %d, %s)", fn, in.opt("frames", 8), in.opt("window", 4), channelCall(in))
			}
			return "", fmt.Sprintf("networking.%s(%d, %s)", fn, in.opt("frames", 8), channelCall(in))
		}
		return e
	}
	saw := arqEntry("stop-and-wait", "Stop-and-Wait ARQ", "StopAndWait",
		"Send one frame, wait for its ack, resend on a timeout: reliable over a lossy link, but one frame per round trip.", "O(ticks)", false,
		func(in Input) (networking.ARQResult, error) {
			return networking.StopAndWait(in.opt("frames", 8), channel(in))
		})
	gbn := arqEntry("go-back-n", "Go-Back-N ARQ", "GoBackN",
		"Keep a window of frames in flight; the receiver takes only the next in order, so one loss resends the whole window after it.", "O(ticks * window)", true,
		func(in Input) (networking.ARQResult, error) {
			return networking.GoBackN(in.opt("frames", 8), in.opt("window", 4), channel(in))
		})
	sr := arqEntry("selective-repeat", "Selective Repeat ARQ", "SelectiveRepeat",
		"A window of frames in flight, buffered out of order and acked one by one, so only the lost frames are resent.", "O(ticks * window)", true,
		func(in Input) (networking.ARQResult, error) {
			return networking.SelectiveRepeat(in.opt("frames", 8), in.opt("window", 4), channel(in))
		})

	cc := entry("tcp-congestion-control", "TCP Congestion Control", "systems/networking/congestion.go", "Congestion")
	cc.Summary, cc.Time, cc.Space, cc.Input = "Slow start doubles the window each round trip, congestion avoidance adds one, and a loss halves the threshold: TCP's sawtooth.", "O(rounds)", "O(rounds)", Simulation
	cc.Options = []Option{{"rounds", "round trips", 20, 1, 1000}, {"ssthresh", "slow-start threshold", 16, 2, 1 << 20},
		{"dupack", "round of a triple duplicate ack (0 none)", 9, 0, 1000}, {"timeout", "round of a timeout (0 none)", 15, 0, 1000}, {"reno", "Reno fast recovery (1) or Tahoe (0)", 1, 0, 1}}
	cc.Example = &Input{Options: map[string]int{"rounds": 20, "ssthresh": 16, "dupack": 9, "timeout": 15, "reno": 1}}
	losses := func(in Input) map[int]networking.LossKind {
		m := map[int]networking.LossKind{}
		if r := in.opt("dupack", 9); r > 0 {
			m[r] = networking.TripleDupAck
		}
		if r := in.opt("timeout", 15); r > 0 {
			m[r] = networking.Timeout
		}
		return m
	}
	cc.Run = func(in Input, _ trace.Recorder) (Output, error) {
		rs, err := networking.Congestion(in.opt("rounds", 20), in.opt("ssthresh", 16), losses(in), in.opt("reno", 1) == 1)
		if err != nil {
			return Output{}, wrap(err)
		}
		out := Output{Table: &TableOutput{Rows: len(rs), Cols: 2, ColLabels: []string{"cwnd", "ssthresh"}}}
		for _, r := range rs {
			out.Values = append(out.Values, r.Cwnd)
			out.Table.Cells = append(out.Table.Cells, r.Cwnd, r.Ssthresh)
			out.Table.RowLabels = append(out.Table.RowLabels, "round "+strconv.Itoa(r.Round))
			out.Lines = append(out.Lines, fmt.Sprintf("round %d: cwnd %d, ssthresh %d (%s)", r.Round, r.Cwnd, r.Ssthresh, r.Phase))
		}
		return out, nil
	}
	cc.Call = func(in Input) (string, string) {
		var parts []string
		if r := in.opt("dupack", 9); r > 0 {
			parts = append(parts, fmt.Sprintf("%d: networking.TripleDupAck", r))
		}
		if r := in.opt("timeout", 15); r > 0 {
			parts = append(parts, fmt.Sprintf("%d: networking.Timeout", r))
		}
		return "", fmt.Sprintf("networking.Congestion(%d, %d, map[int]networking.LossKind{%s}, %v)", in.opt("rounds", 20), in.opt("ssthresh", 16), strings.Join(parts, ", "), in.opt("reno", 1) == 1)
	}

	bucket := func(id, name, fn, summary string, f func([]int, int, int) ([]networking.Decision, error)) Entry {
		e := entry(id, name, "systems/networking/ratelimit.go", fn)
		e.Summary, e.Time, e.Space, e.Input = summary, "O(n log n)", "O(n)", Ints
		e.Options = []Option{{"capacity", "bucket capacity", 3, 1, 10_000}, {"interval", "ticks per token or send", 4, 1, 10_000}}
		e.Example = &Input{Values: []int{0, 0, 0, 0, 1, 2, 9, 10, 11, 20, 21, 22, 23}, Options: map[string]int{"capacity": 3, "interval": 4}}
		e.Limits = "arrival ticks, at most 10000"
		e.Run = func(in Input, _ trace.Recorder) (Output, error) {
			ds, err := f(in.Values, in.opt("capacity", 3), in.opt("interval", 4))
			if err != nil {
				return Output{}, wrap(err)
			}
			out := Output{Table: &TableOutput{Rows: len(ds), Cols: 3, ColLabels: []string{"arrival", "left", "served at"}, Blank: ptr(-1)}}
			allowed := 0
			for i, d := range ds {
				out.Table.Cells = append(out.Table.Cells, d.Arrival, d.Left, d.Served)
				out.Table.RowLabels = append(out.Table.RowLabels, "request "+strconv.Itoa(i+1))
				label, state := "drop", "blocked"
				if d.Allowed {
					allowed++
					label, state = "pass", "running"
				}
				out.Timeline = append(out.Timeline, Span{Lane: "arrivals", Label: label, Start: d.Arrival, End: d.Arrival + 1, State: state})
				if d.Served >= 0 {
					out.Timeline = append(out.Timeline, Span{Lane: "sent", Label: "request " + strconv.Itoa(i+1), Start: d.Served, End: d.Served + 1, State: "done"})
				}
			}
			out.Summary = map[string]float64{"allowed": float64(allowed), "dropped": float64(len(ds) - allowed)}
			return out, nil
		}
		e.Call = func(in Input) (string, string) {
			return "", fmt.Sprintf("networking.%s(%s, %d, %d)", fn, ints(in.Values), in.opt("capacity", 3), in.opt("interval", 4))
		}
		return e
	}
	tb := bucket("token-bucket", "Token Bucket", "TokenBucket", "Tokens drip into a bucket; each request takes one or is rejected: bursts up to the bucket's size pass at once, the long-run rate is capped.", networking.TokenBucket)
	lb := bucket("leaky-bucket", "Leaky Bucket", "LeakyBucket", "Requests wait in a bounded queue that drains at a fixed rate: bursts are smoothed into an even stream, and overflow is dropped.", networking.LeakyBucket)

	const topology = "A B 1\nB C 1\nC D 2\nA D 5"
	dv := entry("distance-vector-routing", "Distance-Vector Routing", "systems/networking/routing.go", "DistanceVector")
	dv.Summary, dv.Time, dv.Space, dv.Input = "Routers swap distance vectors with their neighbors each round (Bellman-Ford, as RIP does); cut a link and, without poisoned reverse, they count to infinity.", "O(rounds * V^2 * degree)", "O(rounds * V^2)", GraphInput
	dv.Params = []string{"key"}
	dv.Options = []Option{{"poison", "poisoned reverse (1 yes, 0 no)", 0, 0, 1}, {"rounds", "most rounds per phase", 30, 1, 100}}
	dv.Example = &Input{Text: "A B 1\nB C 1", Key: "B C", Options: map[string]int{"poison": 0, "rounds": 30}}
	dv.Limits = "2 to 26 routers; Key names a link to cut, \"A B\""
	dv.Run = func(in Input, _ trace.Recorder) (Output, error) {
		t, err := networking.ParseTopology(in.Text)
		if err != nil {
			return Output{}, wrap(err)
		}
		rounds, cutAt, err := networking.DistanceVector(t, in.Key, in.opt("poison", 0) == 1, in.opt("rounds", 30))
		if err != nil {
			return Output{}, wrap(err)
		}
		last := rounds[len(rounds)-1]
		out := Output{Table: distTable(last.Dist, t.Routers)}
		for i, r := range rounds {
			out.Lines = append(out.Lines, fmt.Sprintf("round %d: %s", i, strings.Join(r.Changes, "; ")))
		}
		out.Summary = map[string]float64{"rounds": float64(len(rounds) - 1)}
		if cutAt >= 0 {
			out.Summary["rounds after the cut"] = float64(len(rounds) - 1 - cutAt)
		}
		return out, nil
	}
	dv.Call = func(in Input) (string, string) {
		return fmt.Sprintf("t, err := networking.ParseTopology(%s)\n\t%s", "`"+in.Text+"`", guard),
			fmt.Sprintf("networking.DistanceVector(t, %q, %v, %d)", in.Key, in.opt("poison", 0) == 1, in.opt("rounds", 30))
	}

	ls := entry("link-state-routing", "Link-State Routing", "systems/networking/routing.go", "LinkState")
	ls.Summary, ls.Time, ls.Space, ls.Input = "Every router learns the whole map and runs Dijkstra from itself (as OSPF does) to build its forwarding table.", "O(V^3)", "O(V^2)", GraphInput
	ls.Example = &Input{Text: topology}
	ls.Limits = "2 to 26 routers"
	ls.Run = func(in Input, _ trace.Recorder) (Output, error) {
		t, err := networking.ParseTopology(in.Text)
		if err != nil {
			return Output{}, wrap(err)
		}
		tables := networking.LinkState(t)
		n := len(t.Routers)
		out := Output{Table: &TableOutput{Rows: n, Cols: n, RowLabels: t.Routers, ColLabels: t.Routers, Blank: ptr(-1)}}
		for i, table := range tables {
			var hops []string
			for _, r := range table {
				out.Table.Cells = append(out.Table.Cells, r.Cost)
				if r.Next != "" {
					hops = append(hops, fmt.Sprintf("%s via %s (%d)", r.To, r.Next, r.Cost))
				}
			}
			out.Lines = append(out.Lines, t.Routers[i]+": "+strings.Join(hops, ", "))
		}
		return out, nil
	}
	ls.Call = func(in Input) (string, string) {
		return fmt.Sprintf("t, err := networking.ParseTopology(%s)\n\t%s", "`"+in.Text+"`", guard), "networking.LinkState(t)"
	}

	sub := entry("subnetting", "Subnetting (CIDR)", "systems/networking/ipv4.go", "Split")
	sub.Summary, sub.Time, sub.Space, sub.Input = "A CIDR block's network, broadcast and host range, split into equal subnets by borrowing host bits.", "O(subnets)", "O(subnets)", Text
	sub.Options = []Option{{"subnets", "subnets (a power of two)", 4, 1, 4096}}
	sub.Example = &Input{Text: "192.168.10.0/24", Options: map[string]int{"subnets": 4}}
	sub.Run = func(in Input, _ trace.Recorder) (Output, error) {
		p, err := networking.ParsePrefix(in.Text)
		if err != nil {
			return Output{}, wrap(err)
		}
		subs, err := networking.Split(p, in.opt("subnets", 4))
		if err != nil {
			return Output{}, wrap(err)
		}
		out := Output{Text: p.String(), Summary: map[string]float64{"hosts": float64(p.Hosts()), "subnets": float64(len(subs))}}
		out.Lines = append(out.Lines, fmt.Sprintf("%s: mask %s, network %s, broadcast %s, %d hosts", p, p.Mask(), p.Network(), p.Broadcast(), p.Hosts()))
		for _, s := range subs {
			first, last := s.Network()+1, s.Broadcast()-1
			if s.Bits >= 31 {
				first, last = s.Network(), s.Broadcast()
			}
			out.Lines = append(out.Lines, fmt.Sprintf("%s: hosts %s to %s, broadcast %s (%d hosts)", s, first, last, s.Broadcast(), s.Hosts()))
		}
		return out, nil
	}
	sub.Call = func(in Input) (string, string) {
		return fmt.Sprintf("p, err := networking.ParsePrefix(%q)\n\t%s", in.Text, guard), fmt.Sprintf("networking.Split(p, %d)", in.opt("subnets", 4))
	}

	lpm := entry("longest-prefix-match", "Longest Prefix Match", "systems/networking/ipv4.go", "RoutingTable.Lookup")
	lpm.Summary, lpm.Time, lpm.Space, lpm.Input = "A router forwards each address by the most specific route that contains it, found by walking a binary trie bit by bit.", "O(32) per lookup", "O(routes * 32)", TwoTexts
	lpm.Example = &Input{Text: "0.0.0.0/0 isp\n10.0.0.0/8 core\n10.1.0.0/16 east\n10.1.2.0/24 lab", Other: "10.1.2.3 10.1.9.9 10.200.0.1 8.8.8.8"}
	lpm.Limits = "Text: one route a line, \"prefix/len hop\"; Other: addresses"
	lpm.Run = func(in Input, _ trace.Recorder) (Output, error) {
		rt, err := networking.ParseRoutes(in.Text)
		if err != nil {
			return Output{}, wrap(err)
		}
		var out Output
		for _, s := range strings.Fields(in.Other) {
			a, err := networking.ParseAddr(s)
			if err != nil {
				return Output{}, wrap(err)
			}
			best, matched, ok := rt.Lookup(a)
			var ps []string
			for _, p := range matched {
				ps = append(ps, p.String())
			}
			if !ok {
				out.Lines = append(out.Lines, s+": no route")
				continue
			}
			out.Lines = append(out.Lines, fmt.Sprintf("%s -> %s via %s (matched %s)", s, best.Hop, best.Prefix, strings.Join(ps, ", ")))
		}
		if len(out.Lines) == 0 {
			return Output{}, inputError("give the addresses to look up")
		}
		return out, nil
	}
	lpm.Call = func(in Input) (string, string) {
		addr := ""
		if f := strings.Fields(in.Other); len(f) > 0 {
			addr = f[0]
		}
		return fmt.Sprintf("rt, err := networking.ParseRoutes(%s)\n\t%sa, err := networking.ParseAddr(%q)\n\t%s", "`"+in.Text+"`", guard, addr, guard), "rt.Lookup(a)"
	}

	ck := entry("internet-checksum", "Internet Checksum", "systems/networking/checksum.go", "Checksum")
	ck.Summary, ck.Time, ck.Space, ck.Input = "Add the data as 16-bit words with the carries folded back in, then invert: the check in every IPv4, UDP and TCP header.", "O(n)", "O(1)", Text
	ck.Example = &Input{Text: "Hello, network!"}
	ck.Run = func(in Input, _ trace.Recorder) (Output, error) {
		if len(in.Text) > 1<<16 {
			return Output{}, inputError("at most 64 KiB")
		}
		data := []byte(in.Text)
		c := networking.Checksum(data)
		words, sums := networking.ChecksumSteps(data)
		out := Output{Text: fmt.Sprintf("0x%04x", c), Number: ptr(int(c))}
		for i := range words {
			out.Lines = append(out.Lines, fmt.Sprintf("+ 0x%04x = 0x%04x", words[i], sums[i]))
		}
		out.Lines = append(out.Lines, fmt.Sprintf("invert: 0x%04x", c),
			fmt.Sprintf("check: the data with its checksum sums to 0x%04x", networking.Checksum(append(data, byte(c>>8), byte(c)))))
		return out, nil
	}
	ck.Call = func(in Input) (string, string) { return "", fmt.Sprintf("networking.Checksum([]byte(%q))", in.Text) }

	ng := entry("nagles-algorithm", "Nagle's Algorithm", "systems/networking/nagle.go", "Nagle")
	ng.Summary, ng.Time, ng.Space, ng.Input = "Hold small writes back while data is unacknowledged and send them together: far fewer tiny packets, at the cost of some delay.", "O(ticks)", "O(segments)", Ints
	ng.Options = []Option{{"mss", "maximum segment size (bytes)", 10, 1, 65535}, {"rtt", "round trip (ticks)", 3, 1, 100}, {"nagle", "Nagle on (1) or off (0)", 1, 0, 1}}
	ng.Example = &Input{Values: []int{1, 1, 1, 1, 1, 1, 1, 1, 12, 1, 1, 1}, Options: map[string]int{"mss": 10, "rtt": 3, "nagle": 1}}
	ng.Limits = "bytes written at each tick, at most 1000 writes"
	ng.Run = func(in Input, _ trace.Recorder) (Output, error) {
		res, err := networking.Nagle(in.Values, in.opt("mss", 10), in.opt("rtt", 3), in.opt("nagle", 1) == 1)
		if err != nil {
			return Output{}, wrap(err)
		}
		bytes := 0
		for _, s := range res.Segments {
			bytes += s.Bytes
		}
		return Output{Timeline: netSpans(res.Timeline), Summary: map[string]float64{
			"segments": float64(len(res.Segments)), "bytes": float64(bytes),
			"header bytes (40 each)": float64(40 * len(res.Segments)), "byte-ticks waited": float64(res.Delay),
		}}, nil
	}
	ng.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("networking.Nagle(%s, %d, %d, %v)", ints(in.Values), in.opt("mss", 10), in.opt("rtt", 3), in.opt("nagle", 1) == 1)
	}

	dns := entry("dns-resolution", "DNS Resolution", "systems/networking/dns.go", "Resolver.Resolve")
	dns.Summary, dns.Time, dns.Space, dns.Input = "Resolve names down a simulated hierarchy, root to top-level domain to the domain's server, iteratively or recursively, with a cache.", "O(labels) messages", "O(names)", Text
	dns.Options = []Option{{"recursive", "recursive (1) or iterative (0)", 0, 0, 1}}
	dns.Example = &Input{Text: "www.example.com mail.example.com www.example.com pkg.go.dev nope.example.com", Options: map[string]int{"recursive": 0}}
	dns.Limits = "names in example.com, shop.com, wikipedia.org, go.dev, mit.edu"
	dns.Run = func(in Input, _ trace.Recorder) (Output, error) {
		r := networking.NewResolver(in.opt("recursive", 0) == 1)
		var out Output
		total := 0
		for _, name := range strings.Fields(in.Text) {
			lk, err := r.Resolve(name)
			if err != nil {
				return Output{}, wrap(err)
			}
			total += lk.Messages
			out.Lines = append(out.Lines, fmt.Sprintf("%s = %s (%d messages)", lk.Name, lk.Answer, lk.Messages))
			for _, s := range lk.Steps {
				out.Lines = append(out.Lines, "  "+s)
			}
		}
		if len(out.Lines) == 0 {
			return Output{}, inputError("give the names to resolve")
		}
		out.Timeline = netSpans(r.Timeline)
		out.Summary = map[string]float64{"messages": float64(total)}
		return out, nil
	}
	dns.Call = func(in Input) (string, string) {
		name := ""
		if f := strings.Fields(in.Text); len(f) > 0 {
			name = f[0]
		}
		return fmt.Sprintf("r := networking.NewResolver(%v)\n\t", in.opt("recursive", 0) == 1), fmt.Sprintf("r.Resolve(%q)", name)
	}

	return append([]Entry{saw, gbn, sr, cc, tb, lb, dv, ls, sub, lpm, ck, ng, dns}, moreNetworkingEntries()...)
}

// distTable is a routers-by-destinations table of distances, with
// networking.Infinity shown blank.
func distTable(dist [][]int, names []string) *TableOutput {
	n := len(names)
	t := &TableOutput{Rows: n, Cols: n, RowLabels: names, ColLabels: names, Blank: ptr(networking.Infinity)}
	for _, row := range dist {
		t.Cells = append(t.Cells, row...)
	}
	return t
}
