package networking

import "fmt"

// TCPStep is one transition of one side's TCP state machine.
type TCPStep struct {
	Tick    int    `json:"tick"`
	Side    string `json:"side"` // "client" or "server"
	From    string `json:"from"`
	To      string `json:"to"`
	Segment string `json:"segment"` // what it sent, "" for none
}

// TCPConnection traces a TCP connection's life through the state machine
// of RFC 793: the server waits in LISTEN; the three-way handshake
// (SYN, SYN+ACK, ACK) moves both to ESTABLISHED; data segments of 100 bytes
// are sent and acked; then the side named closer ("client" or "server")
// starts the four-way teardown (FIN, ACK, FIN, ACK), passing through
// FIN_WAIT_1, FIN_WAIT_2 and TIME_WAIT, where it waits two maximum segment
// lifetimes so a lost last ACK can be resent, while the other side goes
// through CLOSE_WAIT and LAST_ACK. The initial sequence numbers are fixed
// (1000 and 5000) instead of random; every segment arrives one tick later.
//
// Time: O(data). Space: O(data).
func TCPConnection(data int, closer string) ([]TCPStep, error) {
	if data < 0 || data > 100 || (closer != "client" && closer != "server") {
		return nil, fmt.Errorf("%w: 0 to 100 data segments, closed by the client or the server", ErrInput)
	}
	state := map[string]string{"client": "CLOSED", "server": "CLOSED"}
	seq := map[string]int{"client": 1000, "server": 5000}
	var out []TCPStep
	t := 0
	move := func(side, to, seg string) {
		out = append(out, TCPStep{Tick: t, Side: side, From: state[side], To: to, Segment: seg})
		state[side] = to
	}
	other := func(s string) string {
		if s == "client" {
			return "server"
		}
		return "client"
	}
	move("server", "LISTEN", "")
	t++
	move("client", "SYN_SENT", fmt.Sprintf("SYN seq=%d", seq["client"]))
	seq["client"]++
	t++
	move("server", "SYN_RCVD", fmt.Sprintf("SYN+ACK seq=%d ack=%d", seq["server"], seq["client"]))
	seq["server"]++
	t++
	move("client", "ESTABLISHED", fmt.Sprintf("ACK seq=%d ack=%d", seq["client"], seq["server"]))
	t++
	move("server", "ESTABLISHED", "")
	for i := 0; i < data; i++ {
		t++
		from := "client"
		if i%2 == 1 {
			from = "server"
		}
		out = append(out, TCPStep{Tick: t, Side: from, From: "ESTABLISHED", To: "ESTABLISHED",
			Segment: fmt.Sprintf("data seq=%d len=100", seq[from])})
		seq[from] += 100
		t++
		out = append(out, TCPStep{Tick: t, Side: other(from), From: "ESTABLISHED", To: "ESTABLISHED",
			Segment: fmt.Sprintf("ACK ack=%d", seq[from])})
	}
	a, b := closer, other(closer)
	t++
	move(a, "FIN_WAIT_1", fmt.Sprintf("FIN seq=%d", seq[a]))
	seq[a]++
	t++
	move(b, "CLOSE_WAIT", fmt.Sprintf("ACK ack=%d", seq[a]))
	t++
	move(a, "FIN_WAIT_2", "")
	move(b, "LAST_ACK", fmt.Sprintf("FIN seq=%d", seq[b]))
	seq[b]++
	t++
	move(a, "TIME_WAIT", fmt.Sprintf("ACK ack=%d", seq[b]))
	t++
	move(b, "CLOSED", "")
	t += 4 // two maximum segment lifetimes, shortened
	move(a, "CLOSED", "")
	return out, nil
}

// FlowTick is one tick of TCP flow control.
type FlowTick struct {
	Tick     int  `json:"tick"`
	Sent     int  `json:"sent"`     // bytes the sender sent this tick
	Buffered int  `json:"buffered"` // bytes in the receiver's buffer after the application read
	Window   int  `json:"window"`   // the window the sender believes in after this tick
	Probe    bool `json:"probe"`    // a zero-window probe went out
}

// FlowControl sends total bytes to a receiver whose buffer holds buffer
// bytes and whose application reads read bytes a tick. The receiver
// advertises the free space as its window in every ack; the sender never
// sends more than the window (one segment of up to mss bytes a tick). When
// the window is zero the sender stops (a receiver advertises zero until
// min(mss, buffer/2) bytes are free, so a slow reader closes the window
// instead of inviting tiny segments), and since the receiver's window
// update can be lost, it sends a zero-window probe every probe ticks to
// learn the window again; here the receiver never sends the update on its
// own, so without probes (probe 0) the connection stalls, which stalled
// reports (the run stops 20 ticks into the stall).
//
// Time: O(ticks). Space: O(ticks).
func FlowControl(total, buffer, read, mss, probe int) (ticks []FlowTick, stalled bool, err error) {
	if total < 1 || total > 1<<20 || buffer < 1 || read < 1 || mss < 1 || probe < 0 {
		return nil, false, fmt.Errorf("%w: positive sizes and a probe interval (0 for none)", ErrInput)
	}
	// Clark's rule against the silly window syndrome: advertise no window
	// until a useful amount, min(mss, buffer/2), is free.
	adv := func(free int) int {
		if free < min(mss, max(1, buffer/2)) {
			return 0
		}
		return free
	}
	out := ticks
	window, buffered, sent, zeroSince := buffer, 0, 0, -1
	for t := 0; t < MaxTicks; t++ {
		ft := FlowTick{Tick: t}
		if sent < total && window > 0 {
			n := min(mss, window, total-sent)
			sent += n
			buffered += n
			ft.Sent = n
			window = adv(buffer - buffered) // the ack carries the new window
		} else if sent < total && probe > 0 && zeroSince >= 0 && (t-zeroSince)%probe == 0 && t > zeroSince {
			ft.Probe = true
			window = adv(buffer - buffered) // the probe's ack reveals the window
		}
		buffered = max(0, buffered-read)
		if window == 0 && zeroSince < 0 {
			zeroSince = t
		} else if window > 0 {
			zeroSince = -1
		}
		ft.Buffered, ft.Window = buffered, window
		out = append(out, ft)
		if sent >= total && buffered == 0 {
			return out, false, nil
		}
		if probe == 0 && window == 0 && sent < total && t >= zeroSince+20 {
			return out, true, nil
		}
	}
	return out, false, fmt.Errorf("%w: not finished in %d ticks", ErrInput, MaxTicks)
}
