# Roadmap

The plan for alGOrithm, for contributors: first the features, by priority, then a curated list of what to learn and teach, by topic, with what alGOrithm has of it. [x] is done (for a topic, in the registry today: `algorithm list`); [ ] is open; "being built" marks work in progress ([rule](../rules/unfinished-work-is-marked-unfinished.md)). To add an entry, follow [adding an entry](../workflows/adding-an-entry.md); tick it here in the same change. A feature that touches the requirements gets a row in the [functional](functional-requirements.md) or [non-functional](non-functional-requirements.md) requirements when it starts.

## Features

### Next

The few things that most improve learning right now.

- [ ] **Exercises with auto-checked solutions.** "Implement insertion sort": the learner writes Go in an editor on the site, it compiles and runs in the browser, and `internal/check` (`check.Sort`, `check.Search`, the structure models) grades it with the failing case shown. Needs: a way to run user Go in the browser (a Go interpreter in wasm, or a small subset compiled by the computing domain's compiler), exercise definitions in the registry, and hints that reveal one step at a time.
- [ ] **Classroom mode.** A teacher makes a seeded quiz or exercise set and shares a link; students play it, and their results stay in their browser or export to a file the teacher collects. No server, keeping [the user's data never leaves the browser](../rules/the-users-data-never-leaves-the-browser.md); a result file is signed with a checksum, never with an account.
- [ ] **Progress per path.** Steps done, quiz scores per path, and "continue where you left off", kept in the browser and exportable.
- [ ] **Side-by-side code stepping.** Highlight the Go line that produced each event, so the animation and the code move together. Needs: events tagged with a source position (file and line), recorded only when tracing.
- [ ] **Complexity race.** Two or more entries animated in lockstep on the same input (bubble against merge on 2000 reversed values), with their operation counters racing; built on `compare` and `measure`.
- [ ] **Learning paths for every domain.** Systems ("How systems talk and agree"), computing ("From gates to a CPU", "How a program runs"), and a "Fundamentals first" path through the beginner entries of every domain.

### Later

- [ ] **Spaced repetition.** Questions missed in the quiz come back on a schedule (Leitner boxes), kept in the browser.
- [ ] **Achievements.** Badges for finishing paths, streaks and perfect games; earned locally and shown on a profile page that is only in the browser.
- [ ] **Offline.** The site as a progressive web app: the wasm, the pages and the user's progress work without a network.
- [ ] **Cheat sheets.** A print-friendly page per category, generated from the registry: every entry's cost, when to use it, and a one-line why; and a printable Big O sheet.
- [ ] **Keyboard-only and screen reader walkthroughs.** Every view narrated step by step in text (already started with step descriptions), checked with automated axe runs and a manual screen reader pass per view.
- [ ] **Embeddable widgets.** A one-line `<iframe>` or script tag for blogs and course pages, showing one entry on a given input, paused, with the site's controls.
- [ ] **Anki export.** Quiz questions and cost cards as an Anki deck, generated from the registry.
- [ ] **More quiz kinds.** "Which algorithm made this trace?", "What is the cost?", "Which structure fits these operations?", and drag-to-order questions (put these steps in order).
- [ ] **Benchmarks in the browser.** Wall time next to operation counts, with a warning about how much browsers vary.

### Someday

- [ ] **Internationalization.** US English first, then translations of the guides, the site and the path notes; entry names stay as they are.
- [ ] **Editor extensions.** A VS Code extension showing an entry's trace for the function under the cursor, and the cost from its doc comment.
- [ ] **Teacher dashboards from files.** Load many students' result files at once and see which questions the class missed, all in the browser.
- [ ] **User-defined entries.** Load a Go function written by the user, traced with `trace.Recorder`, into the visualizer; needs the in-browser Go from the exercises.
- [ ] **Narrated tours.** Recorded walkthroughs (text and timed steps) of a run, made in the site and shared as a link.

## Topics

### Sorting

- [x] Bubble, cocktail shaker, comb, gnome, odd-even sort: neighbor swaps and why they are quadratic
- [x] Insertion, selection, Shell sort: building a sorted prefix; gap sequences
- [x] Merge sort: divide and conquer, stability
- [x] Quick sort: partitioning, pivot choice, the worst case
- [x] Heap sort: a heap in an array
- [x] Counting, bucket, radix sort: beating O(n log n) without comparing
- [x] Cycle sort: the fewest writes; pancake sort: sorting with one move
- [x] Tim sort: runs, galloping, and what real libraries use
- [x] Introsort: quick sort that falls back to heap sort
- [x] Bitonic sort: a sorting network for parallel hardware
- [x] Tree, strand, stooge and bogo sort: what not to do, and why
- [ ] Sorting networks in general: the 0-1 principle, optimal networks for small n
- [ ] Parallel merge sort with goroutines, compared with the sequential one

### Searching

- [x] Linear, binary, ternary, jump, interpolation, exponential, Fibonacci search: what sortedness buys
- [x] Breadth-first search over a graph of nodes: level order, visited sets
- [x] Quickselect and median of medians: the k-th smallest in O(n)
- [x] Lower and upper bound: insertion points
- [ ] Binary search on the answer: searching a monotonic predicate, not a list

### Graphs

- [x] BFS, DFS, Dijkstra, A*, greedy best-first on grids: blind, cost-aware and guided search
- [x] Topological sort: ordering dependencies
- [x] Kruskal's and Prim's minimum spanning trees: greedy choices and union-find
- [x] Bellman-Ford: negative edges and detecting negative cycles
- [x] Floyd-Warshall: all-pairs shortest paths with a table
- [x] Tarjan's and Kosaraju's strongly connected components
- [x] Bipartite check: two-coloring with BFS
- [x] Maximum flow (augmenting paths, Dinic) and minimum cut, Karger's randomized cut
- [x] Articulation points and bridges: what disconnects a network
- [ ] DFS edge classification: tree, back, forward and cross edges
- [ ] Eulerian paths (Hierholzer) and the Chinese postman
- [ ] Traveling salesman: brute force, Held-Karp and the nearest-neighbor heuristic
- [ ] Hungarian algorithm: optimal assignment
- [ ] PageRank: ranking nodes by random walks

### Trees

- [x] Binary search tree, AVL, red-black, B-tree, treap, splay tree, skip list
- [x] Trie, segment tree, Fenwick tree, sparse tables, interval tree
- [x] Tree traversals, construction from traversals, lowest common ancestor, expression trees
- [ ] Heavy-light decomposition and Euler tours
- [ ] Suffix tree (Ukkonen), next to the suffix array

### Strings and text

- [x] Naive search, Knuth-Morris-Pratt, Rabin-Karp, Boyer-Moore, Z-algorithm
- [x] Suffix array, Aho-Corasick, Manacher
- [x] Repeated and common substrings, minimal rotation, Lyndon factorization, periods, fuzzy search (bitap)
- [x] Word frequency, TF-IDF, n-grams, word wrap, justification, Myers diff, Soundex, autocomplete, spelling suggestions
- [ ] Regular expression matching by backtracking against an automaton (pairs with the automata category)
- [ ] Metaphone and Double Metaphone
- [ ] Minimum edit script with moves (patience diff)

### Dynamic programming, greedy and backtracking

- [x] Edit distance, LCS, coin change, LIS (both), knapsack, matrix chain, rod cutting, subset sum, egg drop, unique paths
- [x] Activity selection, interval partitioning, Huffman coding, fractional knapsack, job sequencing
- [x] Permutations, N-queens, sudoku, subsets, combination sum, graph coloring, word search
- [ ] Bitmask DP over subsets; DP on trees
- [ ] Optimal binary search tree
- [ ] Knight's tour and Hamiltonian paths

### Math

- [x] GCD and LCM, extended GCD, modular inverse, Chinese remainder, primality, sieves, factorization, fast powers
- [x] Karatsuba, Strassen, FFT, matrix powers, Horner's rule
- [x] Numerical methods: roots, integration, Gaussian elimination
- [x] Randomized: reservoir sampling, Monte Carlo, randomized quickselect
- [x] Geometry: convex hulls (Graham, Jarvis, monotone chain), segment intersection, closest pair, polygon area, point in polygon
- [ ] Probability and statistics: distributions, the central limit theorem shown by sampling, Bayes' rule, hypothesis tests, bootstrapping
- [ ] Linear algebra visualized: vectors and transforms in the plane, eigenvectors by power iteration, SVD, least squares
- [ ] Information theory: entropy, mutual information, the Huffman bound, Shannon-Fano, arithmetic coding
- [ ] Game theory: minimax, alpha-beta pruning on tic-tac-toe and connect four, Nash equilibria of small games, the prisoner's dilemma tournament
- [ ] Computer graphics: Bresenham's line and circle, polygon fill and rasterization, the z-buffer, ray casting a small scene, Bezier curves (de Casteljau)

### Security

- [x] Classical ciphers and their attacks: Caesar, Vigenère, Hill, affine, Playfair, rail fence, columnar, Enigma, frequency analysis
- [x] Cryptography internals: SHA-256, MD5, HMAC, PBKDF2, AES-128, ChaCha20, RSA, Diffie-Hellman, ElGamal, elliptic curves, Shamir, Merkle trees, Feistel networks, LFSRs, encodings, constant-time compare
- [x] Hash functions and collision resolution
- [ ] Security II, kept safe as simulations: a buffer overflow on a simulated stack (no real memory), SQL injection against a toy query builder with the parameterized fix, a TLS 1.3 handshake walkthrough, password hashing and salts, timing attacks
- [ ] Digital signatures: RSA signatures, Schnorr, and why nonces must never repeat
- [ ] Zero-knowledge proofs, as an interactive game (the cave)

### Data structures

- [x] Dynamic array, JavaScript Array methods, array techniques (two pointers, windows, Kadane, prefix sums, rotation)
- [x] Stack, queue, deque, circular buffer, min stack, monotonic queue, linked lists, sorted list, access orders compared
- [x] Heaps, hash tables, union-find, LRU, LFU and TTL caches
- [x] Bloom and cuckoo filters, Count-Min sketch, HyperLogLog
- [x] k-d tree, quadtree, persistent stack and tree, gap buffer, rope, sparse set, roaring bitmap
- [x] Sets, multisets, ordered sets; lodash-style collections and the JavaScript Object, String, Map, Set and Math methods
- [ ] Fibonacci and pairing heaps: decrease-key and amortized analysis
- [ ] R-tree: rectangles for maps
- [ ] Van Emde Boas tree and x-fast trie
- [ ] Lock-free queue, as a deterministic simulation

### Operating systems

- [x] FCFS, SJF, SRTF, round robin, priority, MLFQ, lottery, EDF and rate-monotonic scheduling
- [x] Page replacement (FIFO, LRU, optimal, clock, LFU, aging), fit allocation, buddy allocation, address translation and TLBs
- [x] Disk scheduling: FCFS, SSTF, SCAN, C-SCAN, LOOK, C-LOOK
- [x] Races, producer-consumer, dining philosophers, readers-writers, sleeping barber, Peterson's algorithm, barriers, the Banker's algorithm
- [x] CPU cache simulation: direct-mapped, set-associative, average memory access time
- [ ] Process states and transitions as a state machine; context switches and what they cost; fork and exec and the process tree
- [ ] Stride scheduling; NRU page replacement
- [ ] Operating systems II: file systems (contiguous, linked/FAT and indexed/inode allocation), free-space bitmaps, directories, journaling and crash recovery, multi-level page tables
- [ ] Deadlock detection with several instances of each resource; cigarette smokers
- [ ] User and kernel threads and the threading models

### Concurrency in Go

- [x] Worker pool, pipeline, semaphore, bounded buffer
- [ ] Fan-in and fan-out with cancellation through `context`
- [ ] errgroup-style error propagation, written with the standard library
- [ ] Memory model lessons: happens-before with channels and `sync`, shown on a simulated trace

### Machine learning

- [x] k-means and k-means++, DBSCAN, hierarchical clustering, silhouette score
- [x] Linear, multiple, polynomial and logistic regression; gradient descent; metrics
- [x] k-NN, naive Bayes (Gaussian and text), perceptron, decision trees, confusion matrices, cross-validation
- [x] A small neural network learning XOR; activations; softmax and cross-entropy
- [ ] Principal component analysis; t-SNE as a seeded demonstration
- [ ] Support vector machines (a linear SVM by gradient descent)
- [ ] Random forests and gradient boosting on small tables
- [ ] Reinforcement learning: multi-armed bandits, Q-learning on a grid world (pairs with the grid view)
- [ ] Convolution on a small image; attention on a short sentence

### Systems

- [x] Stop-and-wait, Go-Back-N and selective repeat; TCP congestion control; token and leaky buckets; Nagle's algorithm
- [x] Distance-vector and link-state routing; subnetting; longest-prefix match; the Internet checksum; DNS resolution
- [x] Lamport and vector clocks, bully and ring elections, two-phase commit, Raft, Paxos, gossip, quorums, CRDTs, Chandy-Lamport snapshots, Byzantine generals
- [x] Databases: write-ahead logging, external merge sort, query planning, relational algebra, isolation anomalies; indexes, joins and transactions (being built)
- [ ] HTTP/1.1 and HTTP/2 framing; QUIC streams; a TLS walkthrough (with security)
- [ ] Load balancing: round robin, least connections, power of two choices
- [ ] MapReduce word count as a deterministic simulation

### Computing fundamentals

Being built: the logic, circuits, representation, automata, compilers and architecture categories.

- [ ] Logic gates and truth tables; Boolean algebra laws; Karnaugh maps; Quine-McCluskey (being built)
- [ ] Half and full adders, ripple-carry and carry-lookahead adders, multiplexers, decoders, latches, flip-flops, counters, registers, an ALU (being built)
- [ ] Bases, two's complement, overflow, IEEE 754 floats, fixed point, ASCII and UTF-8 (being built)
- [ ] DFAs, NFAs, subset construction, minimization, regular expressions to NFAs (Thompson), pushdown automata, context-free grammars, Turing machines (being built)
- [ ] Lexer, recursive-descent and Pratt parsers, the shunting-yard algorithm, syntax trees, tree-walking evaluation, bytecode compilation, a stack VM, a mark-and-sweep garbage collector (being built)
- [ ] A small instruction set, fetch-decode-execute, pipelines and hazards, forwarding, branch prediction (one-bit and two-bit counters) (being built)
- [ ] Computability: the halting problem as a diagonal argument, reductions, P and NP with SAT and 3-coloring examples
- [ ] Quantum computing basics, simulated: qubits as vectors, single-qubit gates, CNOT and entanglement, measurement, Deutsch-Jozsa, Grover search on a few qubits
