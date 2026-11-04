package filesystems

import (
	"errors"
	"reflect"
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

func TestFreeSpace(t *testing.T) {
	r, err := FreeSpace(16, "alloc a 4\nalloc b 4\nalloc c 4\nfree b\n\nalloc d 2")
	if err != nil {
		t.Fatal(err)
	}
	want := []FreeSpaceStep{
		{"alloc a 4", "1111000000000000", []Extent{{4, 12}}},
		{"alloc b 4", "1111111100000000", []Extent{{8, 8}}},
		{"alloc c 4", "1111111111110000", []Extent{{12, 4}}},
		{"free b", "1111000011110000", []Extent{{4, 4}, {12, 4}}},
		{"alloc d 2", "1111110011110000", []Extent{{6, 2}, {12, 4}}},
	}
	if !reflect.DeepEqual(r.Steps, want) {
		t.Errorf("steps %+v\nwant %+v", r.Steps, want)
	}
	// first fit reads 4, 8, 12 and 6 bits
	if r.BitmapBytes != 2 || r.ListBytes != 16 || r.Scanned != 30 {
		t.Errorf("bitmap %d bytes, list %d bytes, scanned %d", r.BitmapBytes, r.ListBytes, r.Scanned)
	}
	if s := r.String(); s != "bitmap 1111110011110000 (2 bytes), free list [{6 2} {12 4}] (16 bytes)" {
		t.Errorf("String %q", s)
	}
	for _, st := range r.Steps {
		free := 0
		for _, e := range st.List {
			free += e.Length
		}
		if free != strings.Count(st.Bitmap, "0") {
			t.Errorf("%s: the list holds %d free blocks, the bitmap %d", st.Call, free, strings.Count(st.Bitmap, "0"))
		}
	}

	t.Run("an allocation that does not fit changes nothing", func(t *testing.T) {
		r, _ := FreeSpace(8, "alloc a 6\nalloc b 3")
		if r.Steps[1].Bitmap != "11111100" {
			t.Errorf("bitmap %s", r.Steps[1].Bitmap)
		}
	})
	if s := (FreeSpaceResult{}).String(); s != "no steps" {
		t.Errorf("empty String %q", s)
	}
	for name, c := range map[string]struct {
		blocks int
		script string
	}{
		"too few blocks":  {7, "alloc a 1"},
		"too many blocks": {257, "alloc a 1"},
		"zero size":       {8, "alloc a 0"},
		"size past disk":  {8, "alloc a 9"},
		"free unknown":    {8, "free a"},
		"unknown call":    {8, "move a"},
		"alloc no size":   {8, "alloc a"},
		"name taken":      {8, "alloc a 2\nalloc a 2"},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := FreeSpace(c.blocks, c.script); !errors.Is(err, ErrInput) {
				t.Errorf("err %v, want ErrInput", err)
			}
		})
	}
}

func TestPageCache(t *testing.T) {
	t.Run("write-through pays a disk write each time and loses nothing", func(t *testing.T) {
		r, err := PageCache(false, "write a\nwrite a\nread a\nread b\ncrash\nread a")
		if err != nil {
			t.Fatal(err)
		}
		if r.DiskWrites != 2 || r.Hits != 1 || r.DiskReads != 2 || r.Lost != 0 {
			t.Errorf("%+v", r)
		}
		if !strings.Contains(r.Log[4], "every write was already on disk") {
			t.Errorf("crash log %q", r.Log[4])
		}
	})
	t.Run("write-back batches writes and loses the dirty pages", func(t *testing.T) {
		r, err := PageCache(true, "write a\nwrite a\nwrite b\nsync\nwrite c\ncrash\nread c")
		if err != nil {
			t.Fatal(err)
		}
		if r.DiskWrites != 2 || r.Lost != 1 || r.DiskReads != 1 || r.Hits != 0 {
			t.Errorf("%+v", r)
		}
		log := strings.Join(r.Log, "\n")
		if !strings.Contains(log, "write 2 dirty pages to disk [a b]") || !strings.Contains(log, "dirty pages [c] never reached the disk") {
			t.Errorf("log:\n%s", log)
		}
		if s := r.String(); s != "disk writes 2, hits 0, lost 1" {
			t.Errorf("String %q", s)
		}
	})
	t.Run("a sync before the crash loses nothing", func(t *testing.T) {
		r, _ := PageCache(true, "write a\nsync\ncrash")
		if r.Lost != 0 || r.DiskWrites != 1 {
			t.Errorf("%+v", r)
		}
	})
	for _, script := range []string{"write", "read", "flush a"} {
		if _, err := PageCache(true, script); !errors.Is(err, ErrInput) {
			t.Errorf("%q: err %v", script, err)
		}
	}
}

