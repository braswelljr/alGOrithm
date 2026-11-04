// Package networking simulates how computers talk over unreliable links, one
// file per idea:
//
//	arq.go         stop-and-wait, Go-Back-N and Selective Repeat over a lossy link
//	congestion.go  TCP congestion control: slow start, avoidance, fast recovery
//	ratelimit.go   the token bucket and the leaky bucket on a request trace
//	routing.go     distance-vector (with count to infinity) and link-state routing
//	ipv4.go        CIDR prefixes, subnetting and longest-prefix-match lookup
//	checksum.go    the Internet checksum (RFC 1071)
//	nagle.go       Nagle's algorithm coalescing small writes
//	dns.go         resolving a name down a simulated DNS hierarchy
//	layers.go      the OSI and TCP/IP models, encapsulation and decapsulation
//	tcpstate.go    TCP's state machine (handshake, teardown) and flow control
//	lan.go         ARP with a cache, DHCP's DORA, NAT with port translation
//	http.go        an HTTP request waterfall, HTTP/1.1 vs HTTP/2 vs HTTP/3,
//	               QUIC's round trips, and the WebSocket upgrade and frame
//	tls.go         the messages of a TLS 1.3 handshake
//	mac.go         Ethernet's CSMA/CD and Wi-Fi's CSMA/CA with hidden terminals
//	stp.go         the spanning tree protocol and BGP's path vector
//	delays.go      circuit vs packet switching, the bandwidth-delay product,
//	               traceroute, IP fragmentation, TCP's retransmission timer
//
// Every simulation counts time in whole ticks and draws its randomness (which
// frames are lost) from a seed, so a run replays exactly. Each returns its
// result and a list of Events (who did what at which tick) for a timeline.
// The models keep what the idea needs and leave the rest out: no real
// sockets, no byte-level headers except where the idea is the header (the
// checksum), and no clocks but the tick counter.
package networking

import "errors"

// ErrInput is returned for a simulation or input that cannot be run.
var ErrInput = errors.New("networking: invalid input")

// MaxTicks bounds every simulation.
const MaxTicks = 20_000

// Event is one thing that happened at Tick on Lane ("sender", "receiver",
// "link", a router or a server). State is "running" for an action, "blocked"
// for a loss or a drop, "idle" for waiting or a timeout, and "done" for a
// delivery or an answer.
type Event struct {
	Tick  int    `json:"tick"`
	Lane  string `json:"lane"`
	What  string `json:"what"`
	State string `json:"state"`
}

// Timeline is the events of a run in tick order.
type Timeline []Event

// add appends an event, joining it to an event already on the same lane at
// the same tick so a timeline never draws two bars in one slot.
func (t *Timeline) add(tick int, lane, what, state string) {
	for i := len(*t) - 1; i >= 0 && (*t)[i].Tick == tick; i-- {
		if (*t)[i].Lane == lane {
			(*t)[i].What += ", " + what
			if rank(state) > rank((*t)[i].State) {
				(*t)[i].State = state
			}
			return
		}
	}
	*t = append(*t, Event{Tick: tick, Lane: lane, What: what, State: state})
}

// rank orders states so the most telling one colors a joined slot.
func rank(s string) int {
	switch s {
	case "blocked":
		return 3
	case "done":
		return 2
	case "running":
		return 1
	default:
		return 0
	}
}
