package core

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/os/filesystems"
)

// filesystemsEntries is every entry of the filesystems category.
func filesystemsEntries() []Entry {
	return simEntries(fsSims())
}

func fsSimAllocation(m filesystems.Method, id, name, summary string, ex string) sim {
	return sim{
		id: id, name: name, file: "os/filesystems/allocation.go", fn: "Allocate", summary: summary,
		time: "O(n * blocks)", room: "O(n * blocks)", input: Text,
		example: Input{Text: ex, Options: map[string]int{"blocks": 12}},
		options: []Option{{"blocks", "disk blocks", 12, 4, 64}},
		run: func(in Input) (Output, error) {
			r, err := filesystems.Allocate(m, in.opt("blocks", 12), in.Text)
			if err != nil {
				return Output{}, err
			}
			ids := map[string]int{}
			n := in.opt("blocks", 12)
			t := &TableOutput{Cols: n, Blank: ptr(0)}
			calls := linesOf(in.Text)
			for i, mp := range r.Maps {
				for _, owner := range mp {
					who := strings.TrimSuffix(owner, "*")
					if who != "" && ids[who] == 0 {
						ids[who] = len(ids) + 1
					}
					t.Cells = append(t.Cells, ids[who])
				}
				t.Rows++
				t.RowLabels = append(t.RowLabels, calls[i])
			}
			for b := 0; b < n; b++ {
				t.ColLabels = append(t.ColLabels, strconv.Itoa(b))
			}
			lines := r.Log
			if m == filesystems.Linked {
				lines = append(lines, fmt.Sprintf("file allocation table (next block, -1 end, -2 free): %v", r.FAT))
			}
			return Output{Lines: lines, Table: t, Summary: map[string]float64{"failed": float64(r.Failed)}}, nil
		},
		args: func(in Input) []any { return []any{string(m), in.opt("blocks", 12), in.Text} },
	}
}

const fsSimFrag = "create a 3\ncreate b 2\ncreate c 4\ndelete a\ndelete c\ncreate d 5\ncreate e 2"

