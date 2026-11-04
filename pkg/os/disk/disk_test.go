package disk_test

import (
	"sort"
	"testing"

	"github.com/braswelljr/alGOrithm/pkg/os/disk"
)

// The textbook queue (Silberschatz): head at 53, tracks 0..199, moving up.
var queue = []int{98, 183, 37, 122, 14, 124, 65, 67}

func TestTextbookMovement(t *testing.T) {
	for name, c := range map[string]struct {
		run  func() (disk.Result, error)
		want int
	}{
		"fcfs":      {func() (disk.Result, error) { return disk.FCFS(queue, 53, 200) }, 640},
		"sstf":      {func() (disk.Result, error) { return disk.SSTF(queue, 53, 200) }, 236},
		"scan":      {func() (disk.Result, error) { return disk.SCAN(queue, 53, 200, false) }, 331},
		"cscan":     {func() (disk.Result, error) { return disk.CSCAN(queue, 53, 200, false) }, 382},
		"look":      {func() (disk.Result, error) { return disk.LOOK(queue, 53, 200, false) }, 299},
		"clook":     {func() (disk.Result, error) { return disk.CLOOK(queue, 53, 200, false) }, 322},
		"scan down": {func() (disk.Result, error) { return disk.SCAN(queue, 53, 200, true) }, 236},
	} {
		r, err := c.run()
		if err != nil {
			t.Fatal(err)
		}
		if r.Movement != c.want {
			t.Errorf("%s moved %d, want %d (path %v)", name, r.Movement, c.want, r.Path)
		}
		served := append([]int(nil), r.Order...)
		sort.Ints(served)
		want := append([]int(nil), queue...)
		sort.Ints(want)
		for i := range want {
			if served[i] != want[i] {
				t.Fatalf("%s served %v", name, r.Order)
			}
		}
	}
	if _, err := disk.SCAN([]int{250}, 53, 200, false); err == nil {
		t.Error("a track past the end was accepted")
	}
}
