package networking

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"
)

func TestStopAndWaitWithoutLoss(t *testing.T) {
	res, err := StopAndWait(3, Channel{Delay: 1, Timeout: 4})
	if err != nil {
		t.Fatal(err)
	}
	if res.Sent != 3 || res.Retransmitted != 0 || !slices.Equal(res.Delivered, []int{0, 1, 2}) {
		t.Fatalf("sent %d, resent %d, delivered %v", res.Sent, res.Retransmitted, res.Delivered)
	}
	// one frame per round trip: send at 0, ack back at 2, and so on
	if res.Ticks != 6 {
		t.Errorf("finished at tick %d, want 6", res.Ticks)
	}
}

func TestGoBackNResendsTheWindowAfterALoss(t *testing.T) {
	// frames 0..5, window 4; the third transmission (frame 2) is lost.
	res, err := GoBackN(6, 4, Channel{Delay: 1, Timeout: 5, Drop: []int{3}})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(res.Delivered, []int{0, 1, 2, 3, 4, 5}) {
		t.Fatalf("delivered %v", res.Delivered)
	}
	// frame 3 (and whatever followed it) arrived out of order and was
	// discarded, so going back resends 2 and everything after it
	if res.Retransmitted < 2 {
		t.Errorf("resent %d frames, want the window after the loss", res.Retransmitted)
	}
	discards := 0
	for _, e := range res.Timeline {
		if e.Lane == "receiver" && strings.Contains(e.What, "discard") {
			discards++
		}
	}
	if discards == 0 {
		t.Error("the receiver discarded nothing after the gap")
	}
}

func TestSelectiveRepeatResendsOnlyTheLostFrame(t *testing.T) {
	res, err := SelectiveRepeat(6, 4, Channel{Delay: 1, Timeout: 5, Drop: []int{3}})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(res.Delivered, []int{0, 1, 2, 3, 4, 5}) {
		t.Fatalf("delivered %v", res.Delivered)
	}
	if res.Retransmitted != 1 {
		t.Errorf("resent %d frames, want only the lost one", res.Retransmitted)
	}
	gbn, _ := GoBackN(6, 4, Channel{Delay: 1, Timeout: 5, Drop: []int{3}})
	if gbn.Sent <= res.Sent {
		t.Errorf("Go-Back-N sent %d, Selective Repeat %d: SR should send fewer", gbn.Sent, res.Sent)
	}
}

func TestARQIsDeterministicAndDeliversInOrder(t *testing.T) {
	for _, f := range []func(int, int, Channel) (ARQResult, error){GoBackN, SelectiveRepeat} {
		for seed := int64(0); seed < 30; seed++ {
			ch := Channel{Loss: 30, Delay: 2, Timeout: 7, Seed: seed}
			a, err := f(20, 5, ch)
			if err != nil {
				t.Fatal(err)
			}
			b, _ := f(20, 5, ch)
			if a.Sent != b.Sent || a.Ticks != b.Ticks {
				t.Fatalf("seed %d: two runs differ", seed)
			}
			want := make([]int, 20)
			for i := range want {
				want[i] = i
			}
			if !slices.Equal(a.Delivered, want) {
				t.Fatalf("seed %d: delivered %v", seed, a.Delivered)
			}
		}
	}
	if _, err := GoBackN(0, 1, Channel{}); !errors.Is(err, ErrInput) {
		t.Errorf("no frames: %v", err)
	}
}

func TestCongestion(t *testing.T) {
	rs, err := Congestion(12, 8, map[int]LossKind{6: TripleDupAck, 10: Timeout}, true)
	if err != nil {
		t.Fatal(err)
	}
	var cwnd []int
	for _, r := range rs {
		cwnd = append(cwnd, r.Cwnd)
	}
	// 1 2 4 8 (slow start to ssthresh), 9 10 (additive), loss at 10 -> ssthresh 5,
	// Reno continues at 5: 5 6 7 8, timeout at 8 -> ssthresh 4, cwnd 1, 2
	want := []int{1, 2, 4, 8, 9, 10, 5, 6, 7, 8, 1, 2}
	if !slices.Equal(cwnd, want) {
		t.Fatalf("cwnd %v, want %v", cwnd, want)
	}
	tahoe, _ := Congestion(8, 8, map[int]LossKind{6: TripleDupAck}, false)
	if tahoe[6].Cwnd != 1 {
		t.Errorf("Tahoe after three duplicate acks: cwnd %d, want 1", tahoe[6].Cwnd)
	}
}