func TestResolve(t *testing.T) {
	setup := "mkdir /a\nmkdir /a/b\ncreate /a/b/f hi there\nsymlink /a/b /l"
	r, err := Resolve(setup, "/l/f")
	if err != nil {
		t.Fatal(err)
	}
	log := strings.Join(r.Log, "\n")
	for _, s := range []string{
		"  start at the root directory, inode 2",
		`  look up "l" in directory inode 2: inode 6 (symlink)`,
		`  inode 6 is a symbolic link to "/a/b": resolve that instead`,
		`  look up "b" in directory inode 3: inode 4 (dir)`,
		`  look up "f" in directory inode 4: inode 5 (file)`,
		"  resolved to inode 5 (file)",
	} {
		if !strings.Contains(log, s) {
			t.Errorf("log lacks %q:\n%s", s, log)
		}
	}
	links := map[int]int{}
	for _, n := range r.Inodes {
		links[n.Number] = n.Links
	}
	// the root has ".", its own "..", and /a's ".."; /a has "a", "." and /a/b's ".."
	if links[2] != 3 || links[3] != 3 || links[4] != 2 || links[5] != 1 {
		t.Errorf("link counts %v", links)
	}
	for path, want := range map[string]string{
		"/l/g":     "ENOENT",
		"/a/b/f/x": "ENOTDIR",
	} {
		r, err := Resolve(setup, path)
		if err != nil || !strings.Contains(strings.Join(r.Log, "\n"), want) {
			t.Errorf("%s: err %v, log %v, want %s", path, err, r.Log, want)
		}
	}
	loop, _ := Resolve("symlink /x /y\nsymlink /y /x", "/x")
	if !strings.Contains(strings.Join(loop.Log, "\n"), "ELOOP") {
		t.Errorf("a symlink loop: %v", loop.Log)
	}
	if _, err := Resolve("mkdir /a\nmkdir /a", "/a"); !errors.Is(err, ErrInput) || !strings.Contains(err.Error(), "line 2") {
		t.Errorf("a bad setup: %v", err)
	}
}

func TestNamespaceErrors(t *testing.T) {
	t.Run("removing the last hard link frees the inode", func(t *testing.T) {
		r, err := Namespace("create /a x\nlink /a /b\nrm /a\nrm /b\nmkdir /d\ncat /d")
		if err != nil {
			t.Fatal(err)
		}
		for _, n := range r.Inodes {
			if n.Kind == "file" {
				t.Errorf("inode %d survived with no names", n.Number)
			}
		}
		log := strings.Join(r.Log, "\n")
		if !strings.Contains(log, "inode and its blocks are freed") || !strings.Contains(log, "  is a directory") {
			t.Errorf("log:\n%s", log)
		}
	})
	for name, script := range map[string]string{
		"mkdir twice":             "mkdir /a\nmkdir /a",
		"hard link to a dir":      "mkdir /a\nlink /a /b",
		"hard link onto a name":   "create /a x\ncreate /b y\nlink /a /b",
		"hard link to nothing":    "link /a /b",
		"rm a directory":          "mkdir /a\nrm /a",
		"rm nothing":              "rm /a",
		"no name to create":       "mkdir /",
		"under a missing dir":     "create /a/b x",
		"under a file":            "create /a x\ncreate /a/b y",
		"unknown command":         "touch /a",
		"rm under a missing dir":  "rm /q/a",
		"link into a missing dir": "create /a x\nlink /a /q/b",
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := Namespace(script); !errors.Is(err, ErrInput) {
				t.Errorf("err %v, want ErrInput", err)
			}
		})
	}
	t.Run("a refused link leaves the counts alone", func(t *testing.T) {
		r, _ := Namespace("create /a x\ncreate /b y\nlink /a /b")
		for _, n := range r.Inodes {
			if n.Kind == "file" && n.Links != 1 {
				t.Errorf("inode %d has %d links", n.Number, n.Links)
			}
		}
	})
}

