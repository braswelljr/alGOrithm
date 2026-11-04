package core

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/systems/networking"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

func phaseSpans(ps []networking.Phase) []Span {
	out := make([]Span, len(ps))
	for i, p := range ps {
		out[i] = Span{Lane: p.Lane, Label: p.Name, Start: p.Start, End: p.End, State: "running"}
	}
	return out
}

func netEntry(id, name, file, fn, summary, time, space string, kind InputKind) Entry {
	e := entry(id, name, "systems/networking/"+file, fn)
	e.Summary, e.Time, e.Space, e.Input = summary, time, space, kind
	return e
}

// moreNetworkingEntries are the protocols from the link layer up: layers,
// TCP's states, the LAN services, the web's protocols, medium access and
// inter-network routing, and the arithmetic of delay.
func moreNetworkingEntries() []Entry {
	var out []Entry

	layers := netEntry("network-layers", "OSI and TCP/IP Layers", "layers.go", "Encapsulate", "Data goes down the stack and each layer wraps it in its own header: HTTP in TCP in IP in an Ethernet frame, stripped again in reverse.", "O(layers)", "O(layers)", Text)
	layers.Options = []Option{{"osi", "OSI's seven layers (1) or TCP/IP's four (0)", 0, 0, 1}}
	layers.Example = &Input{Text: "Hello", Options: map[string]int{"osi": 0}}
	layers.Run = func(in Input, _ trace.Recorder) (Output, error) {
		steps, err := networking.Encapsulate(in.Text, in.opt("osi", 0) == 1)
		if err != nil {
			return Output{}, wrap(err)
		}
		o := Output{Table: &TableOutput{Rows: len(steps), Cols: 3, ColLabels: []string{"header", "trailer", "total bytes"}}}
		for _, s := range steps {
			o.Table.RowLabels = append(o.Table.RowLabels, s.Layer+" ("+s.PDU+")")
			o.Table.Cells = append(o.Table.Cells, s.Header, s.Trailer, s.Total)
			o.Lines = append(o.Lines, fmt.Sprintf("%s: %s. %s", s.Layer, s.Does, s.Shape))
		}
		o.Lines = append(o.Lines, networking.Decapsulate(steps)...)
		return o, nil
	}
	layers.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("networking.Encapsulate(%q, %v)", in.Text, in.opt("osi", 0) == 1)
	}
	out = append(out, layers)

	hs := netEntry("tcp-handshake", "TCP Handshake and Teardown", "tcpstate.go", "TCPConnection", "TCP's state machine: SYN, SYN+ACK, ACK to open; FIN, ACK, FIN, ACK to close, with TIME_WAIT on the side that closed first.", "O(segments)", "O(segments)", Simulation)
	hs.Options = []Option{{"data", "data segments", 2, 0, 100}, {"server", "the server closes (1) or the client (0)", 0, 0, 1}}
	hs.Example = &Input{Options: map[string]int{"data": 2, "server": 0}}
	closer := func(in Input) string {
		if in.opt("server", 0) == 1 {
			return "server"
		}
		return "client"
	}
	hs.Run = func(in Input, _ trace.Recorder) (Output, error) {
		steps, err := networking.TCPConnection(in.opt("data", 2), closer(in))
		if err != nil {
			return Output{}, wrap(err)
		}
		var o Output
		last := map[string]int{}
		cur := map[string]string{}
		end := steps[len(steps)-1].Tick + 1
		for _, s := range steps {
			line := fmt.Sprintf("t%d %s: %s", s.Tick, s.Side, s.From)
			if s.From != s.To {
				line += " -> " + s.To
			}
			if s.Segment != "" {
				line += ", sends " + s.Segment
			}
			o.Lines = append(o.Lines, line)
			if s.From != s.To {
				if st, ok := cur[s.Side]; ok {
					o.Timeline = append(o.Timeline, Span{Lane: s.Side, Label: st, Start: last[s.Side], End: s.Tick, State: tcpState(st)})
				}
				cur[s.Side], last[s.Side] = s.To, s.Tick
			}
		}
		o.Machine = tcpMachine(steps)
		for _, side := range []string{"client", "server"} {
			if cur[side] != "CLOSED" {
				o.Timeline = append(o.Timeline, Span{Lane: side, Label: cur[side], Start: last[side], End: end, State: tcpState(cur[side])})
			}
		}
		return o, nil
	}
	hs.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("networking.TCPConnection(%d, %q)", in.opt("data", 2), closer(in))
	}
	out = append(out, hs)

	fc := netEntry("tcp-flow-control", "TCP Flow Control", "tcpstate.go", "FlowControl", "The receiver advertises the room left in its buffer and the sender never exceeds it; a shut window is reopened by zero-window probes.", "O(ticks)", "O(ticks)", Simulation)
	fc.Options = []Option{{"bytes", "bytes to send", 100, 1, 1 << 20}, {"buffer", "receive buffer", 20, 1, 1 << 20}, {"read", "bytes the app reads a tick", 2, 1, 1 << 20}, {"mss", "segment size", 10, 1, 65535}, {"probe", "ticks between probes (0 none)", 4, 0, 100}}
	fc.Example = &Input{Options: map[string]int{"bytes": 100, "buffer": 20, "read": 2, "mss": 10, "probe": 4}}
	fc.Run = func(in Input, _ trace.Recorder) (Output, error) {
		ticks, stalled, err := networking.FlowControl(in.opt("bytes", 100), in.opt("buffer", 20), in.opt("read", 2), in.opt("mss", 10), in.opt("probe", 4))
		if err != nil {
			return Output{}, wrap(err)
		}
		o := Output{Table: &TableOutput{Rows: len(ticks), Cols: 3, ColLabels: []string{"sent", "buffered", "window"}}, Bool: ptr(stalled)}
		probes := 0
		for _, ft := range ticks {
			o.Table.RowLabels = append(o.Table.RowLabels, "t"+strconv.Itoa(ft.Tick))
			o.Table.Cells = append(o.Table.Cells, ft.Sent, ft.Buffered, ft.Window)
			switch {
			case ft.Sent > 0:
				o.Timeline = append(o.Timeline, Span{Lane: "sender", Label: fmt.Sprintf("send %d", ft.Sent), Start: ft.Tick, End: ft.Tick + 1, State: "running"})
			case ft.Probe:
				probes++
				o.Timeline = append(o.Timeline, Span{Lane: "sender", Label: "zero-window probe", Start: ft.Tick, End: ft.Tick + 1, State: "idle"})
			default:
				o.Timeline = append(o.Timeline, Span{Lane: "sender", Label: "window shut", Start: ft.Tick, End: ft.Tick + 1, State: "blocked"})
			}
		}
		o.Summary = map[string]float64{"ticks": float64(len(ticks)), "probes": float64(probes), "stalled": b2f(stalled)}
		return o, nil
	}
	fc.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("networking.FlowControl(%d, %d, %d, %d, %d)", in.opt("bytes", 100), in.opt("buffer", 20), in.opt("read", 2), in.opt("mss", 10), in.opt("probe", 4))
	}
	out = append(out, fc)

	arp := netEntry("arp", "ARP", "lan.go", "ARP", "Find the MAC address behind an IP address on the local network: check the cache, else broadcast who-has and cache the reply.", "O(sends)", "O(cache)", Text)
	arp.Options = []Option{{"ttl", "cache lifetime (sends)", 4, 1, 1000}}
	arp.Example = &Input{Text: "10.0.0.3 10.0.0.3 8.8.8.8 1.1.1.1 10.0.0.5 10.0.0.42 10.0.0.3", Options: map[string]int{"ttl": 4}}
	arp.Limits = "addresses; the LAN is 10.0.0.0/24 with hosts .1 to .9, the gateway at .1"
	arp.Run = func(in Input, _ trace.Recorder) (Output, error) {
		res, err := networking.ARP(strings.Fields(in.Text), in.opt("ttl", 4))
		if err != nil {
			return Output{}, wrap(err)
		}
		return Output{Lines: res.Steps, Timeline: netSpans(res.Timeline), Summary: map[string]float64{"broadcasts": float64(res.Broadcasts), "cache hits": float64(res.Hits), "unanswered": float64(res.Failures)}}, nil
	}
	arp.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("networking.ARP(%#v, %d)", strings.Fields(in.Text), in.opt("ttl", 4))
	}
	out = append(out, arp)

	dhcp := netEntry("dhcp", "DHCP (DORA)", "lan.go", "DHCP", "A new device has no address: it broadcasts Discover, the server Offers one, the device Requests it, and the server Acknowledges a lease.", "O(clients * pool)", "O(pool)", Simulation)
	dhcp.Options = []Option{{"clients", "clients", 4, 1, 50}, {"pool", "addresses in the pool", 3, 0, 100}, {"lease", "lease (ticks)", 3600, 1, 1 << 30}, {"release", "client that releases its address halfway (0 none)", 0, 0, 50}}
	dhcp.Example = &Input{Options: map[string]int{"clients": 4, "pool": 3, "lease": 3600, "release": 0}}
	releases := func(in Input) []int {
		if r := in.opt("release", 0); r > 0 {
			return []int{r}
		}
		return nil
	}
	dhcp.Run = func(in Input, _ trace.Recorder) (Output, error) {
		res, err := networking.DHCP(in.opt("clients", 4), in.opt("pool", 3), in.opt("lease", 3600), releases(in))
		if err != nil {
			return Output{}, wrap(err)
		}
		return Output{Lines: res.Steps, Timeline: netSpans(res.Timeline), Summary: map[string]float64{"leases": float64(len(res.Leases))}}, nil
	}
	dhcp.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("networking.DHCP(%d, %d, %d, %#v)", in.opt("clients", 4), in.opt("pool", 3), in.opt("lease", 3600), releases(in))
	}
	out = append(out, dhcp)

	nat := netEntry("nat-pat", "NAT with Port Translation", "lan.go", "NAT", "A home router shares one public address among many devices by giving each connection its own public port, and lets in only replies.", "O(packets)", "O(mappings)", Text)
	nat.Example = &Input{Text: "out 192.168.1.10:5000 93.184.215.14:443\nout 192.168.1.11:5000 93.184.215.14:443\nin 93.184.215.14:443 40000\nin 93.184.215.14:443 40001\nin 198.51.100.9:443 40000\nout 192.168.1.10:5000 93.184.215.14:443"}
	nat.Limits = "one packet a line: \"out IP:PORT IP:PORT\" or \"in IP:PORT PUBLICPORT\""
	nat.Run = func(in Input, _ trace.Recorder) (Output, error) {
		steps, table, err := networking.NAT(in.Text)
		if err != nil {
			return Output{}, wrap(err)
		}
		o := Output{Lines: steps}
		for _, m := range table {
			o.Lines = append(o.Lines, fmt.Sprintf("table: %s <-> 203.0.113.1:%d (to %s)", m.Private, m.Public, m.Remote))
		}
		o.Summary = map[string]float64{"mappings": float64(len(table))}
		return o, nil
	}
	nat.Call = func(in Input) (string, string) { return "", fmt.Sprintf("networking.NAT(%s)", "`"+in.Text+"`") }
	out = append(out, nat)

	wf := netEntry("http-waterfall", "HTTP Request Waterfall", "http.go", "Waterfall", "Where a page load's time goes: DNS, the TCP handshake, TLS, waiting for the first byte, and the download; keep-alive skips the setup.", "O(requests)", "O(requests)", Simulation)
	wf.Options = []Option{{"requests", "requests", 3, 1, 50}, {"rtt", "round trip (ms)", 50, 1, 10_000}, {"server", "server time (ms)", 20, 0, 10_000}, {"transfer", "download (ms)", 30, 0, 10_000}, {"tls", "TLS version: 13, 12 or 0", 13, 0, 13}, {"keepalive", "keep-alive (1 yes)", 1, 0, 1}}
	wf.Example = &Input{Options: map[string]int{"requests": 3, "rtt": 50, "server": 20, "transfer": 30, "tls": 13, "keepalive": 1}}
	wf.Run = func(in Input, _ trace.Recorder) (Output, error) {
		ps, err := networking.Waterfall(in.opt("requests", 3), in.opt("rtt", 50), in.opt("server", 20), in.opt("transfer", 30), in.opt("tls", 13), in.opt("keepalive", 1) == 1)
		if err != nil {
			return Output{}, wrap(err)
		}
		o := Output{Timeline: phaseSpans(ps)}
		for _, p := range ps {
			o.Lines = append(o.Lines, fmt.Sprintf("%s: %s %d-%d ms", p.Lane, p.Name, p.Start, p.End))
		}
		o.Summary = map[string]float64{"total ms": float64(ps[len(ps)-1].End)}
		return o, nil
	}
	wf.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("networking.Waterfall(%d, %d, %d, %d, %d, %v)", in.opt("requests", 3), in.opt("rtt", 50), in.opt("server", 20), in.opt("transfer", 30), in.opt("tls", 13), in.opt("keepalive", 1) == 1)
	}
	out = append(out, wf)

	mux := netEntry("http-multiplexing", "HTTP/1.1 vs HTTP/2 vs HTTP/3", "http.go", "Multiplexing", "One request at a time, six connections, one multiplexed connection, or QUIC streams: and how one lost packet blocks every HTTP/2 stream but one QUIC stream.", "O(resources * size)", "O(resources * size)", Simulation)
	mux.Options = []Option{{"resources", "resources", 6, 1, 12}, {"size", "frames each", 2, 1, 50}, {"rtt", "round trip (ticks)", 6, 1, 50}, {"loss", "tick a frame is lost (0 none)", 7, 0, 10_000}}
	mux.Example = &Input{Options: map[string]int{"resources": 6, "size": 2, "rtt": 6, "loss": 7}}
	mux.Run = func(in Input, _ trace.Recorder) (Output, error) {
		ps, finish, err := networking.Multiplexing(in.opt("resources", 6), in.opt("size", 2), in.opt("rtt", 6), in.opt("loss", 7))
		if err != nil {
			return Output{}, wrap(err)
		}
		sum := map[string]float64{}
		for k, v := range finish {
			sum[k+" done at"] = float64(v)
		}
		return Output{Timeline: phaseSpans(ps), Summary: sum}, nil
	}
	mux.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("networking.Multiplexing(%d, %d, %d, %d)", in.opt("resources", 6), in.opt("size", 2), in.opt("rtt", 6), in.opt("loss", 7))
	}
	out = append(out, mux)

	tls := netEntry("tls-1-3-handshake", "TLS 1.3 Handshake", "tls.go", "TLS13", "The messages of a TLS 1.3 handshake, each with its job and the cryptography inside it: one round trip, or zero when resuming.", "O(1)", "O(1)", Simulation)
	tls.Options = []Option{{"resume", "resume with a ticket and 0-RTT data (1 yes)", 0, 0, 1}}
	tls.Example = &Input{Options: map[string]int{"resume": 0}}
	tls.Run = func(in Input, _ trace.Recorder) (Output, error) {
		var o Output
		for _, m := range networking.TLS13(in.opt("resume", 0) == 1) {
			line := fmt.Sprintf("flight %d, %s: %s: %s", m.Flight, m.From, m.Name, m.Purpose)
			if len(m.Uses) > 0 {
				line += " (see " + strings.Join(m.Uses, ", ") + ")"
			}
			o.Lines = append(o.Lines, line)
			state := "running"
			if m.Encrypted {
				state = "done"
			}
			o.Timeline = append(o.Timeline, Span{Lane: m.From, Label: m.Name, Start: m.Flight - 1, End: m.Flight, State: state})
		}
		return o, nil
	}
	tls.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("networking.TLS13(%v)", in.opt("resume", 0) == 1)
	}
	out = append(out, tls)

	quic := netEntry("quic-round-trips", "QUIC vs TCP + TLS", "http.go", "RoundTripsToData", "Round trips before the first byte of a response: TCP + TLS 1.2 takes four, TCP + TLS 1.3 three, QUIC two, and QUIC 0-RTT one.", "O(1)", "O(1)", Simulation)
	quic.Options = []Option{{"rtt", "round trip (ms)", 40, 1, 10_000}}
	quic.Example = &Input{Options: map[string]int{"rtt": 40}}
	quic.Run = func(in Input, _ trace.Recorder) (Output, error) {
		ps, err := networking.RoundTripsToData(in.opt("rtt", 40))
		if err != nil {
			return Output{}, wrap(err)
		}
		sum := map[string]float64{}
		for _, p := range ps {
			sum[p.Lane+" ms"] = float64(p.End)
		}
		return Output{Timeline: phaseSpans(ps), Summary: sum}, nil
	}
	quic.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("networking.RoundTripsToData(%d)", in.opt("rtt", 40))
	}
	out = append(out, quic)

	ws := netEntry("websocket-upgrade", "WebSocket Upgrade", "http.go", "WebSocketAccept", "An HTTP request asks to upgrade, the server proves it understood with a hash of the key, and the connection carries masked frames from then on.", "O(n)", "O(n)", Text)
	ws.Example = &Input{Text: "Hello"}
	ws.Run = func(in Input, _ trace.Recorder) (Output, error) {
		const key = "dGhlIHNhbXBsZSBub25jZQ=="
		frame, err := networking.WebSocketFrame(in.Text, [4]byte{0x37, 0xfa, 0x21, 0x3d})
		if err != nil {
			return Output{}, wrap(err)
		}
		accept := networking.WebSocketAccept(key)
		return Output{Text: fmt.Sprintf("% x", frame), Lines: []string{
			"GET /chat HTTP/1.1", "Host: example.com", "Upgrade: websocket", "Connection: Upgrade", "Sec-WebSocket-Key: " + key, "Sec-WebSocket-Version: 13", "",
			"HTTP/1.1 101 Switching Protocols", "Upgrade: websocket", "Connection: Upgrade", "Sec-WebSocket-Accept: " + accept, "",
			fmt.Sprintf("client frame for %q, masked with 37 fa 21 3d: % x", in.Text, frame),
		}}, nil
	}
	ws.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("networking.WebSocketFrame(%q, [4]byte{0x37, 0xfa, 0x21, 0x3d})", in.Text)
	}
	out = append(out, ws)

	macOut := func(res networking.MACResult, err error) (Output, error) {
		if err != nil {
			return Output{}, wrap(err)
		}
		use := 0.0
		if res.Slots > 0 {
			use = float64(res.Delivered) / float64(res.Slots)
		}
		return Output{Timeline: netSpans(res.Timeline), Summary: map[string]float64{"slots": float64(res.Slots), "delivered": float64(res.Delivered), "collisions": float64(res.Collisions), "dropped": float64(res.Dropped), "frames per slot": use}}, nil
	}
	cd := netEntry("csma-cd", "CSMA/CD (Ethernet)", "mac.go", "CSMACD", "Stations on a shared cable send when it is free; on a collision each waits a random number of slots, doubling the range after each collision.", "O(slots * stations)", "O(stations + slots)", Simulation)
	cd.Options = []Option{{"stations", "stations", 4, 1, 20}, {"frames", "frames each", 3, 1, 100}, seedOption}
	cd.Example = &Input{Options: map[string]int{"stations": 4, "frames": 3, "seed": 1}}
	cd.Run = func(in Input, _ trace.Recorder) (Output, error) {
		return macOut(networking.CSMACD(in.opt("stations", 4), in.opt("frames", 3), int64(in.opt("seed", 1))))
	}
	cd.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("networking.CSMACD(%d, %d, %d)", in.opt("stations", 4), in.opt("frames", 3), in.opt("seed", 1))
	}
	out = append(out, cd)

	ca := netEntry("csma-ca", "CSMA/CA and Hidden Terminals (Wi-Fi)", "mac.go", "CSMACA", "Two laptops that cannot hear each other both think the air is free and collide at the access point; RTS/CTS lets the access point reserve the air.", "O(slots)", "O(slots)", Simulation)
	ca.Options = []Option{{"frames", "frames each", 4, 1, 50}, {"length", "frame length (slots)", 6, 1, 20}, {"rtscts", "use RTS/CTS (1 yes)", 0, 0, 1}, seedOption}
	ca.Example = &Input{Options: map[string]int{"frames": 4, "length": 6, "rtscts": 0, "seed": 3}}
	ca.Run = func(in Input, _ trace.Recorder) (Output, error) {
		return macOut(networking.CSMACA(in.opt("frames", 4), in.opt("length", 6), in.opt("rtscts", 0) == 1, int64(in.opt("seed", 1))))
	}
	ca.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("networking.CSMACA(%d, %d, %v, %d)", in.opt("frames", 4), in.opt("length", 6), in.opt("rtscts", 0) == 1, in.opt("seed", 1))
	}
	out = append(out, ca)

	stp := netEntry("spanning-tree-protocol", "Spanning Tree Protocol", "stp.go", "SpanningTree", "Switches elect the lowest ID as root, keep the cheapest port toward it, and block the rest, so a looped network forwards without loops.", "O(V^2 + E)", "O(V + E)", GraphInput)
	stp.Example = &Input{Text: "S1 S2 4\nS1 S3 4\nS2 S3 4\nS2 S4 4\nS3 S4 12"}
	stp.Limits = "one link a line, \"A B [cost]\"; the lowest name is the lowest bridge ID"
	stp.Run = func(in Input, _ trace.Recorder) (Output, error) {
		t, err := networking.ParseTopology(in.Text)
		if err != nil {
			return Output{}, wrap(err)
		}
		root, cost, ports := networking.SpanningTree(t)
		o := Output{Text: root, Lines: []string{"root bridge: " + root}}
		blocked := 0
		for _, p := range ports {
			if p.Role == "blocked" {
				blocked++
			}
			o.Lines = append(o.Lines, fmt.Sprintf("%s port to %s: %s (root path cost %d)", p.Bridge, p.Peer, p.Role, cost[p.Bridge]))
		}
		o.Summary = map[string]float64{"blocked ports": float64(blocked)}
		return o, nil
	}
	stp.Call = func(in Input) (string, string) {
		return fmt.Sprintf("t, err := networking.ParseTopology(%s)\n\t%s", "`"+in.Text+"`", guard), "networking.SpanningTree(t)"
	}
	out = append(out, stp)

	bgp := netEntry("bgp-path-vector", "BGP Path Vector", "stp.go", "BGP", "Networks announce whole paths of autonomous systems, prefer the shortest, and refuse any path that already contains themselves: no loops.", "O(rounds * E * V)", "O(V^2)", GraphInput)
	bgp.Params = []string{"key"}
	bgp.Example = &Input{Text: "AS1 AS2\nAS2 AS3\nAS3 AS4\nAS4 AS1\nAS3 AS5", Key: "AS1"}
	bgp.Limits = "one link a line, \"AS1 AS2\"; Key is the origin"
	bgp.Run = func(in Input, _ trace.Recorder) (Output, error) {
		t, err := networking.ParseTopology(in.Text)
		if err != nil {
			return Output{}, wrap(err)
		}
		origin := strings.TrimSpace(in.Key)
		if origin == "" {
			origin = t.Routers[0]
		}
		rounds, err := networking.BGP(t, origin)
		if err != nil {
			return Output{}, wrap(err)
		}
		var o Output
		for i, r := range rounds {
			o.Lines = append(o.Lines, fmt.Sprintf("round %d: %s", i, strings.Join(r.Notes, "; ")))
		}
		for _, as := range t.Routers {
			o.Lines = append(o.Lines, fmt.Sprintf("%s: %s", as, strings.Join(rounds[len(rounds)-1].Paths[as], " ")))
		}
		o.Summary = map[string]float64{"rounds": float64(len(rounds) - 1)}
		return o, nil
	}
	bgp.Call = func(in Input) (string, string) {
		return fmt.Sprintf("t, err := networking.ParseTopology(%s)\n\t%s", "`"+in.Text+"`", guard), fmt.Sprintf("networking.BGP(t, %q)", in.Key)
	}
	out = append(out, bgp)

	sw := netEntry("packet-vs-circuit-switching", "Packet vs Circuit Switching", "delays.go", "Switching", "Reserve a circuit first, store and forward the whole message at each hop, or pipeline small packets: the delay of each, by formula.", "O(1)", "O(1)", Simulation)
	sw.Options = []Option{{"message", "message (KB)", 1000, 1, 1 << 20}, {"packet", "packet (KB)", 10, 1, 1 << 20}, {"hops", "links", 3, 1, 100}, {"bandwidth", "bandwidth (Mbit/s)", 8, 1, 100_000}, {"prop", "propagation per link (ms)", 5, 0, 10_000}, {"setup", "circuit setup (ms)", 200, 0, 100_000}}
	sw.Example = &Input{Options: map[string]int{"message": 1000, "packet": 10, "hops": 3, "bandwidth": 8, "prop": 5, "setup": 200}}
	sw.Run = func(in Input, _ trace.Recorder) (Output, error) {
		c, m, p, n, err := networking.Switching(in.opt("message", 1000), in.opt("packet", 10), in.opt("hops", 3), float64(in.opt("bandwidth", 8)), float64(in.opt("prop", 5)), float64(in.opt("setup", 200)))
		if err != nil {
			return Output{}, wrap(err)
		}
		o := Output{Summary: map[string]float64{"circuit ms": c, "message switching ms": m, "packet switching ms": p, "packets": float64(n)}, Lines: []string{
			fmt.Sprintf("circuit: setup + L/R + hops*prop = %.1f ms", c),
			fmt.Sprintf("message switching: hops*(L/R + prop) = %.1f ms", m),
			fmt.Sprintf("packet switching: (hops + packets - 1)*P/R + hops*prop = %.1f ms", p),
		}}
		hops := in.opt("hops", 3)
		for k := 0; k < min(n, 12); k++ { // the pipeline, in packet times
			for h := 0; h < hops; h++ {
				o.Timeline = append(o.Timeline, Span{Lane: "link " + strconv.Itoa(h+1), Label: "packet " + strconv.Itoa(k+1), Start: k + h, End: k + h + 1, State: "running"})
			}
		}
		return o, nil
	}
	sw.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("networking.Switching(%d, %d, %d, %d, %d, %d)", in.opt("message", 1000), in.opt("packet", 10), in.opt("hops", 3), in.opt("bandwidth", 8), in.opt("prop", 5), in.opt("setup", 200))
	}
	out = append(out, sw)

	bdp := netEntry("bandwidth-delay-product", "Bandwidth-Delay Product", "delays.go", "BandwidthDelay", "How many bytes must be in flight to fill a link, and why a small window caps throughput at window / round trip on a long fat network.", "O(1)", "O(1)", Simulation)
	bdp.Options = []Option{{"bandwidth", "bandwidth (Mbit/s)", 100, 1, 1_000_000}, {"rtt", "round trip (ms)", 100, 1, 10_000}, {"window", "window (KB)", 64, 1, 1 << 20}}
	bdp.Example = &Input{Options: map[string]int{"bandwidth": 100, "rtt": 100, "window": 64}}
	bdp.Run = func(in Input, _ trace.Recorder) (Output, error) {
		b, err := networking.BandwidthDelay(float64(in.opt("bandwidth", 100)), float64(in.opt("rtt", 100)), in.opt("window", 64))
		if err != nil {
			return Output{}, wrap(err)
		}
		return Output{Summary: map[string]float64{"bytes in flight to fill the link": b.Bytes, "throughput Mbit/s": b.Throughput, "link use": b.Use}, Lines: []string{
			fmt.Sprintf("BDP = bandwidth x RTT = %.0f bytes (%.0f KB)", b.Bytes, b.Bytes/1024),
			fmt.Sprintf("a %d KB window sends one window per round trip: %.2f Mbit/s, %.0f%% of the link", in.opt("window", 64), b.Throughput, 100*b.Use),
		}}, nil
	}
	bdp.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("networking.BandwidthDelay(%d, %d, %d)", in.opt("bandwidth", 100), in.opt("rtt", 100), in.opt("window", 64))
	}
	out = append(out, bdp)

	tr := netEntry("traceroute", "Traceroute", "delays.go", "Traceroute", "Send probes with a time to live of 1, 2, 3...: each router that drops one at zero answers with its name, mapping the path.", "O(hops * probes)", "O(hops)", Text)
	tr.Options = []Option{{"probes", "probes per hop", 3, 1, 5}, seedOption}
	tr.Example = &Input{Text: "home-router 1\nisp-gateway 4\nisp-core *\nix-london 9\nexample-edge 3\nexample.com 1", Options: map[string]int{"probes": 3, "seed": 1}}
	tr.Limits = "one hop a line, \"name delay-ms\" or \"name *\"; the last is the destination"
	tr.Run = func(in Input, _ trace.Recorder) (Output, error) {
		hops, err := networking.ParseHops(in.Text)
		if err != nil {
			return Output{}, wrap(err)
		}
		lines, err := networking.Traceroute(hops, in.opt("probes", 3), int64(in.opt("seed", 1)))
		if err != nil {
			return Output{}, wrap(err)
		}
		return Output{Lines: lines, Summary: map[string]float64{"hops": float64(len(hops))}}, nil
	}
	tr.Call = func(in Input) (string, string) {
		return fmt.Sprintf("hops, err := networking.ParseHops(%s)\n\t%s", "`"+in.Text+"`", guard), fmt.Sprintf("networking.Traceroute(hops, %d, %d)", in.opt("probes", 3), in.opt("seed", 1))
	}
	out = append(out, tr)

	frag := netEntry("ip-fragmentation", "MTU and IP Fragmentation", "delays.go", "Fragments", "A packet bigger than a link's MTU is cut into fragments at 8-byte offsets, perhaps cut again later, and put back together only at the destination.", "O(fragments * links)", "O(fragments)", Ints)
	frag.Options = []Option{{"payload", "payload (bytes)", 3980, 1, 65515}, {"df", "don't fragment (1 yes)", 0, 0, 1}}
	frag.Example = &Input{Values: []int{1500, 576}, Options: map[string]int{"payload": 3980, "df": 0}}
	frag.Limits = "the MTU of each link on the path, 68 to 65535"
	frag.Run = func(in Input, _ trace.Recorder) (Output, error) {
		fs, steps, err := networking.Fragments(in.opt("payload", 3980), in.Values, in.opt("df", 0) == 1)
		if err != nil {
			return Output{}, wrap(err)
		}
		o := Output{Lines: steps}
		if fs != nil {
			o.Table = &TableOutput{Rows: len(fs), Cols: 4, ColLabels: []string{"offset bytes", "offset field", "length", "more"}}
			for i, f := range fs {
				o.Table.RowLabels = append(o.Table.RowLabels, "fragment "+strconv.Itoa(i+1))
				more := 0
				if f.More {
					more = 1
				}
				o.Table.Cells = append(o.Table.Cells, f.Offset, f.Offset/8, f.Length, more)
			}
		}
		o.Summary = map[string]float64{"fragments": float64(len(fs))}
		return o, nil
	}
	frag.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("networking.Fragments(%d, %s, %v)", in.opt("payload", 3980), ints(in.Values), in.opt("df", 0) == 1)
	}
	out = append(out, frag)

	rto := netEntry("tcp-retransmission-timer", "TCP Retransmission Timer", "delays.go", "RTO", "Smooth the measured round trips and their variation into a timeout (Jacobson/Karels), skip ambiguous samples (Karn), and double on each timeout.", "O(samples)", "O(samples)", Ints)
	rto.Options = []Option{{"min", "minimum RTO (ms)", 0, 0, 60_000}}
	rto.Example = &Input{Values: []int{100, 120, 90, 300, -1, 110, 105}, Options: map[string]int{"min": 0}}
	rto.Limits = "round-trip samples in ms; -1 marks a timeout of a retransmitted segment"
	rto.Run = func(in Input, _ trace.Recorder) (Output, error) {
		rs, err := networking.RTO(in.Values, in.opt("min", 0))
		if err != nil {
			return Output{}, wrap(err)
		}
		o := Output{Table: &TableOutput{Rows: len(rs), Cols: 4, ColLabels: []string{"sample", "SRTT", "RTTVAR", "RTO"}}}
		for i, r := range rs {
			o.Table.RowLabels = append(o.Table.RowLabels, strconv.Itoa(i+1))
			o.Table.Cells = append(o.Table.Cells, r.Sample, int(r.SRTT+0.5), int(r.RTTVar+0.5), int(r.RTO+0.5))
			o.Values = append(o.Values, int(r.RTO+0.5))
		}
		return o, nil
	}
	rto.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("networking.RTO(%s, %d)", ints(in.Values), in.opt("min", 0))
	}
	out = append(out, rto)

	return out
}

