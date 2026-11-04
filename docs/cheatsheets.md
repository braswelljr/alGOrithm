# Cheat sheets

Every category on one table: what each entry costs and the idea that makes it work. Generated from the registry (`go run ./tools/gendocs`); the site's /cheatsheets page is the same, printable. `algorithm info ID` gives the TL;DR, when to use it and the pitfall.

## Algorithms

### Sorting

Putting values in order, sixteen ways, with every comparison and swap shown.

| Entry | Time | Space | Key idea |
|-------|------|-------|----------|
| Bitonic Sort | O(n log^2 n) | O(log n) | A fixed network of compare-and-swap steps that merges rising-then-falling sequences; every step of a stage can run in parallel. |
| Bogo Sort | O(n * n!) expected, unbounded worst | O(1) | Shuffle until sorted: a joke and a lesson, about n! shuffles expected, so it gives up after a set number of tries. |
| Bubble Sort | O(n^2), O(n) if sorted | O(1) | Each pass carries the largest remaining value to the end, like a bubble rising. |
| Bucket Sort | O(n) average, O(n^2) worst | O(n) | If values are spread evenly, each bucket holds about one value. |
| Cocktail Shaker Sort | O(n^2), O(n) if sorted | O(1) | Small values near the end move left quickly instead of one step per pass. |
| Comb Sort | O(n^2) worst, about O(n log n) random | O(1) | Bubble sort over a gap that shrinks by 1.3 each pass, so small values far down move early. |
| Counting Sort | O(n + k) for a range of k | O(n + k) | With values in a small range, counting replaces comparing. |
| Cycle Sort | O(n^2), at most n writes | O(1) | Put each value straight into the position it belongs in, so every value is written at most once. |
| Gnome Sort | O(n^2), O(n) if sorted | O(1) | Walk forward while neighbors are in order; when they are not, swap them and step back. |
| Heap Sort | O(n log n) | O(1) | A heap finds the maximum in O(1) and restores itself in O(log n). |
| Insertion Sort | O(n^2), O(n) if sorted | O(1) | The prefix is always sorted, so each new value only moves past the larger ones. |
| Introsort | O(n log n) | O(log n) | Watching the depth catches the bad-pivot case and caps the worst case at O(n log n). |
| Merge Sort | O(n log n) | O(n) | Merging two sorted lists takes one pass, and halving gives log n levels of passes. |
| Odd-Even Sort | O(n^2) | O(1) | Alternate passes compare pairs at even and odd positions; each pass is independent, which suits parallel hardware. |
| Pancake Sort | O(n^2), at most 2n - 3 flips | O(1) | The only move is flipping a prefix: flip the largest to the top, then down into place. |
| Quick Sort | O(n log n) average, O(n^2) worst | O(log n) | After partitioning the pivot is in its final place, and the two sides never mix again. |
| Radix Sort | O(n * k) for k digits | O(n + 10) | Stability keeps earlier digit orders intact, so after the last digit everything is in order. |
| Selection Sort | O(n^2) | O(1) | After pass i the first i slots hold the i smallest values, in order, for good. |
| Shell Sort | about O(n^(4/3)), O(n^2) worst | O(1) | Long jumps move values close to their place early, so the final pass has little left to do. |
| Stooge Sort | O(n^2.71) | O(log n) | Order the ends, then sort the first two thirds, the last two thirds, and the first two thirds again: correct and famously slow. |
| Strand Sort | O(n^2), O(n) if sorted | O(n) | Pull out an increasing strand of the remaining values and merge it into the result, until none remain. |
| Tim Sort | O(n log n), O(n) if sorted | O(n) | Real data is often partly sorted, and merging existing runs is cheap. |
| Tree Sort | O(n log n) average, O(n^2) if sorted | O(n) | Insert every value into a binary search tree, then read it back in order. |

### Searching

Finding a value in a list, sorted or not.

| Entry | Time | Space | Key idea |
|-------|------|-------|----------|
| Binary Search | O(log n) | O(1) | Each comparison halves what is left, so n values need about log2 n steps. |
| Exponential Search | O(log i) for a target at i | O(1) | The cost depends on where the target is, not on the length of the list. |
| Fibonacci Search | O(log n) | O(1) | Split a sorted range at Fibonacci numbers instead of halves. |
| Interpolation Search | O(log log n) uniform, O(n) worst | O(1) | On evenly spread values the guess lands close, giving about log log n probes. |
| Jump Search | O(sqrt n) | O(1) | sqrt(n) jumps plus sqrt(n) steps balance out at O(sqrt n). |
| Linear Search | O(n) | O(1) | It needs no order, only a way to compare for equality. |
| Lower Bound | O(log n) | O(1) | It is binary search that keeps going after a match, so it handles duplicates and misses. |
| Median of Medians | O(n) | O(n) | Quickselect with a pivot chosen as the median of the medians of groups of five, which guarantees linear time. |
| Quickselect | O(n) average, O(n^2) worst | O(n) | Only the side holding position k matters, so the work shrinks geometrically: O(n) on average. |
| Ternary Search | O(log n) | O(1) | Comparing two inner points tells you which outer third cannot hold the peak. |
| Upper Bound | O(log n) | O(1) | The first index whose value is greater than the target: the insertion point after any equal values. |

### Pathfinding

Finding paths through grids and mazes, from blind search to A*.

| Entry | Time | Space | Key idea |
|-------|------|-------|----------|
| A* Search | O(V log V) worst | O(V) | A heuristic that never overestimates keeps the path optimal while exploring far less. |
| Breadth-First Search | O(V) | O(V) | A queue visits cells in order of distance, so the first time you reach a cell is the shortest way. |
| Depth-First Search | O(V) | O(V) | A stack (or recursion) remembers where to back up to, using memory only for the current path. |
| Dijkstra's Algorithm | O(V log V) | O(V) | With no negative costs, nothing found later can be cheaper than the current cheapest. |
| Greedy Best-First Search | O(V log V) | O(V) | Trusting only the estimate heads straight for the goal. |

### Dynamic programming

Solving a problem by filling a table of answers to smaller ones.

| Entry | Time | Space | Key idea |
|-------|------|-------|----------|
| Coin Change | O(amount * coins) | O(amount) | The best way to make x uses one coin c plus the best way to make x - c. |
| Edit Distance | O(n * m) | O(n * m) | The answer for two prefixes depends only on the answers for slightly shorter prefixes. |
| Egg Drop | O(eggs * drops) | O(eggs * drops) | The fewest drops that always find the highest safe floor with a given number of eggs, from how many floors t drops can cover. |
| 0/1 Knapsack | O(n * capacity) | O(n * capacity) | For each item and capacity, either skip it or take it and solve the smaller capacity. |
| Longest Increasing Subsequence (n log n) | O(n log n) | O(n) | Patience sorting: keep the smallest tail of every run length and binary search where each value goes. |
| Longest Common Subsequence | O(n * m) | O(n * m) | If the last characters match they extend the answer; otherwise drop one and take the better. |
| Longest Increasing Subsequence | O(n^2) | O(n) | For each value, the best chain ending at it extends the best smaller one before it. |
| Longest Palindromic Subsequence | O(n^2) | O(n^2) | The longest subsequence that reads the same backward, from the answer for every inner range. |
| Matrix Chain Multiplication | O(n^3) | O(n^2) | The best split of a chain is the cheapest of every split point plus the best of each side. |
| Rod Cutting | O(n^2) | O(n) | Cut a rod into pieces for the most money, where piece length i sells for price i, from the best answer for every shorter rod. |
| Subset Sum | O(n * target) | O(n * target) | A table of reachable sums grows one number at a time. |
| Unique Paths | O(rows * cols) | O(rows * cols) | Paths to a cell are the paths to the cell above plus the cell to the left. |

### String matching

Finding a pattern in a text without re-reading it.

| Entry | Time | Space | Key idea |
|-------|------|-------|----------|
| Aho-Corasick | O(n + patterns + matches) | O(total pattern length) | Find many patterns in one pass: a trie of the patterns with failure links, so a mismatch falls back without re-reading the text (1975). |
| Anagram | O(n) | O(1) | Whether two phrases use exactly the same letters, ignoring case and spaces. |
| Boyer-Moore-Horspool | O(n / m) best, O(n * m) worst | O(256) | A mismatch on a character not in the pattern skips the pattern's whole length. |
| Knuth-Morris-Pratt | O(n + m) | O(m) | A prefix table says how far the pattern can shift after a mismatch. |
| Manacher's Algorithm | O(n) | O(n) | The longest palindromic substring in linear time: grow each center's radius from its mirror inside the rightmost palindrome so far (1975). |
| Naive String Search | O(n * m) | O(1) | It is the baseline every faster matcher is measured against. |
| Rabin-Karp | O(n + m) average | O(1) | Sliding the window updates the hash in O(1). |
| Suffix Array | O(n log^2 n) | O(n) | Every suffix in sorted order by prefix doubling, with the LCP array of neighbors' common prefixes (Kasai). |
| Z-Algorithm | O(n + m) | O(n + m) | A window of the rightmost match lets most values be copied instead of recomputed. |

### Recursion

Functions defined in terms of themselves, and what that costs.

| Entry | Time | Space | Key idea |
|-------|------|-------|----------|
| Ackermann Function | faster than exponential | O(A(m, n)) | A total recursive function that grows faster than any primitive recursive one. |
| Factorial | O(n) | O(n) | Every recursion needs a base case that stops it and a step that gets closer to it. |
| Fibonacci (plain recursion) | O(1.618^n) | O(n) | The same subproblems are solved again and again, which is why it is exponential. |
| Fibonacci (memoized) | O(n) | O(n) | Caching subproblem answers turns exponential recursion into linear time. |
| Euclidean GCD | O(log min(a, b)) | O(log min(a, b)) | Any common divisor of a and b also divides a mod b, and the numbers shrink fast. |
| Least Common Multiple | O(log min(a, b)) | O(log min(a, b)) | The smallest number both divide: a / gcd(a, b) * b. |
| McCarthy 91 | O(101 - n) | O(101 - n) | n - 10 above 100, else M(M(n + 11)): 91 for every n <= 101. |
| Palindrome | O(n) | O(n) | Whether a word reads the same backward: its ends match and its middle is a palindrome. |
| Permutations | O(n * n!) | O(n * n!) | Choosing, recursing and un-choosing is the shape of every backtracking search. |
| Tower of Hanoi | O(2^n) | O(n) | To move n disks, move n - 1 out of the way, move the largest, then move the n - 1 back on top. |

### Graphs

Traversals, shortest paths, spanning trees, orderings and flows on weighted graphs.

| Entry | Time | Space | Key idea |
|-------|------|-------|----------|
| A* Search (graph) | O((V + E) log V) | O(V) | Uniform-cost search guided by the straight-line distance to the goal, from "# pos NAME X Y" lines; still the cheapest path when edges are at least as long as the straight line. Name the start node (other) and the end node (key); they default to the first and last nodes. |
| Articulation Points | O(V + E) | O(V) | The nodes whose removal disconnects the graph, from one low-link depth-first search (Tarjan, 1973). |
| Beam Search | O(d * k * b log(k * b)) | O(V) | Keep only the k most promising nodes of each level, by straight-line distance to the goal: little memory, no guarantee. Name the start node (other) and the end node (key); they default to the first and last nodes. |
| Bellman-Ford | O(V * E) | O(V) | A shortest path has at most n - 1 edges, so n - 1 rounds are enough. |
| Bidirectional BFS | O(b^(d/2)) | O(b^(d/2)) | Breadth-first search from both ends at once until the frontiers meet: about 2b^(d/2) nodes instead of b^d. Name the start node (other) and the end node (key); they default to the first and last nodes. |
| Bipartite Check | O(V + E) | O(V) | Two-color by BFS; an edge between same colors means an odd cycle. |
| Boruvka's Minimum Spanning Tree | O(E log V) | O(V) | Every component takes its cheapest outgoing edge at once, round after round (1926); the components at least halve each round. |
| Bridges | O(V + E) | O(V) | The edges whose removal disconnects the graph, from the same low-link search. |
| Critical Path (longest path in a DAG) | O(V + E) | O(V) | The heaviest path in a DAG of tasks with durations: no amount of parallel work finishes the project sooner. |
| Cycle Detection | O(V + E) | O(V) | In a directed graph, an edge back to a node still on the DFS path closes a cycle. |
| Shortest Paths in a DAG | O(V + E) | O(V) | Relax every edge in topological order: linear time, and negative weights are fine. Name the start node (other) and the end node (key); they default to the first and last nodes. |
| Dinic's Maximum Flow | O(V^2 * E) | O(V + E) | Level the residual graph breadth-first, push a blocking flow depth-first, repeat (1970); the same answer as Edmonds-Karp in fewer phases. Name the start node (other) and the end node (key); they default to the first and last nodes. |
| Euler Path (Hierholzer) | O(V + E) | O(V + E) | It exists exactly when zero or two nodes have odd degree, and the graph is connected. |
| Floyd-Warshall | O(V^3) | O(V^2) | The best path from i to j through nodes 1..k either uses k or it does not. |
| Breadth-First Traversal | O(V + E) | O(V) | A queue keeps the frontier in distance order. |
| Depth-First Traversal | O(V + E) | O(V) | Discovery and finish times reveal cycles, components and topological order. |
| Dijkstra's Shortest Paths | O((V + E) log V) | O(V) | The closest unfinished node can never get closer later when edges are non-negative. |
| Hamiltonian Path | O(n!) | O(n) | A path through every node exactly once, by backtracking; NP-complete, so small graphs only. |
| Hopcroft-Karp Bipartite Matching | O(E * sqrt V) | O(V) | The most pairs of a bipartite graph with no node in two pairs, augmenting along many shortest paths per phase (1973). |
| Iterative Deepening DFS | O(b^d) | O(d) | Depth-first searches limited to depth 0, 1, 2, ...: the memory of DFS with the fewest-edges path of BFS. Name the start node (other) and the end node (key); they default to the first and last nodes. |
| Johnson's All-Pairs Shortest Paths | O(V * E log V) | O(V^2) | Reweight with Bellman-Ford so every weight is non-negative, then Dijkstra from every node (1977); negative edges allowed. |
| k-Core Decomposition | O(V + E) | O(V + E) | Peel off the node of least degree, again and again: each node's core number is how deeply embedded it is (1983). |
| Kosaraju's Strongly Connected Components | O(V + E) | O(V + E) | Finish order from one depth-first search, then a second over the reversed edges collects each component (1978). |
| Kruskal's Minimum Spanning Tree | O(E log E) | O(V + E) | The cheapest edge across any cut is always safe to add. |
| Maximum Flow (Edmonds-Karp) | O(V * E^2) | O(V + E) | Residual edges let the algorithm undo earlier choices. |
| PageRank | O(k * (V + E)) | O(V) | A page is important if important pages link to it, solved by repeated averaging. |
| Prim's Minimum Spanning Tree | O(E log V) | O(V + E) | The same cut property as Kruskal, applied to the tree grown so far. |
| Stoer-Wagner Minimum Cut | O(V^3) | O(V^2) | The lightest set of edges that splits the graph in two, by merging the last two nodes of each maximum-adjacency phase (1997). |
| Tarjan's Strongly Connected Components | O(V + E) | O(V) | A node's low-link value finds the root of the component it belongs to. |
| Topological Sort | O(V + E) | O(V) | A node can go next once everything pointing to it has gone. |
| Transitive Closure (Warshall) | O(V^3) | O(V^2) | Which nodes reach which: j is reachable from i through k when i reaches k and k reaches j (1962). |
| Traveling Salesperson (Held-Karp) | O(2^n * n^2) | O(2^n * n) | The cheapest tour through every node, exactly, by dynamic programming over subsets (1962). |
| Traveling Salesperson (nearest neighbor) | O(n^2) | O(n^2) | Always go to the nearest unvisited node, then return: fast, but not always the cheapest tour. Name the start node (other) and the end node (key); they default to the first and last nodes. |
| Uniform-Cost Search | O((V + E) log V) | O(V) | Expand the cheapest frontier node first and stop when the goal is settled: Dijkstra's algorithm as a search. Name the start node (other) and the end node (key); they default to the first and last nodes. |