func TestAllocateIndexedAndErrors(t *testing.T) {
	r, err := Allocate(Indexed, 8, "create a 3\ncreate b 8\ndelete a")
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"a*", "a", "a", "a", "", "", "", ""}; !reflect.DeepEqual(r.Maps[0], want) {
		t.Errorf("map %q, want %q", r.Maps[0], want)
	}
	if r.Failed != 1 || strings.Contains(r.Log[1], "fragmentation") {
		t.Errorf("b needs 9 blocks of 8: failed %d, log %q", r.Failed, r.Log[1])
	}
	for _, o := range r.Maps[2] {
		if o != "" {
			t.Errorf("after deleting a the disk should be empty, index block included: %q", r.Maps[2])
		}
	}
	for name, c := range map[string]struct {
		m      Method
		blocks int
		script string
	}{
		"unknown method": {"buddy", 8, "create a 1"},
		"too few blocks": {Linked, 3, "create a 1"},
		"duplicate":      {Linked, 8, "create a 1\ncreate a 1"},
		"star in name":   {Indexed, 8, "create a* 1"},
		"zero size":      {Contiguous, 8, "create a 0"},
		"delete nothing": {Contiguous, 8, "delete a"},
		"unknown call":   {Contiguous, 8, "grow a 1"},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := Allocate(c.m, c.blocks, c.script); !errors.Is(err, ErrInput) {
				t.Errorf("err %v, want ErrInput", err)
			}
		})
	}
}

func TestLFSAndRAIDErrors(t *testing.T) {
	if _, err := LFS(3, 2, []int{1, 2, 3, 4, 5, 6}); !errors.Is(err, ErrInput) {
		t.Errorf("six live blocks in 3 segments of 2 keep no spare: %v", err)
	}
	for _, c := range []struct {
		segments, size int
		writes         []int
	}{{2, 4, []int{1}}, {3, 1, []int{1}}, {3, 4, nil}, {3, 4, []int{1000}}, {3, 4, []int{-1}}} {
		if _, err := LFS(c.segments, c.size, c.writes); !errors.Is(err, ErrInput) {
			t.Errorf("LFS%v: err %v", c, err)
		}
	}
	r, err := LFS(4, 2, []int{1, 1, 1, 1, 1, 1, 1, 1})
	if err != nil || r.Cleaned == 0 || r.Logical != 8 {
		t.Errorf("one hot block: %+v %v", r, err)
	}

	for _, c := range []struct {
		level, disks int
		data, failed []int
	}{
		{2, 4, []int{1}, nil},
		{5, 2, []int{1}, nil},
		{0, 9, []int{1}, nil},
		{0, 2, nil, nil},
		{0, 2, []int{256}, nil},
		{1, 2, []int{1}, []int{0, 0}},
		{1, 2, []int{1}, []int{2}},
	} {
		if _, err := RAID(c.level, c.disks, c.data, c.failed); !errors.Is(err, ErrInput) {
			t.Errorf("RAID%v: err %v", c, err)
		}
	}
	// zeros exercise the Galois-field arithmetic's zero cases
	data := []int{0, 0, 7, 0, 255, 0}
	for a := 0; a < 5; a++ {
		for b := a + 1; b < 5; b++ {
			if r, _ := RAID(6, 5, data, []int{a, b}); !r.Intact || !reflect.DeepEqual(r.Recovered, data) {
				t.Errorf("RAID 6 failing %d and %d: %v", a, b, r.Recovered)
			}
		}
	}
}