func TestRateLimiters(t *testing.T) {
	burst := []int{0, 0, 0, 0, 0, 10}
	tb, err := TokenBucket(burst, 3, 5)
	if err != nil {
		t.Fatal(err)
	}
	var allowed []bool
	for _, d := range tb {
		allowed = append(allowed, d.Allowed)
	}
	// three tokens pass the burst, two are rejected; by tick 10 two tokens are back
	if !slices.Equal(allowed, []bool{true, true, true, false, false, true}) {
		t.Errorf("token bucket allowed %v", allowed)
	}
	lb, err := LeakyBucket(burst, 3, 5)
	if err != nil {
		t.Fatal(err)
	}
	var served []int
	for _, d := range lb {
		served = append(served, d.Served)
	}
	// the first leaves at once, three wait and leave 5 ticks apart, one finds the queue full
	if !slices.Equal(served, []int{0, 5, 10, 15, -1, 20}) {
		t.Errorf("leaky bucket served at %v", served)
	}
}

const line = "A B 1\nB C 1"

func TestDistanceVectorCountsToInfinity(t *testing.T) {
	topo, _ := ParseTopology(line)
	rounds, cutAt, err := DistanceVector(topo, "B C", false, 40)
	if err != nil {
		t.Fatal(err)
	}
	if got := rounds[cutAt-1].Dist[0][2]; got != 2 {
		t.Fatalf("before the cut A reaches C at %d, want 2", got)
	}
	last := rounds[len(rounds)-1]
	if last.Dist[0][2] != Infinity || last.Dist[1][2] != Infinity {
		t.Errorf("after counting, A and B reach C at %d and %d, want %d", last.Dist[0][2], last.Dist[1][2], Infinity)
	}
	slow := len(rounds) - cutAt
	topo, _ = ParseTopology(line)
	poisoned, cutAt, _ := DistanceVector(topo, "B C", true, 40)
	fast := len(poisoned) - cutAt
	if fast >= slow || slow < 8 {
		t.Errorf("rounds after the cut: %d without poisoned reverse, %d with", slow, fast)
	}
}

func TestLinkState(t *testing.T) {
	topo, err := ParseTopology("A B 1\nB C 2\nA C 5\nC D 1")
	if err != nil {
		t.Fatal(err)
	}
	tables := LinkState(topo)
	got := tables[0][3] // A to D
	if got.Next != "B" || got.Cost != 4 {
		t.Errorf("A to D: %+v, want via B at 4", got)
	}
	if _, err := ParseTopology("A A"); err == nil {
		t.Error("a loop link was accepted")
	}
}

func TestPrefixes(t *testing.T) {
	p, err := ParsePrefix("192.168.10.77/26")
	if err != nil {
		t.Fatal(err)
	}
	if p.Network().String() != "192.168.10.64" || p.Broadcast().String() != "192.168.10.127" || p.Hosts() != 62 {
		t.Errorf("%s: network %s broadcast %s hosts %d", p, p.Network(), p.Broadcast(), p.Hosts())
	}
	subs, err := Split(p, 4)
	if err != nil {
		t.Fatal(err)
	}
	if len(subs) != 4 || subs[3].String() != "192.168.10.112/28" {
		t.Errorf("subnets %v", subs)
	}
	if _, err := Split(p, 3); err == nil {
		t.Error("split into 3 was accepted")
	}
	if _, err := ParseAddr("1.2.3.256"); err == nil {
		t.Error("256 was accepted")
	}
}

func TestLongestPrefixMatch(t *testing.T) {
	rt, err := ParseRoutes("0.0.0.0/0 isp\n10.0.0.0/8 core\n10.1.0.0/16 east\n10.1.2.0/24 lab")
	if err != nil {
		t.Fatal(err)
	}
	for addr, hop := range map[string]string{"10.1.2.3": "lab", "10.1.9.9": "east", "10.200.0.1": "core", "8.8.8.8": "isp"} {
		a, _ := ParseAddr(addr)
		r, _, ok := rt.Lookup(a)
		if !ok || r.Hop != hop {
			t.Errorf("%s goes to %q, want %q", addr, r.Hop, hop)
		}
	}
	noDefault, _ := ParseRoutes("10.0.0.0/8 core")
	a, _ := ParseAddr("8.8.8.8")
	if _, _, ok := noDefault.Lookup(a); ok {
		t.Error("matched without a default route")
	}
}

