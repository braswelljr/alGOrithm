# alGOrithm

[![alGOrithm logo](docs/assets/logo.svg)](https://algorithm-beige.vercel.app)

**Algorithms and data structures in Go, step by step: a library, a command and a visualizer that runs the real Go code in your browser.**

[![Tests](https://github.com/braswelljr/alGOrithm/actions/workflows/go.yml/badge.svg)](https://github.com/braswelljr/alGOrithm/actions/workflows/go.yml)
[![Lint](https://github.com/braswelljr/alGOrithm/actions/workflows/lint.yml/badge.svg)](https://github.com/braswelljr/alGOrithm/actions/workflows/lint.yml)
[![Web](https://github.com/braswelljr/alGOrithm/actions/workflows/web.yml/badge.svg)](https://github.com/braswelljr/alGOrithm/actions/workflows/web.yml)
[![Release](https://github.com/braswelljr/alGOrithm/actions/workflows/release.yml/badge.svg)](https://github.com/braswelljr/alGOrithm/actions/workflows/release.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/braswelljr/alGOrithm.svg)](https://pkg.go.dev/github.com/braswelljr/alGOrithm)
[![Go Report Card](https://goreportcard.com/badge/github.com/braswelljr/alGOrithm)](https://goreportcard.com/report/github.com/braswelljr/alGOrithm)
[![Go 1.27+](https://img.shields.io/badge/go-1.27%2B-00add8?logo=go&logoColor=white)](go.mod)
[![MIT license](https://img.shields.io/github/license/braswelljr/alGOrithm)](LICENSE)
[![The visualizer, live](https://img.shields.io/badge/visualizer-algorithm--beige.vercel.app-00add8)](https://algorithm-beige.vercel.app)

**alGOrithm** is an open-source collection of **696 algorithms, data structures and computer-science
simulations in 59 categories**, written in Go with nothing but the standard library: sorting and
searching, graphs and pathfinding, dynamic programming, string matching, cryptography and
hashing, trees, heaps, tries and caches, CPU scheduling, paging and deadlock, concurrency,
networking, databases, distributed systems, system design, machine learning, digital logic and
compilers. Every entry is documented with its time and space complexity (Big O), tested, and
traced, so the same code can be read, imported, run on your own data and watched step by step.

Try it in the browser, no install: **[algorithm-beige.vercel.app](https://algorithm-beige.vercel.app)**.

- **[Visualize](https://algorithm-beige.vercel.app/visualize/)** a sort, a search, a pathfinder or a scheduler on your data, and step through every comparison and swap.
- **[Read the docs](https://algorithm-beige.vercel.app/docs/)** of every entry: what it does, what it costs, its Go source and how to call it.
- **[Learn](https://algorithm-beige.vercel.app/learn/)** along 20 paths, from your first sort to how an operating system works.
- **[Quiz yourself](https://algorithm-beige.vercel.app/quiz/)** on real runs: predict the next step.
- **[Measure Big O](https://algorithm-beige.vercel.app/big-o/)**: time an entry on growing inputs and fit its curve.
- **[System design](https://algorithm-beige.vercel.app/system-design/)**: case studies with live architectures you can change.
- **[Play the arena](https://algorithm-beige.vercel.app/shortcuts/)**: race the pathfinders, beat the sorts at swaps, and more.

## Features

- **A Go library**: one package per topic under `pkg/` (`sorting`, `graphs`, `cryptography`,
  `structures`, `scheduling`, `concurrency`, ...), generic where it makes sense (`cmp.Ordered`,
  `comparable`), every exported function documented with its cost and what it does to its input,
  and **no third-party dependencies**.
- **A command**: `algorithm` runs any entry on your values, a CSV or JSON file or generated data,
  prints every step, compares entries, prints their source, and serves them as a JSON API
  (`algorithm serve`, [OpenAPI](api/openapi.yaml)) or as tools for AI assistants (`algorithm mcp`).
- **A visualizer**: the same Go code compiled to WebAssembly: animated sorts and searches,
  pathfinding on mazes, dynamic-programming tables, data structures drawn as you operate them,
  Gantt charts of CPU schedules, page frames, disk head paths, threads racing, circuits and
  system architectures. **Nothing you load leaves your browser.**
- **Traced**: every algorithm `X` has a `TraceX` that reports each step to a `trace.Recorder`;
  replaying a sort's events rebuilds its result.
- **Tested**: every entry runs its example in CI, every sort and search runs a shared correctness
  suite, structures are checked against models under random operations, everything runs under the
  race detector on Linux, macOS and Windows, and CI keeps a coverage floor (every `pkg/` package
  at 90% or more, no exported function untested).

## Install

```bash
go install github.com/braswelljr/alGOrithm/cmd/algorithm@latest   # the command
go get github.com/braswelljr/alGOrithm                             # the library
```

Release binaries for Linux, macOS, Windows and FreeBSD are on the releases page.

## Quick start

```bash
algorithm list                                          # every entry, by category
algorithm run quick-sort 5 3 9 1 --trace                # the result, the counts, every step
algorithm run binary-search --file scores.csv --column score --target 90
algorithm run a-star --generate maze --rows 15 --cols 31
algorithm run round-robin --example --option quantum=3  # a Gantt chart and waiting times
algorithm run lru-page-replacement 7 0 1 2 0 3 0 4 --option frames=3
algorithm run race-condition --option threads=3 --option locked=0
algorithm run stack --ops "push 1, push 2, pop, peek"
algorithm compare --generate reversed --size 2000      # every sort on the same data
algorithm quiz bubble-sort --count 5                    # predict the next step, scored against the real run
algorithm source dijkstra                               # the Go code behind it
algorithm serve                                         # the same as a JSON API
```

As a library:

```go
import (
 "github.com/braswelljr/alGOrithm/pkg/algorithms/pathfinding"
 "github.com/braswelljr/alGOrithm/pkg/algorithms/sorting"
 "github.com/braswelljr/alGOrithm/pkg/structures"
)

sorted := sorting.Quick([]string{"pear", "apple", "fig"}) // [apple fig pear]

g, _ := pathfinding.Generate(pathfinding.Maze, 21, 31, 1)
path, cost := pathfinding.AStar(g)

s := structures.NewStack[int]() // LIFO
s.Push(1)
v, ok := s.Pop() // 1, true; on an empty stack: 0, false
```

Every algorithm `X` has a `TraceX` beside it that reports each step to a `trace.Recorder`; that is what the command prints with `--trace` and what the visualizer animates.

The visualizer, locally (needs Node.js 24+ and pnpm):

```bash
make web/install
make dev      # builds everything, starts the site, rebuilds and retests on every Go change
```

## Layout

```text
alGOrithm/
  cmd/
    algorithm/              the command
    algorithm-wasm/         the browser build (globalThis.alGOrithm)
  pkg/                      the library, by domain; one package per topic, one file per algorithm
    algorithms/             sorting searching pathfinding graphs dynamic greedy backtracking
                            strmatch text recursion compression
    math/                   numbers numerical geometry bitwise randomized
    security/               ciphers cryptography hashing
    structures/             stacks, queues, lists; arrays/ sets/ trees/ collections/ and more under it
    os/                     scheduling memory disk deadlock concurrency cache
    ml/                     clustering regression classification neural
    systems/                networking databases distributed
    trace/                  the events every traced algorithm reports
    dataset/                real data in: text, CSV, JSON, generators
  internal/
    core/                   the registry: the one list of entries, Execute, Compare, snippets
    cli/                    the command's subcommands, the HTTP API, MCP
    check/                  the suites every sort and search runs
    version/                the program's name and version
  api/openapi.yaml          the HTTP API served by `algorithm serve`
  examples/                 small programs that use the library
  tools/dev/ tools/gendocs/ the dev loop and the generated docs
  algorithm.go              the embedded source, for the command and the site
  web/                      the site (one Next.js app)
  docs/                     guides, the catalog, rules, workflows, requirements, design
```

## Access principles

How a container hands values back, and what implements each in `structures` and `structures/heaps`:

| Principle | Also called | Structure | Go interface |
| ----------- | ------------- | ----------- | -------------- |
| Last in, first out | LIFO, FILO | `Stack`, a `Deque` used at one end | `structures.LIFO` |
| First in, first out | FIFO, LILO | `Queue`, a `Deque` used at both ends | `structures.FIFO` |
| Most urgent first | priority | `heaps.MinHeap`, `heaps.PriorityHeap` | `heaps.PriorityQueue` |

`algorithm run access-orders 3 1 4 1 5` puts the same values through all three and prints the order each gives back.

## Catalog

<!-- catalog:begin (generated by `go run ./tools/gendocs`; do not edit) -->
696 entries in 59 categories. The full catalog, with every cost and the command to run each, is [docs/catalog.md](docs/catalog.md).

New here? Walk one of the 20 [learning paths](docs/learn.md), or run `algorithm learn`.

### Algorithms

| Category | Entries |
|----------|---------|
| [Sorting](pkg/algorithms/sorting) | Bitonic Sort, Bogo Sort, Bubble Sort, Bucket Sort, Cocktail Shaker Sort, Comb Sort, Counting Sort, Cycle Sort, Gnome Sort, Heap Sort, Insertion Sort, Introsort, Merge Sort, Odd-Even Sort, Pancake Sort, Quick Sort, Radix Sort, Selection Sort, Shell Sort, Stooge Sort, Strand Sort, Tim Sort, Tree Sort |
| [Searching](pkg/algorithms/searching) | Binary Search, Exponential Search, Fibonacci Search, Interpolation Search, Jump Search, Linear Search, Lower Bound, Median of Medians, Quickselect, Ternary Search, Upper Bound |
| [Pathfinding](pkg/algorithms/pathfinding) | A* Search, Breadth-First Search, Depth-First Search, Dijkstra's Algorithm, Greedy Best-First Search |
| [Dynamic programming](pkg/algorithms/dynamic) | Coin Change, Edit Distance, Egg Drop, 0/1 Knapsack, Longest Increasing Subsequence (n log n), Longest Common Subsequence, Longest Increasing Subsequence, Longest Palindromic Subsequence, Matrix Chain Multiplication, Rod Cutting, Subset Sum, Unique Paths |
| [String matching](pkg/algorithms/strmatch) | Aho-Corasick, Anagram, Boyer-Moore-Horspool, Knuth-Morris-Pratt, Manacher's Algorithm, Naive String Search, Rabin-Karp, Suffix Array, Z-Algorithm |
| [Recursion](pkg/algorithms/recursion) | Ackermann Function, Factorial, Fibonacci (plain recursion), Fibonacci (memoized), Euclidean GCD, Least Common Multiple, McCarthy 91, Palindrome, Permutations, Tower of Hanoi |
| [Graphs](pkg/algorithms/graphs) | A* Search (graph), Articulation Points, Beam Search, Bellman-Ford, Bidirectional BFS, Bipartite Check, Boruvka's Minimum Spanning Tree, Bridges, Critical Path (longest path in a DAG), Cycle Detection, Shortest Paths in a DAG, Dinic's Maximum Flow, Euler Path (Hierholzer), Floyd-Warshall, Breadth-First Traversal, Depth-First Traversal, Dijkstra's Shortest Paths, Hamiltonian Path, Hopcroft-Karp Bipartite Matching, Iterative Deepening DFS, Johnson's All-Pairs Shortest Paths, k-Core Decomposition, Kosaraju's Strongly Connected Components, Kruskal's Minimum Spanning Tree, Maximum Flow (Edmonds-Karp), PageRank, Prim's Minimum Spanning Tree, Stoer-Wagner Minimum Cut, Tarjan's Strongly Connected Components, Topological Sort, Transitive Closure (Warshall), Traveling Salesperson (Held-Karp), Traveling Salesperson (nearest neighbor), Uniform-Cost Search |
| [Greedy algorithms](pkg/algorithms/greedy) | Activity Selection, Fractional Knapsack, Greedy Coin Change, Huffman Coding, Interval Partitioning, Job Sequencing with Deadlines |
| [Backtracking](pkg/algorithms/backtracking) | Combination Sum, Graph Coloring, N-Queens, Subset Sum (backtracking), Sudoku Solver, Word Search |
| [Compression](pkg/algorithms/compression) | Burrows-Wheeler Transform, LZ77, LZW, Move-to-Front, Run-Length Encoding |
| [Text algorithms](pkg/algorithms/text) | Autocomplete (trie), Distinct Substrings (suffix automaton), Fuzzy Search (bitap), Longest Common Prefix, Longest Common Substring, Longest Repeated Substring, Lyndon Factorization (Duval), Lexicographically Minimal Rotation (Booth), Line Diff (Myers), Character N-Grams, N-Gram Similarity (Jaccard), Smallest Period, Soundex, Spelling Suggestions, Text Justification, TF-IDF, Word Frequency, Word Wrap (minimum raggedness), Word Wrap (greedy) |

### Mathematics

| Category | Entries |
|----------|---------|
| [Numbers](pkg/math/numbers) | Binary GCD (Stein), Chinese Remainder Theorem, Collatz Sequence, Divisors, Euler's Totient, Even, Extended Euclidean Algorithm, Fast Exponentiation, Fibonacci (Fast Doubling), Fibonacci (matrix power), Karatsuba Multiplication, Miller-Rabin Primality, Modular Inverse, Integer Square Root (Newton), Pascal's Triangle, Perfect Numbers, Primality, Prime Factorization, Segmented Sieve, Sieve of Eratosthenes |
| [Computational geometry](pkg/math/geometry) | Closest Pair of Points, Convex Hull (Graham Scan), Convex Hull (Jarvis March), Convex Hull (Monotone Chain), Point in Polygon, Polygon Area (shoelace), Segment Intersection |
| [Bit manipulation](pkg/math/bitwise) | Submask Enumeration, Gray Code, Population Count (Kernighan), Power of Two, Reverse Bits, Single Number (XOR) |
| [Numerical methods](pkg/math/numerical) | Bisection Method, Fast Fourier Transform, Gaussian Elimination, Horner's Rule, Matrix Power, Newton-Raphson Method, Secant Method, Simpson's Rule, Strassen Multiplication, Trapezoid Rule |
| [Randomized algorithms](pkg/math/randomized) | Karger's Minimum Cut, Monte Carlo Pi, Randomized Quickselect, Reservoir Sampling |

### Security

| Category | Entries |
|----------|---------|
| [Ciphers](pkg/security/ciphers) | Affine Cipher, Atbash, Autokey Cipher, Beaufort Cipher, Caesar Brute Force, Caesar Cipher, Columnar Transposition, Enigma Machine, Frequency Analysis, Hill Cipher, One-Time Pad, Playfair Cipher, Rail Fence Cipher, ROT13, Substitution Cipher, Vigenère Cipher, Vigenère Crack, Vigenère Key Length |
| [Cryptography](pkg/security/cryptography) | AES-128, Base32, Base64, ChaCha20, Constant-Time Compare, Diffie-Hellman, ElGamal, Elliptic Curve, Feistel Network, Hex Encoding, HMAC-SHA256, Linear-Feedback Shift Register, MD5, Merkle Tree, PBKDF2, RSA, SHA-256, Shamir Secret Sharing |
| [Hashing](pkg/security/hashing) | Adler-32, Consistent Hashing, CRC-32, Cuckoo Hashing, DJB2, Double Hashing, FNV-1a (32 and 64 bit), Hash Distribution, Linear Probing, MurmurHash3 (32 bit), Polynomial Rolling Hash, Quadratic Probing, Robin Hood Hashing, SDBM |

### Data structures

| Category | Entries |
|----------|---------|
| [Arrays](pkg/structures/arrays) | Dynamic Array, JavaScript Array, Array.prototype.every, Array.prototype.filter, Array.prototype.find, Array.prototype.findIndex, Array.prototype.flatMap, Array.prototype.map, Array.prototype.reduce, Array.prototype.slice, Array.prototype.some, Array.prototype.sort, Array.prototype.splice, Maximum Subarray (Kadane), Prefix Sums, Reverse (two pointers), Rotate (three reversals), Sliding Window Maximum, Two Pointers (pair sum) |
| [Sets](pkg/structures/sets) | Bitset, Cartesian Product, Jaccard Similarity, Multiset (Bag), Ordered Set, Power Set, Roaring Bitmap, Difference (two pointers), Intersection (two pointers), Symmetric Difference (two pointers), Union (two pointers), Sparse Set |
| [Trees](pkg/structures/trees) | Build a Tree from Two Traversals, Expression Tree, Invert a Tree, Lowest Common Ancestor, Morris Inorder Traversal, Path Sum, Serialize and Deserialize, Height, Diameter and Balance, Inorder Traversal, Level-Order Traversal, Postorder Traversal, Preorder Traversal, Validate a Binary Search Tree |
| [Collections (lodash-style)](pkg/structures/collections) | Object.assign, camelCase, Chunk (_.chunk), Count By (_.countBy), Debounce (_.debounce), Deep Clone (_.cloneDeep), Deep Merge (_.merge), JavaScript Map, JavaScript Set, Escape HTML (_.escape), Flatten Deep (_.flattenDeep), Flatten Object, Object.fromEntries, Get by Path (_.get), Group By (_.groupBy), Deep Equality (_.isEqual), Math.round vs Go's math.Round, String at, charAt, codePointAt, String slice vs substring, String split, kebab-case, Key By (_.keyBy), Memoize (_.memoize), Object.groupBy, Omit (_.omit), Pad, padStart, padEnd, Partition (_.partition), Pick (_.pick), Range (_.range), Sample (_.sample), Set by Path (_.set), Shuffle (Fisher-Yates), snake_case, Order By (_.orderBy), Start Case, Sum, Mean, Median, Throttle (_.throttle), Number() conversion, Truncate (_.truncate), Unflatten Object, Uniq By (_.uniqBy), Unset by Path (_.unset), Words (_.words), Zip (_.zip) |
| [Linear structures](pkg/structures) | Access Orders: LIFO, FIFO, Priority, Circular Buffer, Deque, Doubly Linked List, Singly Linked List, Min Stack, Monotonic Queue, Queue (FIFO), Sorted List, Stack (LIFO) |
| [Heaps](pkg/structures/heaps) | Priority Queue (any order), Min-Heap (priority queue) |
| [Search trees](pkg/structures/searchtrees) | AVL Tree, B-Tree, Binary Search Tree, Interval Tree, Red-Black Tree, Skip List, Splay Tree, Treap |
| [Range queries](pkg/structures/ranges) | Disjoint Sparse Table, Fenwick Tree, Segment Tree, Sparse Table |
| [Lookup and caches](pkg/structures/lookup) | Hash Table, LFU Cache, LRU Cache, Set, Trie, TTL Cache, Union-Find |
| [Probabilistic structures](pkg/structures/probabilistic) | Bloom Filter, Count-Min Sketch, Cuckoo Filter, HyperLogLog |
| [Spatial structures](pkg/structures/spatial) | k-d Tree, Quadtree |
| [Persistent structures](pkg/structures/persistent) | Persistent Stack, Persistent Search Tree |
| [Text buffers](pkg/structures/buffers) | Gap Buffer, Rope |

### Operating systems

| Category | Entries |
|----------|---------|
| [CPU scheduling](pkg/os/scheduling) | Earliest Deadline First, First Come, First Served, Lottery Scheduling, Multilevel Feedback Queue, Priority Scheduling, Rate-Monotonic Scheduling, Round Robin, Shortest Job First, Shortest Remaining Time First |
| [Memory management](pkg/os/memory) | Address Translation (paging and TLB), Aging Page Replacement, Best Fit, Buddy Allocator, Clock (Second Chance), FIFO Page Replacement, First Fit, LFU Page Replacement, LRU Page Replacement, Optimal Page Replacement, Worst Fit |
| [Disk scheduling](pkg/os/disk) | C-LOOK, C-SCAN, FCFS Disk Scheduling, LOOK, SCAN (Elevator), Shortest Seek Time First |
| [Deadlock](pkg/os/deadlock) | Banker's Algorithm |
| [Concurrency](pkg/os/concurrency) | Barrier, Dining Philosophers, Peterson's Algorithm, Producer-Consumer, Race Condition, Readers-Writers, Counting Semaphore, Sleeping Barber |
| [CPU caches](pkg/os/cache) | Average Memory Access Time, CPU Cache Simulator |
| [Processes and threads](pkg/os/processes) | fork, exec, wait and exit, Process Lifecycle (five states), Signals and Handlers, System Calls (user and kernel mode), Thread Models (many-to-one, one-to-one, many-to-many), Threads Share, Processes Copy |
| [Synchronization](pkg/os/synchronization) | The ABA Problem, Cigarette Smokers, Compare-and-Swap Spinlock, Condition Variables: while, not if, Lamport's Bakery Algorithm, Peterson's Algorithm, Model Checked, Priority Inversion and Inheritance, Test-and-Set Spinlock |
| [Interprocess communication](pkg/os/ipc) | Local Socket Server, Message Queue, Pipe, Remote Procedure Call, Shared Memory and Torn Reads |
| [File systems](pkg/os/filesystems) | Contiguous Allocation, Free Space: Bitmap and Free List, Hard and Symbolic Links, Indexed Allocation, Inode Pointers and Maximum File Size, Journaling and Crash Recovery, Linked Allocation (FAT), Log-Structured File System Cleaning, Page Cache: Write-Through and Write-Back, Path Resolution, RAID 0, 1, 5 and 6 |
| [Virtual memory](pkg/os/virtualmemory) | Copy-on-Write Fork, Inverted Page Table, Memory-Mapped Files, Multi-Level Page Table Walk, Thrashing, TLB and Effective Access Time, Working Set |
| [Boot and interrupts](pkg/os/boot) | Boot Sequence, DMA and Programmed I/O, Nested Interrupts and Priorities |

### Computing

| Category | Entries |
|----------|---------|
| [Logic gates and Boolean algebra](pkg/computing/logic) | Canonical SOP and POS, De Morgan's Laws, DPLL SAT Solver, Expression to Circuit, Karnaugh Map, Logical Equivalence, Logic Gates, Quine-McCluskey, Satisfiability by Truth Table, Truth Table |
| [Digital circuits](pkg/computing/circuits) | Arithmetic Logic Unit, Binary Subtractor, Carry-Lookahead Adder, D Flip-Flop, Gated D Latch, Decoder (2 to 4), Demultiplexer (1 to 4), Full Adder, Half Adder, JK Flip-Flop, Magnitude Comparator, Multiplexer (4 to 1), Everything from NAND, Everything from NOR, Priority Encoder, Register with Load, Ripple-Carry Adder, Ripple Counter, Seven-Segment Decoder, Shift Register, SR Latch, Synchronous Counter, T Flip-Flop, Traffic Light State Machine |
| [Data representation](pkg/computing/representation) | Base Conversion, Base to Decimal (Horner), Binary-Coded Decimal, Binary Addition, Endianness, Fixed Point, Decode Float Bits, Why 0.1 + 0.2 != 0.3, Hamming(7,4) Code, IEEE 754 Half Precision, IEEE 754 Single Precision, Parity Bit, Signed Number Encodings, Two's Complement Overflow, UTF-8 Encoding |
| [Theory of computation](pkg/computing/automata) | Busy Beaver (3 states), Context-Free Grammar Derivation, CYK Parsing, Deterministic Finite Automaton, DFA Minimization (table filling), Product of Two DFAs, Elementary Cellular Automaton, Conway's Game of Life, Grammar Ambiguity (counting parse trees), The Halting Problem (diagonal argument), Lambda Calculus (beta reduction), Mealy Machine (vending machine), Moore Machine (vending machine), Nondeterministic Finite Automaton, Pushdown Automaton: a^n b^n, Pushdown Automaton: Balanced Parentheses, Primitive Recursion and Ackermann, Pumping Lemma (a^n b^n is not regular), Pushdown Automaton, Regex Backtracking Blowup, Subset Construction (NFA to DFA), Thompson's Construction (regex to NFA), Turing Machine: Binary Increment, Turing Machine, Turing Machine: Palindromes, Turing Machine: Unary Addition |
| [Compilers and interpreters](pkg/computing/compilers) | Bytecode Compiler, Constant Folding, Copying Collector (Cheney), Dead-Code Elimination, Lexer, Mark and Sweep, Syntax Tree of a Program, Pratt Parser, Recursive Descent Parser, Reference Counting, Register Allocation, Reverse Polish Notation, Shunting-Yard Algorithm, Stack Virtual Machine, Three-Address Code, Tree-Walking Interpreter, Incremental Tri-Color Marking, Type Checking |
| [Computer architecture](pkg/computing/architecture) | Amdahl's Law, Two-Pass Assembler, Call Stack and Frames, Context Switch, Pipeline Control Hazards, Cycles per Instruction, Fetch, Decode, Execute, Gshare Branch Predictor, Interrupt Handling, Memory Hierarchy, 1-Bit Branch Predictor, Pipeline Forwarding, Pipeline Data Hazards, Struct Alignment and Padding, 2-Bit Saturating Predictor, Von Neumann and Harvard |

### Machine learning

| Category | Entries |
|----------|---------|
| [Clustering](pkg/ml/clustering) | DBSCAN, Agglomerative Hierarchical Clustering, K-Means Clustering, K-Means++ Seeding, Silhouette Score |
| [Regression](pkg/ml/regression) | Gradient Descent (linear regression), Linear Regression, Logistic Regression, Multiple Linear Regression, Polynomial Regression, Mean Squared Error and R^2 |
| [Classification](pkg/ml/classification) | Confusion Matrix, Decision Tree (CART), Gaussian Naive Bayes, K-Fold Cross-Validation, K-Nearest Neighbors, Naive Bayes Spam Filter, Perceptron, Precision, Recall and F1, Train/Test Split |
| [Neural networks](pkg/ml/neural) | Activation Functions, Softmax and Cross-Entropy, Neural Network Learning XOR |

### Systems

| Category | Entries |
|----------|---------|
| [Networking](pkg/systems/networking) | ARP, Bandwidth-Delay Product, BGP Path Vector, CSMA/CA and Hidden Terminals (Wi-Fi), CSMA/CD (Ethernet), DHCP (DORA), Distance-Vector Routing, DNS Resolution, Go-Back-N ARQ, HTTP/1.1 vs HTTP/2 vs HTTP/3, HTTP Request Waterfall, Internet Checksum, MTU and IP Fragmentation, Leaky Bucket, Link-State Routing, Longest Prefix Match, Nagle's Algorithm, NAT with Port Translation, OSI and TCP/IP Layers, Packet vs Circuit Switching, QUIC vs TCP + TLS, Selective Repeat ARQ, Spanning Tree Protocol, Stop-and-Wait ARQ, Subnetting (CIDR), TCP Congestion Control, TCP Flow Control, TCP Handshake and Teardown, TCP Retransmission Timer, TLS 1.3 Handshake, Token Bucket, Traceroute, WebSocket Upgrade |
| [Databases](pkg/systems/databases) | B+ Tree Index, Block Nested-Loop Join, Buffer Pool (Clock), Buffer Pool (LRU-K), Conflict Serializability (precedence graph), Extendible Hashing, External Merge Sort, Hash Join, Index Nested-Loop Join, Isolation Levels and Their Anomalies, Linear Hashing, LSM Tree, MVCC (snapshot isolation), Nested-Loop Join, Query Planner (index or full scan), Relational Algebra (select, project, group by), Slotted Page, Sort-Merge Join, Timestamp Ordering, Strict Two-Phase Locking, Write-Ahead Log and Recovery |
| [Distributed systems](pkg/systems/distributed) | Bully Election, Byzantine Generals (OM(m)), Chandy-Lamport Snapshot, G-Counter (CRDT), Gossip Protocol, Lamport Clocks, OR-Set (CRDT), Paxos (single decree), Quorum Reads and Writes, Raft Consensus, Rendezvous Hashing, Ring Election (Chang-Roberts), Two-Phase Commit, Vector Clocks |
| [Infrastructure](pkg/systems/infrastructure) | Autoscaling, Cache Stampede, Cache-Aside, Write-Through, Write-Behind, CDN Edge Caches, Container CPU Shares and Memory Limits, Circuit Breaker, At-Most-Once, At-Least-Once, Exactly-Once, Blue/Green, Canary and Rolling Deployments, Exponential Backoff and Jitter, Least-Connections Load Balancing, Power of Two Choices, Round-Robin Load Balancing, Weighted Round-Robin Load Balancing |
| [System design](pkg/systems/design) | Adaptive Bitrate Streaming, Availability in Series and Parallel, Back-of-the-Envelope Estimation, Backpressure and Load Shedding, Base 62 Encoding, The Bulkhead Pattern, CAP and PACELC, Change Data Capture, Design a Chat System, Consistency Models, CQRS, Design a Distributed Rate Limiter, SLO Error Budget and Burn Rate, Event Sourcing, Geohash, Design a Key-Value Store (Dynamo), Latency Numbers Every Programmer Should Know, Leader-Follower Replication, Design a Real-Time Leaderboard, Little's Law, Multi-Leader Conflict Resolution, Design a News Feed, Design a Notification System, Offset vs Cursor Pagination, Design a Payment System, Powers of Two, Design a Proximity Service, Utilization vs Latency (M/M/1), Resharding Cost, Design Ride Matching, The Saga Pattern, Design Search Autocomplete, Range vs Hash vs Directory Sharding, Snowflake IDs, Ticket Servers, Transactional Outbox, Design a URL Shortener, UUID v4 vs v7 vs ULID, Design a Video Upload Pipeline, Design a Web Crawler |

<!-- catalog:end -->

## Documentation

- [Getting started](docs/guides/getting-started.md): the command, the library and the visualizer.
- [Understanding Big O](docs/guides/big-o.md): what the notation means, why each cost is what it is, and how `algorithm bench` measures it.
- [System design](docs/guides/system-design.md): sizing a problem, drawing the boxes, and the trade-offs.
- [Learning paths](docs/learn.md): 20 ordered walks through the catalog.
- [Catalog](docs/catalog.md): every entry with its cost and how to run it.
- [Cheat sheets](docs/cheatsheets.md): every category's costs and key ideas on one page.
- [API reference](https://pkg.go.dev/github.com/braswelljr/alGOrithm) on pkg.go.dev, and the [HTTP API](api/openapi.yaml).
- [Contributing](CONTRIBUTING.md), and the contributor docs in [`docs/`](docs/README.md): the rules the code follows, how to add an algorithm, testing, CI, and the design of the visualizer.

## Author

Built by **braswelljr**: [braswelljr.vercel.app](https://braswelljr.vercel.app) and [github.com/braswelljr](https://github.com/braswelljr).

## License

[MIT](LICENSE)
