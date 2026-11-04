package filesystems

import (
	"fmt"
	"sort"
	"strings"
)

// CacheResult is a PageCache run: the log and the counts of disk writes,
// cache hits and writes lost in a crash.
type CacheResult struct {
	Log        []string `json:"log"`
	DiskWrites int      `json:"diskWrites"`
	DiskReads  int      `json:"diskReads"`
	Hits       int      `json:"hits"`
	Lost       int      `json:"lost"`
}

// PageCache runs "write PAGE", "read PAGE", "sync" and "crash" through an
// operating system's page cache. Write-through sends every write to disk at
// once: safe, but every write pays for the disk. Write-back only marks the
// cached page dirty and writes it at the next sync (or when evicted, which
// this cache never needs to): repeated writes to a page cost one disk write,
// but a crash loses every dirty page. That is the trade fsync exists for.
//
// Time: O(n log p) for n calls and p pages. Space: O(p).
func PageCache(writeBack bool, script string) (CacheResult, error) {
	cached, dirty := map[string]bool{}, map[string]bool{}
	var res CacheResult
	calls, nums := lines(script)
	for k, f := range calls {
		page := ""
		if len(f) > 1 {
			page = f[1]
		}
		switch f[0] {
		case "write":
			if page == "" {
				return res, inputErr("line %d: write PAGE", nums[k])
			}
			cached[page] = true
			if writeBack {
				dirty[page] = true
				res.Log = append(res.Log, fmt.Sprintf("write %s: in the cache, marked dirty (no disk write yet)", page))
			} else {
				res.DiskWrites++
				res.Log = append(res.Log, fmt.Sprintf("write %s: in the cache and on disk", page))
			}
		case "read":
			if page == "" {
				return res, inputErr("line %d: read PAGE", nums[k])
			}
			if cached[page] {
				res.Hits++
				res.Log = append(res.Log, fmt.Sprintf("read %s: from the cache", page))
			} else {
				res.DiskReads++
				cached[page] = true
				res.Log = append(res.Log, fmt.Sprintf("read %s: miss, read from disk into the cache", page))
			}
		case "sync":
			names := keys(dirty)
			res.DiskWrites += len(names)
			res.Log = append(res.Log, fmt.Sprintf("sync: write %d dirty pages to disk %v", len(names), names))
			dirty = map[string]bool{}
		case "crash":
			names := keys(dirty)
			res.Lost += len(names)
			if len(names) > 0 {
				res.Log = append(res.Log, fmt.Sprintf("crash: the cache is gone; dirty pages %v never reached the disk", names))
			} else {
				res.Log = append(res.Log, "crash: the cache is gone, but every write was already on disk")
			}
			cached, dirty = map[string]bool{}, map[string]bool{}
		default:
			return res, inputErr("line %d: write, read, sync or crash", nums[k])
		}
	}
	return res, nil
}

func keys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// String sums a run up.
func (r CacheResult) String() string {
	return strings.Join([]string{fmt.Sprintf("disk writes %d", r.DiskWrites), fmt.Sprintf("hits %d", r.Hits), fmt.Sprintf("lost %d", r.Lost)}, ", ")
}