func TestChecksum(t *testing.T) {
	// RFC 1071's example: the words 0001 f203 f4f5 f6f7 sum to ddf2.
	data := []byte{0x00, 0x01, 0xf2, 0x03, 0xf4, 0xf5, 0xf6, 0xf7}
	if got := Checksum(data); got != ^uint16(0xddf2) {
		t.Fatalf("checksum %#04x, want %#04x", got, ^uint16(0xddf2))
	}
	c := Checksum(data)
	if Checksum(append(data, byte(c>>8), byte(c))) != 0 {
		t.Error("data with its checksum does not check to zero")
	}
	_, sums := ChecksumSteps(data)
	if sums[len(sums)-1] != 0xddf2 {
		t.Errorf("running sum ends at %#04x", sums[len(sums)-1])
	}
}

func TestNagleCoalescesSmallWrites(t *testing.T) {
	writes := []int{1, 1, 1, 1, 1, 1, 1, 1}
	off, err := Nagle(writes, 100, 4, false)
	if err != nil {
		t.Fatal(err)
	}
	on, _ := Nagle(writes, 100, 4, true)
	if len(off.Segments) != 8 {
		t.Errorf("without Nagle %d segments, want one per write", len(off.Segments))
	}
	if len(on.Segments) >= len(off.Segments) || on.Delay == 0 {
		t.Errorf("with Nagle %d segments and delay %d", len(on.Segments), on.Delay)
	}
	total := 0
	for _, s := range on.Segments {
		total += s.Bytes
	}
	if total != 8 {
		t.Errorf("sent %d bytes, want 8", total)
	}
}

func TestDNS(t *testing.T) {
	it := NewResolver(false)
	lk, err := it.Resolve("www.example.com")
	if err != nil {
		t.Fatal(err)
	}
	// client->resolver, 3 asks and 3 replies, resolver->client
	if lk.Answer != "93.184.215.14" || lk.Messages != 8 {
		t.Errorf("iterative: %s in %d messages\n%s", lk.Answer, lk.Messages, strings.Join(lk.Steps, "\n"))
	}
	again, _ := it.Resolve("www.example.com")
	if again.Messages != 2 {
		t.Errorf("cached lookup took %d messages", again.Messages)
	}
	sibling, _ := it.Resolve("mail.example.com")
	if sibling.Messages != 4 {
		t.Errorf("a sibling name took %d messages, want 4 (the domain's server is cached)", sibling.Messages)
	}
	rec := NewResolver(true)
	r, _ := rec.Resolve("pkg.go.dev")
	if r.Answer != "192.0.2.22" || r.Messages != 8 {
		t.Errorf("recursive: %s in %d messages\n%s", r.Answer, r.Messages, strings.Join(r.Steps, "\n"))
	}
	for _, res := range []*Resolver{NewResolver(false), NewResolver(true)} {
		if nx, _ := res.Resolve("nope.example.com"); nx.Answer != "NXDOMAIN" {
			t.Errorf("missing name: %s", nx.Answer)
		}
		if nx, _ := res.Resolve("www.nowhere.com"); nx.Answer != "NXDOMAIN" {
			t.Errorf("missing domain: %s", nx.Answer)
		}
	}
	if _, err := it.Resolve("localhost"); err == nil {
		t.Error("a single label was accepted")
	}
}

func TestEncapsulate(t *testing.T) {
	tcpip, err := Encapsulate("hello", false)
	if err != nil {
		t.Fatal(err)
	}
	osi, _ := Encapsulate("hello", true)
	if len(tcpip) != 4 || len(osi) != 7 {
		t.Fatalf("%d and %d layers", len(tcpip), len(osi))
	}
	app := len("GET / HTTP/1.1\r\nHost: example.com\r\n\r\n")
	if got, want := tcpip[3].Total, 5+app+20+20+22+4; got != want {
		t.Errorf("frame of %d bytes, want %d", got, want)
	}
	if !strings.HasPrefix(tcpip[3].Shape, "[Eth 22 | [IP 20 | [TCP 20 | [HTTP") {
		t.Errorf("shape %s", tcpip[3].Shape)
	}
	if up := Decapsulate(tcpip); len(up) != 4 || !strings.HasPrefix(up[0], "Link") {
		t.Errorf("decapsulation %v", up)
	}
}