func tcpState(s string) string {
	switch s {
	case "ESTABLISHED":
		return "running"
	case "LISTEN", "TIME_WAIT", "CLOSE_WAIT", "FIN_WAIT_2":
		return "idle"
	default:
		return "blocked"
	}
}

// tcpMachine draws a connection's trace on one TCP state diagram: each step
// is a transition of one side, with both sides' current states active.
func tcpMachine(steps []networking.TCPStep) *MachineOutput {
	m := &MachineOutput{}
	index := map[string]int{}
	state := func(name string) int {
		if i, ok := index[name]; ok {
			return i
		}
		index[name] = len(m.States)
		m.States = append(m.States, MachineState{Name: name, Start: name == "CLOSED"})
		return index[name]
	}
	edge := map[[2]int]int{}
	cur := map[string]int{"client": state("CLOSED"), "server": state("CLOSED")}
	for _, s := range steps {
		if s.From == s.To {
			continue
		}
		from, to := state(s.From), state(s.To)
		label := "-"
		if s.Segment != "" {
			label = "send " + strings.Fields(s.Segment)[0]
		}
		k, ok := edge[[2]int{from, to}]
		if !ok {
			k = len(m.Edges)
			edge[[2]int{from, to}] = k
			m.Edges = append(m.Edges, MachineEdge{From: from, To: to, Label: label})
		}
		cur[s.Side] = to
		note := fmt.Sprintf("t%d %s: %s -> %s", s.Tick, s.Side, s.From, s.To)
		if s.Segment != "" {
			note += ", sends " + s.Segment
		}
		m.Steps = append(m.Steps, MachineStep{Active: []int{cur["client"], cur["server"]}, Edges: []int{k}, Note: note})
	}
	return m
}
