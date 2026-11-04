package ciphers

import "github.com/braswelljr/alGOrithm/pkg/trace"

// RailFence writes text in a zigzag down and up rails rows, then reads the
// rows one after another: a transposition cipher, so the letters are the
// same, only their order changes. Every byte moves, spaces included. One
// rail, or rails >= len(text), leaves the text as it is.
//
// Time: O(n). Space: O(n).
func RailFence(text string, rails int) string { return TraceRailFence(text, rails, trace.Nop{}) }

// TraceRailFence is RailFence, reporting each byte written to the result
// (Set at its new position, Values[0] the byte) in the order the rails are
// read.
func TraceRailFence(text string, rails int, r trace.Recorder) string {
	order := railOrder(len(text), rails)
	out := make([]byte, len(text))
	for to, from := range order {
		out[to] = text[from]
		r.Record(trace.Event{Kind: trace.Set, Indices: []int{to}, Values: []int{int(text[from])}})
	}
	return string(out)
}

// RailFenceDecrypt reverses RailFence for the same number of rails.
func RailFenceDecrypt(text string, rails int) string {
	order := railOrder(len(text), rails)
	out := make([]byte, len(text))
	for to, from := range order {
		out[from] = text[to]
	}
	return string(out)
}

// railOrder returns, for each position of the ciphertext, the position of the
// plaintext byte it holds.
func railOrder(n, rails int) []int {
	if rails < 2 || rails >= n {
		order := make([]int, n)
		for i := range order {
			order[i] = i
		}
		return order
	}
	rows := make([][]int, rails)
	row, step := 0, 1
	for i := 0; i < n; i++ {
		rows[row] = append(rows[row], i)
		switch row {
		case 0:
			step = 1
		case rails - 1:
			step = -1
		}
		row += step
	}
	order := make([]int, 0, n)
	for _, rw := range rows {
		order = append(order, rw...)
	}
	return order
}