### Greedy algorithms

Taking the best-looking choice at every step, and when that is optimal.

| Entry | Time | Space | Key idea |
|-------|------|-------|----------|
| Activity Selection | O(n log n) | O(n) | Ending first leaves the most room for everything after it. |
| Fractional Knapsack | O(n log n) | O(n) | When items can be split, the best ratio first is optimal. |
| Greedy Coin Change | O(k log k + amount / smallest) | O(k) | Pay with the largest coin that fits, again and again; optimal for some coin systems and not for others. |
| Huffman Coding | O(n + k log k) | O(n + k) | Merging the two least frequent symbols is always part of an optimal code. |
| Interval Partitioning | O(n log n) | O(n) | Give every interval a room so overlapping ones never share, with the fewest rooms. |
| Job Sequencing with Deadlines | O(n log n + n * d) | O(d) | The most profit from unit-time jobs with deadlines: richest first, each in the latest free slot. |

### Backtracking

Trying choices one at a time and undoing the ones that fail.

| Entry | Time | Space | Key idea |
|-------|------|-------|----------|
| Combination Sum | exponential in target / smallest | O(target) | Every multiset of candidates (each reusable) that sums to the target. |
| Graph Coloring | O(m^n) worst | O(n) | Try a color, recurse, and back up when a neighbor already has it. |
| N-Queens | O(n!) worst | O(n) | Undo a choice as soon as it conflicts, so dead ends are cut off early. |
| Subset Sum (backtracking) | O(2^n) worst | O(n) | Take or leave each value in turn, backing up when the sum overshoots, until it hits the target. |
| Sudoku Solver | exponential worst, fast with pruning | O(81) | Fill the most constrained empty cell with each allowed digit, undoing it when the rest cannot be filled. |
| Word Search | O(cells * 4^len) | O(len) | Find a word as a path of neighboring letters, no cell twice, backing up at dead ends. |

### Compression

Making data smaller without losing any of it.

| Entry | Time | Space | Key idea |
|-------|------|-------|----------|
| Burrows-Wheeler Transform | O(n^2 log n) here, O(n) with suffix arrays | O(n) | Sort every rotation of the text and keep the last column: repeated contexts bunch equal characters together, and the transform can be undone. |
| LZ77 | O(n * window) | O(tokens) | Replace repeats with (offset, length) pointers back into a sliding window of what came before, plus the next character; the idea under gzip and PNG. |
| LZW | O(n) | O(n) | Emit the code of the longest phrase already in the dictionary and add that phrase plus the next character as a new code; GIF's compressor. |
| Move-to-Front | O(256 n) | O(n) | Replace each byte with its place in a list and move it to the front, so recently seen bytes become small numbers. |
| Run-Length Encoding | O(n) | O(runs) | Write each run of equal characters as the character and a count: great on long runs, twice the size with none. |

### Text algorithms

Algorithms on whole texts: suffix arrays, palindromes, many patterns at once.

| Entry | Time | Space | Key idea |
|-------|------|-------|----------|
| Autocomplete (trie) | O(prefix + matches) | O(total length) | Suggest the words that start with what was typed, most used first, by walking a trie of every word down to the prefix. |
| Distinct Substrings (suffix automaton) | O(n) | O(n) | Count every different substring with a suffix automaton, the smallest machine that recognizes all suffixes, built one character at a time. |
| Fuzzy Search (bitap) | O(n * k) | O(k) | Find a pattern with up to k typos: one machine word per error count tracks which prefixes still match, updated with a shift per character (Wu and Manber, 1992). |
| Longest Common Prefix | O(total length) | O(1) | The longest start every word shares, by trimming a candidate until each word begins with it. |
| Longest Common Substring | O(n * m) | O(m) | The longest run of characters two texts share, from a table of common runs ending at each pair of positions, one row at a time. |
| Longest Repeated Substring | O(n log^2 n) | O(n) | The longest piece of a text that occurs twice: sort the suffixes, and the answer is the longest common prefix of two neighbors. |
| Lyndon Factorization (Duval) | O(n) | O(1) | Split a text into non-increasing Lyndon words, each smaller than all its rotations, in one pass (Duval, 1983). |
| Lexicographically Minimal Rotation (Booth) | O(n) | O(n) | The rotation of a text that sorts first, in linear time, with a failure function over the text written twice (Booth, 1980). |
| Line Diff (Myers) | O((n + m) * d) | O((n + m) * d) | The fewest line insertions and deletions turning one text into another, exploring diagonals one edit at a time (Myers, 1986): the algorithm behind git diff. |
| Character N-Grams | O(n * len) | O(n * len) | Every run of n characters, overlapping: the building block of fuzzy matching, language detection and plagiarism checks. |
| N-Gram Similarity (Jaccard) | O(n * (len a + len b)) | O(n * (len a + len b)) | How alike two texts are, from 0 to 1: the n-grams they share over all the n-grams either has. |
| Smallest Period | O(n) | O(n) | The shortest block that repeats to give the text, from its longest border (the prefix function of Knuth-Morris-Pratt). |
| Soundex | O(n) | O(1) | A letter and three digits per name, so names that sound alike match: Robert and Rupert are both R163 (Russell and Odell, 1918). |
| Spelling Suggestions | O(d * n * m) | O(m) | The dictionary words within k edits of a misspelled word, closest first: Levenshtein distance that skips hopeless words early. |
| Text Justification | O(n) | O(n) | Wrap words greedily, then spread spaces between them so every line but the last is exactly the width, extra spaces going left first. |
| TF-IDF | O(total words + t log t) | O(distinct words) | Score each word of each document (one per line) by how often it appears there times how rare it is elsewhere: the classic relevance measure. |
| Word Frequency | O(n + w log w) | O(w) | Count every word, most frequent first: the first step of search engines, spam filters and word clouds. |
| Word Wrap (minimum raggedness) | O(w^2) | O(w) | Break lines to minimize the squared leftover space, ending a line early when it saves a ragged one later, by dynamic programming (the TeX idea). |
| Word Wrap (greedy) | O(n) | O(n) | Fill each line with as many words as fit, then start the next: what most editors do. |

## Mathematics

### Numbers

Primes, factors and powers.

| Entry | Time | Space | Key idea |
|-------|------|-------|----------|
| Base Conversion | O(n^2) | O(n) | Write a number of any length in another base from 2 to 36. |
| Binary GCD (Stein) | O(log^2 n) bit operations | O(1) | The greatest common divisor with only shifts and subtraction, taking out shared factors of two first. |
| Chinese Remainder Theorem | O(k log M) | O(1) | The x that leaves given remainders for given moduli, combining one congruence at a time; the moduli need not be coprime. |
| Collatz Sequence | O(steps) | O(1) | Halve even numbers, take 3n+1 of odd ones, until 1: every start tried so far gets there. |
| Divisors | O(sqrt n) | O(d(n)) | Every divisor of n, found in pairs d and n/d while d*d <= n. |
| Euler's Totient | O(sqrt n) | O(1) | How many of 1..n share no factor with n: n times (1 - 1/p) for each prime p dividing it. |
| Even | O(1) | O(1) | Whether n is even. |
| Extended Euclidean Algorithm | O(log min(a, b)) | O(1) | Run Euclid's algorithm backward to express the gcd as a combination. |
| Fast Exponentiation | O(log exp) | O(1) | x^n is (x^(n/2))^2 for even n, so it takes about log2 n multiplications. |
| Fibonacci (Fast Doubling) | O(log n) | O(1) | F(n) from F(2k) = F(k)(2F(k+1) - F(k)) and F(2k+1) = F(k)^2 + F(k+1)^2, reading the bits of n from the top. |
| Fibonacci (matrix power) | O(log n) | O(1) | F(n) from the nth power of [[1 1] [1 0]] by repeated squaring: O(log n) instead of n steps. |
| Karatsuba Multiplication | O(n^1.585) | O(n) | (a + b)(c + d) - ac - bd gives the middle term with one multiplication. |
| Miller-Rabin Primality | O(k log^3 n) | O(1) | Write n-1 = d * 2^s and test bases: a composite fails for one; with twelve bases it is exact for every 64-bit n. |
| Modular Inverse | O(log m) | O(1) | The x with a*x = 1 (mod m), from the extended Euclidean algorithm; none when a and m share a factor. |
| Integer Square Root (Newton) | O(log n) | O(1) | x = (x + n/x) / 2 until it stops shrinking: the digits right double each step. |
| Pascal's Triangle | O(rows^2) | O(1) | Each value is the sum of the two above it, so row n holds the binomial coefficients C(n, k). |
| Perfect Numbers | O(n log n) | O(n) | The numbers up to a limit equal to the sum of their smaller divisors (6 = 1 + 2 + 3), with a sieve over divisor sums. |
| Primality | O(sqrt n) | O(1) | If n = a * b, one of a and b is at most sqrt(n). |
| Prime Factorization | O(sqrt n) | O(log n) | The prime factors of n, by trial division. |
| Segmented Sieve | O((hi - lo) log log hi + sqrt hi) | O(sqrt hi) | The primes in a range [lo, hi], sieving one window of about sqrt(hi) numbers at a time with the primes up to sqrt(hi). |
| Sieve of Eratosthenes | O(n log log n) | O(n) | Start crossing out at p * p, because smaller multiples were crossed out by smaller primes. |

### Computational geometry

Points, lines and polygons: hulls, intersections, areas.

| Entry | Time | Space | Key idea |
|-------|------|-------|----------|
| Closest Pair of Points | O(n log^2 n) | O(n) | The two points nearest each other, by divide and conquer: solve each half, then check a narrow strip. |
| Convex Hull (Graham Scan) | O(n log n) | O(n) | Sort the points by angle around the lowest one, then keep a stack that only turns left (Graham, 1972). |
| Convex Hull (Jarvis March) | O(n * h) for h hull points | O(n) | Gift wrapping: from the lowest point, repeatedly take the point with every other to its left (Jarvis, 1973). |
| Convex Hull (Monotone Chain) | O(n log n) | O(n) | Sort by x, then build the lower and upper hulls with a left-turn stack (Andrew, 1979). |
| Point in Polygon | O(n) per point | O(1) | Whether points lie inside a polygon, by casting a ray to the right and counting the edges it crosses. |
| Polygon Area (shoelace) | O(n) | O(1) | The area of a polygon from its vertices in order, exactly, with the shoelace formula. |
| Segment Intersection | O(n^2) | O(k) | Which segments cross or touch, tested pair by pair with orientation (cross product) tests. |

### Bit manipulation

Tricks with the bits of an integer.

| Entry | Time | Space | Key idea |
|-------|------|-------|----------|
| Submask Enumeration | O(2^k) for k set bits | O(1) | Every submask of a mask with sub = (sub - 1) & mask, from the mask itself down to 0; the core of subset dynamic programming. |
| Gray Code | O(2^bits) | O(1) | Number the values so neighbors differ in one bit: x ^ (x >> 1), decoded by XOR-folding the shifts. |
| Population Count (Kernighan) | O(set bits) | O(1) | Count the set bits by clearing the lowest one with x & (x - 1) until none remain: one step per set bit. |
| Power of Two | O(1) | O(1) | x is a power of two when x & (x - 1) is zero; the next power of two comes from copying the top bit into every lower bit. |
| Reverse Bits | O(width) | O(1) | Reverse the lowest bits of x, moving one bit at a time from the bottom of x to the bottom of the result. |
| Single Number (XOR) | O(n) | O(1) | The one value without a pair: XOR everything, and every pair cancels itself (v ^ v = 0). |

### Numerical methods

Roots, integrals, linear systems and transforms, computed numerically.

| Entry | Time | Space | Key idea |
|-------|------|-------|----------|
| Bisection Method | O(log(width / tol)) | O(steps) | Halve an interval whose ends have opposite signs, keeping the half where the sign changes: slow, one bit a step, but sure. |
| Fast Fourier Transform | O(n log n) | O(n log n) | The frequencies in a signal, from the transforms of its even and odd positions combined with twiddle factors (Cooley-Tukey radix 2). |
| Gaussian Elimination | O(n^3) | O(n^2) | Solve A x = b: eliminate below each pivot, swapping in the largest pivot first (partial pivoting), then substitute back. |
| Horner's Rule | O(n) | O(1) | Evaluate a polynomial as a + x(b + x(c + ...)): one multiply and one add per coefficient, no powers. |
| Matrix Power | O(n^3 log k) | O(n^2) | A square matrix to the kth power by repeated squaring: O(log k) products. [[1 1] [1 0]]^n holds Fibonacci numbers. |
| Newton-Raphson Method | O(steps) | O(steps) | Follow the tangent line to where it crosses zero, x - f(x)/f'(x); near a root the correct digits double each step. |
| Secant Method | O(steps) | O(steps) | Newton's method with the derivative replaced by the slope through the last two guesses; no derivative needed. |
| Simpson's Rule | O(n) | O(1) | Integrate by fitting a parabola through each pair of slices (weights 1, 4, 2, ..., 4, 1); exact for cubics. |
| Strassen Multiplication | O(n^2.81) | O(n^2) | Multiply square matrices with seven block products instead of eight, recursively: under n^3. |
| Trapezoid Rule | O(n) | O(1) | Integrate by joining neighboring points with straight lines and adding the trapezoids' areas; exact for lines. |

### Randomized algorithms

Algorithms that flip coins, and what that buys.

