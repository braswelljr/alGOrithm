package networking

import (
	"fmt"
	"math/rand"
)

// Channel is the lossy link an ARQ protocol runs over.
type Channel struct {
	// Loss is the percent chance (0 to 100) that a frame or an ack is lost.
	Loss int
	// Delay is the ticks a frame or an ack takes one way (at least 1).
	Delay int
	// Timeout is the ticks the sender waits for an ack before resending
	// (more than a round trip, 2*Delay, or every frame is resent).
	Timeout int
	// Seed picks which frames the Loss percent drops.
	Seed int64
	// Drop lists data transmissions to lose on purpose, counted from 1 in
	// the order they are sent, on top of Loss: a scripted loss for teaching.
	Drop []int
}

// ARQResult is the outcome of an ARQ run.
type ARQResult struct {
	Timeline Timeline `json:"timeline"`
	// Sent counts every data transmission, Retransmitted those that repeat a
	// frame, Lost the frames and acks the link dropped.
	Sent          int `json:"sent"`
	Retransmitted int `json:"retransmitted"`
	Lost          int `json:"lost"`
	// Delivered is the frames handed to the receiving application, in order.
	Delivered []int `json:"delivered"`
	// Ticks is when the last frame was acknowledged.
	Ticks int `json:"ticks"`
}

// StopAndWait sends frames 0..n-1, one at a time: the next frame goes only
// when the last is acknowledged, and a frame is resent when its timer runs
// out. It is Go-Back-N with a window of one.
//
// Time: O(ticks). Space: O(n + ticks).
func StopAndWait(n int, ch Channel) (ARQResult, error) { return arq(n, 1, false, ch) }

// GoBackN sends frames 0..n-1 keeping up to window unacknowledged. The
// receiver accepts only the next frame in order and acks cumulatively
// ("everything before k arrived"); when the oldest frame's timer runs out
// the sender resends it and every frame after it: one loss costs a window.
//
// Time: O(ticks * window). Space: O(n + ticks).
func GoBackN(n, window int, ch Channel) (ARQResult, error) { return arq(n, window, false, ch) }

// SelectiveRepeat sends frames 0..n-1 keeping up to window unacknowledged.
// The receiver buffers frames that arrive out of order and acks each one; the
// sender keeps a timer per frame and resends only the frames whose timer runs
// out. Sequence numbers are not wrapped, so the window-size limit of a real
// protocol (window <= half the sequence space) does not arise.
//
// Time: O(ticks * window). Space: O(n + ticks).
func SelectiveRepeat(n, window int, ch Channel) (ARQResult, error) {
	return arq(n, window, true, ch)
}

type packet struct {
	at, seq int
	ack     bool
}

func arq(n, window int, selective bool, ch Channel) (ARQResult, error) {
	if n < 1 || n > 1000 || window < 1 || window > 64 || ch.Loss < 0 || ch.Loss > 90 {
		return ARQResult{}, fmt.Errorf("%w: 1 to 1000 frames, a window of 1 to 64, loss 0 to 90 percent", ErrInput)
	}
	s := newARQSim(n, window, selective, ch)
	for t := 0; t < MaxTicks; t++ {
		s.arrivals(t)
		if s.base >= n {
			s.res.Ticks = t
			return s.res, nil
		}
		s.timeouts(t)
		s.transmit(t)
	}
	return s.res, fmt.Errorf("%w: not finished in %d ticks; lower the loss", ErrInput, MaxTicks)
}

// arqSim is one ARQ run: the sender's window and timers, the receiver's
// state, and the frames and acks in flight. Go-Back-N and Selective Repeat
// differ only where selective is checked.
type arqSim struct {
	n, window int
	selective bool
	ch        Channel
	r         *rand.Rand
	drop      map[int]bool
	res       ARQResult
	flight    []packet
	acked     []bool // sender: frame acknowledged (selective)
	timer     []int  // sender: tick a frame's timer was started, -1 when off
	got       []bool // receiver: frame buffered (selective)
	everSent  []bool // sender: frame went out before
	base      int    // sender window is [base, next)
	next      int
	expected  int   // receiver's next in-order frame
	resend    []int // frames queued to go again
}

func newARQSim(n, window int, selective bool, ch Channel) *arqSim {
	if ch.Delay < 1 {
		ch.Delay = 1
	}
	if ch.Timeout <= 2*ch.Delay {
		ch.Timeout = 2*ch.Delay + 2
	}
	s := &arqSim{n: n, window: window, selective: selective, ch: ch, r: rand.New(rand.NewSource(ch.Seed)),
		drop: map[int]bool{}, acked: make([]bool, n), timer: make([]int, n), got: make([]bool, n), everSent: make([]bool, n)}
	for _, d := range ch.Drop {
		s.drop[d] = true
	}
	for i := range s.timer {
		s.timer[i] = -1
	}
	return s
}

func (s *arqSim) lose() bool { return s.ch.Loss > 0 && s.r.Intn(100) < s.ch.Loss }

