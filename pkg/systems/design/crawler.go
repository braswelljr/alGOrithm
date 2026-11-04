package design

import (
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"strings"
)

// CrawlerConfig is a synthetic web to crawl.
type CrawlerConfig struct {
	Hosts        int   // web sites
	PagesPerHost int   // pages on each site
	LinksPerPage int   // links on each page, to any page of any site
	Politeness   int   // ticks to wait between two fetches from the same host
	Fetchers     int   // fetches the crawler makes each tick, across hosts
	Ticks        int   // how long to crawl
	Seed         int64 // the links
}

// WebCrawler models a polite web crawler: a URL frontier with one queue per
// host, so no host is fetched more than once every Politeness ticks; each
// tick the fetchers take the next URL from hosts that are allowed again,
// fetch it, parse its links, and add each one not seen before (a seen set of
// URL fingerprints, which at web scale is a Bloom filter). It reports pages
// fetched, links skipped as already seen, and the busiest host's fetch rate,
// which politeness caps. It leaves out robots.txt, priorities and recrawls.
//
// Time: O(Ticks * Fetchers * LinksPerPage). Space: O(Hosts * PagesPerHost).
func WebCrawler(c CrawlerConfig) (Report, error) {
	if err := check(c.Hosts >= 1 && c.Hosts <= 1000 && c.PagesPerHost >= 1 && c.PagesPerHost <= 1000 && c.LinksPerPage >= 0 && c.LinksPerPage <= 50 &&
		c.Politeness >= 1 && c.Fetchers >= 1 && c.Ticks >= 1 && c.Ticks <= 10_000, "1 to 1,000 hosts and pages per host, 0 to 50 links a page"); err != nil {
		return Report{}, err
	}
	rng := rand.New(rand.NewSource(c.Seed))
	url := func(h, p int) string { return fmt.Sprintf("https://site%d.example/page%d", h, p) }
	links := func(u string) []string {
		r := rand.New(rand.NewSource(int64(hash64(u)) ^ c.Seed))
		out := make([]string, c.LinksPerPage)
		for i := range out {
			h := r.Intn(c.Hosts)
			if r.Intn(3) > 0 { // most links stay on the same site
				h = hostOf(u)
			}
			out[i] = url(h, r.Intn(c.PagesPerHost))
		}
		return out
	}
	a := arch([]Component{
		{ID: "frontier", Kind: Queue, Label: "URL frontier", Notes: "one queue per host"},
		{ID: "fetch", Kind: Worker, Label: "Fetchers"},
		{ID: "web", Kind: Client, Label: "The web"},
		{ID: "parse", Kind: Worker, Label: "Parser"},
		{ID: "seen", Kind: Cache, Label: "Seen URLs", Notes: "fingerprints or a Bloom filter"},
		{ID: "store", Kind: Blob, Label: "Page store"},
	}, "frontier>fetch", "fetch>web:GET", "fetch>store", "fetch>parse", "parse>seen:new?", "seen>frontier:enqueue")
	queues := make([][]string, c.Hosts)
	nextOK := make([]int, c.Hosts)
	seen := map[uint64]bool{}
	start := url(rng.Intn(c.Hosts), 0)
	seen[hash64(start)] = true
	h0 := hostOf(start)
	queues[h0] = append(queues[h0], start)
	perHost := make([]int, c.Hosts)
	fetched, dupes, id := 0, 0, 0
	for t := 0; t < c.Ticks; t++ {
		var ready []int
		for h := range queues {
			if len(queues[h]) > 0 && nextOK[h] <= t {
				ready = append(ready, h)
			}
		}
		sort.Ints(ready)
		var f *Frame
		if a.recording() && len(ready) > 0 {
			f = a.frame("")
		}
		for i := 0; i < len(ready) && i < c.Fetchers; i++ {
			h := ready[i]
			u := queues[h][0]
			queues[h] = queues[h][1:]
			nextOK[h] = t + c.Politeness
			fetched++
			perHost[h]++
			id++
			newLinks := 0
			for _, l := range links(u) {
				k := hash64(l)
				if seen[k] {
					dupes++
					continue
				}
				seen[k] = true
				newLinks++
				lh := hostOf(l)
				queues[lh] = append(queues[lh], l)
			}
			if f != nil {
				f.Note = fmt.Sprintf("tick %d: fetched %s, %d new links", t, u, newLinks)
				a.send(f, "frontier", "fetch", id, OK)
				a.send(f, "fetch", "web", id, OK)
				a.send(f, "fetch", "store", id, OK)
				a.send(f, "fetch", "parse", id, OK)
				a.send(f, "parse", "seen", id, map[bool]Status{true: Miss, false: Hit}[newLinks > 0])
				if newLinks > 0 {
					a.send(f, "seen", "frontier", id, Queued)
				}
			}
		}
	}
	waiting := 0
	for _, q := range queues {
		waiting += len(q)
	}
	busiest := 0
	for _, n := range perHost {
		busiest = max(busiest, n)
	}
	last := a.frame("totals")
	last.load("frontier", waiting, waiting, -1)
	last.load("seen", len(seen), 0, percent(dupes, dupes+len(seen)))
	var r Report
	r.Arch = a
	r.say("%d pages fetched in %d ticks; %d links skipped as already seen; %d URLs waiting", fetched, c.Ticks, dupes, waiting)
	r.say("politeness of %d ticks caps any host at %d fetches; the busiest made %d", c.Politeness, (c.Ticks+c.Politeness-1)/c.Politeness, busiest)
	r.put("fetched", float64(fetched))
	r.put("duplicates skipped", float64(dupes))
	r.put("frontier size", float64(waiting))
	r.put("busiest host fetches", float64(busiest))
	return r, nil
}

// hostOf is the site number of a crawler URL ("https://site3.example/page1" is 3).
func hostOf(u string) int {
	s := strings.TrimPrefix(u, "https://site")
	if i := strings.IndexByte(s, '.'); i >= 0 {
		s = s[:i]
	}
	n, _ := strconv.Atoi(s)
	return n
}
