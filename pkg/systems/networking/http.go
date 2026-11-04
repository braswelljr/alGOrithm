package networking

import (
	"crypto/sha1" //nolint:gosec // RFC 6455 fixes SHA-1 for the WebSocket accept key; it is not used for security here
	"encoding/base64"
	"fmt"
	"strconv"
)

// Phase is one bar of a request waterfall: Lane did Name from Start to End
// (milliseconds, or ticks).
type Phase struct {
	Lane  string `json:"lane"`
	Name  string `json:"name"`
	Start int    `json:"start"`
	End   int    `json:"end"`
}

// Waterfall times requests HTTP requests to one server, one after another,
// as a browser's network panel shows them: DNS lookup (one round trip of
// rtt, the first time only), the TCP handshake (one round trip), the TLS
// handshake (tls 13: one round trip, tls 12: two, 0: plain HTTP), then the
// request and the wait for the first byte (one round trip plus server
// milliseconds), then the download (transfer milliseconds). With keepAlive
// the later requests reuse the open connection and pay only the last two;
// without it each opens a new connection.
//
// Time: O(requests). Space: O(requests).
func Waterfall(requests, rtt, server, transfer, tls int, keepAlive bool) ([]Phase, error) {
	if requests < 1 || requests > 50 || rtt < 1 || server < 0 || transfer < 0 || (tls != 0 && tls != 12 && tls != 13) {
		return nil, fmt.Errorf("%w: 1 to 50 requests, positive times, TLS 0, 12 or 13", ErrInput)
	}
	var out []Phase
	t := 0
	for r := 1; r <= requests; r++ {
		lane := "request " + strconv.Itoa(r)
		add := func(name string, d int) {
			out = append(out, Phase{Lane: lane, Name: name, Start: t, End: t + d})
			t += d
		}
		if r == 1 {
			add("DNS", rtt)
		}
		if r == 1 || !keepAlive {
			add("TCP connect", rtt)
			switch tls {
			case 13:
				add("TLS 1.3", rtt)
			case 12:
				add("TLS 1.2", 2*rtt)
			}
		}
		add("wait (TTFB)", rtt+server)
		add("download", transfer)
	}
	return out, nil
}

// Multiplexing loads resources responses of size frames each from one
// server over a link that carries one frame a tick, with a round trip of
// rtt ticks, four ways: HTTP/1.1 on one connection (one request at a time),
// HTTP/1.1 on six connections (the browser's usual limit, sharing the
// link), HTTP/2 (every request at once on one connection, frames of the
// streams interleaved), and HTTP/3 over QUIC (the same, on independent
// streams). If lossAt is positive the frame sent at that tick is lost and
// arrives rtt ticks late: TCP delivers bytes in order, so every later frame
// on that connection waits for it (head-of-line blocking, which hits all of
// HTTP/2's streams at once), while QUIC holds back only the lost frame's
// stream. It returns each resource's waterfall bar per way.
//
// Time: O(resources * size). Space: O(resources * size).
func Multiplexing(resources, size, rtt, lossAt int) ([]Phase, map[string]int, error) {
	if resources < 1 || resources > 12 || size < 1 || size > 50 || rtt < 1 || rtt > 50 || lossAt < 0 {
		return nil, nil, fmt.Errorf("%w: 1 to 12 resources of 1 to 50 frames, a round trip of 1 to 50", ErrInput)
	}
	ways := []httpWay{{"HTTP/1.1", 1, false, false}, {"HTTP/1.1 x6", 6, false, false}, {"HTTP/2", 1, true, false}, {"HTTP/3", 1, true, true}}
	var out []Phase
	finish := map[string]int{}
	for _, w := range ways {
		start, done := w.load(resources, size, rtt, lossAt)
		for r := 0; r < resources; r++ {
			out = append(out, Phase{Lane: w.name + " #" + strconv.Itoa(r+1), Name: "resource " + strconv.Itoa(r+1), Start: start[r], End: done[r]})
			finish[w.name] = max(finish[w.name], done[r])
		}
	}
	return out, finish, nil
}

// httpWay is one way of loading the resources.
type httpWay struct {
	name   string
	conns  int  // connections (each a TCP or QUIC connection)
	all    bool // every request goes at once
	perStr bool // loss blocks only its stream (QUIC)
}

// httpLoad is one way's run: which resources each connection carries, and
// each resource's frames left, request time, first sendable tick and finish.
type httpLoad struct {
	httpWay
	rtt                      int
	conn                     [][]int
	left, start, ready, done []int
	blockedUntil             map[int]int // connection or stream -> tick held back to
	lost                     bool
}

// load runs the way and returns when each resource was requested and when
// its last frame arrived.
func (w httpWay) load(resources, size, rtt, lossAt int) (start, done []int) {
	l := &httpLoad{httpWay: w, rtt: rtt, conn: make([][]int, w.conns), left: make([]int, resources),
		start: make([]int, resources), ready: make([]int, resources), done: make([]int, resources), blockedUntil: map[int]int{}}
	for r := 0; r < resources; r++ {
		l.conn[r%w.conns] = append(l.conn[r%w.conns], r)
		l.left[r], l.start[r], l.ready[r] = size, -1, -1
	}
	for c := range l.conn {
		for i, r := range l.conn[c] {
			if w.all || i == 0 {
				l.start[r], l.ready[r] = 0, rtt // the request takes a round trip to come back as data
			}
		}
	}
	for t := 0; t < MaxTicks; t++ {
		active := l.active(t)
		if len(active) == 0 {
			if l.finished() {
				break
			}
			continue
		}
		l.send(t, active[t%len(active)], lossAt)
	}
	return l.start, l.done
}

