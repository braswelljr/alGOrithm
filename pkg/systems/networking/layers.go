package networking

import (
	"fmt"
	"strings"
)

// LayerStep is one layer's part in sending data down the stack.
type LayerStep struct {
	Layer   string `json:"layer"`
	PDU     string `json:"pdu"`     // what the unit is called at this layer
	Header  int    `json:"header"`  // bytes this layer adds in front
	Trailer int    `json:"trailer"` // bytes it adds behind (Ethernet's FCS)
	Total   int    `json:"total"`   // bytes of the unit once this layer is done
	Shape   string `json:"shape"`   // the unit drawn as nested brackets
	Does    string `json:"does"`
}

type layer struct {
	name, pdu, tag string
	header, tail   int
	does           string
}

var osiLayers = []layer{
	{"7 Application", "data", "HTTP", -1, 0, "the request itself: GET / HTTP/1.1 and its headers"},
	{"6 Presentation", "data", "TLS", 5, 0, "encoding and encryption: a TLS record header (and, in reality, a tag)"},
	{"5 Session", "data", "", 0, 0, "keeping a conversation going; on the Internet this lives in the application"},
	{"4 Transport", "segment", "TCP", 20, 0, "ports, sequence numbers, acks and windows"},
	{"3 Network", "packet", "IP", 20, 0, "addresses and routing from host to host"},
	{"2 Data link", "frame", "Eth", 14, 4, "MAC addresses on one link, and a checksum (FCS) at the end"},
	{"1 Physical", "bits", "PHY", 8, 0, "a preamble to sync the clocks, then the bits on the wire"},
}

var tcpipLayers = []layer{
	{"Application (OSI 5-7)", "message", "HTTP", -1, 0, "the request, and TLS around it when encrypted"},
	{"Transport (OSI 4)", "segment", "TCP", 20, 0, "ports, sequence numbers, acks and windows"},
	{"Internet (OSI 3)", "packet", "IP", 20, 0, "addresses and routing from host to host"},
	{"Link (OSI 1-2)", "frame", "Eth", 14 + 8, 4, "MAC addresses, preamble and checksum on one link"},
}

// Encapsulate sends payload down the OSI model (osi true) or the TCP/IP
// model as an HTTP request: each layer wraps what the one above handed it
// in its own header (and trailer), and the receiver strips them in reverse.
// Header sizes are the usual minimums: TCP and IPv4 without options, an
// untagged Ethernet frame.
//
// Time: O(layers). Space: O(layers).
func Encapsulate(payload string, osi bool) ([]LayerStep, error) {
	if len(payload) > 1<<16 {
		return nil, fmt.Errorf("%w: at most 64 KiB of payload", ErrInput)
	}
	layers := tcpipLayers
	if osi {
		layers = osiLayers
	}
	app := "GET / HTTP/1.1\r\nHost: example.com\r\n\r\n"
	size := len(payload)
	shape := fmt.Sprintf("data %d", size)
	var out []LayerStep
	for _, l := range layers {
		h := l.header
		if h < 0 {
			h = len(app)
		}
		size += h + l.tail
		if l.tag != "" {
			inner := shape
			shape = fmt.Sprintf("[%s %d | %s", l.tag, h, inner)
			if l.tail > 0 {
				shape += fmt.Sprintf(" | FCS %d", l.tail)
			}
			shape += "]"
		}
		out = append(out, LayerStep{Layer: l.name, PDU: l.pdu, Header: h, Trailer: l.tail, Total: size, Shape: shape, Does: l.does})
	}
	return out, nil
}

// Decapsulate lists what the receiver strips at each layer, bottom up.
func Decapsulate(steps []LayerStep) []string {
	var out []string
	for i := len(steps) - 1; i >= 0; i-- {
		s := steps[i]
		strip := s.Header + s.Trailer
		left := s.Total - strip
		out = append(out, fmt.Sprintf("%s strips %d bytes, passes %d up", strings.TrimSpace(s.Layer), strip, left))
	}
	return out
}
