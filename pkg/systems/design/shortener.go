package design

import (
	"fmt"
	"math/rand"
	"strings"
)

const base62Digits = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// Base62 writes n in base 62 (0-9, a-z, A-Z), the alphabet of short links:
// seven characters cover 62^7, about 3.5 trillion links. It is an error for
// a negative n.
//
// Time: O(log n). Space: O(log n).
func Base62(n int64) (string, error) {
	if err := check(n >= 0, "base 62 of a negative number"); err != nil {
		return "", err
	}
	if n == 0 {
		return "0", nil
	}
	var b []byte
	for n > 0 {
		b = append(b, base62Digits[n%62])
		n /= 62
	}
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}
	return string(b), nil
}

// FromBase62 reads a base-62 string back into a number.
//
// Time: O(len(s)). Space: O(1).
func FromBase62(s string) (int64, error) {
	if err := check(s != "", "an empty code"); err != nil {
		return 0, err
	}
	var n int64
	for _, c := range s {
		d := strings.IndexRune(base62Digits, c)
		if err := check(d >= 0, "%q is not a base-62 digit", c); err != nil {
			return 0, err
		}
		n = n*62 + int64(d)
	}
	return n, nil
}

// ShortenerConfig is a URL shortener's workload.
type ShortenerConfig struct {
	URLs      int   // links created
	Reads     int   // redirects served
	CacheSize int   // links the cache holds
	CodeLen   int   // characters of a hashed code
	Hashed    bool  // hash the long URL instead of counting
	Seed      int64 // which links are read
}

// URLShortener models a URL shortener (create a short code for a long URL,
// redirect from the code). Codes come from a counter written in base 62,
// which never collides but needs a counter shared by every server (a ticket
// server or a range per server), or from the first CodeLen base-62 digits of
// a hash of the URL, which needs no coordination but collides more as links
// grow and must retry. Reads outnumber writes and follow a few popular links,
// so an LRU cache in front of the database serves most redirects. It leaves
// out expiry, analytics and custom aliases.
//
// Time: O(URLs + Reads). Space: O(URLs).
func URLShortener(c ShortenerConfig) (Report, error) {
	if err := check(c.URLs >= 1 && c.URLs <= 1_000_000 && c.Reads >= 0 && c.Reads <= 1_000_000 && c.CacheSize >= 0 && c.CodeLen >= 1 && c.CodeLen <= 11,
		"1 to 1,000,000 links and reads, code length 1 to 11"); err != nil {
		return Report{}, err
	}
	a := arch([]Component{
		{ID: "client", Kind: Client, Label: "Browser"},
		{ID: "lb", Kind: Balancer, Label: "Load balancer"},
		{ID: "api", Kind: Service, Label: "Shortener service", Notes: "create and redirect"},
		{ID: "ids", Kind: Service, Label: "ID range service", Notes: "hands each server a block of counter values"},
		{ID: "cache", Kind: Cache, Label: "Redirect cache", Notes: fmt.Sprintf("LRU of %d links", c.CacheSize)},
		{ID: "db", Kind: Database, Label: "Links database", Notes: "code -> long URL"},
	}, "client>lb", "lb>api", "api>ids:next id", "api>cache:lookup", "api>db:read or write")
	codes := map[string]int{}
	collisions := 0
	for i := 0; i < c.URLs; i++ {
		long := fmt.Sprintf("https://example.com/articles/%d", i)
		var code string
		if c.Hashed {
			for salt := 0; ; salt++ {
				h := hash64(fmt.Sprintf("%s#%d", long, salt))
				code = hashCode(h, c.CodeLen)
				if _, taken := codes[code]; !taken {
					break
				}
				collisions++
			}
		} else {
			code, _ = Base62(int64(i + 1))
		}
		codes[code] = i
		if a.recording() && i < 3 {
			f := a.frame(fmt.Sprintf("create %s -> %s", long, code))
			a.send(f, "client", "lb", i, OK)
			a.send(f, "lb", "api", i, OK)
			if !c.Hashed {
				a.send(f, "api", "ids", i, OK)
			}
			a.send(f, "api", "db", i, OK)
		}
	}
	rng := rand.New(rand.NewSource(c.Seed))
	next := popular(rng, c.URLs)
	cache := newLRU(c.CacheSize)
	hits := 0
	for i := 0; i < c.Reads; i++ {
		k := next()
		hit := cache.get(k)
		if hit {
			hits++
		} else {
			cache.put(k)
		}
		if a.recording() {
			f := a.frame(fmt.Sprintf("redirect link %d: %s", k, map[bool]string{true: "cache hit", false: "cache miss, read the database"}[hit]))
			id := c.URLs + i
			a.send(f, "client", "lb", id, OK)
			a.send(f, "lb", "api", id, OK)
			if hit {
				a.send(f, "api", "cache", id, Hit)
			} else {
				a.send(f, "api", "cache", id, Miss)
				a.send(f, "api", "db", id, OK)
			}
			f.load("cache", cache.order.Len(), 0, percent(hits, i+1))
		}
	}
	last := a.frame("totals")
	last.load("cache", cache.order.Len(), 0, percent(hits, c.Reads))
	last.load("db", c.Reads-hits, 0, -1)
	var r Report
	r.Arch = a
	if c.Hashed {
		r.say("hashed codes of %d characters: %d collisions retried among %d links (62^%d codes)", c.CodeLen, collisions, c.URLs, c.CodeLen)
	} else {
		r.say("counter codes in base 62: link %d is %q, no collisions, but the counter must be shared", c.URLs, mustBase62(int64(c.URLs)))
	}
	r.say("%d redirects, %d%% from the cache of %d links: the database sees only the misses", c.Reads, percent(hits, c.Reads), c.CacheSize)
	r.put("collisions", float64(collisions))
	r.put("cache hit rate (%)", float64(percent(hits, c.Reads)))
	r.put("database reads", float64(c.Reads-hits))
	return r, nil
}

func hashCode(h uint64, n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = base62Digits[h%62]
		h /= 62
	}
	return string(b)
}

func mustBase62(n int64) string {
	s, _ := Base62(n)
	return s
}