// active lists the resources that may send a frame at t, round robin; an
// HTTP/1.1 connection carries one response at a time.
func (l *httpLoad) active(t int) []int {
	var active []int
	for c := range l.conn {
		for _, r := range l.conn[c] {
			if l.left[r] > 0 && l.ready[r] >= 0 && l.ready[r] <= t {
				active = append(active, r)
				if !l.all {
					break
				}
			}
		}
	}
	return active
}

func (l *httpLoad) finished() bool {
	for _, left := range l.left {
		if left > 0 {
			return false
		}
	}
	return true
}

// send puts one frame of resource r on the link at t. The frame sent at
// lossAt is lost and comes rtt late, holding back what follows it on the
// same connection (TCP) or stream (QUIC).
func (l *httpLoad) send(t, r, lossAt int) {
	l.left[r]--
	arrive := t + 1
	key := r % l.conns
	if l.perStr {
		key = 1000 + r
	}
	if lossAt > 0 && t == lossAt && !l.lost {
		l.lost = true
		arrive = t + 1 + l.rtt
		l.blockedUntil[key] = arrive
	}
	if b, ok := l.blockedUntil[key]; ok && arrive < b {
		arrive = b // held behind the lost frame
	}
	if l.left[r] > 0 {
		l.done[r] = max(l.done[r], arrive)
		return
	}
	l.done[r] = arrive
	if !l.all {
		l.requestNext(r, arrive)
	}
}

// requestNext sends the request after r on its HTTP/1.1 connection once r
// has arrived.
func (l *httpLoad) requestNext(r, arrive int) {
	c := r % l.conns
	for i, q := range l.conn[c] {
		if q == r && i+1 < len(l.conn[c]) {
			nq := l.conn[c][i+1]
			l.start[nq], l.ready[nq] = arrive, arrive+l.rtt
		}
	}
}

// RoundTripsToData compares how many round trips each way of starting a
// secure connection takes before the response's first byte arrives: TCP
// (1) then TLS 1.2 (2) then the request (1); TCP then TLS 1.3 (1) then the
// request; QUIC, which does the transport and TLS 1.3 handshakes together
// (1) then the request; and QUIC resuming with 0-RTT, which sends the
// request in its first flight (keys from a ticket of an earlier
// connection; such early data can be replayed, so only safe requests use it).
//
// Time: O(1). Space: O(1).
func RoundTripsToData(rtt int) ([]Phase, error) {
	if rtt < 1 || rtt > 10_000 {
		return nil, fmt.Errorf("%w: a round trip of 1 to 10000 ms", ErrInput)
	}
	ways := []struct {
		name  string
		steps []string
	}{
		{"TCP + TLS 1.2", []string{"TCP handshake", "TLS 1.2 (1)", "TLS 1.2 (2)", "request"}},
		{"TCP + TLS 1.3", []string{"TCP handshake", "TLS 1.3", "request"}},
		{"QUIC", []string{"QUIC + TLS 1.3 handshake", "request"}},
		{"QUIC 0-RTT", []string{"request with the first flight"}},
	}
	var out []Phase
	for _, w := range ways {
		for i, s := range w.steps {
			out = append(out, Phase{Lane: w.name, Name: s, Start: i * rtt, End: (i + 1) * rtt})
		}
	}
	return out, nil
}

// WebSocketAccept is the Sec-WebSocket-Accept a server answers an upgrade
// with key: base64(SHA-1(key + the fixed GUID of RFC 6455)), proving it
// understood the WebSocket handshake rather than being a confused cache.
//
// Time: O(1). Space: O(1).
func WebSocketAccept(key string) string {
	h := sha1.Sum([]byte(key + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11")) //nolint:gosec // fixed by RFC 6455
	return base64.StdEncoding.EncodeToString(h[:])
}

// WebSocketFrame builds the frame a client sends for a text message: FIN
// and opcode 1 (text), the mask bit and length, the 4-byte mask, and the
// payload XORed with the mask (clients must mask, so a proxy cannot be
// tricked by bytes that look like HTTP). Payloads over 125 bytes use the
// 16-bit length form; the 64-bit form is left out.
//
// Time: O(n). Space: O(n).
func WebSocketFrame(text string, mask [4]byte) ([]byte, error) {
	n := len(text)
	if n > 65535 {
		return nil, fmt.Errorf("%w: at most 65535 bytes", ErrInput)
	}
	frame := []byte{0x81}
	if n <= 125 {
		frame = append(frame, 0x80|byte(n))
	} else {
		frame = append(frame, 0x80|126, byte(n>>8), byte(n))
	}
	frame = append(frame, mask[:]...)
	for i := 0; i < n; i++ {
		frame = append(frame, text[i]^mask[i%4])
	}
	return frame, nil
}
