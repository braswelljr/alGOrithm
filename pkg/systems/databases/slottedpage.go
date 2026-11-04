package databases

import (
	"fmt"
	"slices"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// slotBytes is the size of one slot directory entry (an offset and a length),
// headerBytes the size of the page header (the slot count and the free-space
// pointer).
const (
	slotBytes   = 4
	headerBytes = 4
)

type slot struct {
	offset, length int // length 0 marks a free slot
}

// SlottedPage is the layout most databases give a page of variable-length
// records: a header and a slot directory growing from the front, records
// growing from the back, and free space between. A record is named by its
// slot number, which never changes, so indexes can point at it while
// compaction moves its bytes. Deleting a record frees its slot for reuse and
// leaves a hole; compaction slides the records together to turn the holes
// back into free space.
//
// Simplified: records are strings held as their lengths and positions in a
// page of size bytes. The zero SlottedPage is not ready; use NewSlottedPage.
type SlottedPage struct {
	size    int
	slots   []slot
	records map[int]string
	end     int // where the record area starts; it grows down from size
}

// NewSlottedPage returns an empty page of size bytes.
func NewSlottedPage(size int) *SlottedPage {
	return &SlottedPage{size: max(16, size), records: map[int]string{}, end: max(16, size)}
}

// Free is the bytes between the slot directory and the records; Holes the
// bytes of deleted records not yet compacted away.
func (p *SlottedPage) Free() int { return p.end - headerBytes - slotBytes*len(p.slots) }

// Holes is the bytes left by deleted records until the next compaction.
func (p *SlottedPage) Holes() int {
	used := 0
	for _, s := range p.slots {
		used += s.length
	}
	return p.size - p.end - used
}

// Insert stores record and returns its slot, reusing a free slot when there
// is one, and compacting first when only the holes have room. It returns an
// error when the page is full. Time: O(s) for s slots, plus O(n) bytes moved
// by a compaction.
func (p *SlottedPage) Insert(record string) (int, error) {
	if record == "" {
		return 0, fmt.Errorf("%w: an empty record", ErrInput)
	}
	id := slices.IndexFunc(p.slots, func(s slot) bool { return s.length == 0 })
	need := len(record)
	if id < 0 {
		need += slotBytes
	}
	if p.Free() < need && p.Free()+p.Holes() >= need {
		p.Compact()
	}
	if p.Free() < need {
		return 0, fmt.Errorf("%w: the page is full (%d bytes free, %d in holes, %d needed)", ErrInput, p.Free(), p.Holes(), need)
	}
	p.end -= len(record)
	s := slot{offset: p.end, length: len(record)}
	if id < 0 {
		id = len(p.slots)
		p.slots = append(p.slots, s)
	} else {
		p.slots[id] = s
	}
	p.records[id] = record
	return id, nil
}

// Get returns the record in slot id, and whether there is one. Time: O(1).
func (p *SlottedPage) Get(id int) (string, bool) {
	r, ok := p.records[id]
	return r, ok
}

// Delete removes the record in slot id and reports whether there was one;
// its bytes become a hole. Time: O(1).
func (p *SlottedPage) Delete(id int) bool {
	if _, ok := p.records[id]; !ok {
		return false
	}
	delete(p.records, id)
	p.slots[id] = slot{}
	return true
}

// Compact moves the records to the back of the page, in their order there,
// so the holes join the free space; slot numbers stay the same. Time: O(n)
// bytes moved.
func (p *SlottedPage) Compact() {
	ids := p.live()
	slices.SortFunc(ids, func(a, b int) int { return p.slots[b].offset - p.slots[a].offset })
	p.end = p.size
	for _, id := range ids {
		p.end -= p.slots[id].length
		p.slots[id].offset = p.end
	}
}

func (p *SlottedPage) live() []int {
	var ids []int
	for id, s := range p.slots {
		if s.length > 0 {
			ids = append(ids, id)
		}
	}
	return ids
}

// Snapshot draws the page front to back: the header, each slot pointing at
// its record, the free space, and the records and holes by position.
func (p *SlottedPage) Snapshot() trace.Snapshot {
	s := trace.Snapshot{Layout: "row", Nodes: []trace.Node{}}
	add := func(label, role string) int {
		id := len(s.Nodes)
		s.Nodes = append(s.Nodes, trace.Node{ID: id, Label: label, Role: role, Parent: -1})
		return id
	}
	add(fmt.Sprintf("header: %d slots", len(p.slots)), "head")
	slotNode := make([]int, len(p.slots))
	for i, sl := range p.slots {
		if sl.length == 0 {
			slotNode[i] = add(fmt.Sprintf("slot %d: free", i), "")
		} else {
			slotNode[i] = add(fmt.Sprintf("slot %d: @%d len %d", i, sl.offset, sl.length), "")
		}
	}
	add(fmt.Sprintf("free %d bytes", p.Free()), "")
	ids := p.live()
	slices.SortFunc(ids, func(a, b int) int { return p.slots[a].offset - p.slots[b].offset })
	at := p.end
	for _, id := range ids {
		if gap := p.slots[id].offset - at; gap > 0 {
			add(fmt.Sprintf("hole %d bytes", gap), "")
		}
		n := add(fmt.Sprintf("%q", p.records[id]), "")
		s.Edges = append(s.Edges, trace.Edge{From: slotNode[id], To: n})
		at = p.slots[id].offset + p.slots[id].length
	}
	if gap := p.size - at; gap > 0 && len(ids) > 0 {
		add(fmt.Sprintf("hole %d bytes", gap), "")
	}
	return s
}