| Entry | Time | Space | Key idea |
|-------|------|-------|----------|
| Karger's Minimum Cut | O(trials * m) | O(n + m) | Merge the ends of random edges until two groups remain; the edges between them are a cut, and the smallest over many trials is probably the minimum. |
| Monte Carlo Pi | O(n) | O(1) | Throw random points into a square and count those inside the quarter circle: pi is about 4 * inside / all. The error shrinks like 1/sqrt(n). |
| Randomized Quickselect | O(n) expected | O(n) | The kth smallest value by quickselect around random pivots: always exact (Las Vegas), linear time expected for every input. |
| Reservoir Sampling | O(n) | O(k) | Choose k values uniformly from a stream read once: keep the first k, then let the ith replace a random kept one with probability k/i. |

## Security

### Ciphers

Classical encryption, and why it is not enough.

| Entry | Time | Space | Key idea |
|-------|------|-------|----------|
| Affine Cipher | O(n) | O(n) | Map each letter x to (a*x + b) mod 26; a must share no factor with 26. |
| Atbash | O(n) | O(n) | Mirror the alphabet: A becomes Z, B becomes Y. Its own inverse. |
| Autokey Cipher | O(n) | O(n) | Vigenère with a key that runs on into the plaintext itself, so it never repeats. |
| Beaufort Cipher | O(n) | O(n) | Subtract each letter from the key letter (key - plain): the same operation encrypts and decrypts. |
| Caesar Brute Force | O(26n) | O(26n) | Try all 26 Caesar shifts and rank them by how English the letters look (chi-squared). |
| Caesar Cipher | O(n) | O(n) | Shift every letter forward by a fixed number of places (A becomes D for 3). |
| Columnar Transposition | O(n log k) | O(n) | Write the text in rows under a keyword, then read the columns in the keyword's alphabetical order. |
| Enigma Machine | O(n) | O(n) | The German Enigma I: three stepping rotors, a reflector and a plugboard; the setup is "rotors positions rings pairs". |
| Frequency Analysis | O(26n) | O(1) | Count the letters and pick the Caesar shift whose letter frequencies best match English. |
| Hill Cipher | O(n * k) | O(n) | Multiply blocks of letters by a key matrix modulo 26; decrypt with its inverse. |
| One-Time Pad | O(n) | O(n) | XOR the message with a truly random pad as long as it, used once: unbreakable, and impractical. |
| Playfair Cipher | O(n) | O(n) | Encrypt pairs of letters by their places in a 5x5 square built from a keyword (J counts as I). |
| Rail Fence Cipher | O(n) | O(n) | Write the text in a zigzag over a number of rails, then read the rails one after another. |
| ROT13 | O(n) | O(n) | Shift every letter by 13, so applying it twice gives the text back. |
| Substitution Cipher | O(n) | O(n) | Replace each letter by the one in its place in a scrambled alphabet: 26! keys, broken by letter frequencies. |
| Vigenère Cipher | O(n) | O(n) | Shift each letter by the letter of a repeating keyword under it. |
| Vigenère Crack | O(n * max) | O(n) | Break a Vigenère cipher: find the key length, then each key letter as the best Caesar shift of its column. |
| Vigenère Key Length | O(n * max + r^2) | O(n) | Find a Vigenère key's length from repeated trigrams (Kasiski) and the columns' index of coincidence (Friedman). |

### Cryptography

How modern cryptography works inside: hashes, block and stream ciphers, public keys, signatures, secret sharing. For learning, never for protecting data.

| Entry | Time | Space | Key idea |
|-------|------|-------|----------|
| AES-128 | O(1) | O(1) | Encrypt one 16-byte block (text, in hex) with a 16-byte key: ten rounds of SubBytes, ShiftRows, MixColumns, AddRoundKey. |
| Base32 | O(n) | O(n) | Write bytes as A-Z and 2-7, 5 bytes to 8 characters: case-insensitive, used in TOTP secrets. |
| Base64 | O(n) | O(n) | Write bytes as text: every 3 bytes become 4 of 64 safe characters. An encoding, not encryption. |
| ChaCha20 | O(n) | O(n) | A stream cipher of additions, XORs and rotations; the key field is "KEY NONCE" in hex. Shows the state after each double round. |
| Constant-Time Compare | O(n) | O(1) | Compare two secrets looking at every byte, so the time taken leaks nothing about where they differ. |
| Diffie-Hellman | O(log p) | O(1) | g^(ab) is easy for both sides to compute and hard for an eavesdropper. |
| ElGamal | O(log p) | O(1) | Encrypt a number with a public key y = g^x mod p and a one-time k, then decrypt it with the secret x. |
| Elliptic Curve | O(p^2) to list, O(log k) to multiply | O(p) | Points on y^2 = x^3 + ax + b mod p: list them, and compute k*G by double-and-add. |
| Feistel Network | O(rounds) | O(rounds) | Split a 32-bit block in halves; each round swaps them and mixes one into the other. Decrypting is the same in reverse. |
| Hex Encoding | O(n) | O(n) | Write each byte as two hex digits, 0-9 and a-f. |
| HMAC-SHA256 | O(n) | O(n) | Authenticate a message with a key: two nested SHA-256 hashes, safe where hash(key + message) is not. |
| Linear-Feedback Shift Register | O(n * taps) | O(n) | Shift a register right and feed the XOR of its tapped bits back in at the top: a fast, predictable bit stream. |
| MD5 | O(n) | O(n) | The 128-bit hash of the 1990s, now broken: collisions take seconds. Shown to learn from, never to use. |
| Merkle Tree | O(n) | O(n) | Hash blocks (one a line) pairwise up to one root, and prove a block is in the tree with log n hashes. |
| PBKDF2 | O(iterations * length) | O(length) | Stretch a password (the text) with a salt (the key) into a key, with many HMAC rounds to slow down guessing. |
| RSA | O(bits^4) to make the key | O(bits) | Exponents e and d undo each other modulo n, and finding d requires factoring n. |
| SHA-256 | O(n) | O(n) | Each round mixes bits with additions, rotations and constants until every output bit depends on every input bit. |
| Shamir Secret Sharing | O(n k) | O(n) | Split a secret into n shares so any k rebuild it and k-1 reveal nothing: points on a random polynomial. |

### Hashing

Hash functions and the ways a hash table resolves collisions.

