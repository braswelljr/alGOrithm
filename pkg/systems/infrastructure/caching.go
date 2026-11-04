package infrastructure

import (
	"fmt"
	"strings"
)

// CachePolicy is how an application keeps a cache in front of a database.
type CachePolicy string

// The three policies CacheStrategy knows.
const (
	CacheAside   CachePolicy = "cache-aside"   // read: cache, else DB and fill; write: DB, then delete the cache entry
	WriteThrough CachePolicy = "write-through" // write: cache and DB together
	WriteBehind  CachePolicy = "write-behind"  // write: cache now, DB at the next flush
)

// CacheRunResult is a CacheStrategy run.
type CacheRunResult struct {
	Log      []string `json:"log"`
	Hits     int      `json:"hits"`
	DBReads  int      `json:"dbReads"`
	DBWrites int      `json:"dbWrites"`
	Lost     int      `json:"lost"` // writes lost in a crash
}

// CacheStrategy runs "read KEY", "write KEY VALUE", "flush" and "crash"
// (the cache server dies) against a cache in front of a database.
// Cache-aside is the common default: simple, and the database is always
// right, but the first read after a write misses. Write-through keeps the
// cache warm at the cost of a database write every time. Write-behind
// batches writes, so a key written ten times costs one database write, but
// a crash before the flush loses them.
//
// Time: O(n). Space: O(keys).
func CacheStrategy(policy CachePolicy, script string) (CacheRunResult, error) {
	if policy != CacheAside && policy != WriteThrough && policy != WriteBehind {
		return CacheRunResult{}, inputErr("unknown policy %q", policy)
	}
	cache, db, dirty := map[string]string{}, map[string]string{}, map[string]bool{}
	var res CacheRunResult
	for n, line := range strings.Split(script, "\n") {
		f := strings.Fields(line)
		if len(f) == 0 {
			continue
		}
		switch {
		case f[0] == "read" && len(f) == 2:
			if v, ok := cache[f[1]]; ok {
				res.Hits++
				res.Log = append(res.Log, fmt.Sprintf("read %s: cache hit, %q", f[1], v))
				continue
			}
			res.DBReads++
			v, ok := db[f[1]]
			if !ok {
				res.Log = append(res.Log, fmt.Sprintf("read %s: miss, and not in the database either", f[1]))
				continue
			}
			cache[f[1]] = v
			res.Log = append(res.Log, fmt.Sprintf("read %s: miss, read %q from the database and cache it", f[1], v))
		case f[0] == "write" && len(f) >= 3:
			k, v := f[1], strings.Join(f[2:], " ")
			switch policy {
			case CacheAside:
				db[k] = v
				delete(cache, k)
				res.DBWrites++
				res.Log = append(res.Log, fmt.Sprintf("write %s: to the database, then delete the cache entry", k))
			case WriteThrough:
				db[k], cache[k] = v, v
				res.DBWrites++
				res.Log = append(res.Log, fmt.Sprintf("write %s: to the cache and the database", k))
			default:
				cache[k] = v
				dirty[k] = true
				res.Log = append(res.Log, fmt.Sprintf("write %s: to the cache only, marked dirty", k))
			}
		case f[0] == "flush":
			for k := range dirty {
				db[k] = cache[k]
				res.DBWrites++
			}
			res.Log = append(res.Log, fmt.Sprintf("flush: %d dirty keys written to the database", len(dirty)))
			dirty = map[string]bool{}
		case f[0] == "crash":
			res.Lost += len(dirty)
			res.Log = append(res.Log, fmt.Sprintf("crash: the cache is empty again; %d unflushed writes are gone", len(dirty)))
			cache, dirty = map[string]string{}, map[string]bool{}
		default:
			return res, inputErr("line %d: read KEY, write KEY VALUE, flush or crash", n+1)
		}
	}
	return res, nil
}