// arrivals delivers this tick's packets: acks to the sender, frames to the
// receiver, whose acks join the packets still in flight.
func (s *arqSim) arrivals(t int) {
	var still []packet
	for _, p := range s.flight {
		switch {
		case p.at != t:
			still = append(still, p)
		case p.ack:
			s.onAck(t, p.seq)
		default:
			ack := s.onFrame(t, p.seq)
			if s.lose() {
				s.res.Lost++
				s.res.Timeline.add(t, "link", fmt.Sprintf("ack %d lost", ack), "blocked")
			} else {
				still = append(still, packet{at: t + s.ch.Delay, seq: ack, ack: true})
				s.res.Timeline.add(t, "receiver", fmt.Sprintf("ack %d", ack), "running")
			}
		}
	}
	s.flight = still
}

// onAck slides the window: past every acknowledged frame (selective), or up
// to the cumulative ack, restarting the timer on the new oldest frame.
func (s *arqSim) onAck(t, seq int) {
	s.res.Timeline.add(t, "sender", fmt.Sprintf("got ack %d", seq), "done")
	if s.selective {
		s.acked[seq], s.timer[seq] = true, -1
		for s.base < s.n && s.acked[s.base] {
			s.base++
		}
		return
	}
	if seq <= s.base { // cumulative: every frame before seq arrived
		return
	}
	for i := s.base; i < seq; i++ {
		s.timer[i] = -1
	}
	s.base = seq
	if s.base < s.next {
		s.timer[s.base] = t
	}
}

// onFrame is the receiver taking frame seq; it returns the ack to send.
func (s *arqSim) onFrame(t, seq int) int {
	if !s.selective {
		if seq == s.expected {
			s.deliver(t, seq)
		} else {
			s.res.Timeline.add(t, "receiver", fmt.Sprintf("discard %d", seq), "idle")
		}
		return s.expected
	}
	if seq < s.expected || seq >= s.expected+s.window || s.got[seq] {
		s.res.Timeline.add(t, "receiver", fmt.Sprintf("duplicate %d", seq), "idle")
		return seq
	}
	s.got[seq] = true
	s.res.Timeline.add(t, "receiver", fmt.Sprintf("buffer %d", seq), "running")
	for s.expected < s.n && s.got[s.expected] {
		s.deliver(t, s.expected)
	}
	return seq
}

// deliver hands the next in-order frame to the application.
func (s *arqSim) deliver(t, seq int) {
	s.res.Delivered = append(s.res.Delivered, seq)
	s.res.Timeline.add(t, "receiver", fmt.Sprintf("deliver %d", seq), "done")
	s.expected++
}

// timeouts queues what to resend: each timed-out frame (selective), or the
// whole window behind the oldest frame's timer (Go-Back-N).
func (s *arqSim) timeouts(t int) {
	if s.selective {
		for i := s.base; i < s.next; i++ {
			if s.timer[i] >= 0 && t-s.timer[i] >= s.ch.Timeout && !s.acked[i] {
				s.res.Timeline.add(t, "sender", fmt.Sprintf("timeout %d", i), "idle")
				s.timer[i] = -1
				s.resend = append(s.resend, i)
			}
		}
		return
	}
	if s.base < s.next && s.timer[s.base] >= 0 && t-s.timer[s.base] >= s.ch.Timeout {
		s.res.Timeline.add(t, "sender", fmt.Sprintf("timeout %d", s.base), "idle")
		s.timer[s.base] = -1
		s.resend = s.resend[:0]
		for i := s.base; i < s.next; i++ {
			s.resend = append(s.resend, i)
		}
	}
}

// transmit sends one frame a tick: a queued resend first, else a new frame
// if the window has room.
func (s *arqSim) transmit(t int) {
	seq := s.pick()
	if seq < 0 {
		return
	}
	s.res.Sent++
	verb := "send"
	if s.everSent[seq] {
		s.res.Retransmitted++
		verb = "resend"
	}
	s.everSent[seq] = true
	if s.selective {
		s.timer[seq] = t // a timer per frame
	} else if s.timer[s.base] < 0 {
		s.timer[s.base] = t // one timer, on the oldest frame
	}
	s.res.Timeline.add(t, "sender", fmt.Sprintf("%s %d", verb, seq), "running")
	if s.drop[s.res.Sent] || s.lose() {
		s.res.Lost++
		s.res.Timeline.add(t, "link", fmt.Sprintf("frame %d lost", seq), "blocked")
		return
	}
	s.flight = append(s.flight, packet{at: t + s.ch.Delay, seq: seq})
}

// pick is the frame to send this tick, or -1 for none.
func (s *arqSim) pick() int {
	for len(s.resend) > 0 {
		f := s.resend[0]
		s.resend = s.resend[1:]
		if f >= s.base && (!s.selective || !s.acked[f]) {
			return f
		}
	}
	if s.next < s.n && s.next < s.base+s.window {
		s.next++
		return s.next - 1
	}
	return -1
}
