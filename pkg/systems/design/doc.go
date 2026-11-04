// Package design teaches system design with small, deterministic models you
// can run: back-of-the-envelope estimates, unique ID schemes, reliability
// math, the classic case studies (a URL shortener, a key-value store, a news
// feed, chat, notifications, a crawler, autocomplete, proximity search, ride
// matching, payments, a distributed rate limiter, video delivery, a
// leaderboard) and the data patterns behind them (replication, sharding,
// CAP and PACELC, consistency models, CQRS, event sourcing, sagas, the
// outbox, change data capture, pagination, backpressure, bulkheads and
// request coalescing).
//
// Every model is a simulation with a seed, so a run replays exactly. Each one
// returns a Report: lines in words, named metrics, an optional table, and for
// the case studies an Architecture, the components and connections of the
// design with the requests in flight at each step, which the visualizer
// animates. Each doc comment says what the model keeps and what it leaves
// out: these are teaching models, not capacity planners.
//
// The neighbors: load balancing, circuit breakers, retries, caching
// strategies and CDNs are in systems/infrastructure; rate-limiting buckets,
// routing and the transport protocols in systems/networking; clocks,
// consensus, quorums and CRDTs in systems/distributed; consistent hashing in
// security/hashing; k-d trees and quadtrees in structures/spatial.
package design
