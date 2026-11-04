package lookup_test

import (
	"math/rand"
	"reflect"
	"testing"

	"github.com/braswelljr/alGOrithm/pkg/structures/lookup"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

func labels(s trace.Snapshot) []string {
	out := make([]string, len(s.Nodes))
	for i, n := range s.Nodes {
		out[i] = n.Label
	}
	return out
}

func TestSet(t *testing.T) {
	a := lookup.NewSet(1, 2, 3, 3)
	b := lookup.NewSet(3, 4)
	if a.Len() != 3 || a.Add(2) || !a.Add(9) || !a.Remove(9) || a.Remove(9) {
		t.Fatal("Add/Remove wrong")
	}
	if u := a.Union(b); u.Len() != 4 {
		t.Errorf("union has %d", u.Len())
	}
	if i := a.Intersection(b); i.Len() != 1 || !i.Contains(3) {
		t.Error("intersection wrong")
	}
	if d := a.Difference(b); d.Len() != 2 || d.Contains(3) {
		t.Error("difference wrong")
	}
	if !lookup.NewSet(1).SubsetOf(a) || b.SubsetOf(a) {
		t.Error("SubsetOf wrong")
	}
	if got := labels(a.Snapshot()); !reflect.DeepEqual(got, []string{"1", "2", "3"}) {
		t.Errorf("snapshot %v", got)
	}
}

func TestTrie(t *testing.T) {
	tr := lookup.NewTrie("car", "cart", "care", "cat", "dog", "car")
	if tr.Len() != 5 || !tr.Contains("cart") || tr.Contains("ca") {
		t.Fatal("Contains wrong")
	}
	if got := tr.WithPrefix("car"); !reflect.DeepEqual(got, []string{"car", "care", "cart"}) {
		t.Errorf("WithPrefix = %v", got)
	}
	nodes := len(tr.Snapshot().Nodes)
	if !tr.Remove("cart") || tr.Remove("cart") || len(tr.Snapshot().Nodes) != nodes-1 {
		t.Error("Remove did not prune")
	}
}

func TestHashTable(t *testing.T) {
	h := lookup.NewHashTable[string, int](2)
	words := []string{"a", "b", "c", "d", "e", "f", "g"}
	for i, w := range words {
		h.Put(w, i)
	}
	if h.Len() != 7 || h.Buckets() <= 2 {
		t.Fatalf("len %d buckets %d", h.Len(), h.Buckets())
	}
	for i, w := range words {
		if v, ok := h.Get(w); !ok || v != i {
			t.Fatalf("Get(%q) = %d, %v", w, v, ok)
		}
	}
	if h.Put("a", 100) || !h.Remove("a") || h.Remove("a") {
		t.Error("Put/Remove wrong")
	}
	if len(lookup.NewHashTable[int, int](3).Snapshot().Nodes) != 4 {
		t.Error("empty snapshot wrong")
	}
}

func TestUnionFind(t *testing.T) {
	u := lookup.NewUnionFind(0, 1, 2, 3, 4, 5)
	u.Union(0, 1)
	u.Union(2, 3)
	u.Union(1, 3)
	if u.Union(0, 2) || !u.Connected(0, 3) || u.Connected(0, 5) || u.Groups() != 3 {
		t.Errorf("groups %d", u.Groups())
	}
	if _, ok := u.Find(42); ok {
		t.Error("found an unknown element")
	}
	roots := 0
	for _, n := range u.Snapshot().Nodes {
		if n.Role == "root" {
			roots++
		}
	}
	if roots != u.Groups() {
		t.Errorf("%d roots, %d groups", roots, u.Groups())
	}
}

func TestLRU(t *testing.T) {
	c := lookup.NewLRU[int, string](2)
	c.Put(1, "one")
	c.Put(2, "two")
	c.Get(1)
	if k, ok := c.Put(3, "three"); !ok || k != 2 {
		t.Errorf("evicted %d, %v; want 2", k, ok)
	}
	if _, ok := c.Get(2); ok {
		t.Error("2 still cached")
	}
	if !reflect.DeepEqual(c.Keys(), []int{3, 1}) {
		t.Errorf("keys %v", c.Keys())
	}
	if _, ok := c.Put(1, "uno"); ok || c.Len() != 2 {
		t.Error("update evicted")
	}
}

func TestLFUCache(t *testing.T) {
	c := lookup.NewLFUCache[int, string](2)
	c.Put(1, "a")
	c.Put(2, "b")
	c.Get(1)
	c.Get(1)
	if k, ok := c.Put(3, "c"); !ok || k != 2 {
		t.Errorf("evicted %d, %v; want 2, the least used", k, ok)
	}
	c.Get(3)
	c.Get(3) // 1 and 3 used three times each now; 1 less recently
	if k, ok := c.Put(4, "d"); !ok || k != 1 {
		t.Errorf("tie: evicted %d, want 1, the least recently used of the least used", k)
	}
	if f, _ := c.Frequency(3); f != 3 || c.Len() != 2 {
		t.Errorf("freq %d len %d", f, c.Len())
	}
	if got := c.Keys(); !reflect.DeepEqual(got, []int{3, 4}) {
		t.Errorf("keys %v", got)
	}
	if _, ok := c.Get(1); ok {
		t.Error("1 still cached")
	}
	// Against a slow model.
	r := rand.New(rand.NewSource(2))
	big := lookup.NewLFUCache[int, int](5)
	freq, last := map[int]int{}, map[int]int{}
	for step := 0; step < 3000; step++ {
		k := r.Intn(12)
		if r.Intn(2) == 0 {
			_, okModel := freq[k]
			if _, ok := big.Get(k); ok != okModel {
				t.Fatalf("Get(%d) presence disagreed", k)
			}
			if okModel {
				freq[k]++
				last[k] = step
			}
			continue
		}
		var want int
		wantEvict := false
		if _, ok := freq[k]; !ok && len(freq) == 5 {
			first := true
			for kk := range freq {
				if first || freq[kk] < freq[want] || (freq[kk] == freq[want] && last[kk] < last[want]) {
					want, first = kk, false
				}
			}
			wantEvict = true
		}
		got, ev := big.Put(k, step)
		if ev != wantEvict || (ev && got != want) {
			t.Fatalf("Put(%d) evicted %d,%v; want %d,%v", k, got, ev, want, wantEvict)
		}
		if ev {
			delete(freq, want)
			delete(last, want)
		}
		freq[k]++
		last[k] = step
	}
}