func states(steps []TCPStep, side string) []string {
	var out []string
	for _, s := range steps {
		if s.Side == side && s.From != s.To {
			out = append(out, s.To)
		}
	}
	return out
}

func TestTCPConnection(t *testing.T) {
	steps, err := TCPConnection(2, "client")
	if err != nil {
		t.Fatal(err)
	}
	if got := states(steps, "client"); !slices.Equal(got, []string{"SYN_SENT", "ESTABLISHED", "FIN_WAIT_1", "FIN_WAIT_2", "TIME_WAIT", "CLOSED"}) {
		t.Errorf("client states %v", got)
	}
	if got := states(steps, "server"); !slices.Equal(got, []string{"LISTEN", "SYN_RCVD", "ESTABLISHED", "CLOSE_WAIT", "LAST_ACK", "CLOSED"}) {
		t.Errorf("server states %v", got)
	}
	if steps[2].Segment != "SYN+ACK seq=5000 ack=1001" {
		t.Errorf("SYN+ACK is %q", steps[2].Segment)
	}
	server, _ := TCPConnection(0, "server")
	if got := states(server, "server"); got[len(got)-2] != "TIME_WAIT" {
		t.Errorf("the closer waits in TIME_WAIT: %v", got)
	}
}

func TestFlowControl(t *testing.T) {
	ticks, stalled, err := FlowControl(100, 20, 2, 10, 4)
	if err != nil || stalled {
		t.Fatal(err, stalled)
	}
	sent, window, probes := 0, 20, 0
	for _, ft := range ticks {
		if ft.Sent > window {
			t.Fatalf("tick %d: sent %d over a window of %d", ft.Tick, ft.Sent, window)
		}
		sent += ft.Sent
		window = ft.Window
		if ft.Probe {
			probes++
		}
	}
	if sent != 100 || probes == 0 {
		t.Errorf("sent %d with %d probes", sent, probes)
	}
	if _, stalled, _ := FlowControl(100, 20, 2, 10, 0); !stalled {
		t.Error("without probes a closed window should stall")
	}
}

func TestARP(t *testing.T) {
	res, err := ARP([]string{"10.0.0.3", "10.0.0.3", "8.8.8.8", "1.1.1.1", "10.0.0.42"}, 10)
	if err != nil {
		t.Fatal(err)
	}
	if res.Broadcasts != 3 || res.Hits != 2 || res.Failures != 1 {
		t.Errorf("broadcasts %d hits %d failures %d\n%s", res.Broadcasts, res.Hits, res.Failures, strings.Join(res.Steps, "\n"))
	}
	expired, _ := ARP([]string{"10.0.0.3", "10.0.0.4", "10.0.0.3"}, 1)
	if expired.Broadcasts != 3 {
		t.Errorf("an expired entry was used: %d broadcasts", expired.Broadcasts)
	}
}

func TestDHCP(t *testing.T) {
	full, err := DHCP(3, 2, 60, nil)
	if err != nil {
		t.Fatal(err)
	}
	if full.Leases["client 1"] != "192.168.1.100" || !strings.HasPrefix(full.Leases["client 3"], "169.254.") {
		t.Errorf("leases %v", full.Leases)
	}
	reuse, _ := DHCP(3, 2, 60, []int{1})
	if reuse.Leases["client 2"] != "192.168.1.100" || reuse.Leases["client 3"] != "192.168.1.101" {
		t.Errorf("after a release: %v", reuse.Leases)
	}
}