var fsSimTable = []sim{
	fsSimAllocation(filesystems.Contiguous, "contiguous-allocation", "Contiguous Allocation", "Each file in one run of blocks: fast to read, but deletes leave holes too small to reuse.", fsSimFrag),
	fsSimAllocation(filesystems.Linked, "linked-allocation", "Linked Allocation (FAT)", "Each block points to the next in a file allocation table: no fragmentation, but reaching block k takes k hops.", fsSimFrag),
	fsSimAllocation(filesystems.Indexed, "indexed-allocation", "Indexed Allocation", "Each file has an index block listing its blocks: any block in one lookup, for one extra block per file.", fsSimFrag),
	{
		id: "inode-pointers", name: "Inode Pointers and Maximum File Size", file: "os/filesystems/inode.go", fn: "Inode",
		summary: "Direct, single, double and triple indirect pointers: how big a file can get, and how many reads reach a byte.",
		time:    "O(1)", room: "O(1)", input: Simulation,
		example: Input{Options: map[string]int{"block": 4096, "pointer": 4, "direct": 12, "offset": 5000000}},
		options: []Option{{"block", "block size in bytes", 4096, 64, 65536}, {"pointer", "pointer size in bytes", 4, 4, 8}, {"direct", "direct pointers", 12, 0, 32}, {"offset", "byte offset to find", 5000000, 0, 2_000_000_000}},
		run: func(in Input) (Output, error) {
			l, err := filesystems.Inode(in.opt("block", 4096), in.opt("pointer", 4), in.opt("direct", 12), int64(in.opt("offset", 5000000)))
			if err != nil {
				return Output{}, err
			}
			t := &TableOutput{Cols: 2, ColLabels: []string{"blocks", "MiB"}}
			lines := []string{fmt.Sprintf("%d pointers per indirect block; largest file %d bytes (%.1f GiB)", l.PointersPerBlock, l.MaxBytes, float64(l.MaxBytes)/(1<<30))}
			for _, lv := range l.Levels {
				t.Rows++
				t.Cells = append(t.Cells, int(lv.Blocks), int(lv.Bytes>>20))
				t.RowLabels = append(t.RowLabels, lv.Name)
			}
			return Output{Lines: append(lines, l.Lookup...), Table: t, Number: ptr(int(l.MaxBytes)), Summary: map[string]float64{"disk reads": float64(l.DiskReads)}}, nil
		},
		args: func(in Input) []any {
			return []any{in.opt("block", 4096), in.opt("pointer", 4), in.opt("direct", 12), in.opt("offset", 5000000)}
		},
	},
	{
		id: "hard-and-symbolic-links", name: "Hard and Symbolic Links", file: "os/filesystems/namespace.go", fn: "Namespace",
		summary: "Names, inodes and link counts: a hard link keeps a file alive, a symbolic link only remembers its path.",
		time:    "O(n * d)", room: "O(inodes)", input: Text,
		example: Input{Text: "mkdir /docs\ncreate /docs/notes.txt buy milk\nlink /docs/notes.txt /backup\nsymlink /docs/notes.txt /shortcut\ncat /shortcut\nrm /docs/notes.txt\ncat /backup\ncat /shortcut"},
		run: func(in Input) (Output, error) {
			r, err := filesystems.Namespace(in.Text)
			if err != nil {
				return Output{}, err
			}
			return Output{Lines: r.Log, Table: inodeTable(r.Inodes)}, nil
		},
		args: func(in Input) []any { return []any{in.Text} },
	},
	{
		id: "path-resolution", name: "Path Resolution", file: "os/filesystems/namespace.go", fn: "Resolve",
		summary: "Follow a path component by component through directory inodes and symbolic links, as the kernel does on every open.",
		time:    "O(n * d)", room: "O(inodes)", input: TwoTexts,
		example: Input{Text: "mkdir /home\nmkdir /home/ada\nmkdir /home/ada/code\ncreate /home/ada/code/main.go package main\nsymlink /home/ada/code /src", Other: "/src/main.go"},
		run: func(in Input) (Output, error) {
			r, err := filesystems.Resolve(in.Text, in.Other)
			if err != nil {
				return Output{}, err
			}
			return Output{Lines: r.Log, Table: inodeTable(r.Inodes)}, nil
		},
		args: func(in Input) []any { return []any{in.Text, in.Other} },
	},
	{
		id: "free-space-management", name: "Free Space: Bitmap and Free List", file: "os/filesystems/freespace.go", fn: "FreeSpace",
		summary: "Track free blocks with a bitmap (fixed size, easy to scan) or a list of free runs (small until space fragments).",
		time:    "O(n * blocks)", room: "O(n * blocks)", input: Text,
		example: Input{Text: "alloc a 4\nalloc b 3\nalloc c 5\nfree b\nalloc d 2\nfree a", Options: map[string]int{"blocks": 16}},
		options: []Option{{"blocks", "disk blocks", 16, 8, 256}},
		run: func(in Input) (Output, error) {
			r, err := filesystems.FreeSpace(in.opt("blocks", 16), in.Text)
			if err != nil {
				return Output{}, err
			}
			n := in.opt("blocks", 16)
			t := &TableOutput{Cols: n}
			var lines []string
			for _, s := range r.Steps {
				t.Rows++
				for _, c := range s.Bitmap {
					t.Cells = append(t.Cells, int(c-'0'))
				}
				t.RowLabels = append(t.RowLabels, s.Call)
				lines = append(lines, fmt.Sprintf("%s: bitmap %s, free list %v", s.Call, s.Bitmap, s.List))
			}
			return Output{Lines: lines, Table: t, Summary: map[string]float64{"bitmap bytes": float64(r.BitmapBytes), "free list bytes": float64(r.ListBytes), "bits scanned": float64(r.Scanned)}}, nil
		},
		args: func(in Input) []any { return []any{in.opt("blocks", 16), in.Text} },
	},
	{
		id: "journaling", name: "Journaling and Crash Recovery", file: "os/filesystems/journal.go", fn: "Journaling",
		summary: "Crash in the middle of appending a block, with and without a journal, and see what recovery finds.",
		time:    "O(1)", room: "O(1)", input: Simulation,
		example: Input{Options: map[string]int{"crash": 2, "journal": 0}},
		options: []Option{{"crash", "crash after this many writes", 2, 0, 10}, {"journal", "journaled (1) or not (0)", 0, 0, 1}},
		run: func(in Input) (Output, error) {
			r, err := filesystems.Journaling(in.opt("crash", 2), on(in, "journal", 0))
			if err != nil {
				return Output{}, err
			}
			lines := append(append(r.Writes, "recovery:"), r.Recovery...)
			return Output{Lines: lines, Bool: ptr(r.Consistent)}, nil
		},
		args: func(in Input) []any { return []any{in.opt("crash", 2), on(in, "journal", 0)} },
	},
	{
		id: "page-cache", name: "Page Cache: Write-Through and Write-Back", file: "os/filesystems/pagecache.go", fn: "PageCache",
		summary: "Write-through pays the disk on every write; write-back batches writes and loses the dirty ones in a crash.",
		time:    "O(n log p)", room: "O(p)", input: Text,
		example: Input{Text: "write a\nwrite a\nwrite b\nread a\nread c\nsync\nwrite c\nwrite a\ncrash\nread a", Options: map[string]int{"back": 1}},
		options: []Option{{"back", "write-back (1) or write-through (0)", 1, 0, 1}},
		run: func(in Input) (Output, error) {
			r, err := filesystems.PageCache(on(in, "back", 1), in.Text)
			if err != nil {
				return Output{}, err
			}
			return Output{Lines: r.Log, Summary: map[string]float64{"disk writes": float64(r.DiskWrites), "disk reads": float64(r.DiskReads), "cache hits": float64(r.Hits), "writes lost": float64(r.Lost)}}, nil
		},
		args: func(in Input) []any { return []any{on(in, "back", 1), in.Text} },
	},
	{
		id: "raid", name: "RAID 0, 1, 5 and 6", file: "os/filesystems/raid.go", fn: "RAID",
		summary: "Stripe, mirror, or add parity across disks; fail one or two and rebuild the data with XOR and Reed-Solomon.",
		time:    "O(n * disks)", room: "O(n)", input: Ints,
		example: Input{Values: []int{10, 20, 30, 40, 50, 60, 70, 80, 90}, Options: map[string]int{"level": 5, "disks": 4, "fail": 1, "fail2": -1}},
		options: []Option{{"level", "RAID level (0, 1, 5, 6)", 5, 0, 6}, {"disks", "disks", 4, 2, 8}, {"fail", "a failed disk (-1 none)", 1, -1, 7}, {"fail2", "a second failed disk (-1 none)", -1, -1, 7}},
		run: func(in Input) (Output, error) {
			r, err := filesystems.RAID(in.opt("level", 5), in.opt("disks", 4), in.Values, failedDisks(in))
			if err != nil {
				return Output{}, err
			}
			var t *TableOutput
			if len(r.Disks) > 0 && len(r.Disks[0]) > 0 {
				t = &TableOutput{Rows: len(r.Disks), Cols: len(r.Disks[0])}
				down := map[int]bool{}
				for _, d := range r.Failed {
					down[d] = true
				}
				for d, blocks := range r.Disks {
					t.Cells = append(t.Cells, blocks...)
					label := fmt.Sprintf("disk %d", d)
					if down[d] {
						label += " (failed)"
					}
					t.RowLabels = append(t.RowLabels, label)
				}
				for s := range r.Disks[0] {
					var roles []string
					for d := range r.Disks {
						roles = append(roles, r.Roles[d][s])
					}
					t.ColLabels = append(t.ColLabels, "stripe "+strconv.Itoa(s))
					r.Log = append(r.Log, fmt.Sprintf("stripe %d: %s", s, strings.Join(roles, " | ")))
				}
			}
			return Output{Lines: r.Log, Table: t, Values: r.Recovered, Bool: ptr(r.Intact), Summary: map[string]float64{"usable fraction": r.Usable}}, nil
		},
		args: func(in Input) []any {
			return []any{in.opt("level", 5), in.opt("disks", 4), in.Values, rawInts(failedDisks(in))}
		},
	},
	{
		id: "log-structured-fs", name: "Log-Structured File System Cleaning", file: "os/filesystems/lfs.go", fn: "LFS",
		summary: "Every write appends to a log; overwrites leave dead blocks, and the cleaner copies live ones out of the emptiest segment.",
		time:    "O(n * segments * size)", room: "O(segments * size)", input: Ints,
		example: Input{Values: []int{1, 2, 3, 4, 1, 2, 5, 1, 2, 6, 1, 2, 7, 1, 2, 8}, Options: map[string]int{"segments": 4, "size": 4}},
		options: []Option{{"segments", "segments", 4, 3, 16}, {"size", "blocks per segment", 4, 2, 16}},
		run: func(in Input) (Output, error) {
			r, err := filesystems.LFS(in.opt("segments", 4), in.opt("size", 4), in.Values)
			if err != nil {
				return Output{}, err
			}
			t := &TableOutput{Blank: ptr(-1)}
			for s, slots := range r.Segments {
				t.Rows++
				t.Cols = len(slots)
				t.Cells = append(t.Cells, slots...)
				t.RowLabels = append(t.RowLabels, "segment "+strconv.Itoa(s))
			}
			return Output{Lines: r.Log, Table: t, Summary: map[string]float64{"logical writes": float64(r.Logical), "disk writes": float64(r.Physical), "cleanings": float64(r.Cleaned), "write amplification": float64(r.Physical) / float64(r.Logical)}}, nil
		},
		args: func(in Input) []any { return []any{in.opt("segments", 4), in.opt("size", 4), in.Values} },
	},
}

func fsSims() []sim {
	return fsSimTable
}

func failedDisks(in Input) []int {
	var out []int
	for _, k := range []string{"fail", "fail2"} {
		if d := in.opt(k, -1); d >= 0 {
			out = append(out, d)
		}
	}
	return out
}

func inodeTable(inodes []filesystems.INode) *TableOutput {
	t := &TableOutput{Cols: 1, ColLabels: []string{"links"}}
	for _, n := range inodes {
		t.Rows++
		t.Cells = append(t.Cells, n.Links)
		label := fmt.Sprintf("inode %d (%s)", n.Number, n.Kind)
		if n.Kind == "dir" {
			var names []string
			for name := range n.Entries {
				if name != "." && name != ".." {
					names = append(names, name)
				}
			}
			sort.Strings(names)
			label += ": " + strings.Join(names, ", ")
		} else if n.Data != "" {
			label += fmt.Sprintf(": %q", n.Data)
		}
		t.RowLabels = append(t.RowLabels, label)
	}
	return t
}