| Entry | Time | Space | Key idea |
|-------|------|-------|----------|
| Adler-32 | O(n) | O(1) | Two running sums modulo 65521, the second summing the first (zlib's checksum). |
| Consistent Hashing | O(log n) per lookup | O(n * replicas) | Adding or removing a server moves only the keys next to it. |
| CRC-32 | O(n) | O(1) | The remainder of the data as a polynomial over a fixed one, a table lookup a byte: catches accidental corruption. |
| Cuckoo Hashing | O(1) lookups, O(1) amortized inserts | O(n) | Lookups check at most two places, so they are O(1) in the worst case. |
| DJB2 | O(n) | O(1) | Daniel Bernstein's h = h*33 + byte, from 5381. |
| Double Hashing | O(1) average at a bounded load | O(n) | On a collision step by a second hash of the key, so keys that collide once usually part ways. Deletes leave tombstones. |
| FNV-1a (32 and 64 bit) | O(n) | O(1) | For each byte, XOR it in and multiply by the FNV prime: simple, fast and well spread. The 64-bit hash is FNV1a64. |
| Hash Distribution | O(keys + buckets) | O(buckets) | How evenly a hash spreads keys over buckets: counts per bucket and a chi-square near buckets - 1 for a good hash. |
| Linear Probing | O(1) average at a bounded load | O(n) | Neighboring slots share cache lines, so probing is fast in practice. |
| MurmurHash3 (32 bit) | O(n) | O(1) | Mix four bytes at a time with multiplies and rotations, then avalanche: every input bit affects every output bit. |
| Polynomial Rolling Hash | O(n) | O(1) | sum of byte * base^position mod a prime: the hash a sliding window updates in O(1). |
| Quadratic Probing | O(1) average at a bounded load | O(n) | On a collision try 1, 4, 9, ... slots on: breaks up runs; the table stays at most half full. Deletes leave tombstones. |
| Robin Hood Hashing | O(1) expected | O(n) | Linear probing where a key far from home takes the slot of a key nearer home, keeping every probe distance close to the average. |
| SDBM | O(n) | O(1) | The sdbm database's h = byte + (h << 6) + (h << 16) - h. |

## Data structures

### Arrays

The array underneath everything: growth, shifting, and the classic techniques.

| Entry | Time | Space | Key idea |
|-------|------|-------|----------|
| Dynamic Array | O(1) amortized append and pop, O(1) get and set, O(n) insert and delete | O(n) | Doubling makes the copying cost O(1) amortized per append. |
| JavaScript Array | O(1) push, pop, at; O(n) shift, unshift, splice | O(n) | Every JavaScript array method on a live array, with JS semantics: negative indexes, clamped ranges, mutating and copying methods. |
| Array.prototype.every | O(n) | O(n) | Whether the callback accepts every element; it stops at the first refusal (true for []). The callback, by name (even, greater, negative, odd, positive), goes in the keyword. |
| Array.prototype.filter | O(n) | O(n) | A new array of the elements the callback keeps. The callback, by name (even, greater, negative, odd, positive), goes in the keyword. |
| Array.prototype.find | O(n) | O(n) | The first element the callback accepts, stopping there (undefined if none; here ok is false). The callback, by name (even, greater, negative, odd, positive), goes in the keyword. |
| Array.prototype.findIndex | O(n) | O(n) | The index of the first element the callback accepts, or -1. The callback, by name (even, greater, negative, odd, positive), goes in the keyword. |
| Array.prototype.flatMap | O(n) | O(n) | Map each element to an array, then join them one level deep. The callback, by name (duplicate, range), goes in the keyword. |
| Array.prototype.map | O(n) | O(n) | It transforms without changing the original. |
| Array.prototype.reduce | O(n) | O(n) | Each step combines the accumulator with the next element. |
| Array.prototype.slice | O(end - start) | O(end - start) | A copy of the elements from start up to (not including) end; negative indexes count from the end, and both are clamped. The array is not changed. |
| Array.prototype.some | O(n) | O(n) | Whether the callback accepts at least one element; it stops at the first (false for []). The callback, by name (even, greater, negative, odd, positive), goes in the keyword. |
| Array.prototype.sort | O(n log n) | O(n) | JavaScript sorts by string form unless given a compare function: [1, 10, 2].sort() stays [1, 10, 2]. Keyword "numeric" sorts as numbers; "default" (or empty) as JS does. |
| Array.prototype.splice | O(n) | O(n) | Remove count elements from start (negative counts from the end) and insert the new items there, in place; returns what was removed. The array goes in the first field and the items to insert in the second. |
| Maximum Subarray (Kadane) | O(n) | O(1) | A run worth extending never has a negative prefix, so restart when the sum drops below zero. |
| Prefix Sums | O(n) | O(1) | sum(i..j) = prefix[j + 1] - prefix[i]. |
| Reverse (two pointers) | O(n) | O(1) | Swap the ends and walk two pointers inward. |
| Rotate (three reversals) | O(n) | O(1) | Reversing twice puts each part back in order at its new place. |
| Sliding Window Maximum | O(n) | O(1) | A deque keeps only values that could still be a maximum. |
| Two Pointers (pair sum) | O(n) | O(1) | If the sum is too small move the left pointer up; if too large move the right down. |

### Sets

Collections without duplicates: bitsets, multisets, ordered sets and the set operations.

| Entry | Time | Space | Key idea |
|-------|------|-------|----------|
| Bitset | O(1) set, clear, test; O(n/64) count | O(max / 64) | Union and intersection become single machine instructions on whole words. |
| Cartesian Product | O(n * m) | O(n * m) | Every pair (x, y) with x from the first set and y from the second. |
| Jaccard Similarity | O(n + m) | O(n + m) | How alike two sets are: the size of their intersection over the size of their union. |
| Multiset (Bag) | O(1) average add, remove, count | O(distinct) | A set where a value can occur more than once: each value with a count. |
| Ordered Set | O(log n) search, O(n) insert | O(n) | A sorted set: binary search for membership, rank, floor, ceiling and ranges. |
| Power Set | O(n * 2^n) | O(n * 2^n) | Each subset is a binary number saying which members are in. |
| Roaring Bitmap | O(log c) to find a group, then O(log 4096) or O(1) | about 2 bytes per sparse member, 8 KB per dense group | A compressed set of 32-bit integers: grouped by their high half, each group a sorted array while small and a bitmap once dense. |
| Difference (two pointers) | O(n + m) | O(n + m) | The values in the first set but not the second. |
| Intersection (two pointers) | O(n + m) | O(n + m) | The values in both sets, walking both sorted sets at once. |
| Symmetric Difference (two pointers) | O(n + m) | O(n + m) | The values in exactly one of the two sets. |
| Union (two pointers) | O(n + m) | O(n + m) | The values in either set, walking both sorted sets at once like a merge. |
| Sparse Set | O(1) add, remove, check and clear | O(universe) | Small integers in two arrays pointing at each other: add, remove, check and even clear in O(1), with no array ever zeroed. |

### Trees

Walking, measuring and building trees.

| Entry | Time | Space | Key idea |
|-------|------|-------|----------|
| Build a Tree from Two Traversals | O(n) | O(n) | Rebuild a tree from its preorder and inorder values: the first preorder value is the root, and it splits the inorder list in two. |
| Expression Tree | O(n) | O(n) | Parse arithmetic into a tree with the shunting-yard algorithm, evaluate it in postorder, and print it in prefix and postfix. |
| Invert a Tree | O(n) | O(h) | Mirror the tree by swapping every node's children. |
| Lowest Common Ancestor | O(n) | O(h) | Lift both nodes to the same depth, then lift them together until they meet. |
| Morris Inorder Traversal | O(n) | O(1) | Inorder with no stack: thread each node's predecessor back to it, follow the threads, then remove them (Morris, 1979). |
| Path Sum | O(n * h) | O(h) | Every root-to-leaf path whose values add up to a target, by depth-first search keeping the current path. |
| Serialize and Deserialize | O(n) | O(n) | Write a tree as its level order with null for missing children, and read it back: the same tree, node for node. |
| Height, Diameter and Balance | O(n) | O(h) | The longest path through a node is its two deepest subtrees joined. |
| Inorder Traversal | O(n) | O(h) | On a binary search tree it yields the keys in sorted order. |
| Level-Order Traversal | O(n) | O(w) for the widest level | The queue holds one level while the next is being added. |
| Postorder Traversal | O(n) | O(h) | Visit both subtrees, then the root: the order to delete a tree or evaluate an expression. |
| Preorder Traversal | O(n) | O(h) | Visit the root, then the left subtree, then the right: the order to copy a tree. |
| Validate a Binary Search Tree | O(n) | O(h) | Each node must respect all of its ancestors, not only its parent. |

### Collections (lodash-style)

The utilities every codebase rewrites: deep merge, deep clone, path get and set, group by, chunk, zip, debounce.

| Entry | Time | Space | Key idea |
|-------|------|-------|----------|
| Object.assign | O(keys) | O(keys) | Copy the second object's keys onto a copy of the first, one level only: unlike deep merge, a nested object is replaced, not merged. |
| camelCase | O(n) | O(n) | Write text as camelCase, from its words. |
| Chunk (_.chunk) | O(n) | O(n / size) | Split the values into groups of a size; the last may be shorter. |
| Count By (_.countBy) | O(n) | O(n) | Count the words or numbers by a callback. Callback: even-odd, sign, mod3 (numbers), length, first-letter (words). |
| Debounce (_.debounce) | O(calls) | O(calls) | Calls at the given milliseconds: f runs once a burst has been quiet for wait (and at its start with leading). |
| Deep Clone (_.cloneDeep) | O(size) | O(size) | Copy a JSON document so the copy shares no object or array with the original. |
| Deep Merge (_.merge) | O(total size) | O(total size) | Merge the second JSON document into a copy of the first: objects key by key, arrays index by index, null never overwrites. |
| JavaScript Map | O(1) set, get, has, delete | O(n) | A map that remembers insertion order, as JavaScript's Map does (Go's map has none): set keeps a key's place, delete and set moves it to the end. |
| JavaScript Set | O(1) add, has, delete; O(n + m) set methods | O(n) | A set in insertion order with the ES2025 set methods: union, intersection, difference, symmetricDifference, isSubsetOf, isSupersetOf, isDisjointFrom. |
| Escape HTML (_.escape) | O(n) | O(n) | Replace & < > " ' with HTML entities, so text is shown, not run, in a page. |
| Flatten Deep (_.flattenDeep) | O(size) | O(size) | Flatten every level of a nested JSON array, as flat(Infinity) does. |
| Flatten Object | O(size) | O(size) | Turn nested JSON into one level of paths to its leaves (a.b[0].c), and back with unflatten-object. |
| Object.fromEntries | O(n) | O(n) | Build an object from [key, value] pairs; a later pair wins a repeated key. |
| Get by Path (_.get) | O(path length) | O(1) | Read the value at a path like a.b[0].c in a JSON document, or a default when any step is missing. |
| Group By (_.groupBy) | O(n) | O(n) | Group the words or numbers by a callback, each group in order. Callback: even-odd, sign, mod3 (numbers), length, first-letter (words). |
| Deep Equality (_.isEqual) | O(size) | O(depth) | Whether two JSON documents are deeply equal, and the path of the first difference. |
| Math.round vs Go's math.Round | O(n) | O(1) | Round each number both ways: JavaScript sends halves up (-2.5 to -2), Go away from zero (-2.5 to -3). |
| String at, charAt, codePointAt | O(n) | O(n) | Read one character: at counts negatives from the end, charAt does not, codePointAt gives its number. |
| String slice vs substring | O(n) | O(n) | JavaScript's two ways to cut a string: slice counts negatives from the end; substring treats them as 0 and swaps start and end. |
| String split | O(n) | O(n) | Split text at a separator (the second text; empty splits into characters), keeping at most limit pieces as JavaScript does (Go's SplitN keeps the rest instead). |
| kebab-case | O(n) | O(n) | Write text as kebab-case, from its words. |
| Key By (_.keyBy) | O(n) | O(n) | Index a JSON array of objects by a field: a later object wins a repeated key. |
| Memoize (_.memoize) | O(n) memoized, O(phi^n) plain | O(n) | Count the calls naive recursive Fibonacci makes, and the same function memoized: exponential against linear. |
| Object.groupBy | O(n) | O(n) | Group a JSON array of objects by the value of one field (the second text), as Object.groupBy does. |
| Omit (_.omit) | O(size) | O(size) | A copy of the object without the given paths (comma-separated). |
| Pad, padStart, padEnd | O(length) | O(length) | Pad text to a length on both sides, the start and the end, with the characters in the key field (a space when empty). |
| Partition (_.partition) | O(n) | O(n) | Split the values into those the callback accepts and the rest, both in order. Callback: even, negative, odd, positive. |
| Pick (_.pick) | O(size) | O(size) | A new object of only the given paths (comma-separated). |
| Range (_.range) | O(n) | O(n) | start, start+step, ... up to but not including end; give start, end and step (step 0 means 1 toward end). |
| Sample (_.sample) | O(1) | O(1) | One value picked at random, from a seed. |
| Set by Path (_.set) | O(size) | O(size) | Write a JSON value (the key field) at a path in a copy of the document, creating the objects and arrays on the way. |
| Shuffle (Fisher-Yates) | O(n) | O(n) | A random order of the values, every order equally likely: swap each position with a random earlier one, from the end. |
| snake_case | O(n) | O(n) | Write text as snake_case, from its words. |
| Order By (_.orderBy) | O(k n log n) | O(n) | Sort a JSON array of objects by several fields, each "asc" or "desc" ("user asc, age desc"), stable. |
| Start Case | O(n) | O(n) | Write text as Start Case, each word capitalized. |
| Sum, Mean, Median | O(n log n) | O(n) | The sum, the mean and the median (middle value) of the values. |
| Throttle (_.throttle) | O(calls) | O(calls) | Calls at the given milliseconds: f runs at most once per wait, at the start of a window and, with trailing, at its end. |
| Number() conversion | O(n) | O(1) | Convert each line the way JavaScript's Number() does, beside what Go's strconv.ParseFloat says: they disagree on "", hex, "1_000" and "inf". |
| Truncate (_.truncate) | O(n) | O(n) | Shorten text to a length ending in ...; with a separator (the second text, e.g. a space) it cuts at the last one that fits. |
| Unflatten Object | O(size) | O(size) | Rebuild nested JSON from an object of paths to values. |
| Uniq By (_.uniqBy) | O(n) | O(n) | Drop the values whose key a value earlier in the list already had. Callback: abs, identity, mod10. |
| Unset by Path (_.unset) | O(size) | O(size) | Remove the value at a path from a copy of the document. |
| Words (_.words) | O(n) | O(n) | Split text into words at spaces and punctuation, case changes, acronyms and digits: what every case conversion starts from. |
| Zip (_.zip) | O(n) | O(n) | Pair the i-th word of the first list with the i-th of the second; a shorter list gives empty strings. |

### Linear structures

Stacks, queues, deques, lists and buffers: values in a line, and the LIFO and FIFO principles.

| Entry | Time | Space | Key idea |
|-------|------|-------|----------|
| Access Orders: LIFO, FIFO, Priority | O(n log n) | O(n) | Put the same values into a stack, a queue and a min-heap, and see the order each gives them back. |
| Circular Buffer | O(1) push and pop | O(capacity) | Head and tail indices modulo the size make a queue with no shifting. |
| Deque | O(1) at both ends | O(n) | It is a stack and a queue at once. |
| Doubly Linked List | O(1) at both ends, O(n) search | O(n) | Removing a node you hold is O(1) because you can reach its predecessor. |
| Singly Linked List | O(1) append and prepend, O(n) search and remove | O(n) | Inserting or deleting next to a node you hold is O(1), with no shifting. |
| Min Stack | O(1) push, pop and min | O(n) | Each entry remembers the minimum at the time it was pushed. |
| Monotonic Queue | O(1) amortized enqueue, dequeue and max | O(n) | A newer, larger value makes every older smaller one useless for maxima. |
| Queue (FIFO) | O(1) enqueue, dequeue and peek | O(n) | Things are served in the order they arrived. |
| Sorted List | O(log n) find, O(n) insert and remove | O(n) | Values kept in order as they are inserted, found by binary search. |
| Stack (LIFO) | O(1) push, pop and peek | O(n) | Only the top is reachable, which matches nested work such as function calls. |

### Heaps

Priority queues: the most urgent value first.

| Entry | Time | Space | Key idea |
|-------|------|-------|----------|
| Priority Queue (any order) | O(log n) push and pop, O(1) top | O(n) | A binary heap ordered by a function you give it: smallest first, largest first, or by deadline. |
| Min-Heap (priority queue) | O(log n) push and pop, O(1) top | O(n) | Only the root is guaranteed smallest, so keeping that property costs just O(log n). |

### Search trees

Ordered values in a tree that stays shallow: BST, AVL, red-black, B-tree, treap, splay, skip list.

| Entry | Time | Space | Key idea |
|-------|------|-------|----------|
| AVL Tree | O(log n) insert, remove and search | O(n) | Balance keeps the height at about 1.44 log n, so every operation is O(log n). |
| B-Tree | O(t log_t n) insert, remove and search | O(n) | Wide nodes make the tree shallow, so a lookup reads only a few disk pages. |
| Binary Search Tree | O(h) for height h: O(log n) balanced, O(n) worst | O(n) | Each comparison chooses one subtree, so search follows one path. |
| Interval Tree | O(h) insert and remove, O(h + k) to report k overlaps | O(n) | A search tree of intervals where each node knows the latest end below it, so finding every clash with a booking skips whole subtrees. |
| Red-Black Tree | O(log n) insert, remove and search | O(n) | No red node has a red child, so the longest path is at most twice the shortest. |
| Skip List | O(log n) expected insert, remove and search | O(n) expected | Each level skips about half the nodes below, giving O(log n) expected search. |
| Splay Tree | O(log n) amortized insert, remove and search | O(n) | A search tree that moves every value it touches to the root: no balance bookkeeping, O(log n) amortized, and recently used values stay on top. |
| Treap | O(log n) expected insert, remove and search | O(n) | A search tree by value and a heap by random priority: balanced in expectation, with no bookkeeping. |

### Range queries

The sum or minimum of any range of an array, fast.

| Entry | Time | Space | Key idea |
|-------|------|-------|----------|
| Disjoint Sparse Table | O(1) range query after O(n log n) build | O(n log n) | Answers for every block around every midpoint, so the sum, product, min or gcd of any range is two lookups, no overlap needed. |
| Fenwick Tree | O(log n) update and prefix sum | O(n) | The lowest set bit of an index says how many values that slot covers. |
| Segment Tree | O(log n) update and range query, O(n) build | O(n) | Each node stores the answer for its range, combined from its two halves. |
| Sparse Table | O(1) range minimum after O(n log n) build | O(n log n) | The minimum of every run of 2^j values, so any range minimum is two lookups. |

### Lookup and caches

Finding by key: hash tables, sets, tries, union-find, and caches that evict.

| Entry | Time | Space | Key idea |
|-------|------|-------|----------|
| Hash Table | O(1) average put, get and remove | O(n) | A good hash spreads keys evenly, so each slot holds about one key. |
| LFU Cache | O(1) get and put | O(capacity) | Evicts the entry used the fewest times (the least recent of those on a tie), in O(1). |
| LRU Cache | O(1) get and put | O(capacity) | A hash map finds items and a linked list keeps them in use order, both O(1). |
| Set | O(1) average add, remove and contains | O(n) | Each value at most once; adding one already there changes nothing. |
| Trie | O(length) insert, search and remove | O(total letters) | Keys with the same prefix share a path, so prefix queries are fast. |
| TTL Cache | O(1) get and put, O(n) when a put sweeps expired entries | O(capacity) | An LRU cache whose entries also expire a fixed time after they were written, like a DNS or session cache; time is a clock you advance. |
| Union-Find | O(alpha(n)) find and union | O(n) | Path compression and union by rank make operations nearly O(1). |

### Probabilistic structures

Approximate answers in a fraction of the memory: Bloom and cuckoo filters, Count-Min, HyperLogLog.

| Entry | Time | Space | Key idea |
|-------|------|-------|----------|
| Bloom Filter | O(k) add and check | O(m) bits | Several hashes set bits; a member has all of them set. |
| Count-Min Sketch | O(depth) add and estimate | O(width * depth) | Counts items in a stream in fixed memory: a few rows of counters, one hash each, and the smallest counter wins; never an underestimate. |
| Cuckoo Filter | O(1) check and remove, O(1) amortized add | O(buckets) | A membership filter like Bloom's that can also delete: fingerprints in one of two buckets, kicking residents to their other bucket when full. |
| HyperLogLog | O(1) add, O(m) count | O(m) registers | Estimates how many distinct items a stream has from the longest run of leading zeros in their hashes, within a few percent in a few kilobytes. |

### Spatial structures

Points in the plane, found by where they are: k-d trees and quadtrees.

| Entry | Time | Space | Key idea |
|-------|------|-------|----------|
| k-d Tree | O(h) insert, about O(log n) nearest on spread-out points | O(n) | A binary tree of points that splits on x, then y, then x: the nearest point to a query without checking every one. |
| Quadtree | O(depth) insert, query in proportion to the cells overlapped | O(n) | Splits a square into four quarters whenever one gets crowded, so a region query looks only at the cells it overlaps. |

### Persistent structures

Structures that never change: every update is a new version sharing the old one's parts.

| Entry | Time | Space | Key idea |
|-------|------|-------|----------|
| Persistent Stack | O(1) push and pop | O(1) per push, shared between versions | An immutable stack: push and pop make new versions and every old one still works, sharing its nodes instead of copying them. |
| Persistent Search Tree | O(h) insert, remove and search | O(h) new nodes per update | An immutable search tree: an insert copies only the path from the root and shares the rest, so every version stays usable. |

### Text buffers

How an editor holds the text you type: gap buffers and ropes.

| Entry | Time | Space | Key idea |
|-------|------|-------|----------|
| Gap Buffer | O(1) amortized typing at the cursor, O(d) to move it d places | O(n) | Text in one array with free space at the cursor: typing fills the gap, and only moving the cursor shifts text. |
| Rope | O(depth) insert, delete and index | O(n) | Text as a tree of short strings with left-subtree lengths, so an edit in the middle splits and joins trees instead of copying the text. |

## Operating systems

### CPU scheduling

How an operating system decides which process runs next.

| Entry | Time | Space | Key idea |
|-------|------|-------|----------|
| Earliest Deadline First | O(H * n) for a hyperperiod of H | O(H) | Periodic real-time tasks: always run the job whose deadline is nearest; optimal on one CPU when utilization is at most 1. |
| First Come, First Served | O(T * n) for T ticks | O(T) | Simple and fair in arrival order, with no preemption. |
| Lottery Scheduling | O(T * n) | O(T) | Each tick, draw a ticket at random: a process gets the CPU in proportion to its tickets (the priority column). |
| Multilevel Feedback Queue | O(T * n) for T ticks | O(T) | Round-robin queues by urgency; using a whole quantum moves a process down, so short jobs finish fast. |
| Priority Scheduling | O(T * n) for T ticks | O(T) | Importance, not arrival, decides who runs. |
| Rate-Monotonic Scheduling | O(H * n) for a hyperperiod of H | O(H) | Periodic real-time tasks with fixed priorities, shorter period first; safe below the Liu-Layland bound. |
| Round Robin | O(T * n) for T ticks | O(T) | No process waits more than (n - 1) quanta for its next turn. |
| Shortest Job First | O(T * n) for T ticks | O(T) | Short jobs first minimizes the average waiting time. |
| Shortest Remaining Time First | O(T * n) for T ticks | O(T) | Always running the job closest to finishing keeps average waiting lowest. |

### Memory management

Which page to evict, and where a process fits.

| Entry | Time | Space | Key idea |
|-------|------|-------|----------|
| Address Translation (paging and TLB) | O(n * TLB size) | O(n) | Split virtual addresses into page and offset, look the page up (in a TLB first), and see faults and TLB hits. |
| Aging Page Replacement | O(n * frames) | O(n * frames) | An 8-bit counter per frame, shifted right on every reference with the used frame's top bit set: LRU on a budget. |
| Best Fit | O(n * m) | O(n * m) | The tightest fit leaves the least waste in that block. |
| Buddy Allocator | O(log(size / smallest block)) alloc and free | O(blocks) | Round each request up to a power of two, split bigger blocks in halves to fit, and merge buddies back when both are free. |
| Clock (Second Chance) | O(n * frames) | O(n * frames) | One bit per page approximates LRU cheaply. |
| FIFO Page Replacement | O(n * frames) | O(n * frames) | A queue of pages is all it needs. |
| First Fit | O(n * m) | O(n * m) | Stopping at the first fit keeps allocation fast. |
| LFU Page Replacement | O(n * frames) | O(n * frames) | Evict the page used least often since it was loaded; keeps popular pages, adapts slowly. |
| LRU Page Replacement | O(n * frames) | O(n * frames) | The recent past predicts the near future. |
| Optimal Page Replacement | O(n * frames) | O(n * frames) | It sees the future, so no policy can do better. |
| Worst Fit | O(n * m) | O(n * m) | Place each process in the largest free partition, so the leftover stays useful. |

### Disk scheduling

In what order a disk's head serves its requests.

| Entry | Time | Space | Key idea |
|-------|------|-------|----------|
| C-LOOK | O(n log n) | O(n) | Serving in one direction only gives more uniform waiting times. |
| C-SCAN | O(n log n) | O(n) | Sweep one way to the end, jump back to the start, sweep again: more even waits. |
| FCFS Disk Scheduling | O(n) | O(n) | Serve requests in arrival order: fair, but the head swings back and forth. |
| LOOK | O(n log n) | O(n) | SCAN that turns at the last request instead of the disk's end. |
| SCAN (Elevator) | O(n log n) | O(n) | Every request is reached within two sweeps, so none starve. |
| Shortest Seek Time First | O(n^2) | O(n) | Minimizing each seek reduces total head movement. |

### Deadlock

Granting resources only when everyone can still finish.

| Entry | Time | Space | Key idea |
|-------|------|-------|----------|
| Banker's Algorithm | O(n^2 * m) | O(n * m) | Find an order in which every process can finish with what is free, or show the state is unsafe (Dijkstra, 1965). |

### Concurrency

Threads, races, locks, and the classic synchronization problems.

| Entry | Time | Space | Key idea |
|-------|------|-------|----------|
| Barrier | O(ticks) | O(ticks) | Threads work, then wait at a barrier until all arrive, round after round: the fast wait for the slow. |
| Dining Philosophers | O(n * meals) | O(n * meals) | Philosophers need both neighboring forks to eat; picking up the left fork first can deadlock, ordering the forks cannot. |
| Peterson's Algorithm | O(ticks) | O(ticks) | Mutual exclusion for two threads with only shared flags and a turn; a check-then-set version lets both in. |
| Producer-Consumer | O(items) | O(items) | Producers fill a bounded buffer and consumers empty it; each waits when the buffer is full or empty. |
| Race Condition | O(threads * increments) | O(threads * increments) | Threads add to a shared counter in three steps (read, add, write); without a lock, updates are lost. |
| Readers-Writers | O(ticks) | O(ticks) | Readers share the data, a writer needs it alone; with reader preference writers can starve. |
| Counting Semaphore | O(ticks) | O(ticks) | Threads wait (P) for one of a few slots, work, and signal (V); never more inside than slots. |
| Sleeping Barber | O(ticks) | O(ticks) | A barber sleeps until a customer arrives; customers wait in a few chairs or leave when they are full. |

### CPU caches

How a CPU cache maps addresses to lines, and which line to evict.

| Entry | Time | Space | Key idea |
|-------|------|-------|----------|
| Average Memory Access Time | O(1) | O(1) | AMAT = hit time + miss rate x miss penalty, for one cache level or two. |
| CPU Cache Simulator | O(n * ways) | O(lines) | Run an address trace through a direct-mapped, set-associative or fully associative cache, and see every hit, miss and eviction. |

### Processes and threads

What the OS does with a running program: states, fork and exec, system calls, signals, threads.

| Entry | Time | Space | Key idea |
|-------|------|-------|----------|
| fork, exec, wait and exit | O(n * p) | O(p) | Build a process tree with fork and exec, and watch exit leave zombies until a parent waits, and orphans adopted by init. |
| Process Lifecycle (five states) | O(n) | O(n) | Drive a process through new, ready, running, waiting and terminated, and see which events each state allows. |
| Signals and Handlers | O(n) | O(signals) | Install handlers, block and ignore signals, and see why SIGKILL always wins. |
| System Calls (user and kernel mode) | O(ticks) | O(ticks) | Run a program's work and system calls, and see every trap into kernel mode and the time it costs. |
| Thread Models (many-to-one, one-to-one, many-to-many) | O(ticks * threads) | O(ticks * threads) | Map user threads onto kernel threads three ways, and see what one blocking system call does to the others. |
| Threads Share, Processes Copy | O(workers * increments) | O(workers) | Workers add to one counter: as threads they share it, as forked processes each changes its own copy. |

### Synchronization

Taking turns safely: Peterson, the bakery, spinlocks, semaphores, condition variables, and the classic problems.

| Entry | Time | Space | Key idea |
|-------|------|-------|----------|
| The ABA Problem | O(1) | O(1) | A lock-free stack's compare-and-swap succeeds on a pointer that changed and changed back; a tag counter catches it. |
| Cigarette Smokers | O(rounds) | O(rounds) | An agent puts out two of three ingredients; grabbing them one at a time deadlocks, a pusher that sees the pair does not. |
| Compare-and-Swap Spinlock | O(threads * rounds * cs) | O(threads) | A spinlock on compare-and-swap, the instruction lock-free code is built on, counting the spins a waiter burns. |
| Condition Variables: while, not if | O(items * consumers) | O(consumers) | Consumers wait on a condition variable for items; waiting in an if lets a woken consumer take from an empty buffer. |
| Lamport's Bakery Algorithm | O(threads^2 * rounds) | O(threads) | Mutual exclusion for any number of threads with only reads and writes: take a ticket, wait for every smaller one. |
| Peterson's Algorithm, Model Checked | O(states) | O(states) | Check every interleaving of two threads: Peterson's algorithm is safe and live; the two simpler attempts are not, and here is how they fail. |
| Priority Inversion and Inheritance | O(ticks) | O(ticks) | A low-priority task holds a lock the high one needs while a medium one hogs the CPU: the Mars Pathfinder bug, and its fix. |
| Test-and-Set Spinlock | O(threads * rounds * cs) | O(threads) | A spinlock on one atomic instruction, and the same lock split into a read and a write, which lets two threads in. |

### Interprocess communication

How processes talk: pipes, message queues, shared memory, sockets, RPC.

| Entry | Time | Space | Key idea |
|-------|------|-------|----------|
| Local Socket Server | O(T + clients) | O(clients) | Clients connect to a Unix-domain socket: queued in the listen backlog until accepted, refused when it is full. |
| Message Queue | O(n log c) | O(c) | Messages keep their boundaries and come out by priority, oldest first among equals; a full queue says EAGAIN. |
| Pipe | O(bytes + calls) | O(capacity) | A writer and a reader on a small pipe: writes block when it is full, reads when it is empty, and close means end of file. |
| Remote Procedure Call | O(retries) | O(retries) | Stubs marshal a call over a lossy network; a lost reply makes a retry run it twice, unless the server dedupes by request id. |
| Shared Memory and Torn Reads | O(rounds) | O(rounds) | Two processes share a record: without a semaphore the reader can copy it half written. |

### File systems

Blocks, inodes, links, journaling, the page cache, RAID and log-structured storage.

| Entry | Time | Space | Key idea |
|-------|------|-------|----------|
| Contiguous Allocation | O(n * blocks) | O(n * blocks) | Each file in one run of blocks: fast to read, but deletes leave holes too small to reuse. |
| Free Space: Bitmap and Free List | O(n * blocks) | O(n * blocks) | Track free blocks with a bitmap (fixed size, easy to scan) or a list of free runs (small until space fragments). |
| Hard and Symbolic Links | O(n * d) | O(inodes) | Names, inodes and link counts: a hard link keeps a file alive, a symbolic link only remembers its path. |
| Indexed Allocation | O(n * blocks) | O(n * blocks) | Each file has an index block listing its blocks: any block in one lookup, for one extra block per file. |
| Inode Pointers and Maximum File Size | O(1) | O(1) | Direct, single, double and triple indirect pointers: how big a file can get, and how many reads reach a byte. |
| Journaling and Crash Recovery | O(1) | O(1) | Crash in the middle of appending a block, with and without a journal, and see what recovery finds. |
| Linked Allocation (FAT) | O(n * blocks) | O(n * blocks) | Each block points to the next in a file allocation table: no fragmentation, but reaching block k takes k hops. |
| Log-Structured File System Cleaning | O(n * segments * size) | O(segments * size) | Every write appends to a log; overwrites leave dead blocks, and the cleaner copies live ones out of the emptiest segment. |
| Page Cache: Write-Through and Write-Back | O(n log p) | O(p) | Write-through pays the disk on every write; write-back batches writes and loses the dirty ones in a crash. |
| Path Resolution | O(n * d) | O(inodes) | Follow a path component by component through directory inodes and symbolic links, as the kernel does on every open. |
| RAID 0, 1, 5 and 6 | O(n * disks) | O(n) | Stripe, mirror, or add parity across disks; fail one or two and rebuild the data with XOR and Reed-Solomon. |

### Virtual memory

Multi-level page tables, the TLB, working sets and thrashing, copy-on-write, mmap.

| Entry | Time | Space | Key idea |
|-------|------|-------|----------|
| Copy-on-Write Fork | O(n + pages) | O(n * pages) | After fork, parent and child share every frame read-only; a write faults and copies just that page. |
| Inverted Page Table | O(n) average, O(n * frames) worst | O(n * frames) | One entry per physical frame, found by hashing (process, page): a table the size of memory, not of address spaces. |
| Memory-Mapped Files | O(accesses) | O(pages) | Read a file with read() (a system call and a copy each time) or map it with mmap() (a fault on first touch, then plain memory). |
| Multi-Level Page Table Walk | O(levels + mappings) | O(mappings * levels) | Split a virtual address into per-level indices and walk the tables to a frame, or to a page fault. |
| Thrashing | O(processes) | O(processes) | More processes keep the CPU busier until their working sets stop fitting; then utilization collapses. |
| TLB and Effective Access Time | O(n * size) | O(n * size) | Run page references through a small TLB and compute the effective access time: why locality makes paging cheap. |
| Working Set | O(n * window) | O(n * window) | The pages used in the last few references: how many frames a process needs right now, phase by phase. |

### Boot and interrupts

How a machine starts, and how devices get the CPU's attention.

| Entry | Time | Space | Key idea |
|-------|------|-------|----------|
| Boot Sequence | O(services) | O(services) | From the reset vector through firmware, the boot loader and the kernel to init and a login prompt, BIOS or UEFI. |
| DMA and Programmed I/O | O(bytes / word) | O(bytes / word) | Copy a device's data word by word with the CPU, or let a DMA controller move it while the CPU does other work. |
| Nested Interrupts and Priorities | O(ticks * irqs) | O(ticks) | Devices raise interrupts: the CPU saves its state, looks up the vector table, and lets higher lines preempt lower handlers. |

## Computing

### Logic gates and Boolean algebra

Gates, truth tables, and simplifying Boolean expressions.

| Entry | Time | Space | Key idea |
|-------|------|-------|----------|
| Canonical SOP and POS | O(2^v * v) | O(2^v * v) | Read a function straight off its truth table: OR together one product per true row (sum of minterms), or AND one sum per false row (product of maxterms). |
| De Morgan's Laws | O(size) | O(size) | Push every NOT down to the variables: NOT (A AND B) is (NOT A) OR (NOT B), and NOT (A OR B) is (NOT A) AND (NOT B), one law at a time. |
| DPLL SAT Solver | O(2^v * size) worst case, far less in practice | O(v + size) | Decide satisfiability without trying every row: assign forced literals (unit clauses) and one-signed variables (pure literals), guess the rest, and backtrack on a conflict. The core of every modern SAT solver. |
| Expression to Circuit | O(size) | O(size) | Turn a Boolean expression into gates, one per operator, and watch the inputs ripple through them to the output one gate delay at a time. |
| Karnaugh Map | O(1) for up to 4 variables | O(1) | Lay the truth table out so neighbors differ in one variable, then circle the largest blocks of 1s (1, 2, 4 or 8 cells, wrapping around): each block is one product term of the simplest sum of products. |
| Logical Equivalence | O(2^v * size) | O(v) | Two expressions are the same function when they agree on every row of the truth table; when they do not, the first row where they differ is the proof. |
| Logic Gates | O(1) | O(1) | The seven gates every computer is made of, NOT, AND, OR, NAND, NOR, XOR and XNOR, and what each outputs for every pair of inputs. |
| Quine-McCluskey | O(3^n), and exponential in the worst case | O(3^n) | Minimize a function of up to 8 variables mechanically: combine terms that differ in one bit until nothing combines, keep the essential prime implicants, and let Petrick's method choose the rest. |
| Satisfiability by Truth Table | O(2^v * size) | O(v) | Is there any input that makes the expression true (satisfiable), are all of them (a tautology), or none (a contradiction)? Try every row and see. |
| Truth Table | O(2^v * size) | O(2^v) | Evaluate a Boolean expression on every combination of its inputs, one row at a time: the complete description of what it computes. |

### Digital circuits

Adders, multiplexers, latches, flip-flops and an ALU, built from gates.

| Entry | Time | Space | Key idea |
|-------|------|-------|----------|
| Arithmetic Logic Unit | O(n) | O(1) | The calculator inside a CPU: add, subtract, AND, OR or XOR two words, chosen by an operation code, and set the zero, negative, carry and overflow flags that branches test. |
| Binary Subtractor | O(n) gates, O(n) delay | O(n) | Subtract by adding: A - B is A + NOT B + 1 in two's complement, so the adder does the work and the carry out says whether there was a borrow. |
| Carry-Lookahead Adder | O(n^2) gates, O(1) carry delay | O(n^2) | Compute every carry at once from generate (A AND B) and propagate (A XOR B) signals, so the sum settles in a fixed number of gate delays whatever the width. |
| D Flip-Flop | O(cycles) | O(cycles) | Store D at the instant the clock rises, and only then: the edge-triggered memory cell that registers, pipelines and counters are made of. |
| Gated D Latch | O(1) per input change | O(1) | An SR latch with a gate in front: while enable is high the output follows D, and when enable falls it keeps the last value. |
| Decoder (2 to 4) | O(1) | O(1) | Turn a 2-bit number into one of four lines: how a CPU picks one memory chip or register from an address. |
| Demultiplexer (1 to 4) | O(1) | O(1) | The reverse switch: send one data input to the output the select lines name, and hold the others low. |
| Full Adder | O(1) | O(1) | Add three bits (two plus a carry in) with two half adders and an OR: the building block of every adder. |
| Half Adder | O(1) | O(1) | Add two bits: the sum is A XOR B and the carry is A AND B. One XOR and one AND, and the start of all arithmetic. |
| JK Flip-Flop | O(cycles) | O(cycles) | A flip-flop with no forbidden input: J sets, K resets, both hold, and J and K together toggle the stored bit. |
| Magnitude Comparator | O(n^2) gates, O(1) delay | O(n^2) | Compare two numbers from the top bit down: they are equal when every bit pair is (XNOR), and the first bit where they differ decides which is greater. |
| Multiplexer (4 to 1) | O(1) | O(1) | A digital switch: two select lines choose which of four data inputs reaches the output, built as one AND per input and an OR. |
| Everything from NAND | O(1) | O(1) | NAND is universal: NOT, AND, OR and XOR built from NAND gates alone, which is why chips can be made of one kind of gate. |
| Everything from NOR | O(1) | O(1) | NOR is universal too: NOT, OR, AND and XOR from NOR gates alone (the Apollo Guidance Computer was built only of NOR gates). |
| Priority Encoder | O(1) | O(1) | Report the number of the highest request line that is on, and whether any is: how an interrupt controller decides who goes first. |
| Register with Load | O(cycles * n) | O(cycles * n) | Four flip-flops sharing a clock and a load line: on a rising edge with load high they store the word on the data lines, and otherwise keep what they hold. |
| Ripple-Carry Adder | O(n) gates, O(n) delay | O(n) | Add two n-bit numbers with n full adders in a chain: simple, but each carry waits for the one before, so the delay grows with the width. |
| Ripple Counter | O(cycles * n) | O(cycles * n) | Chain toggle flip-flops so each one is clocked by the one before it falling: it counts in binary with almost no logic, but the change ripples through the bits one after another. |
| Seven-Segment Decoder | O(1) | O(1) | Turn a 4-bit number into the segments of a digital display: seven Boolean functions of four inputs, one per segment. |
| Shift Register | O(cycles * n) | O(cycles * n) | A chain of flip-flops that moves every bit one place per clock: serial data in, parallel data out, the way UARTs and SPI turn one wire into a byte. |
| SR Latch | O(1) per input change | O(1) | Two NOR gates feeding each other remember one bit: S sets it, R resets it, and with both low the loop holds whatever was last stored. |
| Synchronous Counter | O(cycles * n) | O(cycles * n) | Clock every flip-flop together and toggle bit i when all the bits below it are 1: every bit changes on the same edge, so the count is never wrong in between. |
| T Flip-Flop | O(cycles) | O(cycles) | Toggle on every rising edge while T is high: with T held high, the output runs at half the clock's frequency, the first bit of a counter. |
| Traffic Light State Machine | O(cycles) | O(cycles) | A finite state machine in hardware: two flip-flops hold green, yellow or red, a timer counts the cycles, and the lamps are decoded from the state alone. |

### Data representation

Bases, two's complement, floating point and text, bit by bit.

| Entry | Time | Space | Key idea |
|-------|------|-------|----------|
| Base Conversion | O(log n) | O(log n) | Write a number in binary, octal, hex or any base by repeated division: the remainders, read from last to first, are the digits. |
| Base to Decimal (Horner) | O(digits) | O(digits) | Read a number written in any base: start at zero and, for each digit from the left, multiply by the base and add the digit. |
| Binary-Coded Decimal | O(digits) | O(digits) | Store each decimal digit in its own 4 bits, so a clock or calculator can show a number without dividing it: 409 is 0100 0000 1001. |
| Binary Addition | O(n) | O(n) | Add two binary numbers the way you add on paper, column by column from the right: 1 + 1 is 0, carry 1. |
| Endianness | O(1) | O(1) | The same 32-bit number laid out in memory two ways: big-endian puts the most significant byte first (network order), little-endian the least (x86, ARM). |
| Fixed Point | O(bits) | O(bits) | Store a fraction as an integer with an imaginary binary point: fast and exact in steps of 2^-f, the way DSPs, audio and old games do arithmetic. |
| Decode Float Bits | O(1) | O(1) | Go the other way: give 32 bits and read off the sign, exponent and fraction, including zeros, subnormals, infinities and NaN. |
| Why 0.1 + 0.2 != 0.3 | O(1) | O(1) | Decimal fractions like 0.1 have no exact binary form, so each is stored as the nearest double; the small errors add up and 0.1 + 0.2 lands one step away from 0.3. |
| Hamming(7,4) Code | O(1) | O(1) | Protect 4 data bits with 3 parity bits placed at positions 1, 2 and 4: when one bit flips, the failing checks spell out its position, so the receiver fixes it. |
| IEEE 754 Half Precision | O(1) | O(1) | The 16-bit float of GPUs and machine learning: 5 exponent bits, 10 fraction bits, a range up to 65504, and rounding you can see. |
| IEEE 754 Single Precision | O(1) | O(1) | How a float is stored in 32 bits: a sign, an 8-bit exponent with a bias of 127, and 23 fraction bits after a hidden leading 1, and the exact value that really holds. |
| Parity Bit | O(n) | O(1) | Add one bit that makes the count of 1s even: the receiver recounts, and any single flipped bit shows up (two flipped bits do not). |
| Signed Number Encodings | O(n) | O(n) | Write a negative number three ways: sign-magnitude, ones' complement and two's complement, and see why two's complement won (one zero, and the same adder for signed and unsigned). |
| Two's Complement Overflow | O(n) | O(n) | Add two signed numbers on a fixed number of bits: past the largest value the result wraps around to a negative, and the overflow flag says so. A carry out is not the same thing. |
| UTF-8 Encoding | O(n) | O(n) | How text becomes bytes: ASCII letters stay one byte, and every other character takes 2 to 4, with leading bits that say how many. |

### Theory of computation

Finite automata, regular expressions, grammars and Turing machines.

| Entry | Time | Space | Key idea |
|-------|------|-------|----------|
| Busy Beaver (3 states) | O(1): 14 moves | O(1) | The 3-state, 2-symbol busy beaver: started on a blank tape it writes 6 ones and halts after 14 moves, the most ones any 3-state machine can leave. Busy beaver numbers grow faster than anything computable. |
| Context-Free Grammar Derivation | exponential worst case (bounded) | the same | Find a leftmost derivation of a string: start from the start symbol and keep replacing the leftmost nonterminal by one of its rules until only the string is left. |
| CYK Parsing | O(n^3 * \|G\|) | O(n^2 * \|N\|) | Decide whether a grammar in Chomsky normal form derives a string by filling a triangle: which nonterminals derive each substring, from every way to split it in two. |
| Deterministic Finite Automaton | O(n) | O(1) | A machine with finitely many states reads a string one character at a time, each character moving it along exactly one arrow; it accepts if it ends in an accepting state. |
| DFA Minimization (table filling) | O(k * n^3) | O(n^2) | Find the smallest DFA for the same language: mark every accepting and rejecting pair as different, keep marking pairs whose moves lead to marked pairs, then merge what is left. |
| Product of Two DFAs | O(\|A\| * \|B\| * k) | O(\|A\| * \|B\|) | Run two DFAs side by side as one whose states are pairs: accept when both accept (intersection) or, with union on, when either does. Regular languages are closed under both. |
| Elementary Cellular Automaton | O(gens * width) | O(gens * width) | A row of cells, each alive or dead; each generation a cell looks at itself and its two neighbors and the rule number's bits say what it becomes. Rule 30 is chaos from one cell; rule 110 can compute anything. |
| Conway's Game of Life | O(gens * rows * cols) | O(gens * rows * cols) | A grid of cells: a live cell with two or three live neighbors survives, a dead cell with exactly three is born, everything else dies. Gliders walk, oscillators blink, and it can compute anything. |
| Grammar Ambiguity (counting parse trees) | O(n^3 * \|G\|) | O(n^2 * \|N\|) | A grammar is ambiguous when one string has two parse trees. E -> E + E parses a+a+a two ways ((a+a)+a and a+(a+a)); CYK can count the trees. |
| The Halting Problem (diagonal argument) | O(p^2 * budget) | O(p^2) | No program can decide whether every program halts. Play it out: list programs, ask a would-be decider about each on its own number, and build D that does the opposite. D differs from every program on the list. |
| Lambda Calculus (beta reduction) | O(steps * term size) | O(term size) | Computation as substitution: (λx.M) N becomes M with N for x, the leftmost outermost first. Numbers, booleans and arithmetic are all functions (Church encodings): PLUS 2 3 reduces to 5. |
| Mealy Machine (vending machine) | O(n) | O(n) | A finite machine that writes on its transitions: a vending machine for a 15-cent drink, with n a nickel and d a dime, writes "drink" on the move that reaches 15. |
| Moore Machine (vending machine) | O(n) | O(n) | The same vending machine as a Moore machine: the output belongs to the states (15 and 20 drop a drink), so it needs more states than the Mealy machine but its output never depends on the move. |
| Nondeterministic Finite Automaton | O(n * (V + E)) | O(V) | A machine that may have several arrows on one character, or arrows taken without reading (ε). Simulate it by keeping the set of every state it could be in. |
| Pushdown Automaton: a^n b^n | O(n) | O(n) | Accept a's followed by the same number of b's: push one A per a, pop one per b. No DFA can do it (see the pumping lemma), a PDA can. |
| Pushdown Automaton: Balanced Parentheses | O(n) | O(n) | A finite machine plus a stack: push on "(", pop on ")", and accept when only the bottom marker $ is left. The stack is the memory a DFA lacks. |
| Primitive Recursion and Ackermann | O(value) applications | O(1) | Build add, multiply and power from zero, successor and a loop whose count is fixed in advance (primitive recursion), then watch Ackermann's function outgrow them all: computable, but not primitive recursive. See also the ackermann entry. |
| Pumping Lemma (a^n b^n is not regular) | O(p^3) | O(p^2) | If a DFA with p states accepted a^n b^n, every long string could be split x y z with y repeatable. Try every split of a^p b^p: pumping y always breaks the balance, so no DFA exists. |
| Pushdown Automaton | O(n) | O(n) | Run your own deterministic PDA: each rule reads a character (or _ for none), pops the top (or _), pushes a string and moves. |
| Regex Backtracking Blowup | O(2^n) backtracking, O(n^2) with the NFA | O(n) | Match (a?)^n a^n against a^n both ways: a backtracking engine tries about 2^n choices, while simulating the Thompson NFA takes n * m steps. The reason some regex engines can be frozen by one input. |
| Subset Construction (NFA to DFA) | O(2^V * k) worst, usually far less | O(2^V) | Turn an NFA into a DFA that accepts the same strings: each DFA state is a set of NFA states, found by following every character from every set found so far. |
| Thompson's Construction (regex to NFA) | O(m) to build, O(n * m) to match | O(m) | Build an NFA from a regular expression, one small machine per character, joined with ε-moves for \|, * and +; then match by simulating it. No backtracking, so never exponential. |
| Turing Machine: Binary Increment | O(moves) | O(moves) | Add 1 to a binary number on the tape: run right to the end, then carry left, turning 1s into 0s until a 0 or a blank takes the 1. |
| Turing Machine | O(moves) | O(moves) | Run your own Turing machine: a finite control, a tape that never ends and a head that reads, writes and moves. Each rule is "STATE READ -> NEXT WRITE MOVE". |
| Turing Machine: Palindromes | O(moves) | O(moves) | Decide whether a string of a and b reads the same both ways: erase the first symbol, remember it in the state, check and erase the last, and walk back. O(n^2) moves. |
| Turing Machine: Unary Addition | O(moves) | O(moves) | Add two numbers written in unary (111+11 is 3 + 2): turn the + into a 1 and erase the last 1. |

### Compilers and interpreters

From source text to tokens, trees, bytecode and a running program.

| Entry | Time | Space | Key idea |
|-------|------|-------|----------|
| Bytecode Compiler | O(n) | O(n) | Compiles a program to bytecode for a stack machine: operands are pushed, operators pop them, and if and while become jumps. |
| Constant Folding | O(n) | O(n) | Computes at compile time every expression whose operands are constants, so 2 * 3 + 4 becomes 10 before the program ever runs. |
| Copying Collector (Cheney) | O(live objects) | O(n) | Copies the live objects to a fresh space, breadth first, using the new space itself as the queue, and drops the old space whole: garbage costs nothing. |
| Dead-Code Elimination | O(n) | O(n) | Removes code that can never run: branches ruled out by constant conditions, loops that never start, statements after a return. Folds constants first. |
| Lexer | O(n) | O(n) | Splits source code into tokens (numbers, names, keywords, operators), each with its line and column: the first stage of every compiler. |
| Mark and Sweep | O(n + e) | O(n) | Finds every object reachable from the roots, marking as it goes, then sweeps the heap and frees the rest, cycles included. |
| Syntax Tree of a Program | O(n) | O(n) | Parses a whole program (let, if, while, fn, print) into its syntax tree by recursive descent, the tree every later stage works on. |
| Pratt Parser | O(n) | O(n) | Parses an expression with a table of binding powers and one loop: an operator takes the value on its left only if it binds tighter than the caller allows. |
| Recursive Descent Parser | O(n) | O(n) | Parses an expression with one function per grammar rule (sum calls product calls unary calls primary), so precedence falls out of the call structure. |
| Reference Counting | O(1) per reference change | O(n) | Frees an object the moment nothing references it, by keeping a count per object; simple and immediate, but a cycle of garbage keeps its counts up and leaks. |
| Register Allocation | O(n * v + v^2) | O(v^2) | Gives variables CPU registers by coloring their interference graph (Chaitin): variables live at the same time get different colors, and what does not fit spills to memory. |
| Reverse Polish Notation | O(n) | O(n) | Evaluates postfix like 3 4 + 2 * with one stack and no parentheses: push numbers, and let each operator pop two and push its result. |
| Shunting-Yard Algorithm | O(n) | O(n) | Turns infix like 3 + 4 * 2 into reverse Polish notation (3 4 2 * +) with an operator stack, as Dijkstra did for ALGOL compilers. |
| Stack Virtual Machine | O(s) for s instructions run | O(d + h) for d calls and a stack h deep | Runs bytecode one instruction at a time on an operand stack, with a frame of variables per call: how Python, Java and Lua run code. |
| Three-Address Code | O(n) | O(n) | Lowers a program to instructions with at most one operator each (t1 = b * c), with temporaries and labeled jumps: the form optimizers work on. |
| Tree-Walking Interpreter | O(s) for s statements run | O(d) for d nested calls | Runs a program by walking its syntax tree, statement by statement, keeping variables in a map per function call. |
| Incremental Tri-Color Marking | O(n + e) | O(n) | Marks in small steps while the program keeps running; a write that hides a white object behind a black one loses it, unless a write barrier shades it gray. |
| Type Checking | O(n) | O(n) | Checks every name is defined and every operator gets the right types, reporting each mistake at its line and column without running anything. |

### Computer architecture

How a CPU fetches, decodes and executes, and how pipelines speed it up.

| Entry | Time | Space | Key idea |
|-------|------|-------|----------|
| Amdahl's Law | O(1) | O(1) | How much faster a program gets on n processors when only part of it runs in parallel: the serial part caps the speedup, however many cores you add. |
| Two-Pass Assembler | O(n) | O(n) | Turns assembly into machine instructions in two passes: the first notes where each label is, so the second can resolve jumps to labels further down. |
| Call Stack and Frames | O(n) | O(n) | Watches the stack during recursive factorial: each call pushes a frame with its argument and return address, and each return pops one. |
| Context Switch | O(t / q) | O(t / q) | Time-slices one CPU between processes; every switch saves one process's registers into its control block and loads the next one's, which is pure overhead. |
| Pipeline Control Hazards | O(s * c) for s instructions and c cycles | O(s * c) | A taken branch is only known in the execute stage, so the two instructions fetched behind it are flushed: why loops cost more than their instructions. |
| Cycles per Instruction | O(c) for c classes | O(1) | Averages the cycles per instruction over an instruction mix; with the clock rate it gives CPU time = instructions x CPI / clock rate. |
| Fetch, Decode, Execute | O(s) for s instructions run | O(r + m) | Runs a program on a tiny CPU one phase at a time: fetch the instruction, decode it and read registers, execute it in the ALU, write the result back. |
| Gshare Branch Predictor | O(n) | O(b) for b branches | Indexes 2-bit counters by the branch address XOR the last few outcomes, so a branch learns patterns: a loop of four becomes perfectly predictable. |
| Interrupt Handling | O(1) | O(1) | A device interrupts a running program: the CPU finishes the instruction, saves its state, runs the handler from the vector table, and resumes. |
| Memory Hierarchy | O(1) | O(1) | Registers, caches, RAM and disks from fastest to slowest, with how long each would feel if a CPU cycle took one second. |
| 1-Bit Branch Predictor | O(n) | O(b) for b branches | Guesses each branch does what it did last time; a loop's branch misses twice per loop, at the exit and again on the way back in. |
| Pipeline Forwarding | O(s * c) for s instructions and c cycles | O(s * c) | The same pipeline with forwarding: a result goes straight from the ALU to the next instruction, so only a value loaded from memory still costs a bubble. |
| Pipeline Data Hazards | O(s * c) for s instructions and c cycles | O(s * c) | Overlaps five stages of five instructions at once; an instruction that needs a result not yet written back waits, leaving bubbles in the pipeline. |
| Struct Alignment and Padding | O(n log n) | O(n) | Lays out a struct's fields the way Go and C do, each at an offset that is a multiple of its alignment, and shows how reordering them removes padding. |
| 2-Bit Saturating Predictor | O(n) | O(b) for b branches | A 2-bit counter per branch needs two surprises in a row to change its mind, so a loop's branch misses once per loop instead of twice. |
| Von Neumann and Harvard | O(s) for s instructions run | O(1) | One memory for code and data shares one bus, so every load or store delays the next fetch; separate memories (Harvard) do both at once. |

## Machine learning

### Clustering

Grouping points that are close together, without labels.

| Entry | Time | Space | Key idea |
|-------|------|-------|----------|
| DBSCAN | O(n^2 * d) | O(n) | Clusters by density: points with enough neighbors within a radius grow a cluster; the rest is noise. Finds the number of clusters itself. |
| Agglomerative Hierarchical Clustering | O(n^3) | O(n^2) | Start with every point alone and keep joining the two nearest clusters; the merges form a dendrogram you can cut into k clusters. |
| K-Means Clustering | O(n * k * d) per round | O(n + k * d) | Assigning points and moving centers each lower the total distance, so it converges. |
| K-Means++ Seeding | O(n * k * d) | O(n + k * d) | Pick starting centers far apart: each next one with probability proportional to its squared distance from the nearest chosen. |
| Silhouette Score | O(n^2 * d) | O(n) | How well each point sits in its cluster, from -1 to 1, here for k-means at every k from 2 to the option: the best k scores highest. |

### Regression

Fitting a line or curve that predicts a number.

| Entry | Time | Space | Key idea |
|-------|------|-------|----------|
| Gradient Descent (linear regression) | O(n * epochs) | O(epochs) | Learn a line by stepping downhill on the error, epoch after epoch; the loss curve shows learning, or diverging when the rate is too high. |
| Linear Regression | O(n) | O(1) | The best slope and intercept have a closed form from means and covariances. |
| Logistic Regression | O(n * d * epochs) | O(d + epochs) | A yes-or-no model: the sigmoid of a weighted sum gives P(yes), learned by gradient descent on the log loss. |
| Multiple Linear Regression | O(n * d^2 + d^3) | O(d^2) | A plane (or hyperplane) through data with several features, by solving the normal equations (X^T X) c = X^T y. |
| Polynomial Regression | O(n * degree^2 + degree^3) | O(n * degree) | A curve c0 + c1 x + c2 x^2 + ... by least squares on the powers of x; raise the degree to see overfitting. |
| Mean Squared Error and R^2 | O(n) | O(1) | How good predictions are: the mean squared error, and R^2, the share of the variance they explain. |

### Classification

Learning classes from labeled examples.

| Entry | Time | Space | Key idea |
|-------|------|-------|----------|
| Confusion Matrix | O(n + classes^2) | O(classes^2) | A table of the true class against the predicted one: the diagonal is right, everything else shows which classes get mixed up. |
| Decision Tree (CART) | O(d * n^2 * depth) | O(n) per level | Ask the yes-or-no question that best splits the classes (lowest Gini impurity), then ask again on each side. |
| Gaussian Naive Bayes | O(n * d) to fit, O(classes * d) to predict | O(classes * d) | Per class, a bell curve for each feature; the most probable class wins, pretending the features are independent. |
| K-Fold Cross-Validation | O(n^2 * d) for k-NN | O(n) | Split the data into k folds, test on each in turn while training on the rest, and average: every example is tested exactly once. |
| K-Nearest Neighbors | O(n * d + n log n) per query | O(n) | Classify a new point by a vote of the k training points nearest it. No training at all: every query does the work. |
| Naive Bayes Spam Filter | O(words) to fit and to predict | O(classes * vocabulary) | Count words per class, then score a new message by how likely each class is to produce its words (with add-one smoothing). |
| Perceptron | O(n * d * epochs) | O(d + epochs) | The first learning machine: nudge a line toward every example it gets wrong. It stops if a line separates the classes, and never does on XOR. |
| Precision, Recall and F1 | O(classes^2) | O(classes) | Per class: of the predictions, how many were right (precision); of the truth, how much was found (recall); and their balance (F1). |
| Train/Test Split | O(n) | O(n) | Shuffle the examples and hold some back, so a model is judged on data it never trained on. |

### Neural networks

Tiny networks built from scratch, to see how learning works.

| Entry | Time | Space | Key idea |
|-------|------|-------|----------|
| Activation Functions | O(1) each | O(1) | Sigmoid, tanh and ReLU, and their slopes, from x = -4 to 4: why ReLU keeps gradients alive and sigmoid lets them fade. |
| Softmax and Cross-Entropy | O(n) | O(n) | Turn scores into probabilities that sum to 1, and score them against the right answer: the loss a classifier learns from. |
| Neural Network Learning XOR | O(epochs * n * weights) | O(units + epochs) | A two-layer network learns XOR by backpropagation: errors flow backward through the weights by the chain rule. |

## Systems

### Networking

Talking over unreliable links: sliding windows, retransmission, routing, rate limits.

| Entry | Time | Space | Key idea |
|-------|------|-------|----------|
| ARP | O(sends) | O(cache) | Find the MAC address behind an IP address on the local network: check the cache, else broadcast who-has and cache the reply. |
| Bandwidth-Delay Product | O(1) | O(1) | How many bytes must be in flight to fill a link, and why a small window caps throughput at window / round trip on a long fat network. |
| BGP Path Vector | O(rounds * E * V) | O(V^2) | Networks announce whole paths of autonomous systems, prefer the shortest, and refuse any path that already contains themselves: no loops. |
| CSMA/CA and Hidden Terminals (Wi-Fi) | O(slots) | O(slots) | Two laptops that cannot hear each other both think the air is free and collide at the access point; RTS/CTS lets the access point reserve the air. |
| CSMA/CD (Ethernet) | O(slots * stations) | O(stations + slots) | Stations on a shared cable send when it is free; on a collision each waits a random number of slots, doubling the range after each collision. |
| DHCP (DORA) | O(clients * pool) | O(pool) | A new device has no address: it broadcasts Discover, the server Offers one, the device Requests it, and the server Acknowledges a lease. |
| Distance-Vector Routing | O(rounds * V^2 * degree) | O(rounds * V^2) | Routers swap distance vectors with their neighbors each round (Bellman-Ford, as RIP does); cut a link and, without poisoned reverse, they count to infinity. |
| DNS Resolution | O(labels) messages | O(names) | Resolve names down a simulated hierarchy, root to top-level domain to the domain's server, iteratively or recursively, with a cache. |
| Go-Back-N ARQ | O(ticks * window) | O(frames + ticks) | Keep a window of frames in flight; the receiver takes only the next in order, so one loss resends the whole window after it. |
| HTTP/1.1 vs HTTP/2 vs HTTP/3 | O(resources * size) | O(resources * size) | One request at a time, six connections, one multiplexed connection, or QUIC streams: and how one lost packet blocks every HTTP/2 stream but one QUIC stream. |
| HTTP Request Waterfall | O(requests) | O(requests) | Where a page load's time goes: DNS, the TCP handshake, TLS, waiting for the first byte, and the download; keep-alive skips the setup. |
| Internet Checksum | O(n) | O(1) | Add the data as 16-bit words with the carries folded back in, then invert: the check in every IPv4, UDP and TCP header. |
| MTU and IP Fragmentation | O(fragments * links) | O(fragments) | A packet bigger than a link's MTU is cut into fragments at 8-byte offsets, perhaps cut again later, and put back together only at the destination. |
| Leaky Bucket | O(n log n) | O(n) | Requests wait in a bounded queue that drains at a fixed rate: bursts are smoothed into an even stream, and overflow is dropped. |
| Link-State Routing | O(V^3) | O(V^2) | Every router learns the whole map and runs Dijkstra from itself (as OSPF does) to build its forwarding table. |
| Longest Prefix Match | O(32) per lookup | O(routes * 32) | A router forwards each address by the most specific route that contains it, found by walking a binary trie bit by bit. |
| Nagle's Algorithm | O(ticks) | O(segments) | Hold small writes back while data is unacknowledged and send them together: far fewer tiny packets, at the cost of some delay. |
| NAT with Port Translation | O(packets) | O(mappings) | A home router shares one public address among many devices by giving each connection its own public port, and lets in only replies. |
| OSI and TCP/IP Layers | O(layers) | O(layers) | Data goes down the stack and each layer wraps it in its own header: HTTP in TCP in IP in an Ethernet frame, stripped again in reverse. |
| Packet vs Circuit Switching | O(1) | O(1) | Reserve a circuit first, store and forward the whole message at each hop, or pipeline small packets: the delay of each, by formula. |
| QUIC vs TCP + TLS | O(1) | O(1) | Round trips before the first byte of a response: TCP + TLS 1.2 takes four, TCP + TLS 1.3 three, QUIC two, and QUIC 0-RTT one. |
| Selective Repeat ARQ | O(ticks * window) | O(frames + ticks) | A window of frames in flight, buffered out of order and acked one by one, so only the lost frames are resent. |
| Spanning Tree Protocol | O(V^2 + E) | O(V + E) | Switches elect the lowest ID as root, keep the cheapest port toward it, and block the rest, so a looped network forwards without loops. |
| Stop-and-Wait ARQ | O(ticks) | O(frames + ticks) | Send one frame, wait for its ack, resend on a timeout: reliable over a lossy link, but one frame per round trip. |
| Subnetting (CIDR) | O(subnets) | O(subnets) | The prefix length sets how many bits name the network and how many name hosts. |
| TCP Congestion Control | O(rounds) | O(rounds) | Additive increase and multiplicative decrease share bandwidth fairly. |
| TCP Flow Control | O(ticks) | O(ticks) | The receiver advertises the room left in its buffer and the sender never exceeds it; a shut window is reopened by zero-window probes. |
| TCP Handshake and Teardown | O(segments) | O(segments) | TCP's state machine: SYN, SYN+ACK, ACK to open; FIN, ACK, FIN, ACK to close, with TIME_WAIT on the side that closed first. |
| TCP Retransmission Timer | O(samples) | O(samples) | Smooth the measured round trips and their variation into a timeout (Jacobson/Karels), skip ambiguous samples (Karn), and double on each timeout. |
| TLS 1.3 Handshake | O(1) | O(1) | The messages of a TLS 1.3 handshake, each with its job and the cryptography inside it: one round trip, or zero when resuming. |
| Token Bucket | O(n log n) | O(n) | Tokens drip into a bucket; each request takes one or is rejected: bursts up to the bucket's size pass at once, the long-run rate is capped. |
| Traceroute | O(hops * probes) | O(hops) | Send probes with a time to live of 1, 2, 3...: each router that drops one at zero answers with its name, mapping the path. |
| WebSocket Upgrade | O(n) | O(n) | An HTTP request asks to upgrade, the server proves it understood with a hash of the key, and the connection carries masked frames from then on. |

### Databases

Storing, indexing and querying data, and keeping it safe.

| Entry | Time | Space | Key idea |
|-------|------|-------|----------|
| B+ Tree Index | O(log_m n) search, insert and delete; O(log_m n + k) for a range of k keys | O(n) | Inner nodes only guide; linked leaves make range scans sequential. |
| Block Nested-Loop Join | O(n * m), with n / block inner scans | O(k) for k joined rows | A nested-loop join reading the outer table a block at a time, so the inner table is scanned once per block instead of once per row. |
| Buffer Pool (Clock) | O(n) per miss for n frames, O(n) per hit to find the frame | O(n) | Disk pages cached in a fixed number of frames: hits, misses, pins, dirty write-backs, and a clock hand giving each page a second chance. |
| Buffer Pool (LRU-K) | O(n) per miss for n frames | O(n + pages seen) | The same buffer pool evicting the page whose K-th most recent use is oldest, so one big scan cannot push out the pages used again and again. |
| Conflict Serializability (precedence graph) | O(n^2) | O(n) | Draw an edge for every pair of conflicting operations; the schedule equals some serial order exactly when the graph has no cycle. |
| Extendible Hashing | O(1) search, O(1) amortized insert | O(n + 2^g) | A directory of 2^g pointers to buckets read by a key's low bits: a full bucket splits on one more bit, and the directory doubles only when it must. |
| External Merge Sort | O(N log N), 1 + ceil(log_k(N / M)) passes | O(N) | Sort more data than fits in memory: sorted runs the size of memory, then k-way merges with a heap, pass after pass, until one run is left. |
| Hash Join | O(n + m + k) expected | O(k) for k joined rows | One pass to build and one to probe beats comparing every pair. |
| Index Nested-Loop Join | O(n log m + k) | O(k) for k joined rows | For each outer row, look the key up in an index on the inner table instead of scanning it. |
| Isolation Levels and Their Anomalies | O(1) | O(1) | Dirty reads, non-repeatable reads, phantoms, lost updates and write skew, each run at all four isolation levels to see which ones each level stops. |
| Linear Hashing | O(1) amortized search and insert | O(n) | A hash table that grows one bucket at a time in a fixed order, with overflow chains until each bucket's turn to split comes. |
| LSM Tree | O(1) amortized writes; O(r log n) reads over r runs, most skipped by Bloom filters | O(n) | Writes into a memtable flushed as sorted runs, reads newest first with a Bloom filter per run, and compaction merging runs: the engine of LevelDB and RocksDB. |
| MVCC (snapshot isolation) | O(n v) for v versions per item | O(n) | Writes make new versions instead of overwriting, each transaction reads its own snapshot, and of two writers of one item the first to commit wins. |
| Nested-Loop Join | O(n * m) | O(k) for k joined rows | Join two tables by comparing every row of one with every row of the other: simple, and slow on anything big. |
| Query Planner (index or full scan) | O(n) | O(n) | A cost-based optimizer in miniature: estimate how many rows a condition matches, price an index scan and a full scan in page reads, and pick the cheaper. |
| Relational Algebra (select, project, group by) | O(n log n) per sort, O(n) per other stage | O(n) | Run a pipeline of relational operators over a table: select rows, project columns, group and aggregate, sort and limit, seeing the table after each. |
| Slotted Page | O(s) insert for s slots, O(1) get and delete, O(n) compaction | O(page size) | How a page holds variable-length records: a slot directory from the front, records from the back, holes left by deletes, and compaction. |
| Sort-Merge Join | O(n log n + m log m + k) | O(k) for k joined rows | Sort both tables on the join column, then walk them together like a merge, pairing runs of equal keys. |
| Timestamp Ordering | O(n) | O(n) | No locks: every transaction gets a timestamp, and an operation that arrives too late for timestamp order aborts its transaction, unless the Thomas write rule can skip it. |
| Strict Two-Phase Locking | O(n^2 t) at worst | O(n) | Shared locks to read, exclusive to write, all held to the end: blocked operations wait, and a cycle in the wait-for graph is a deadlock broken by aborting the youngest. |
| Write-Ahead Log and Recovery | O(n) | O(n) | After a crash, replaying the log redoes committed work and undoes the rest. |

### Distributed systems

Machines agreeing over a network: clocks, elections, consensus.

| Entry | Time | Space | Key idea |
|-------|------|-------|----------|
| Bully Election | O(n^2) messages | O(n^2) | A process that notices the leader is gone challenges every higher id; the highest one alive bullies its way to leader. |
| Byzantine Generals (OM(m)) | O(n^(m+1)) messages | O(n^(m+1)) | Generals relay orders through rounds of oral messages and take majorities; loyal ones agree despite m traitors only when n >= 3m + 1. |
| Chandy-Lamport Snapshot | O(transfers + n^2) messages | O(n^2) | Record a consistent global state while money keeps moving: markers split each channel into before and after the snapshot. |
| G-Counter (CRDT) | O(replicas) per merge | O(replicas) | Each replica counts its own increments; merging takes the larger count per replica, so replicas agree in any merge order. |
| Gossip Protocol | O(rounds * n * fanout) | O(n) | Each node that knows a rumor tells a few random others every round: everyone knows in about log n rounds, and pulling finishes the stragglers. |
| Lamport Clocks | O(events) | O(events) | Count events per process and jump past the time a message carries: if a happened before b, a's timestamp is smaller. |
| OR-Set (CRDT) | O(tags) per merge | O(tags) | A set whose adds carry unique tags and whose removes delete only the tags they saw, so a concurrent add survives a remove: add wins. |
| Paxos (single decree) | O(attempts * proposers * acceptors) messages | O(messages) | Proposers race with numbered ballots; acceptors promise and accept, and once a majority accepts a value every later ballot carries it. |
| Quorum Reads and Writes | O(trials * n) | O(n) | Write to W of N replicas and read R: when R + W > N every read meets the latest write; below that, some reads come back stale. |
| Raft Consensus | O(ticks * nodes^2) | O(ticks * nodes) | An entry is committed once a majority stores it, and any two majorities overlap. |
| Rendezvous Hashing | O(keys * nodes) | O(nodes) | Every node scores every key and the highest score owns it: a new node takes only its share, while hash mod n reshuffles nearly all. |
| Ring Election (Chang-Roberts) | O(n log n) average, O(n^2) worst messages | O(n) | Every process sends its id round a ring; larger ids are passed on, smaller dropped, and the id that comes home is the leader. |
| Two-Phase Commit | O(n) ticks, 4n messages | O(n) | Nobody commits until everyone has promised they can. |
| Vector Clocks | O(events * processes) | O(events * processes) | One counter per process in every timestamp: a happened before b exactly when a's vector is below b's, so concurrent events show. |

### Infrastructure

Keeping services up: load balancing, circuit breakers, retries, caches, delivery guarantees, autoscaling, rollouts.

| Entry | Time | Space | Key idea |
|-------|------|-------|----------|
| Autoscaling | O(n) | O(n) | Add and remove instances to follow a load curve, with boot time and a cooldown, and count the requests dropped. |
| Cache Stampede | O(ticks * rate) | O(ticks) | A hot key expires and every request hits the database at once; a lock or early refresh keeps it to one. |
| Cache-Aside, Write-Through, Write-Behind | O(n) | O(keys) | Three ways to keep a cache in front of a database, traded off in database writes, misses and lost writes. |
| CDN Edge Caches | O(requests) | O(regions * objects) | Requests from several regions for popular and rare objects: how much each edge serves and how much reaches the origin. |
| Container CPU Shares and Memory Limits | O(n^2) | O(n) | Containers split the CPUs by shares only under contention, and one over its memory limit meets the OOM killer. |
| Circuit Breaker | O(n) | O(n) | After repeated failures stop calling a sick service for a while, then let one trial call decide whether to resume. |
| At-Most-Once, At-Least-Once, Exactly-Once | O(messages * retries) | O(messages) | Send messages over a lossy link: lose some, duplicate some, or dedupe with idempotency keys. |
| Blue/Green, Canary and Rolling Deployments | O(steps) | O(steps) | Roll out a bad build three ways and count the failed requests before the rollback. |
| Exponential Backoff and Jitter | O(clients * attempts + T) | O(T) | Many clients retry after an outage: without jitter their retries arrive in synchronized spikes, with jitter they spread out. |
| Least-Connections Load Balancing | O(T * (servers + open requests)) | O(servers * T) | Send each request to the server with the fewest open requests for its speed. |
| Power of Two Choices | O(T * (servers + open requests)) | O(servers * T) | Sample two servers at random and take the less loaded: nearly as good as checking them all. |
| Round-Robin Load Balancing | O(T * (servers + open requests)) | O(servers * T) | Send each request to the next server in turn, fast or slow: simple, and unfair to slow servers. |
| Weighted Round-Robin Load Balancing | O(T * (servers + open requests)) | O(servers * T) | Send requests in proportion to each server's speed. |

### System design

Designing real systems: estimates, IDs, case studies, data patterns and reliability math.

| Entry | Time | Space | Key idea |
|-------|------|-------|----------|
| Adaptive Bitrate Streaming | O(seconds) | O(seconds) | Each second the player picks the best bitrate the network allows, dropping a step when the buffer runs thin, to avoid stalls. |
| Availability in Series and Parallel | O(n) | O(n) | Every dependency in series multiplies availability down; every redundant copy in parallel adds a nine. Nines and downtime a year. |
| Back-of-the-Envelope Estimation | O(1) | O(1) | Size a service before designing it: reads and writes a second, peak traffic, storage over the years, bandwidth. |
| Backpressure and Load Shedding | O(ticks) | O(ticks) | A burst bigger than the service: an unbounded queue's latency grows without limit, shedding drops the excess, backpressure pushes it back. |
| Base 62 Encoding | O(log n) | O(log n) | Write a number with 0-9, a-z and A-Z: seven characters cover 3.5 trillion short links. |
| The Bulkhead Pattern | O(ticks x requests) | O(threads) | One slow dependency takes every thread of a shared pool and the healthy one fails too; a pool per dependency contains it. |
| CAP and PACELC | O(ops) | O(1) | During a partition, choose consistency (refuse writes) or availability (diverge, reconcile later); else, choose latency or consistency. |
| Change Data Capture | O(changes) | O(rows) | Tail the database's log and stream every change, in order per key, to search indexes and caches, so the app writes once. |
| Design a Chat System | O(m log m) | O(m) | WebSocket gateways, a sequence number per conversation to keep order through network reordering, presence, and the offline inbox. |
| Consistency Models | O(n!) | O(n) | Is a history of reads and writes linearizable, sequential, causal or only eventual? A search for an order each model allows. |
| CQRS | O(commands + queries) | O(commands) | Separate the write model from a read model shaped for queries, kept up to date by events: scalable, but reads can be stale. |
| Design a Distributed Rate Limiter | O(requests x nodes) | O(nodes) | One limit across many API nodes: a central counter, local shares, or synced local counters, and what each lets through. |
| SLO Error Budget and Burn Rate | O(1) | O(1) | A 99.9% SLO allows 0.1% failures: the budget, how much is spent, and the burn rate that says when to stop shipping or page someone. |
| Event Sourcing | O(events) | O(events) | Store every change as an event and rebuild state by replaying them; snapshots keep the replay short. |
| Geohash | O(precision) | O(precision) | Interleave the bits of latitude and longitude and write them in base 32: nearby places share a prefix. |
| Design a Key-Value Store (Dynamo) | O(ops x nodes) | O(keys x nodes) | A consistent-hashing ring with N replicas, R and W quorums, sloppy quorums with hinted handoff while a node is down, and read repair. |
| Latency Numbers Every Programmer Should Know | O(1) | O(1) | From an L1 cache hit (half a nanosecond) to a packet across the Atlantic (150 ms): the ratios behind every design choice. |
| Leader-Follower Replication | O(writes) | O(followers) | Writes to the leader, reads from lagging followers: how often a user misses their own write, and how read-your-writes routing fixes it. |
| Design a Real-Time Leaderboard | O(updates x players) | O(players) | A sorted set (Redis ZSET) in front of the database: add points, the top k and any player's rank without scanning every player. |
| Little's Law | O(1) | O(1) | Requests in flight = arrival rate x time in the system: how many threads or connections a service needs, and why latency costs capacity. |
| Multi-Leader Conflict Resolution | O(rounds x writes) | O(keys) | Two datacenters each taking writes: last write wins silently drops conflicting updates; version vectors keep both for a merge. |
| Design a News Feed | O(posts x followers + reads x follows) | O(users x follows) | Fan-out on write, fan-out on read, or a hybrid that pulls only celebrities: the work each does, and the celebrity problem. |
| Design a Notification System | O(n x retries) | O(n) | A queue, idempotency keys that drop the producer's retries, exponential backoff on provider failures, and a dead-letter queue. |
| Offset vs Cursor Pagination | O(pages x offset) | O(rows) | Offset pages repeat items when new ones arrive and read every skipped row; cursor pages never repeat and read only the page. |
| Design a Payment System | O(payments) | O(payments) | Idempotency keys so a retried payment is never charged twice, and a double-entry ledger that always sums to zero. |
| Powers of Two | O(1) | O(1) | 2^10 is about a thousand, 2^20 a million, 2^30 a billion: the table for sizing data in your head. |
| Design a Proximity Service | O(candidates) per search | O(businesses) | Index businesses by geohash and search the user's cell and its eight neighbors: a few candidates instead of a city-wide scan. |
| Utilization vs Latency (M/M/1) | O(customers) | O(1) | Why services run well below 100%: an M/M/1 queue's latency is 2x at half load, 10x at 90%, with a simulation beside the formula. |
| Thundering Herd and Request Coalescing | O(ticks x clients) | O(1) | A hot key expires and every request hits the database at once; coalescing sends one read and the rest wait for it. |
| Resharding Cost | O(keys log shards) | O(shards) | Grow from 4 to 5 shards: hash mod n moves 80% of the keys, consistent hashing about a third, close to the least possible 20%. |
| Design Ride Matching | O(riders x rings^2) | O(drivers) | Drivers in a grid of cells, and each rider matched to the nearest free driver by searching rings of cells outward. |
| The Saga Pattern | O(steps) | O(steps) | A transaction across services as local steps, each with a compensation; a failure undoes the finished steps in reverse order. |
| Design Search Autocomplete | O(1) per keystroke | O(total length x k) | Precompute the top k searches for every prefix, so each keystroke is one lookup instead of a scan of every matching term. |
| Range vs Hash vs Directory Sharding | O(keys + reads) | O(shards) | Range sharding piles new and hot keys on one shard; hashing spreads them but scatters scans; a directory moves ranges where load is. |
| Snowflake IDs | O(workers x ids) | O(ids) | 64-bit IDs that sort by time, unique across workers with no coordination: milliseconds, worker and sequence bits, and what a clock going back does. |
| Ticket Servers | O(ids) | O(ids) | Flickr's unique IDs: n database servers each counting by n from their own offset, so they never collide, even when one is down. |
| Transactional Outbox | O(orders) | O(orders) | Write and publish in two steps and a crash loses the event; write the event to an outbox in the same transaction and nothing is lost. |
| Design a URL Shortener | O(links + reads) | O(links) | Short codes from a base-62 counter or a hash (with collisions), and a cache in front of the database for the read-heavy redirects. |
| UUID v4 vs v7 vs ULID | O(n^2) | O(n) | Random UUIDs scatter inserts across an index; time-ordered UUID v7 and ULIDs append at the end like an auto-increment key. |
| Design a Video Upload Pipeline | O(jobs x workers) | O(jobs) | Split uploads into chunks, transcode every chunk to every resolution in parallel on a pool of workers, then serve from a CDN. |
| Design a Web Crawler | O(ticks x fetchers x links) | O(pages) | A URL frontier with a queue per host for politeness, fetchers, a parser, and a seen set so no page is fetched twice. |
