package design

import (
	"fmt"
	"math/rand"
	"sort"
)

// Leaderboard is a sorted set of players by score, the structure behind a
// Redis ZSET: a map from player to score and the players kept in score
// order, so adding points, a player's rank and the top k are all quick.
// Redis uses a skip list for the order (O(log n) updates); this one keeps a
// sorted slice, O(log n) to find and O(n) to move, which is fine for
// teaching.
type Leaderboard struct {
	score map[string]int
	order []string // highest score first, ties by name
}

// NewLeaderboard returns an empty leaderboard.
func NewLeaderboard() *Leaderboard { return &Leaderboard{score: map[string]int{}} }

func (l *Leaderboard) less(a, b string) bool {
	if l.score[a] != l.score[b] {
		return l.score[a] > l.score[b]
	}
	return a < b
}

func (l *Leaderboard) find(p string) int {
	return sort.Search(len(l.order), func(i int) bool { return !l.less(l.order[i], p) })
}

// Add adds points to player (a new player starts at 0) and returns the new score.
//
// Time: O(log n) to find, O(n) to move in the slice. Space: O(1).
func (l *Leaderboard) Add(player string, points int) int {
	if _, ok := l.score[player]; ok {
		i := l.find(player)
		l.order = append(l.order[:i], l.order[i+1:]...)
	}
	l.score[player] += points
	i := l.find(player)
	l.order = append(l.order, "")
	copy(l.order[i+1:], l.order[i:])
	l.order[i] = player
	return l.score[player]
}

// Rank returns player's rank, 1 for the top, and false for an unknown player.
//
// Time: O(log n). Space: O(1).
func (l *Leaderboard) Rank(player string) (int, bool) {
	if _, ok := l.score[player]; !ok {
		return 0, false
	}
	return l.find(player) + 1, true
}

// Top returns the k best players, best first.
//
// Time: O(k). Space: O(k).
func (l *Leaderboard) Top(k int) []string {
	return append([]string(nil), l.order[:min(k, len(l.order))]...)
}

// Score returns player's score.
func (l *Leaderboard) Score(player string) int { return l.score[player] }

// LeaderboardConfig is a game's players and score updates.
type LeaderboardConfig struct {
	Players int   // players
	Updates int   // score updates (a match won)
	TopK    int   // how many to show at the top
	Seed    int64 // who scores
}

// LeaderboardDemo models a real-time game leaderboard: every match result
// adds points through the score service to a sorted set in a cache, so the
// top k and any player's rank are answered without the database, which
// would otherwise count every player with a higher score (a scan per rank
// query). The database keeps the durable copy. It leaves out sharding the
// set by region and ties broken by time.
//
// Time: O(Updates * Players) here, O(Updates log Players) with a skip list. Space: O(Players).
func LeaderboardDemo(c LeaderboardConfig) (Report, error) {
	if err := check(c.Players >= 1 && c.Players <= 100_000 && c.Updates >= 0 && c.Updates <= 1_000_000 && c.TopK >= 1, "1 to 100,000 players, a top k of at least 1"); err != nil {
		return Report{}, err
	}
	rng := rand.New(rand.NewSource(c.Seed))
	next := popular(rng, c.Players)
	lb := NewLeaderboard()
	a := arch([]Component{
		{ID: "game", Kind: Client, Label: "Game servers"},
		{ID: "score", Kind: Service, Label: "Score service"},
		{ID: "zset", Kind: Cache, Label: "Sorted set", Notes: "ZINCRBY, ZREVRANK, ZREVRANGE"},
		{ID: "db", Kind: Database, Label: "Scores database"},
		{ID: "player", Kind: Client, Label: "Player"},
	}, "game>score:match won", "score>zset:add points", "score>db:durable copy", "player>score:top and my rank", "score>zset:read")
	for i := 0; i < c.Updates; i++ {
		p := fmt.Sprintf("player%d", next())
		pts := 1 + rng.Intn(10)
		s := lb.Add(p, pts)
		if a.recording() {
			rank, _ := lb.Rank(p)
			f := a.frame(fmt.Sprintf("%s +%d points: score %d, rank %d", p, pts, s, rank))
			a.send(f, "game", "score", i, OK)
			a.send(f, "score", "zset", i, OK)
			a.send(f, "score", "db", i, Queued)
		}
	}
	f := a.frame("a player opens the leaderboard")
	a.send(f, "player", "score", 0, OK)
	a.send(f, "score", "zset", 0, Hit)
	var r Report
	r.Arch = a
	t := r.table("score")
	for i, p := range lb.Top(c.TopK) {
		t.add(fmt.Sprintf("%d. %s", i+1, p), lb.Score(p))
	}
	r.say("%d updates across %d players; top %d answered from the sorted set in O(k)", c.Updates, len(lb.score), c.TopK)
	r.say("a rank query without it scans all %d players for higher scores", len(lb.score))
	r.put("players on the board", float64(len(lb.score)))
	if top := lb.Top(1); len(top) > 0 {
		r.put("top score", float64(lb.Score(top[0])))
	}
	return r, nil
}
