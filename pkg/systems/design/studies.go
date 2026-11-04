package design

import "slices"

// Study is a system design case study: what the design must do and how
// well, the estimate that sizes it (options of the back-of-the-envelope
// entry), its trade-offs and failure modes, and the entries to run next. ID
// is the registry entry that simulates it.
type Study struct {
	ID            string         `json:"id"`
	Title         string         `json:"title"`
	Level         string         `json:"level"`
	Pitch         string         `json:"pitch"`
	Functional    []string       `json:"functional"`
	NonFunctional []string       `json:"nonFunctional"`
	Estimate      map[string]int `json:"estimate"`
	Tradeoffs     []Tradeoff     `json:"tradeoffs"`
	Failures      []Failure      `json:"failures"`
	Related       []string       `json:"related"`
}

// Tradeoff is one design choice, what it gains and what it costs.
type Tradeoff struct {
	Choice string `json:"choice"`
	Gain   string `json:"gain"`
	Cost   string `json:"cost"`
}

// Failure is something that goes wrong and how the design handles it.
type Failure struct {
	What     string `json:"what"`
	Handling string `json:"handling"`
}

var studyTable = []Study{
	{
		ID: "url-shortener", Title: "URL shortener", Level: "beginner",
		Pitch:         "Turn a long URL into a short code, and redirect from the code, fast.",
		Functional:    []string{"Create a short code for a long URL", "Redirect a code to its URL", "Codes never collide"},
		NonFunctional: []string{"Redirects in under 50 ms at the 99th percentile", "Reads outnumber writes 100 to 1", "Links last for years"},
		Estimate:      map[string]int{"users": 100000000, "reads": 1, "writes": 0, "bytes": 500, "peak": 3, "years": 10, "replicas": 3},
		Tradeoffs: []Tradeoff{
			{"Counter in base 62", "No collisions, short codes", "A shared counter or ID ranges per server; codes are guessable"},
			{"Hash of the URL", "No coordination", "Collisions to retry, growing with the number of links"},
			{"Cache in front of the database", "Most redirects never touch the database", "Memory, and stale entries after a delete"},
		},
		Failures: []Failure{
			{"The cache is cold after a restart", "Reads fall through to replicas; warm the hottest codes first"},
			{"A popular link floods one cache node", "Replicate hot keys, or keep a small cache in each server"},
		},
		Related: []string{"base62", "ticket-server", "back-of-the-envelope"},
	},
	{
		ID: "key-value-store", Title: "Key-value store (Dynamo)", Level: "advanced",
		Pitch:         "A store that stays writable when machines fail, by keeping N copies of every key.",
		Functional:    []string{"put(key, value) and get(key)", "Survive node failures", "Grow by adding nodes"},
		NonFunctional: []string{"Always writable", "Tunable consistency with R and W", "Millisecond latency"},
		Estimate:      map[string]int{"users": 10000000, "reads": 50, "writes": 10, "bytes": 1000, "peak": 2, "years": 3, "replicas": 3},
		Tradeoffs: []Tradeoff{
			{"R + W > N", "Every read sees the latest write", "Slower, and fails with fewer nodes up"},
			{"Sloppy quorum", "Writes succeed while a home node is down", "Reads may miss them until handoff"},
			{"Last write wins", "Simple", "Concurrent writes are lost; vector clocks keep both"},
		},
		Failures: []Failure{
			{"A node is down", "Hinted handoff: a neighbor holds its writes and hands them back"},
			{"Replicas drift apart", "Read repair on every read, and Merkle-tree anti-entropy in the background"},
		},
		Related: []string{"quorum-consistency", "consistent-hashing", "merkle-tree", "vector-clock"},
	},
	{
		ID: "news-feed", Title: "News feed", Level: "intermediate",
		Pitch:         "Show each user the latest posts from the accounts they follow.",
		Functional:    []string{"Post", "Read a feed of followees' posts, newest first", "Follow and unfollow"},
		NonFunctional: []string{"A feed loads in under 200 ms", "Posts appear within seconds", "Celebrities have millions of followers"},
		Estimate:      map[string]int{"users": 300000000, "reads": 10, "writes": 1, "bytes": 1000, "peak": 3, "years": 5, "replicas": 3},
		Tradeoffs: []Tradeoff{
			{"Fan-out on write (push)", "A feed read is one cache lookup", "A celebrity post is millions of writes"},
			{"Fan-out on read (pull)", "A post is one write", "Every read merges many timelines"},
			{"Hybrid", "Push for most, pull only celebrities", "Two code paths to keep correct"},
		},
		Failures: []Failure{
			{"The fan-out queue backs up after a viral post", "Workers scale out; the feed falls back to pulling meanwhile"},
			{"Feed cache loss", "Rebuild on demand by pulling, then push again"},
		},
		Related: []string{"back-of-the-envelope", "cache-stampede", "backpressure"},
	},
	{
		ID: "chat-system", Title: "Chat system", Level: "intermediate",
		Pitch:         "One-to-one messages, delivered live or later, always in order.",
		Functional:    []string{"Send and receive messages", "Offline delivery", "Presence"},
		NonFunctional: []string{"Messages in order per conversation", "Delivery in under a second when online", "Never lose a message"},
		Estimate:      map[string]int{"users": 50000000, "reads": 40, "writes": 40, "bytes": 200, "peak": 4, "years": 5, "replicas": 3},
		Tradeoffs: []Tradeoff{
			{"WebSockets", "Low-latency push both ways", "Long-lived connections to balance and drain"},
			{"Sequence numbers per conversation", "Order without global clocks", "A counter per conversation"},
		},
		Failures: []Failure{
			{"The recipient is offline", "Store in the inbox, deliver on reconnect"},
			{"Messages arrive out of order", "Hold back until the gap fills, by sequence number"},
		},
		Related: []string{"lamport-clock", "notification-system"},
	},
	{
		ID: "notification-system", Title: "Notification system", Level: "intermediate",
		Pitch:         "Send push, email and SMS reliably, exactly once as far as the user can tell.",
		Functional:    []string{"Queue notifications", "Send through providers", "Retry and give up cleanly"},
		NonFunctional: []string{"No duplicates for the user", "Survive provider outages", "Millions a day"},
		Estimate:      map[string]int{"users": 20000000, "reads": 0, "writes": 5, "bytes": 500, "peak": 5, "years": 1, "replicas": 3},
		Tradeoffs: []Tradeoff{
			{"Idempotency keys", "Producer retries never notify twice", "A store of sent keys"},
			{"Exponential backoff", "Does not hammer a struggling provider", "Later delivery"},
		},
		Failures: []Failure{
			{"A provider is down", "Retries with backoff, then the dead-letter queue"},
			{"The producer retries a request", "Dropped by its idempotency key"},
		},
		Related: []string{"transactional-outbox", "backpressure"},
	},
	{
		ID: "web-crawler", Title: "Web crawler", Level: "intermediate",
		Pitch:         "Fetch billions of pages without fetching any twice or overloading any site.",
		Functional:    []string{"Fetch pages from seed URLs", "Follow links", "Store pages"},
		NonFunctional: []string{"Polite to every host", "No duplicate fetches", "Billions of pages"},
		Estimate:      map[string]int{"users": 1, "reads": 0, "writes": 1000000000, "bytes": 100000, "peak": 1, "years": 1, "replicas": 1},
		Tradeoffs: []Tradeoff{
			{"A queue per host", "Politeness is easy to enforce", "Many queues to schedule"},
			{"Bloom filter for seen URLs", "Fits in memory at web scale", "Rare false positives skip a page"},
		},
		Failures: []Failure{
			{"A spider trap of endless URLs", "Depth and per-host page limits"},
			{"A host goes slow", "Its queue waits; others continue"},
		},
		Related: []string{"bloom-filter", "breadth-first-search"},
	},
	{
		ID: "search-autocomplete", Title: "Search autocomplete", Level: "intermediate",
		Pitch:         "Suggest the most popular completions as the user types.",
		Functional:    []string{"Top k suggestions per prefix", "Update from query logs"},
		NonFunctional: []string{"Under 100 ms per keystroke", "Fresh within a day"},
		Estimate:      map[string]int{"users": 100000000, "reads": 50, "writes": 5, "bytes": 50, "peak": 3, "years": 1, "replicas": 3},
		Tradeoffs: []Tradeoff{
			{"Precomputed top k per prefix", "One lookup per keystroke", "Memory, and rebuilds to stay fresh"},
			{"Scan at read time", "Always current", "Too slow for popular prefixes"},
		},
		Failures: []Failure{
			{"A trending term", "A fast path for recent counts beside the nightly rebuild"},
		},
		Related: []string{"autocomplete", "trie"},
	},
	{
		ID: "proximity-service", Title: "Proximity service", Level: "intermediate",
		Pitch:         "Find businesses near a user, by geohash, without scanning the whole city.",
		Functional:    []string{"Search within a radius", "Add and update businesses"},
		NonFunctional: []string{"Searches in under 100 ms", "Read heavy"},
		Estimate:      map[string]int{"users": 100000000, "reads": 5, "writes": 0, "bytes": 1000, "peak": 3, "years": 5, "replicas": 3},
		Tradeoffs: []Tradeoff{
			{"Geohash cells", "A prefix lookup in any database", "Edge cases at cell borders: read the neighbors"},
			{"Quadtree", "Adapts to dense areas", "An in-memory structure to rebuild"},
		},
		Failures: []Failure{
			{"A radius larger than the cells", "Use a shorter geohash"},
		},
		Related: []string{"geohash", "quadtree", "ride-matching"},
	},
	{
		ID: "ride-matching", Title: "Ride matching", Level: "intermediate",
		Pitch:         "Match each rider to the nearest free driver, from live locations.",
		Functional:    []string{"Drivers report locations", "Riders request rides", "Match nearest free driver"},
		NonFunctional: []string{"Locations every few seconds", "Match in under a second"},
		Estimate:      map[string]int{"users": 5000000, "reads": 20, "writes": 1000, "bytes": 50, "peak": 3, "years": 1, "replicas": 2},
		Tradeoffs: []Tradeoff{
			{"Ring search over cells", "Few cells read", "Sparse areas search far"},
		},
		Failures: []Failure{
			{"No free driver nearby", "Widen the search, or queue the request"},
		},
		Related: []string{"proximity-service", "geohash"},
	},
	{
		ID: "payment-system", Title: "Payment system", Level: "advanced",
		Pitch:         "Charge customers exactly once and keep a ledger that always balances.",
		Functional:    []string{"Charge a payment", "Record it in a ledger", "Reconcile with the processor"},
		NonFunctional: []string{"Never charge twice", "Every cent accounted for", "Auditable"},
		Estimate:      map[string]int{"users": 1000000, "reads": 2, "writes": 1, "bytes": 2000, "peak": 10, "years": 7, "replicas": 3},
		Tradeoffs: []Tradeoff{
			{"Idempotency keys", "Retries are safe", "A key store with the results"},
			{"Double-entry ledger", "Errors show as an imbalance", "Two writes per movement"},
		},
		Failures: []Failure{
			{"The client times out and retries", "The key returns the first result"},
			{"The processor charged but the reply was lost", "Reconcile against the processor's records daily"},
		},
		Related: []string{"transactional-outbox", "saga-pattern"},
	},
	{
		ID: "distributed-rate-limiter", Title: "Distributed rate limiter", Level: "advanced",
		Pitch:         "Enforce one limit per user across many API nodes.",
		Functional:    []string{"Allow up to N requests per window", "Reject the rest with 429"},
		NonFunctional: []string{"Adds under a millisecond", "Close to the limit, never far over"},
		Estimate:      map[string]int{"users": 10000000, "reads": 100, "writes": 0, "bytes": 1, "peak": 5, "years": 1, "replicas": 1},
		Tradeoffs: []Tradeoff{
			{"Central counter", "Exact", "A round trip per request"},
			{"Synced local counters", "Cheap", "Lets a little over the limit through"},
		},
		Failures: []Failure{
			{"The central counter is down", "Fail open with local limits, or fail closed"},
		},
		Related: []string{"token-bucket", "leaky-bucket"},
	},
	{
		ID: "video-pipeline", Title: "Video upload and streaming", Level: "intermediate",
		Pitch:         "Take an upload, transcode it to every resolution, and stream it smoothly.",
		Functional:    []string{"Upload", "Transcode to several resolutions", "Stream adaptively"},
		NonFunctional: []string{"Ready minutes after upload", "Few stalls on slow networks"},
		Estimate:      map[string]int{"users": 1000000, "reads": 0, "writes": 1, "bytes": 500000000, "peak": 2, "years": 10, "replicas": 3},
		Tradeoffs: []Tradeoff{
			{"Chunked parallel transcoding", "Long videos ready fast", "Orchestration of many jobs"},
			{"Adaptive bitrate", "Smooth playback", "Many renditions to store"},
		},
		Failures: []Failure{
			{"A transcoder dies mid-job", "The job returns to the queue"},
		},
		Related: []string{"adaptive-bitrate", "backpressure"},
	},
	{
		ID: "leaderboard", Title: "Real-time leaderboard", Level: "beginner",
		Pitch:         "Show the top players and any player's rank, live.",
		Functional:    []string{"Add points", "Top k", "My rank"},
		NonFunctional: []string{"Updates visible in under a second", "Millions of players"},
		Estimate:      map[string]int{"users": 5000000, "reads": 10, "writes": 10, "bytes": 50, "peak": 4, "years": 1, "replicas": 2},
		Tradeoffs: []Tradeoff{
			{"Sorted set in a cache", "O(log n) rank and update", "Memory; the database keeps the durable copy"},
		},
		Failures: []Failure{
			{"The cache restarts", "Rebuild the set from the database"},
		},
		Related: []string{"skip-list", "heap-sort"},
	},
}

// Studies are the case studies, easiest first within each group. The
// site's system design hub shows them; each ID is an entry of this package.
//
// Time: O(1). Space: O(1).
func Studies() []Study {
	return slices.Clone(studyTable)
}

// Patterns are the data-pattern entries the case studies are built from.
var Patterns = []string{"leader-follower-replication", "multi-leader-replication", "sharding-strategies", "resharding", "cap-pacelc", "consistency-models", "cqrs", "event-sourcing", "saga-pattern", "transactional-outbox", "change-data-capture", "pagination-strategies", "backpressure", "bulkhead-pattern", "cache-stampede"}

// Foundations are the sizing and reliability entries.
var Foundations = []string{"back-of-the-envelope", "latency-numbers", "powers-of-two", "availability-nines", "error-budget", "littles-law", "queueing-latency", "snowflake-id", "uuid-ordering", "ticket-server"}
