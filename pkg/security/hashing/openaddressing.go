package hashing

import "errors"

// Probe names an open-addressing probe sequence.
type Probe int

// The probe sequences, for a key hashing to h in a table of m slots.
const (
	// Linear tries h, h+1, h+2, ...: simple and cache friendly, but runs of
	// full slots grow together (primary clustering).
	Linear Probe = iota
	// Quadratic tries h, h+1, h+4, h+9, ...: breaks up runs, but keys with
	// the same h still share a path (secondary clustering). With a prime
	// table size it reaches half the table, so the table resizes at 50%.
	Quadratic
	// Double tries h, h+d, h+2d, ... with d from a second hash, so keys that
	// collide once usually part ways.
	Double
)

func (p Probe) String() string {
	return [...]string{"linear", "quadratic", "double"}[p]
}

// Slot states.
const (
	Free = iota
	Used
	Tombstone // a deleted key: searches go past it, inserts may reuse it
)

// ErrFull is returned when no free slot was found (it cannot happen while the
// table resizes, but the probe loop is bounded all the same).
var ErrFull = errors.New("hashing: no free slot found")

// OpenTable is a hash table of int keys with open addressing: every key lives
// in the slots array itself, and a collision moves on along the probe sequence.
// Deleting leaves a tombstone so later searches still find keys placed past
// it. The table doubles (to the next prime) when used and tombstone slots pass
// MaxLoad. The zero OpenTable is not ready; use NewOpenTable.
type OpenTable struct {
	probe   Probe
	keys    []int
	state   []int
	size    int // keys stored
	tombs   int
	MaxLoad float64
}

// NewOpenTable returns an empty table of at least capacity slots (rounded up
// to a prime) using probe.
func NewOpenTable(capacity int, probe Probe) *OpenTable {
	m := nextPrime(max(capacity, 3))
	load := 0.7
	if probe == Quadratic {
		load = 0.5
	}
	return &OpenTable{probe: probe, keys: make([]int, m), state: make([]int, m), MaxLoad: load}
}

// Len is the number of keys; Cap the number of slots. Time: O(1).
func (t *OpenTable) Len() int { return t.size }

// Cap is the number of slots.
func (t *OpenTable) Cap() int { return len(t.keys) }

// Slot returns the key and state of slot i.
func (t *OpenTable) Slot(i int) (key, state int) { return t.keys[i], t.state[i] }

// LoadFactor is the fraction of slots holding keys.
func (t *OpenTable) LoadFactor() float64 { return float64(t.size) / float64(len(t.keys)) }

func hash1(key, m int) int {
	h := uint64(key) * 0x9E3779B97F4A7C15
	return int((h >> 32) % uint64(m))
}

// hash2 is the step for double hashing: never 0, and below m, which is
// prime, so every slot is reachable.
func hash2(key, m int) int {
	h := uint64(key)*0xC2B2AE3D27D4EB4F + 0x165667B19E3779F9
	return 1 + int((h>>33)%uint64(m-1))
}

// at returns the i-th slot of key's probe sequence.
func (t *OpenTable) at(key, i int) int {
	m := len(t.keys)
	h := hash1(key, m)
	switch t.probe {
	case Quadratic:
		return (h + i*i) % m
	case Double:
		return (h + i*hash2(key, m)) % m
	default: // Linear
		return (h + i) % m
	}
}

// find returns the slot holding key and the slots probed, or -1.
func (t *OpenTable) find(key int) (int, []int) {
	var probed []int
	for i := 0; i < len(t.keys); i++ {
		s := t.at(key, i)
		probed = append(probed, s)
		switch t.state[s] {
		case Free:
			return -1, probed
		case Used:
			if t.keys[s] == key {
				return s, probed
			}
		}
	}
	return -1, probed
}

// Search reports whether key is in the table and the slots probed to find out.
// Time: O(1) on average at a bounded load, O(n) in the worst case.
func (t *OpenTable) Search(key int) (found bool, probed []int) {
	s, probed := t.find(key)
	return s >= 0, probed
}

// Insert adds key and returns the slots probed; added is false when key was
// already there. It reuses the first tombstone on the way.
// Time: O(1) on average at a bounded load, O(n) for a resize.
func (t *OpenTable) Insert(key int) (added bool, probed []int, err error) {
	if s, p := t.find(key); s >= 0 {
		return false, p, nil
	}
	if float64(t.size+t.tombs+1) > t.MaxLoad*float64(len(t.keys)) {
		t.resize(nextPrime(2 * len(t.keys)))
	}
	for i := 0; i < len(t.keys); i++ {
		s := t.at(key, i)
		probed = append(probed, s)
		if t.state[s] != Used {
			if t.state[s] == Tombstone {
				t.tombs--
			}
			t.keys[s], t.state[s] = key, Used
			t.size++
			return true, probed, nil
		}
	}
	return false, probed, ErrFull
}

// Delete removes key, leaving a tombstone, and returns the slots probed;
// removed is false when key was not there. Time: O(1) on average.
func (t *OpenTable) Delete(key int) (removed bool, probed []int) {
	s, probed := t.find(key)
	if s < 0 {
		return false, probed
	}
	t.state[s] = Tombstone
	t.size--
	t.tombs++
	return true, probed
}

// Keys returns the keys in slot order.
func (t *OpenTable) Keys() []int {
	var out []int
	for i, st := range t.state {
		if st == Used {
			out = append(out, t.keys[i])
		}
	}
	return out
}

func (t *OpenTable) resize(m int) {
	old := t.Keys()
	t.keys, t.state, t.size, t.tombs = make([]int, m), make([]int, m), 0, 0
	for _, k := range old {
		_, _, _ = t.Insert(k)
	}
}

func nextPrime(n int) int {
	for ; ; n++ {
		prime := n >= 2
		for d := 2; d*d <= n; d++ {
			if n%d == 0 {
				prime = false
				break
			}
		}
		if prime {
			return n
		}
	}
}
