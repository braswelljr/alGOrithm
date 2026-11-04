package design

import (
	"fmt"
	"math/rand"
)

// FeedMode is how a news feed is built.
type FeedMode int

// The ways to build a feed.
const (
	FanOutOnWrite FeedMode = iota // push each post into every follower's feed
	FanOutOnRead                  // pull the followees' posts when a feed is read
	Hybrid                        // push for most authors, pull for celebrities
)

// FeedConfig is a social network's shape and traffic.
type FeedConfig struct {
	Users       int // accounts
	Follows     int // accounts each user follows
	Celebrities int // accounts that everyone follows
	Posts       int // posts written
	Reads       int // feeds read
	Mode        FeedMode
	Threshold   int   // followers above which Hybrid pulls instead of pushing
	Seed        int64 // who follows whom and who posts
}

// NewsFeed compares the three ways to build a news feed by the work they do.
// Fan-out on write copies each post into every follower's feed: reads are one
// lookup, but a celebrity's post costs a write per follower (millions).
// Fan-out on read stores each post once and merges the followees' posts at
// read time: cheap writes, expensive reads. Hybrid pushes for ordinary
// authors and pulls only celebrities, as Twitter does. Writes count feed
// insertions; reads count timelines fetched. It leaves out ranking and
// pagination of the feed.
//
// Time: O(Posts * followers + Reads * Follows). Space: O(Users * Follows).
func NewsFeed(c FeedConfig) (Report, error) {
	if err := check(c.Users >= 2 && c.Users <= 100_000 && c.Follows >= 1 && c.Follows < c.Users && c.Celebrities >= 0 && c.Celebrities < c.Users &&
		c.Posts >= 0 && c.Reads >= 0 && c.Mode >= FanOutOnWrite && c.Mode <= Hybrid && c.Threshold >= 0,
		"2 to 100,000 users, follows fewer than users, a known mode"); err != nil {
		return Report{}, err
	}
	rng := rand.New(rand.NewSource(c.Seed))
	followers := make([][]int, c.Users)
	following := make([][]int, c.Users)
	for u := 0; u < c.Users; u++ {
		seen := map[int]bool{u: true}
		for celeb := 0; celeb < c.Celebrities && len(following[u]) < c.Follows; celeb++ {
			if !seen[celeb] {
				seen[celeb] = true
				following[u] = append(following[u], celeb)
			}
		}
		for len(following[u]) < c.Follows {
			v := rng.Intn(c.Users)
			if !seen[v] {
				seen[v] = true
				following[u] = append(following[u], v)
			}
		}
		for _, v := range following[u] {
			followers[v] = append(followers[v], u)
		}
	}
	push := func(author int) bool {
		return c.Mode == FanOutOnWrite || (c.Mode == Hybrid && len(followers[author]) <= c.Threshold)
	}
	a := arch([]Component{
		{ID: "client", Kind: Client, Label: "App"},
		{ID: "post", Kind: Service, Label: "Post service"},
		{ID: "posts", Kind: Database, Label: "Posts store"},
		{ID: "queue", Kind: Queue, Label: "Fan-out queue"},
		{ID: "fanout", Kind: Worker, Label: "Fan-out workers"},
		{ID: "feeds", Kind: Cache, Label: "Feed cache", Notes: "one list of post ids per user"},
		{ID: "feed", Kind: Service, Label: "Feed service"},
	}, "client>post:post", "post>posts", "post>queue", "queue>fanout", "fanout>feeds", "client>feed:read feed", "feed>feeds", "feed>posts:pull followees")
	var feedWrites, maxWrites, pulls int
	for p := 0; p < c.Posts; p++ {
		author := rng.Intn(c.Users)
		if p%4 == 0 && c.Celebrities > 0 {
			author = rng.Intn(c.Celebrities)
		}
		w := 0
		if push(author) {
			w = len(followers[author])
		}
		feedWrites += w
		maxWrites = max(maxWrites, w)
		if a.recording() && p < 6 {
			f := a.frame(fmt.Sprintf("user %d posts (%d followers): %d feed writes", author, len(followers[author]), w))
			a.send(f, "client", "post", p, OK)
			a.send(f, "post", "posts", p, OK)
			if w > 0 {
				a.send(f, "post", "queue", p, Queued)
				a.send(f, "queue", "fanout", p, OK)
				a.send(f, "fanout", "feeds", p, OK)
			}
		}
	}
	for q := 0; q < c.Reads; q++ {
		u := rng.Intn(c.Users)
		n := 0
		for _, v := range following[u] {
			if !push(v) {
				n++
			}
		}
		pulls += n
		if a.recording() {
			f := a.frame(fmt.Sprintf("user %d reads the feed: 1 cache read and %d followee timelines pulled", u, n))
			a.send(f, "client", "feed", c.Posts+q, OK)
			a.send(f, "feed", "feeds", c.Posts+q, Hit)
			if n > 0 {
				a.send(f, "feed", "posts", c.Posts+q, Miss)
			}
		}
	}
	last := a.frame("totals")
	last.load("feeds", feedWrites, 0, -1)
	last.load("posts", pulls, 0, -1)
	var r Report
	r.Arch = a
	r.say("%s: %d feed writes for %d posts (worst post: %d), %d timelines pulled for %d reads",
		[]string{"fan-out on write", "fan-out on read", "hybrid"}[c.Mode], feedWrites, c.Posts, maxWrites, pulls, c.Reads)
	if c.Celebrities > 0 {
		r.say("each of the %d celebrities has about %d followers: pushing their posts is the celebrity problem", c.Celebrities, len(followers[0]))
	}
	r.put("feed writes", float64(feedWrites))
	r.put("worst post (writes)", float64(maxWrites))
	r.put("timelines pulled", float64(pulls))
	r.put("work", float64(feedWrites+pulls))
	return r, nil
}
