package hashing_test

import (
	"fmt"
	"hash/adler32"
	"hash/crc32"
	"hash/fnv"
	"math"
	"math/rand"
	"sort"
	"testing"

	"github.com/braswelljr/alGOrithm/pkg/security/hashing"
)

func TestFunctionsAgainstOracles(t *testing.T) {
	inputs := []string{"", "a", "hello", "The quick brown fox jumps over the lazy dog", string(make([]byte, 1000))}
	for _, s := range inputs {
		b := []byte(s)
		f32 := fnv.New32a()
		f32.Write(b)
		f64 := fnv.New64a()
		f64.Write(b)
		if hashing.FNV1a32(b) != f32.Sum32() || hashing.FNV1a64(b) != f64.Sum64() {
			t.Errorf("FNV-1a(%q) differs from hash/fnv", s[:min(len(s), 10)])
		}
		if hashing.CRC32(b) != crc32.ChecksumIEEE(b) {
			t.Errorf("CRC32(%q) differs from hash/crc32", s[:min(len(s), 10)])
		}
		if hashing.Adler32(b) != adler32.Checksum(b) {
			t.Errorf("Adler32(%q) differs from hash/adler32", s[:min(len(s), 10)])
		}
	}
	// Published MurmurHash3 x86_32 vectors.
	for _, v := range []struct {
		in   string
		seed uint32
		want uint32
	}{{"", 0, 0}, {"", 1, 0x514E28B7}, {"hello", 0, 0x248bfa47}, {"Hello, world!", 1234, 0xfaf6cdb3}, {"The quick brown fox jumps over the lazy dog", 0, 0x2e4ff723}} {
		if got := hashing.Murmur3_32([]byte(v.in), v.seed); got != v.want {
			t.Errorf("Murmur3(%q, %d) = %#x, want %#x", v.in, v.seed, got, v.want)
		}
	}
	if hashing.DJB2([]byte("a")) != 5381*33+'a' || hashing.SDBM([]byte("a")) != 'a' {
		t.Error("DJB2 or SDBM wrong")
	}
	if hashing.Polynomial([]byte("ab"), 31, 1_000_000_007) != 'a'*31+'b' {
		t.Error("Polynomial wrong")
	}
}

func TestDistribution(t *testing.T) {
	var keys []string
	for i := 0; i < 10000; i++ {
		keys = append(keys, fmt.Sprintf("key-%d", i))
	}
	d := hashing.Distribution(keys, 100, hashing.FNV1a32)
	sum := 0
	for _, c := range d.Counts {
		sum += c
	}
	if sum != len(keys) || d.Empty != 0 || d.ChiSquare > 200 {
		t.Errorf("FNV-1a spread: sum %d, empty %d, chi-square %.1f", sum, d.Empty, d.ChiSquare)
	}
	bad := hashing.Distribution(keys, 100, func(b []byte) uint32 { return uint32(len(b)) })
	if bad.ChiSquare < 10*d.ChiSquare {
		t.Errorf("a length hash should spread badly: chi-square %.1f", bad.ChiSquare)
	}
}

func TestOpenAddressingAgainstAMap(t *testing.T) {
	for _, probe := range []hashing.Probe{hashing.Linear, hashing.Quadratic, hashing.Double} {
		t.Run(probe.String(), func(t *testing.T) {
			tab := hashing.NewOpenTable(5, probe)
			model := map[int]bool{}
			r := rand.New(rand.NewSource(1))
			for i := 0; i < 5000; i++ {
				k := r.Intn(400)
				switch r.Intn(3) {
				case 0:
					removed, _ := tab.Delete(k)
					if removed != model[k] {
						t.Fatalf("Delete(%d) = %v", k, removed)
					}
					delete(model, k)
				default:
					added, probed, err := tab.Insert(k)
					if err != nil || added == model[k] || len(probed) == 0 {
						t.Fatalf("Insert(%d) = %v, %v", k, added, err)
					}
					model[k] = true
				}
				if found, _ := tab.Search(k); found != model[k] {
					t.Fatalf("Search(%d) = %v", k, found)
				}
				if tab.LoadFactor() > tab.MaxLoad+0.01 {
					t.Fatalf("load %.2f above %.2f", tab.LoadFactor(), tab.MaxLoad)
				}
			}
			if tab.Len() != len(model) || len(tab.Keys()) != len(model) {
				t.Fatalf("len %d, model %d", tab.Len(), len(model))
			}
		})
	}
}

func TestCuckooNeverLosesAKey(t *testing.T) {
	c := hashing.NewCuckoo(4)
	model := map[int]bool{}
	r := rand.New(rand.NewSource(2))
	for i := 0; i < 3000; i++ {
		k := r.Intn(2000)
		if r.Intn(4) == 0 {
			if c.Delete(k) != model[k] {
				t.Fatalf("Delete(%d)", k)
			}
			delete(model, k)
			continue
		}
		added, _ := c.Insert(k)
		if added == model[k] {
			t.Fatalf("Insert(%d) = %v", k, added)
		}
		model[k] = true
	}
	for k := range model {
		if !c.Contains(k) {
			t.Fatalf("lost key %d", k)
		}
	}
	if c.Len() != len(model) || len(c.Keys()) != len(model) {
		t.Fatalf("len %d keys %d, model %d", c.Len(), len(c.Keys()), len(model))
	}
	if c.Rehashes == 0 {
		t.Error("expected at least one rehash while growing from 4 slots")
	}
}

func TestRobinHood(t *testing.T) {
	rh := hashing.NewRobinHood(4)
	model := map[int]bool{}
	r := rand.New(rand.NewSource(3))
	for i := 0; i < 5000; i++ {
		k := r.Intn(500)
		if r.Intn(3) == 0 {
			ok, _ := rh.Delete(k)
			if ok != model[k] {
				t.Fatalf("Delete(%d) = %v", k, ok)
			}
			delete(model, k)
		} else {
			added, _ := rh.Insert(k)
			if added == model[k] {
				t.Fatalf("Insert(%d) = %v", k, added)
			}
			model[k] = true
		}
		if ok, _ := rh.Search(k); ok != model[k] {
			t.Fatalf("Search(%d) = %v", k, ok)
		}
	}
	got := rh.Keys()
	sort.Ints(got)
	if len(got) != len(model) || rh.MaxDistance() > 20 {
		t.Errorf("%d keys (want %d), max distance %d", len(got), len(model), rh.MaxDistance())
	}
}

func TestConsistentHashingMovesAboutOneNth(t *testing.T) {
	g := hashing.NewRing(100)
	for i := 0; i < 5; i++ {
		g.Add(fmt.Sprintf("node-%d", i))
	}
	var keys []string
	for i := 0; i < 20000; i++ {
		keys = append(keys, fmt.Sprintf("user-%d", i))
	}
	before := g.Assign(keys)
	g.Add("node-5")
	after := g.Assign(keys)
	frac := float64(hashing.Moved(before, after)) / float64(len(keys))
	if math.Abs(frac-1.0/6) > 0.06 {
		t.Errorf("adding a 6th node moved %.3f of the keys, want about 1/6", frac)
	}
	for k, n := range after {
		if before[k] != n && n != "node-5" {
			t.Fatalf("key %s moved to %s, not to the new node", k, n)
		}
	}
	g.Remove("node-5")
	if hashing.Moved(before, g.Assign(keys)) != 0 || len(g.Nodes()) != 5 {
		t.Error("removing the new node should restore every key")
	}
	if _, ok := hashing.NewRing(1).Locate("x"); ok {
		t.Error("an empty ring located a key")
	}
}
