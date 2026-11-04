# System design

How to design a system, for an interview or for real, one step at a time, with every step something you can run. The models live in [`pkg/systems/design`](../../pkg/systems/design); each is a seeded simulation, so a run replays exactly. On the site, [/system-design](https://algorithm-beige.vercel.app/system-design/) shows each case study's architecture with requests moving through it and knobs to change. The learning path is `algorithm learn system-design`.

These are teaching models. Each doc comment says what it keeps and what it leaves out; none of them is a capacity planner.

## The approach

```text
1. Requirements      what it must do (functional) and how well (non-functional)
        |
2. Estimate          reads/s, writes/s, storage, bandwidth: is this one box or a thousand?
        |
3. API and data      the calls, the records, the IDs
        |
4. High-level design clients -> load balancer -> services -> caches -> databases, queues
        |
5. Deep dives        the hard part: the hot key, the fan-out, the failure
        |
6. Trade-offs        what each choice gains and costs, and what fails
```

Spend the first minutes on 1 and 2. A design for 2,000 reads a second and one for 2 million share almost nothing.

## 1 and 2: requirements and the estimate

Write the functional requirements as verbs (create a short link, redirect) and the non-functional ones as numbers (p99 under 50 ms, 99.9% available, links kept 10 years). Then estimate:

```bash
algorithm run back-of-the-envelope --example
```

```text
reads: 10000000 users x 20 a day / 86,400 s = 2314.8 a second on average, 6944.4 at peak
writes: 10000000 users x 2 a day / 86,400 s = 231.5 a second on average, 694.4 at peak
read:write ratio 10:1, so read heavy: cache and replicate reads
storage: 20.0 GB a day, 109.5 TB over 5 years with 3 replicas
```

The shortcuts: a day is about 10^5 seconds, so a million requests a day is about 12 a second. Powers of two size the data (`algorithm run powers-of-two --example`). The latency numbers say where time goes: memory is 100 times an L1 hit, a datacenter round trip 5,000 times, a packet across the Atlantic 1.5 million times (`algorithm run latency-numbers --example`). Little's law sizes the pools: requests in flight = arrival rate x latency, so 500 a second at 80 ms keeps 40 in flight (`algorithm run littles-law --example`).

## 3: IDs

Every record needs an ID, and how you make them matters at scale:

| Scheme | Entry | Gains | Costs |
|--------|-------|-------|-------|
| Auto-increment on one database | | Dense, ordered | One database is the bottleneck and the single point of failure |
| Ticket servers counting by n from their own offsets | `ticket-server` | No collisions, simple | Gaps, not ordered across servers |
| Snowflake: ms timestamp, worker, sequence | `snowflake-id` | Sorts by time, no coordination | Needs worker numbers; a clock going back must refuse IDs |
| UUID v4 | `uuid-ordering` | No coordination at all | Random inserts scatter across the index |
| UUID v7 or ULID | `uuid-ordering` | Time ordered, no coordination | Leak the creation time |

`algorithm run snowflake-id --example` makes a worker's clock jump back 3 ms: the generator refuses IDs until the clock passes its last one again.

## 4 and 5: the case studies

Each is an entry; its knobs are options (`algorithm info ID` lists them).

| Design | Entry | The deep dive |
|--------|-------|---------------|
| URL shortener | `url-shortener`, `base62` | Counter or hash for codes; a cache for the read-heavy redirects |
| Key-value store (Dynamo) | `key-value-store` | A ring, N replicas, R and W quorums, sloppy quorums and hinted handoff, read repair |
| News feed | `news-feed` | Fan-out on write, on read, or hybrid; the celebrity problem |
| Chat | `chat-system` | Sequence numbers per conversation, presence, the offline inbox |
| Notifications | `notification-system` | Idempotency keys, exponential backoff, the dead-letter queue |
| Web crawler | `web-crawler` | A queue per host for politeness, a seen set |
| Autocomplete | `search-autocomplete` | Top k precomputed per prefix |
| Proximity search | `proximity-service`, `geohash` | Geohash cells and their eight neighbors |
| Ride matching | `ride-matching` | Rings of cells around the rider |
| Payments | `payment-system` | Idempotency keys and a double-entry ledger |
| Distributed rate limiter | `distributed-rate-limiter` | A central counter, local shares, or synced local counters |
| Video | `video-pipeline`, `adaptive-bitrate` | Chunked parallel transcoding; bitrate chosen by the buffer |
| Leaderboard | `leaderboard` | A sorted set in a cache |

Run one, then change one knob and run it again: that is the trade-off, measured. The news feed with fan-out on write:

```bash
algorithm run news-feed --example --option mode=0
```

```text
fan-out on write: 52511 feed writes for 200 posts (worst post: 999), 0 timelines pulled for 500 reads
```

The hybrid, the example's own setting, pulls only the three celebrities:

```text
hybrid: 2561 feed writes for 200 posts (worst post: 28), 1500 timelines pulled for 500 reads
```

The payment system without idempotency keys (`--option idempotent=0`) charges 80 customers twice; with them, none, and the ledger's 546 entries sum to 0.

## The data patterns

The case studies are built from these. Each has an entry:

```text
Replication     leader-follower-replication   followers lag: read-your-writes routing
                multi-leader-replication      last write wins loses updates; version vectors keep both
                key-value-store               leaderless quorums (and quorum-consistency in systems/distributed)
Partitioning    sharding-strategies           range piles hot keys on one shard; hash spreads them
                resharding                    mod n moves 80% of keys from 4 to 5 shards; a ring about 32% (20% is the least)
Consistency     cap-pacelc                    partitioned: refuse or diverge; else: latency or staleness
                consistency-models            is a history linearizable, sequential, causal?
Events          cqrs, event-sourcing          a read model fed by events; state as a replay
                transactional-outbox          two writes lose events; one transaction does not
                change-data-capture           stream the database's log to search and caches
                saga-pattern                  compensations undo finished steps in reverse
Under load      backpressure                  unbounded queues, load shedding, pushing back
                bulkhead-pattern              a pool per dependency contains a slow one
                request-coalescing            one database read for a hot key, not hundreds
                pagination-strategies         offset repeats rows as new ones arrive; cursors do not
```

`algorithm run consistency-models --example` checks a history where B reads 0 after A's write of 1 finished: sequential but not linearizable, because a read missed a write that had already completed in real time.

The neighbors: load balancing, circuit breakers, retries, caching strategies, CDNs and autoscaling are in [`pkg/systems/infrastructure`](../../pkg/systems/infrastructure); clocks, elections, Raft, Paxos, quorums and CRDTs in [`pkg/systems/distributed`](../../pkg/systems/distributed); sliding windows, congestion control and rate-limiting buckets in [`pkg/systems/networking`](../../pkg/systems/networking); B+ trees, LSM trees, joins and transactions in [`pkg/systems/databases`](../../pkg/systems/databases).

## 6: trade-offs and reliability

Say what each choice gains and costs, then what fails and how the design handles it. Availability math keeps it honest:

```bash
algorithm run availability-nines --example
```

```text
2: in series 98.0100% (10459 min down a year), in parallel 99.990000% (52.6 min)
```

Two 99% components in series are down five days a year; side by side, under an hour, if they fail independently, which shared deploys and regions often break. Then the budget: `algorithm run error-budget --example` shows a 99.9% SLO's allowed failures and the burn rate (14.4 or more spends a 30-day budget in two days: page someone). And why services run at 60% to 70% busy, not 100%: `algorithm run queueing-latency --example` shows an M/M/1 queue's latency at 2 times the service time at half load and 10 times at 90%.

## Checklist

- [ ] Requirements written as verbs and numbers
- [ ] An estimate: reads/s, writes/s, storage, bandwidth, the read:write ratio
- [ ] IDs chosen, and why
- [ ] The boxes, and the request path through them
- [ ] The hot spot found and handled: a hot key, a fan-out, a big tenant
- [ ] Every component's failure named, with what happens then
- [ ] The trade-offs said out loud
