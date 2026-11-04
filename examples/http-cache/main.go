// Command http-cache caches API responses in an LRU cache whose entries
// also expire after a time to live, like a CDN edge or an in-process cache
// in front of a slow backend: hot responses stay, stale ones are refetched.
package main

import (
	"fmt"

	"github.com/braswelljr/alGOrithm/pkg/structures/lookup"
)

func main() {
	const ttl = 30 // seconds
	cache := lookup.NewTTLCache[string, string](2, ttl)
	backendCalls := 0
	get := func(path string, now int64) string {
		if body, ok := cache.Get(path, now); ok {
			return "HIT  " + body
		}
		backendCalls++
		body := fmt.Sprintf("{\"path\":%q,\"fetchedAt\":%d}", path, now)
		cache.Put(path, body, now)
		return "MISS " + body
	}
	requests := []struct {
		path string
		at   int64
	}{
		{"/users/1", 0}, {"/users/1", 5}, {"/orders", 10}, {"/users/2", 12},
		{"/users/1", 15}, {"/orders", 50}, {"/orders", 55},
	}
	fmt.Println("serving 7 requests through a 2-entry cache with a 30s TTL")
	for _, r := range requests {
		fmt.Printf("t=%2ds GET %-9s %s\n", r.at, r.path, get(r.path, r.at))
	}
	fmt.Printf("backend calls: %d of %d requests\n", backendCalls, len(requests))
}