func TestNAT(t *testing.T) {
	steps, table, err := NAT(`out 192.168.1.10:5000 93.184.215.14:443
out 192.168.1.11:5000 93.184.215.14:443
in 93.184.215.14:443 40001
in 198.51.100.9:443 40000
in 93.184.215.14:443 40999`)
	if err != nil {
		t.Fatal(err)
	}
	if len(table) != 2 || table[0].Public == table[1].Public {
		t.Fatalf("two hosts with one port need two public ports: %v", table)
	}
	if !strings.HasSuffix(steps[2], "-> 192.168.1.11:5000") || !strings.Contains(steps[3], "dropped") || !strings.Contains(steps[4], "dropped") {
		t.Errorf("steps:\n%s", strings.Join(steps, "\n"))
	}
}

func TestWaterfallAndMultiplexing(t *testing.T) {
	end := func(ps []Phase) int { return ps[len(ps)-1].End }
	keep, err := Waterfall(3, 50, 20, 30, 13, true)
	if err != nil {
		t.Fatal(err)
	}
	fresh, _ := Waterfall(3, 50, 20, 30, 13, false)
	// first request: DNS 50 + TCP 50 + TLS 50 + 70 + 30 = 250; then 100 each, or 200 each
	if end(keep) != 450 || end(fresh) != 650 {
		t.Errorf("keep-alive %d, new connections %d", end(keep), end(fresh))
	}
	_, clean, err := Multiplexing(6, 4, 6, 0)
	if err != nil {
		t.Fatal(err)
	}
	if clean["HTTP/2"] >= clean["HTTP/1.1"] {
		t.Errorf("without loss HTTP/2 %d should beat HTTP/1.1 %d", clean["HTTP/2"], clean["HTTP/1.1"])
	}
	phases, lossy, _ := Multiplexing(6, 2, 6, 7)
	if lossy["HTTP/3"] > lossy["HTTP/2"] {
		t.Errorf("with a loss QUIC %d should not trail HTTP/2 %d", lossy["HTTP/3"], lossy["HTTP/2"])
	}
	// HTTP/2's head-of-line blocking delays streams other than the lost one
	h2, h3 := 0, 0
	for _, p := range phases {
		if strings.HasPrefix(p.Lane, "HTTP/2 ") {
			h2 += p.End
		}
		if strings.HasPrefix(p.Lane, "HTTP/3 ") {
			h3 += p.End
		}
	}
	if h3 >= h2 {
		t.Errorf("total completion: HTTP/3 %d, HTTP/2 %d", h3, h2)
	}
	rt, _ := RoundTripsToData(30)
	last := map[string]int{}
	for _, p := range rt {
		last[p.Lane] = p.End
	}
	if last["TCP + TLS 1.2"] != 120 || last["QUIC 0-RTT"] != 30 {
		t.Errorf("round trips %v", last)
	}
}

func TestWebSocket(t *testing.T) {
	// RFC 6455 section 1.3 and 5.7
	if got := WebSocketAccept("dGhlIHNhbXBsZSBub25jZQ=="); got != "s3pPLMBiTxaQ9kYGzzhZRbK+xOo=" {
		t.Errorf("accept %s", got)
	}
	f, err := WebSocketFrame("Hello", [4]byte{0x37, 0xfa, 0x21, 0x3d})
	if err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("% x", f); got != "81 85 37 fa 21 3d 7f 9f 4d 51 58" {
		t.Errorf("frame %s", got)
	}
}

func TestTLS13(t *testing.T) {
	full := TLS13(false)
	if full[0].Name != "ClientHello + key_share" || full[len(full)-1].Flight != 3 {
		t.Errorf("full handshake %v", full)
	}
	for _, m := range full[2:] {
		if !m.Encrypted {
			t.Errorf("%s after ServerHello should be encrypted", m.Name)
		}
	}
	resume := TLS13(true)
	if !strings.Contains(resume[1].Name, "early data") || resume[1].Flight != 1 {
		t.Errorf("0-RTT data should go in the first flight: %v", resume[1])
	}
}

