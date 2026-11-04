package filesystems

import (
	"strings"
	"testing"
)

func TestInodeMaxFileSize(t *testing.T) {
	l, err := Inode(4096, 4, 12, 0)
	if err != nil {
		t.Fatal(err)
	}
	// (12 + 1024 + 1024^2 + 1024^3) * 4096 bytes: the classic ext2 figure.
	want := int64(12+1024+1024*1024+1024*1024*1024) * 4096
	if l.MaxBytes != want {
		t.Errorf("max %d, want %d", l.MaxBytes, want)
	}
	if l.DiskReads != 2 {
		t.Errorf("byte 0 takes %d reads, want 2 (inode, data)", l.DiskReads)
	}
	if l, _ := Inode(4096, 4, 12, 12*4096+1024*4096); l.DiskReads != 4 {
		t.Errorf("the first double-indirect byte takes %d reads, want 4", l.DiskReads)
	}
}

func TestRAIDRebuilds(t *testing.T) {
	data := []int{10, 20, 30, 40, 50, 60, 70, 80, 90}
	for _, c := range []struct {
		level, disks int
		failed       []int
		intact       bool
	}{
		{0, 3, nil, true}, {0, 3, []int{1}, false},
		{1, 2, []int{0}, true},
		{5, 4, []int{2}, true}, {5, 4, []int{1, 2}, false},
		{6, 5, []int{0, 3}, true}, {6, 5, []int{1, 2}, true}, {6, 6, []int{4, 5}, true},
	} {
		r, err := RAID(c.level, c.disks, data, c.failed)
		if err != nil {
			t.Fatal(err)
		}
		if r.Intact != c.intact {
			t.Errorf("RAID %d on %d disks, %v failed: intact %v, recovered %v", c.level, c.disks, c.failed, r.Intact, r.Recovered)
		}
	}
	// Every pair of failures on RAID 6.
	for a := 0; a < 5; a++ {
		for b := a + 1; b < 5; b++ {
			if r, _ := RAID(6, 5, data, []int{a, b}); !r.Intact {
				t.Errorf("RAID 6 failing %d and %d: %v", a, b, r.Recovered)
			}
		}
	}
}

func TestAllocationFragmentation(t *testing.T) {
	script := "create a 3\ncreate b 2\ncreate c 3\ndelete a\ndelete c\ncreate d 5"
	cont, _ := Allocate(Contiguous, 9, script)
	linked, _ := Allocate(Linked, 9, script)
	if cont.Failed != 1 || linked.Failed != 0 {
		t.Errorf("contiguous failed %d, linked %d", cont.Failed, linked.Failed)
	}
	if d := linked.Files["d"]; len(d) != 5 || linked.FAT[d[4]] != -1 || linked.FAT[d[0]] != d[1] {
		t.Errorf("FAT chain %v %v", d, linked.FAT)
	}
}

func TestNamespaceLinks(t *testing.T) {
	r, err := Namespace("mkdir /docs\ncreate /docs/a hello\nlink /docs/a /b\nsymlink /docs/a /s\nrm /docs/a\ncat /b\ncat /s")
	if err != nil {
		t.Fatal(err)
	}
	log := strings.Join(r.Log, "\n")
	if !strings.Contains(log, `"hello"`) || !strings.Contains(log, "ENOENT") {
		t.Errorf("the hard link should still read and the symlink dangle:\n%s", log)
	}
}

func TestJournaling(t *testing.T) {
	for crash := 0; crash <= 8; crash++ {
		if r, _ := Journaling(crash, true); !r.Consistent {
			t.Errorf("journaled crash after %d: inconsistent", crash)
		}
	}
	if r, _ := Journaling(2, false); r.Consistent {
		t.Error("no journal, crash between inode and bitmap: should be inconsistent")
	}
}

func TestLFS(t *testing.T) {
	writes := []int{1, 2, 3, 4, 1, 2, 1, 2, 1, 2, 5, 6, 1, 2, 7}
	r, err := LFS(4, 4, writes)
	if err != nil {
		t.Fatal(err)
	}
	live := map[int]int{}
	for _, s := range r.Segments {
		for _, b := range s {
			if b >= 0 {
				live[b]++
			}
		}
	}
	for b := 1; b <= 7; b++ {
		if live[b] != 1 {
			t.Errorf("block %d has %d live copies", b, live[b])
		}
	}
	if r.Physical < r.Logical {
		t.Error("cleaning only adds writes")
	}
}
