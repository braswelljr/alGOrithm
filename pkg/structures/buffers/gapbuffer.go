package buffers

import (
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// GapBuffer is text in one array with a gap of free space at the cursor:
// typing fills the gap, backspace widens it, and only moving the cursor
// shifts text, by the distance moved. That is cheap because people type
// where they are and move a little at a time. When the gap fills up the
// array doubles.
//
// It works on bytes; multi-byte UTF-8 characters count as several positions.
//
// Time: O(1) amortized Insert and Delete at the cursor; O(d) to move the
// cursor d positions.
// Space: O(n).
type GapBuffer struct {
	buf        []byte
	start, end int // the gap is buf[start:end]; the cursor is at start
}

// NewGapBuffer returns a buffer holding text with the cursor at the end.
func NewGapBuffer(text string) *GapBuffer {
	capacity := max(2*len(text), 8)
	g := &GapBuffer{buf: make([]byte, capacity)}
	copy(g.buf, text)
	g.start, g.end = len(text), capacity
	return g
}

// Len returns the length of the text.
func (g *GapBuffer) Len() int { return len(g.buf) - (g.end - g.start) }

// Cursor returns the cursor's position, 0..Len.
func (g *GapBuffer) Cursor() int { return g.start }

// Gap returns how many free bytes the gap holds.
func (g *GapBuffer) Gap() int { return g.end - g.start }

// Move puts the cursor at pos, clamped to 0..Len, and returns how many bytes
// it shifted.
func (g *GapBuffer) Move(pos int) int {
	pos = min(max(pos, 0), g.Len())
	moved := 0
	for g.start > pos { // move left: the byte before the gap goes after it
		g.start--
		g.end--
		g.buf[g.end] = g.buf[g.start]
		moved++
	}
	for g.start < pos { // move right: the byte after the gap goes before it
		g.buf[g.start] = g.buf[g.end]
		g.start++
		g.end++
		moved++
	}
	return moved
}

// Insert types s at the cursor and moves the cursor past it.
func (g *GapBuffer) Insert(s string) {
	if len(s) > g.Gap() {
		g.grow(len(s))
	}
	copy(g.buf[g.start:], s)
	g.start += len(s)
}

func (g *GapBuffer) grow(need int) {
	capacity := max(2*len(g.buf), len(g.buf)+need)
	buf := make([]byte, capacity)
	copy(buf, g.buf[:g.start])
	tail := len(g.buf) - g.end
	copy(buf[capacity-tail:], g.buf[g.end:])
	g.buf, g.end = buf, capacity-tail
}

// Backspace deletes up to n bytes before the cursor and returns how many it deleted.
func (g *GapBuffer) Backspace(n int) int {
	n = min(max(n, 0), g.start)
	g.start -= n
	return n
}

// Delete deletes up to n bytes after the cursor and returns how many it deleted.
func (g *GapBuffer) Delete(n int) int {
	n = min(max(n, 0), len(g.buf)-g.end)
	g.end += n
	return n
}

// String returns the text.
func (g *GapBuffer) String() string { return string(g.buf[:g.start]) + string(g.buf[g.end:]) }

// Snapshot describes the array cell by cell, the gap's cells as "_" with the
// cursor on the first.
func (g *GapBuffer) Snapshot() trace.Snapshot {
	cells := make([]string, len(g.buf))
	for i := range g.buf {
		if i >= g.start && i < g.end {
			cells[i] = "_"
		} else {
			cells[i] = strings.ReplaceAll(string(g.buf[i]), " ", "␠")
		}
	}
	s := trace.Row(cells, "", "", "")
	for i := range s.Nodes {
		switch {
		case i == g.start && g.start < g.end:
			s.Nodes[i].Role = "cursor"
		case i >= g.start && i < g.end:
			s.Nodes[i].Role = "gap"
		}
	}
	return s
}