func TestCSMA(t *testing.T) {
	one, _ := CSMACD(1, 5, 1)
	if one.Collisions != 0 || one.Delivered != 5 || one.Slots != 5 {
		t.Errorf("alone: %+v", one)
	}
	res, err := CSMACD(4, 5, 1)
	if err != nil {
		t.Fatal(err)
	}
	if res.Delivered+res.Dropped != 20 || res.Collisions == 0 {
		t.Errorf("four stations: %+v", res)
	}
	garbled := func(r MACResult) int {
		n := 0
		for _, e := range r.Timeline {
			if strings.Contains(e.What, "garbled") {
				n++
			}
		}
		return n
	}
	hidden, err := CSMACA(10, 8, false, 3)
	if err != nil {
		t.Fatal(err)
	}
	rts, _ := CSMACA(10, 8, true, 3)
	if hidden.Delivered != 20 || rts.Delivered != 20 {
		t.Fatalf("delivered %d and %d", hidden.Delivered, rts.Delivered)
	}
	if garbled(hidden) == 0 || garbled(rts) != 0 {
		t.Errorf("garbled data frames: %d without RTS/CTS, %d with", garbled(hidden), garbled(rts))
	}
}

func TestSpanningTree(t *testing.T) {
	topo, _ := ParseTopology("A B 1\nA C 1\nB C 1")
	root, cost, ports := SpanningTree(topo)
	if root != "A" || cost["C"] != 1 {
		t.Fatalf("root %s, costs %v", root, cost)
	}
	roles := map[string]string{}
	for _, p := range ports {
		roles[p.Bridge+">"+p.Peer] = p.Role
	}
	want := map[string]string{"A>B": "designated", "A>C": "designated", "B>A": "root", "C>A": "root", "B>C": "designated", "C>B": "blocked"}
	for k, v := range want {
		if roles[k] != v {
			t.Errorf("port %s is %s, want %s", k, roles[k], v)
		}
	}
}

func TestBGP(t *testing.T) {
	topo, _ := ParseTopology("AS1 AS2\nAS2 AS3\nAS3 AS4\nAS4 AS1")
	rounds, err := BGP(topo, "AS1")
	if err != nil {
		t.Fatal(err)
	}
	last := rounds[len(rounds)-1]
	if got := last.Paths["AS3"]; !slices.Equal(got, []string{"AS3", "AS2", "AS1"}) {
		t.Errorf("AS3's path %v", got)
	}
	loops := 0
	for _, r := range rounds {
		for _, n := range r.Notes {
			if strings.Contains(n, "loop") {
				loops++
			}
		}
	}
	if loops == 0 {
		t.Error("no looping path was refused")
	}
}

func TestDelays(t *testing.T) {
	// 1000 KB over 3 hops at 8 Mbit/s: 1000 ms to transmit the message
	c, m, p, n, err := Switching(1000, 10, 3, 8, 5, 200)
	if err != nil {
		t.Fatal(err)
	}
	if n != 100 || c != 1215 || m != 3015 || p != 1035 {
		t.Errorf("circuit %v message %v packet %v (%d packets)", c, m, p, n)
	}
	b, _ := BandwidthDelay(100, 100, 64)
	if b.Bytes != 1.25e6 || b.Throughput > 5.3 || b.Throughput < 5.2 {
		t.Errorf("BDP %+v", b)
	}
	hops, _ := ParseHops("gw 1\nisp *\ncore 5\nexample.com 10")
	lines, _ := Traceroute(hops, 3, 1)
	if len(lines) != 4 || !strings.Contains(lines[1], "* * *") && !strings.Contains(lines[1], "*  *") || !strings.Contains(lines[3], "port unreachable") {
		t.Errorf("traceroute:\n%s", strings.Join(lines, "\n"))
	}
	frags, _, err := Fragments(3980, []int{1500}, false)
	if err != nil {
		t.Fatal(err)
	}
	want := []Fragment{{0, 1480, true}, {1480, 1480, true}, {2960, 1020, false}}
	if !slices.Equal(frags, want) {
		t.Errorf("fragments %v", frags)
	}
	again, _, _ := Fragments(3980, []int{1500, 600}, false)
	if len(again) != 8 || again[len(again)-1].More || again[2].Length != 328 {
		t.Errorf("refragmented into %v", again)
	}
	if f, steps, _ := Fragments(3980, []int{1500}, true); f != nil || !strings.Contains(steps[0], "fragmentation needed") {
		t.Errorf("DF: %v %v", f, steps)
	}
	rto, _ := RTO([]int{100, -1, 120}, 0)
	if rto[0].RTO != 300 || rto[1].RTO != 600 || rto[2].SRTT != 102.5 {
		t.Errorf("RTO %+v", rto)
	}
}
