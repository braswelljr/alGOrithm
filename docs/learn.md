# Learning paths

Ordered walks through the catalog, generated from `core.Paths()` (`go run ./tools/gendocs`). Each step is an entry to run, what to notice while it runs, and the command that shows it; `algorithm learn ID` prints the same in the terminal, and the site's Learn page opens each step in the visualizer.

| Path | Level | For | Steps | First |
|------|-------|-----|-------|-------|
| [Big O by example](#big-o-by-example) | beginner | anyone new to algorithms | 8 | none |
| [Sorting from zero](#sorting-from-zero) | beginner | first algorithms course | 10 | [Big O by example](#big-o-by-example) |
| [Searching, and why sorted data matters](#searching-and-why-sorted-data-matters) | beginner | first algorithms course | 8 | [Sorting from zero](#sorting-from-zero) |
| [Data structures: from arrays to trees](#data-structures-from-arrays-to-trees) | beginner | first data structures course | 11 | [Big O by example](#big-o-by-example) |
| [JavaScript arrays and lodash, inside](#javascript-arrays-and-lodash-inside) | beginner | web developers | 12 | none |
| [Thinking recursively, then dynamic programming](#thinking-recursively-then-dynamic-programming) | intermediate | second algorithms course | 10 | [Big O by example](#big-o-by-example) |
| [Greedy choices, and when they fail](#greedy-choices-and-when-they-fail) | intermediate | second algorithms course | 8 | [Thinking recursively, then dynamic programming](#thinking-recursively-then-dynamic-programming) |
| [Pathfinding on grids](#pathfinding-on-grids) | beginner | game and robotics beginners | 5 | [Data structures: from arrays to trees](#data-structures-from-arrays-to-trees) |
| [Graphs step by step](#graphs-step-by-step) | intermediate | second algorithms course | 11 | [Pathfinding on grids](#pathfinding-on-grids) |
| [Searching text](#searching-text) | intermediate | anyone who has used grep | 10 | [Searching, and why sorted data matters](#searching-and-why-sorted-data-matters) |
| [Words and documents](#words-and-documents) | intermediate | anyone building search or text tools | 8 | [Searching text](#searching-text) |
| [Making data smaller](#making-data-smaller) | intermediate | anyone curious about zip files | 6 | [Greedy choices, and when they fail](#greedy-choices-and-when-they-fail) |
| [Numbers and the math under algorithms](#numbers-and-the-math-under-algorithms) | intermediate | learners who like math | 10 | [Big O by example](#big-o-by-example) |
| [Classical ciphers and how they break](#classical-ciphers-and-how-they-break) | beginner | puzzle lovers and security beginners | 9 | none |
| [Crypto internals](#crypto-internals) | advanced | developers who use cryptography libraries | 13 | [Classical ciphers and how they break](#classical-ciphers-and-how-they-break), [Numbers and the math under algorithms](#numbers-and-the-math-under-algorithms) |
| [Hashing inside out](#hashing-inside-out) | intermediate | developers who use maps every day | 10 | [Data structures: from arrays to trees](#data-structures-from-arrays-to-trees) |
| [Machine learning from scratch](#machine-learning-from-scratch) | intermediate | developers curious about ML | 18 | [Numbers and the math under algorithms](#numbers-and-the-math-under-algorithms) |
| [How an operating system works](#how-an-operating-system-works) | intermediate | operating systems course | 17 | [Data structures: from arrays to trees](#data-structures-from-arrays-to-trees) |
| [How systems talk and agree](#how-systems-talk-and-agree) | intermediate | backend developers | 11 | [How an operating system works](#how-an-operating-system-works) |
| [System design from estimates to architectures](#system-design-from-estimates-to-architectures) | intermediate | engineers preparing a design or an interview | 14 | [How systems talk and agree](#how-systems-talk-and-agree) |

## Big O by example

What O(1), O(log n), O(n), O(n log n), O(n^2) and O(2^n) feel like, seen in the counts of real runs.

- Level: beginner, for anyone new to algorithms.
- Read: [docs/guides/big-o.md](../docs/guides/big-o.md).
- In the terminal: `algorithm learn big-o-by-example`.

1. **Hash Table** (`hash-table`): A lookup costs about the same however many keys there are: O(1) on average.

   ```bash
   algorithm run hash-table --example
   ```

2. **Binary Search** (`binary-search`): Count the comparisons: halving the range each time means at most about log2(n) of them: 3 to find 51 among 14 values.

   ```bash
   algorithm run binary-search --example
   ```

3. **Linear Search** (`linear-search`): The same data, one value at a time: the comparisons grow with n, O(n).

   ```bash
   algorithm run linear-search --example
   ```

4. **Merge Sort** (`merge-sort`): Halving and merging: about n log2(n) comparisons, the best a comparison sort can do.

   ```bash
   algorithm run merge-sort 12 11 10 9 8 7 6 5 4 3 2 1
   ```

5. **Bubble Sort** (`bubble-sort`): Every pair compared: on 12 reversed values that is 66 comparisons and 66 swaps, O(n^2).

   ```bash
   algorithm run bubble-sort 12 11 10 9 8 7 6 5 4 3 2 1
   ```

6. **Fibonacci (plain recursion)** (`fibonacci`): Plain recursion recomputes the same values: the calls grow like 1.6^n, so 25 is already slow.

   ```bash
   algorithm run fibonacci 25
   ```

7. **Fibonacci (memoized)** (`fibonacci-memo`): Remember each answer once and the same question takes n steps: O(n) instead of O(1.6^n).

   ```bash
   algorithm run fibonacci-memo 25
   ```

8. **Tower of Hanoi** (`tower-of-hanoi`): Each extra disk doubles the moves: 2^n - 1, the cost of a truly exponential problem.

   ```bash
   algorithm run tower-of-hanoi --example
   ```


## Sorting from zero

From the simplest sorts to the ones in standard libraries, and why input order changes the cost.

- Level: beginner, for first algorithms course.
- Walk first: [Big O by example](#big-o-by-example).
- In the terminal: `algorithm learn sorting-from-zero`.

1. **Bubble Sort** (`bubble-sort`): Watch how many swaps bubble sort makes on reversed input: every pair is out of order.

   ```bash
   algorithm run bubble-sort 12 11 10 9 8 7 6 5 4 3 2 1
   ```

2. **Selection Sort** (`selection-sort`): The same 66 comparisons as bubble sort, but at most one swap per pass: only 6 here.

   ```bash
   algorithm run selection-sort 12 11 10 9 8 7 6 5 4 3 2 1
   ```

3. **Insertion Sort** (`insertion-sort`): On nearly sorted input almost nothing moves: insertion sort is close to O(n) here.

   ```bash
   algorithm run insertion-sort 1 2 3 5 4 6 7 8 10 9 11 12
   ```

4. **Shell Sort** (`shell-sort`): Insertion sort on far-apart elements first, so the last pass has little left to do.

   ```bash
   algorithm run shell-sort --example
   ```

5. **Merge Sort** (`merge-sort`): Split, sort the halves, merge: O(n log n) whatever the input, at the cost of extra space.

   ```bash
   algorithm run merge-sort --example
   ```

6. **Quick Sort** (`quick-sort`): Partition around a pivot: fast on average, and the swaps happen in place.

   ```bash
   algorithm run quick-sort --example
   ```

7. **Heap Sort** (`heap-sort`): Build a heap, then pop the largest again and again: O(n log n) with O(1) extra space.

   ```bash
   algorithm run heap-sort --example
   ```

8. **Counting Sort** (`counting-sort`): No comparisons at all: count each value, then write them out in order.

   ```bash
   algorithm run counting-sort --example
   ```

9. **Radix Sort** (`radix-sort`): Sort by one digit at a time with a stable pass: linear in n for fixed-width keys.

   ```bash
   algorithm run radix-sort --example
   ```

10. **Tim Sort** (`tim-sort`): What Python and Java use: find the runs already sorted and merge them.

   ```bash
   algorithm run tim-sort --example
   ```


## Searching, and why sorted data matters

What sorting buys you: from scanning every value to finding one in a handful of steps.

- Level: beginner, for first algorithms course.
- Walk first: [Sorting from zero](#sorting-from-zero).
- In the terminal: `algorithm learn searching-and-sorted-data`.

1. **Linear Search** (`linear-search`): Works on any data, and looks at every value in the worst case.

   ```bash
   algorithm run linear-search --example
   ```

2. **Binary Search** (`binary-search`): Needs sorted data, and halves the range with each comparison.

   ```bash
   algorithm run binary-search --example
   ```

3. **Lower Bound** (`lower-bound`): Binary search for the first value not less than the target: where it would be inserted.

   ```bash
   algorithm run lower-bound --example
   ```

4. **Jump Search** (`jump-search`): Jump ahead sqrt(n) at a time, then step back: between linear and binary.

   ```bash
   algorithm run jump-search --example
   ```

5. **Interpolation Search** (`interpolation-search`): Guess where the target should be from its value: very fast on evenly spread data.

   ```bash
   algorithm run interpolation-search --example
   ```

6. **Exponential Search** (`exponential-search`): Double the range until it passes the target: useful when the end is unknown.

   ```bash
   algorithm run exponential-search --example
   ```

7. **Quickselect** (`quickselect`): Find the k-th smallest without sorting everything: O(n) on average.

   ```bash
   algorithm run quickselect --example
   ```

8. **Median of Medians** (`median-of-medians`): The same answer with a guaranteed O(n): a careful pivot choice.

   ```bash
   algorithm run median-of-medians --example
   ```


## Data structures: from arrays to trees

How the way data is held decides which operations are fast.

- Level: beginner, for first data structures course.
- Walk first: [Big O by example](#big-o-by-example).
- In the terminal: `algorithm learn arrays-to-trees`.

1. **Dynamic Array** (`dynamic-array`): Appends are cheap until the array is full; then it doubles, so they stay O(1) on average.

   ```bash
   algorithm run dynamic-array --example
   ```

2. **Singly Linked List** (`linked-list`): Inserting at the front is O(1), but finding the k-th value means walking k links.

   ```bash
   algorithm run linked-list --example
   ```

3. **Stack (LIFO)** (`stack`): Last in, first out: the shape of function calls and undo.

   ```bash
   algorithm run stack --example
   ```

4. **Queue (FIFO)** (`queue`): First in, first out: the shape of waiting lines and breadth-first search.

   ```bash
   algorithm run queue --example
   ```

5. **Access Orders: LIFO, FIFO, Priority** (`access-orders`): The same values through a stack, a queue and a heap: three orders back.

   ```bash
   algorithm run access-orders --example
   ```

6. **Hash Table** (`hash-table`): Keys spread over buckets: lookups in O(1) on average.

   ```bash
   algorithm run hash-table --example
   ```

7. **Binary Search Tree** (`binary-search-tree`): Smaller to the left, larger to the right: fast while the tree stays bushy.

   ```bash
   algorithm run binary-search-tree --example
   ```

8. **AVL Tree** (`avl-tree`): Rotations keep the tree balanced, so every operation stays O(log n).

   ```bash
   algorithm run avl-tree --example
   ```

9. **Min-Heap (priority queue)** (`min-heap`): Only the smallest value is easy to find, and that is all a priority queue needs.

   ```bash
   algorithm run min-heap --example
   ```

10. **Trie** (`trie`): A tree of characters: every word with a prefix is below one node.

   ```bash
   algorithm run trie --example
   ```

11. **Union-Find** (`union-find`): Which group is each element in? Nearly O(1) per question with path compression.

   ```bash
   algorithm run union-find --example
   ```


## JavaScript arrays and lodash, inside

The array methods and utility functions you use every day, implemented and traced.

- Level: beginner, for web developers.
- In the terminal: `algorithm learn javascript-and-lodash`.

1. **JavaScript Array** (`js-array`): push, pop, shift and unshift: why shift and unshift move every element.

   ```bash
   algorithm run js-array --example
   ```

2. **Array.prototype.map** (`js-map`): A new array of the same length, one call per element.

   ```bash
   algorithm run js-map --example
   ```

3. **Array.prototype.filter** (`js-filter`): A new array of only the elements that pass.

   ```bash
   algorithm run js-filter --example
   ```

4. **Array.prototype.reduce** (`js-reduce`): Everything above can be written as a reduce.

   ```bash
   algorithm run js-reduce --example
   ```

5. **Array.prototype.splice** (`js-splice`): Removing from the middle shifts everything after it: O(n).

   ```bash
   algorithm run js-splice --example
   ```

6. **Array.prototype.sort** (`js-sort`): The comparator decides the order, and sort changes the array in place.

   ```bash
   algorithm run js-sort --example
   ```

7. **Deep Clone (_.cloneDeep)** (`deep-clone`): A copy with no shared nested objects, so changing one never changes the other.

   ```bash
   algorithm run deep-clone --example
   ```

8. **Deep Merge (_.merge)** (`deep-merge`): Nested objects merged key by key, with the later value winning.

   ```bash
   algorithm run deep-merge --example
   ```

9. **Group By (_.groupBy)** (`group-by`): One pass, one bucket per key.

   ```bash
   algorithm run group-by --example
   ```

10. **Memoize (_.memoize)** (`memoize-demo`): Cache results by argument: the second call with the same input is free.

   ```bash
   algorithm run memoize-demo --example
   ```

11. **Debounce (_.debounce)** (`debounce-demo`): Only the last call in a burst runs.

   ```bash
   algorithm run debounce-demo --example
   ```

12. **Throttle (_.throttle)** (`throttle-demo`): At most one call per interval, however many arrive.

   ```bash
   algorithm run throttle-demo --example
   ```


## Thinking recursively, then dynamic programming

Solve a problem in terms of smaller ones, then stop solving the same smaller ones twice.

- Level: intermediate, for second algorithms course.
- Walk first: [Big O by example](#big-o-by-example).
- In the terminal: `algorithm learn recursion-and-dynamic-programming`.

1. **Factorial** (`factorial`): The smallest recursion: a base case and one smaller call.

   ```bash
   algorithm run factorial --example
   ```

2. **Tower of Hanoi** (`tower-of-hanoi`): Trust the recursion: move n - 1 disks, move one, move n - 1 again.

   ```bash
   algorithm run tower-of-hanoi --example
   ```

3. **Permutations** (`permutations`): Choose each first element, then permute the rest.

   ```bash
   algorithm run permutations --example
   ```

4. **N-Queens** (`n-queens`): Recursion that undoes a choice when it leads nowhere: backtracking.

   ```bash
   algorithm run n-queens --example
   ```

5. **Fibonacci (plain recursion)** (`fibonacci`): Overlapping subproblems: the same calls again and again.

   ```bash
   algorithm run fibonacci 20
   ```

6. **Fibonacci (memoized)** (`fibonacci-memo`): Memoization: the same recursion, each answer computed once.

   ```bash
   algorithm run fibonacci-memo 20
   ```

7. **Coin Change** (`coin-change`): The table fills from small amounts up: each cell uses ones already filled.

   ```bash
   algorithm run coin-change --example
   ```

8. **Longest Common Subsequence** (`longest-common-subsequence`): A two-dimensional table: each cell looks left, up and diagonally.

   ```bash
   algorithm run longest-common-subsequence --example
   ```

9. **Edit Distance** (`edit-distance`): The same table shape, counting insertions, deletions and substitutions.

   ```bash
   algorithm run edit-distance --example
   ```

10. **0/1 Knapsack** (`knapsack-01`): Take an item or leave it: the table remembers the best of both.

   ```bash
   algorithm run knapsack-01 --example
   ```


## Greedy choices, and when they fail

Taking the best-looking choice is fast, and sometimes exactly right; these pairs show when.

- Level: intermediate, for second algorithms course.
- Walk first: [Thinking recursively, then dynamic programming](#thinking-recursively-then-dynamic-programming).
- In the terminal: `algorithm learn greedy-versus-optimal`.

1. **Greedy Coin Change** (`greedy-coin-change`): Coins 1, 3 and 4 for 6: greedy takes 4 + 1 + 1, three coins.

   ```bash
   algorithm run greedy-coin-change --example
   ```

2. **Coin Change** (`coin-change`): Dynamic programming finds 3 + 3, two coins: greedy was not optimal.

   ```bash
   algorithm run coin-change 1 3 4 --amount 6
   ```

3. **Activity Selection** (`activity-selection`): Here greedy is optimal: always take the meeting that ends first.

   ```bash
   algorithm run activity-selection --example
   ```

4. **Fractional Knapsack** (`fractional-knapsack`): When items can be cut, the best value per weight first is optimal.

   ```bash
   algorithm run fractional-knapsack --example
   ```

5. **0/1 Knapsack** (`knapsack-01`): When they cannot, greedy can be beaten: this needs the table.

   ```bash
   algorithm run knapsack-01 --example
   ```

6. **Huffman Coding** (`huffman-coding`): Merge the two rarest symbols first: greedy, and it gives the shortest code.

   ```bash
   algorithm run huffman-coding --example
   ```

7. **Traveling Salesperson (nearest neighbor)** (`tsp-nearest-neighbor`): Always visit the closest city next: quick, and often a poor tour.

   ```bash
   algorithm run tsp-nearest-neighbor --example
   ```

8. **Traveling Salesperson (Held-Karp)** (`tsp-held-karp`): The exact tour, in O(2^n n^2): fine for ten cities, hopeless for a hundred.

   ```bash
   algorithm run tsp-held-karp --example
   ```


## Pathfinding on grids

How a search decides which cell to explore next, and why that changes everything.

- Level: beginner, for game and robotics beginners.
- Walk first: [Data structures: from arrays to trees](#data-structures-from-arrays-to-trees).
- In the terminal: `algorithm learn pathfinding-on-grids`.

1. **Breadth-First Search** (`breadth-first-search`): A queue explores in rings: the first path to the end is a shortest one.

   ```bash
   algorithm run breadth-first-search --example
   ```

2. **Depth-First Search** (`depth-first-search`): A stack dives deep: it finds a path, rarely the shortest.

   ```bash
   algorithm run depth-first-search --example
   ```

3. **Dijkstra's Algorithm** (`dijkstra`): A priority queue by cost so far: shortest paths when cells cost different amounts.

   ```bash
   algorithm run dijkstra --example
   ```

4. **Greedy Best-First Search** (`greedy-best-first`): By distance to the goal only: fast, and fooled by walls.

   ```bash
   algorithm run greedy-best-first --example
   ```

5. **A* Search** (`a-star`): Cost so far plus distance to go: as fast as greedy, as right as Dijkstra.

   ```bash
   algorithm run a-star --example
   ```


## Graphs step by step

Traversals, orders, shortest paths, spanning trees, components and flows on weighted graphs.

- Level: intermediate, for second algorithms course.
- Walk first: [Pathfinding on grids](#pathfinding-on-grids).
- In the terminal: `algorithm learn graphs-step-by-step`.

1. **Breadth-First Traversal** (`graph-bfs`): Visit by distance in edges: the frontier is a queue.

   ```bash
   algorithm run graph-bfs --example
   ```

2. **Depth-First Traversal** (`graph-dfs`): Visit as deep as possible first: the frontier is a stack.

   ```bash
   algorithm run graph-dfs --example
   ```

3. **Topological Sort** (`topological-sort`): An order where every edge points forward: only a graph without cycles has one.

   ```bash
   algorithm run topological-sort --example
   ```

4. **Dijkstra's Shortest Paths** (`graph-dijkstra`): Shortest paths with weights, as long as none is negative.

   ```bash
   algorithm run graph-dijkstra --example
   ```

5. **Bellman-Ford** (`bellman-ford`): Negative weights allowed: relax every edge n - 1 times, and detect negative cycles.

   ```bash
   algorithm run bellman-ford --example
   ```

6. **Floyd-Warshall** (`floyd-warshall`): Every pair at once, in a table: O(n^3).

   ```bash
   algorithm run floyd-warshall --example
   ```

7. **Kruskal's Minimum Spanning Tree** (`kruskal`): Cheapest edges first, skipping any that would close a cycle: a minimum spanning tree.

   ```bash
   algorithm run kruskal --example
   ```

8. **Prim's Minimum Spanning Tree** (`prim`): The same tree grown from one node, always by its cheapest edge out.

   ```bash
   algorithm run prim --example
   ```

9. **Tarjan's Strongly Connected Components** (`tarjan-scc`): Groups where every node reaches every other, in one depth-first pass.

   ```bash
   algorithm run tarjan-scc --example
   ```

10. **Maximum Flow (Edmonds-Karp)** (`max-flow`): Push flow along paths with room left until none is left.

   ```bash
   algorithm run max-flow --example
   ```

11. **PageRank** (`pagerank`): Importance as the chance a random walk is at a node.

   ```bash
   algorithm run pagerank --example
   ```


## Searching text

Finding patterns in text without reading it twice, and the structures that make it fast.

- Level: intermediate, for anyone who has used grep.
- Walk first: [Searching, and why sorted data matters](#searching-and-why-sorted-data-matters).
- In the terminal: `algorithm learn text-and-patterns`.

1. **Naive String Search** (`naive-search`): Try every position: O(n m) when the text is full of near matches.

   ```bash
   algorithm run naive-search --example
   ```

2. **Knuth-Morris-Pratt** (`knuth-morris-pratt`): Never move back in the text: the pattern's own structure says how far to jump.

   ```bash
   algorithm run knuth-morris-pratt --example
   ```

3. **Z-Algorithm** (`z-algorithm`): For each position, how long a match with the start: another linear-time view.

   ```bash
   algorithm run z-algorithm --example
   ```

4. **Boyer-Moore-Horspool** (`boyer-moore-horspool`): Compare from the end of the pattern and skip ahead on a mismatch.

   ```bash
   algorithm run boyer-moore-horspool --example
   ```

5. **Rabin-Karp** (`rabin-karp`): Compare hashes of windows instead of the windows themselves.

   ```bash
   algorithm run rabin-karp --example
   ```

6. **Aho-Corasick** (`aho-corasick`): Many patterns at once in a single pass over the text.

   ```bash
   algorithm run aho-corasick --example
   ```

7. **Suffix Array** (`suffix-array`): Every suffix, sorted: any substring search becomes a binary search.

   ```bash
   algorithm run suffix-array --example
   ```

8. **Longest Repeated Substring** (`longest-repeated-substring`): With the suffixes sorted, repeats sit side by side: compare neighbors only.

   ```bash
   algorithm run longest-repeated-substring --example
   ```

9. **Longest Common Substring** (`longest-common-substring`): The same trick across two texts.

   ```bash
   algorithm run longest-common-substring --example
   ```

10. **Manacher's Algorithm** (`manacher`): Every longest palindrome around every center, in linear time.

   ```bash
   algorithm run manacher --example
   ```


## Words and documents

From counting words to ranking documents, suggesting spellings and showing what changed.

- Level: intermediate, for anyone building search or text tools.
- Walk first: [Searching text](#searching-text).
- In the terminal: `algorithm learn words-and-documents`.

1. **Word Frequency** (`word-frequency`): The first thing any text tool does: count the words.

   ```bash
   algorithm run word-frequency --example
   ```

2. **Character N-Grams** (`n-grams`): Runs of n words or characters: the context a single word lacks.

   ```bash
   algorithm run n-grams --example
   ```

3. **TF-IDF** (`tf-idf`): A word matters in a document if it is frequent there and rare elsewhere.

   ```bash
   algorithm run tf-idf --example
   ```

4. **Soundex** (`soundex`): Names that sound alike get the same code, however they are spelled.

   ```bash
   algorithm run soundex --example
   ```

5. **Fuzzy Search (bitap)** (`fuzzy-search`): Find matches that are close, not exact: edit distance at work.

   ```bash
   algorithm run fuzzy-search --example
   ```

6. **Spelling Suggestions** (`spelling-suggestions`): The nearest dictionary words to a misspelling.

   ```bash
   algorithm run spelling-suggestions --example
   ```

7. **Line Diff (Myers)** (`myers-diff`): The fewest insertions and deletions between two texts: what diff and git show.

   ```bash
   algorithm run myers-diff --example
   ```

8. **Text Justification** (`text-justification`): Break lines and spread spaces so both edges line up.

   ```bash
   algorithm run text-justification --example
   ```


## Making data smaller

Lossless compression: repeated runs, frequent symbols, repeated phrases, and transforms that help.

- Level: intermediate, for anyone curious about zip files.
- Walk first: [Greedy choices, and when they fail](#greedy-choices-and-when-they-fail).
- In the terminal: `algorithm learn making-data-smaller`.

1. **Run-Length Encoding** (`run-length-encoding`): Runs of the same symbol become a count: great for some data, worse for most.

   ```bash
   algorithm run run-length-encoding --example
   ```

2. **Huffman Coding** (`huffman-coding`): Frequent symbols get short codes.

   ```bash
   algorithm run huffman-coding --example
   ```

3. **LZ77** (`lz77`): Repeated phrases become a pointer back to where they appeared.

   ```bash
   algorithm run lz77 --example
   ```

4. **LZW** (`lzw`): A dictionary of phrases built as the text is read.

   ```bash
   algorithm run lzw --example
   ```

5. **Burrows-Wheeler Transform** (`burrows-wheeler`): A reversible shuffle that groups similar characters together.

   ```bash
   algorithm run burrows-wheeler --example
   ```

6. **Move-to-Front** (`move-to-front`): After that shuffle, recent symbols become small numbers that compress well.

   ```bash
   algorithm run move-to-front --example
   ```


## Numbers and the math under algorithms

Divisibility, primes, modular arithmetic and fast arithmetic: the toolkit behind cryptography.

- Level: intermediate, for learners who like math.
- Walk first: [Big O by example](#big-o-by-example).
- In the terminal: `algorithm learn numbers-and-math`.

1. **Euclidean GCD** (`gcd`): Euclid's algorithm: the remainder shrinks quickly, about log steps.

   ```bash
   algorithm run gcd --example
   ```

2. **Extended Euclidean Algorithm** (`extended-gcd`): The same steps also give the coefficients of the gcd.

   ```bash
   algorithm run extended-gcd --example
   ```

3. **Sieve of Eratosthenes** (`sieve-of-eratosthenes`): Every prime up to n by crossing out multiples.

   ```bash
   algorithm run sieve-of-eratosthenes --example
   ```

4. **Fast Exponentiation** (`fast-power`): Square and multiply: x^n in log2(n) multiplications.

   ```bash
   algorithm run fast-power --example
   ```

5. **Modular Inverse** (`mod-inverse`): Division in modular arithmetic, from the extended gcd.

   ```bash
   algorithm run mod-inverse --example
   ```

6. **Miller-Rabin Primality** (`miller-rabin`): Is a huge number prime? A few random witnesses make an error very unlikely.

   ```bash
   algorithm run miller-rabin --example
   ```

7. **Chinese Remainder Theorem** (`chinese-remainder`): One number from its remainders by several moduli.

   ```bash
   algorithm run chinese-remainder --example
   ```

8. **Karatsuba Multiplication** (`karatsuba`): Multiply big numbers with three half-size products instead of four.

   ```bash
   algorithm run karatsuba --example
   ```

9. **Newton-Raphson Method** (`newton-raphson`): Follow the tangent to a root: the correct digits roughly double each step.

   ```bash
   algorithm run newton-raphson --example
   ```

10. **Fast Fourier Transform** (`fft`): Multiply polynomials in O(n log n) by working with their values instead.

   ```bash
   algorithm run fft --example
   ```


## Classical ciphers and how they break

Two thousand years of secret writing, and the attacks that make each one obsolete.

- Level: beginner, for puzzle lovers and security beginners.
- In the terminal: `algorithm learn classical-ciphers`.

1. **Caesar Cipher** (`caesar-cipher`): Shift every letter by the same amount.

   ```bash
   algorithm run caesar-cipher --example
   ```

2. **Caesar Brute Force** (`caesar-brute-force`): Only 25 keys: try them all and read the one that makes sense.

   ```bash
   algorithm run caesar-brute-force --example
   ```

3. **Substitution Cipher** (`substitution-cipher`): A shuffled alphabet: too many keys to try.

   ```bash
   algorithm run substitution-cipher --example
   ```

4. **Frequency Analysis** (`frequency-analysis`): But letter frequencies survive, and give the key away.

   ```bash
   algorithm run frequency-analysis --example
   ```

5. **Vigenère Cipher** (`vigenere-cipher`): Several shifts in turn hide the frequencies.

   ```bash
   algorithm run vigenere-cipher --example
   ```

6. **Vigenère Key Length** (`vigenere-key-length`): Repeated fragments reveal the length of the key.

   ```bash
   algorithm run vigenere-key-length --example
   ```

7. **Vigenère Crack** (`vigenere-crack`): Then each position is a Caesar cipher again.

   ```bash
   algorithm run vigenere-crack --example
   ```

8. **One-Time Pad** (`one-time-pad`): A random key as long as the message: unbreakable, if never reused.

   ```bash
   algorithm run one-time-pad --example
   ```

9. **Enigma Machine** (`enigma`): Rotors that turn with every letter: a machine that changes its cipher each keypress.

   ```bash
   algorithm run enigma --example
   ```


## Crypto internals

What happens inside hashes, ciphers and key exchange. For learning only: never protect data with these.

- Level: advanced, for developers who use cryptography libraries.
- Walk first: [Classical ciphers and how they break](#classical-ciphers-and-how-they-break), [Numbers and the math under algorithms](#numbers-and-the-math-under-algorithms).
- In the terminal: `algorithm learn crypto-internals`.

1. **Base64** (`base64`): Encoding is not encryption: anyone can reverse it.

   ```bash
   algorithm run base64 --example
   ```

2. **SHA-256** (`sha-256`): A fixed-size fingerprint: change one bit of input and half the output bits change.

   ```bash
   algorithm run sha-256 --example
   ```

3. **HMAC-SHA256** (`hmac-sha256`): A hash with a key: proves who wrote the message.

   ```bash
   algorithm run hmac-sha256 --example
   ```

4. **PBKDF2** (`pbkdf2`): A hash run thousands of times, so guessing passwords is slow.

   ```bash
   algorithm run pbkdf2 --example
   ```

5. **Constant-Time Compare** (`constant-time-compare`): Compare secrets without leaking how many bytes matched.

   ```bash
   algorithm run constant-time-compare --example
   ```

6. **Feistel Network** (`feistel-network`): The round structure behind DES: encryption and decryption share the code.

   ```bash
   algorithm run feistel-network --example
   ```

7. **AES-128** (`aes-128`): Substitute, shift, mix and add the key, ten rounds over.

   ```bash
   algorithm run aes-128 --example
   ```

8. **ChaCha20** (`chacha20`): A stream cipher built only from add, rotate and xor.

   ```bash
   algorithm run chacha20 --example
   ```

9. **Diffie-Hellman** (`diffie-hellman`): Two parties agree on a secret over a public channel.

   ```bash
   algorithm run diffie-hellman --example
   ```

10. **RSA** (`rsa`): Easy to multiply two primes, hard to factor the product.

   ```bash
   algorithm run rsa --example
   ```

11. **Elliptic Curve** (`elliptic-curve`): The same idea on a curve, with much smaller keys.

   ```bash
   algorithm run elliptic-curve --example
   ```

12. **Merkle Tree** (`merkle-tree`): Hashes of hashes: prove one block is in a large set with a few hashes.

   ```bash
   algorithm run merkle-tree --example
   ```

13. **Shamir Secret Sharing** (`shamir-secret-sharing`): Split a secret so any k of n shares recover it, and fewer reveal nothing.

   ```bash
   algorithm run shamir-secret-sharing --example
   ```


## Hashing inside out

Hash functions, and what a hash table does when two keys land in the same place.

- Level: intermediate, for developers who use maps every day.
- Walk first: [Data structures: from arrays to trees](#data-structures-from-arrays-to-trees).
- In the terminal: `algorithm learn hashing-inside-out`.

1. **DJB2** (`djb2`): A classic string hash: multiply by 33 and add each character.

   ```bash
   algorithm run djb2 --example
   ```

2. **FNV-1a (32 and 64 bit)** (`fnv1a`): Xor then multiply: simple, fast and well spread.

   ```bash
   algorithm run fnv1a --example
   ```

3. **Hash Distribution** (`hash-distribution`): How evenly a hash spreads keys over buckets.

   ```bash
   algorithm run hash-distribution --example
   ```

4. **Linear Probing** (`linear-probing`): On a collision, take the next free slot: fast until clusters form.

   ```bash
   algorithm run linear-probing --example
   ```

5. **Quadratic Probing** (`quadratic-probing`): Jump further each time to break up clusters.

   ```bash
   algorithm run quadratic-probing --example
   ```

6. **Double Hashing** (`double-hashing`): A second hash decides the jump.

   ```bash
   algorithm run double-hashing --example
   ```

7. **Robin Hood Hashing** (`robin-hood-hashing`): Rich keys give their slot to poor ones: probe lengths even out.

   ```bash
   algorithm run robin-hood-hashing --example
   ```

8. **Cuckoo Hashing** (`cuckoo-hashing`): Two places per key, and kicking out whoever is there: O(1) worst-case lookups.

   ```bash
   algorithm run cuckoo-hashing --example
   ```

9. **Bloom Filter** (`bloom-filter`): A few bits per key: never a false no, sometimes a false yes.

   ```bash
   algorithm run bloom-filter --example
   ```

10. **Consistent Hashing** (`consistent-hashing`): Add a server and only a few keys move.

   ```bash
   algorithm run consistent-hashing --example
   ```


## Machine learning from scratch

Fit a line, then classify, cluster and train a tiny neural network, with every number visible.

- Level: intermediate, for developers curious about ML.
- Walk first: [Numbers and the math under algorithms](#numbers-and-the-math-under-algorithms).
- In the terminal: `algorithm learn machine-learning-from-scratch`.

1. **Linear Regression** (`linear-regression`): The line that makes the squared errors smallest, solved in one formula.

   ```bash
   algorithm run linear-regression --example
   ```

2. **Gradient Descent (linear regression)** (`gradient-descent`): The same line found by stepping downhill on the error: how most models learn.

   ```bash
   algorithm run gradient-descent --example
   ```

3. **Polynomial Regression** (`polynomial-regression`): More terms fit the data better, and past a point fit the noise too.

   ```bash
   algorithm run polynomial-regression --example
   ```

4. **Mean Squared Error and R^2** (`regression-metrics`): How wrong is the model? Mean squared error, mean absolute error and R squared.

   ```bash
   algorithm run regression-metrics --example
   ```

5. **Train/Test Split** (`train-test-split`): Judge a model on data it has not seen, or the score means little.

   ```bash
   algorithm run train-test-split --example
   ```

6. **K-Nearest Neighbors** (`k-nearest-neighbors`): No training at all: a new point takes the class of its nearest neighbors.

   ```bash
   algorithm run k-nearest-neighbors --example
   ```

7. **Logistic Regression** (`logistic-regression`): A line that outputs a probability, for yes or no questions.

   ```bash
   algorithm run logistic-regression --example
   ```

8. **Confusion Matrix** (`confusion-matrix`): Which classes the model mixes up, not just how often it is right.

   ```bash
   algorithm run confusion-matrix --example
   ```

9. **Precision, Recall and F1** (`precision-recall-f1`): When one class is rare, accuracy hides the misses; precision and recall do not.

   ```bash
   algorithm run precision-recall-f1 --example
   ```

10. **Decision Tree (CART)** (`decision-tree`): A sequence of yes or no questions, each chosen to split the classes best.

   ```bash
   algorithm run decision-tree --example
   ```

11. **Naive Bayes Spam Filter** (`naive-bayes-text`): Spam or not, by multiplying how likely each word is in each class.

   ```bash
   algorithm run naive-bayes-text --example
   ```

12. **K-Means Clustering** (`k-means`): Move each center to the mean of its points until nothing changes.

   ```bash
   algorithm run k-means --example
   ```

13. **K-Means++ Seeding** (`k-means-plus-plus`): Spread the first centers out, and k-means settles faster and better.

   ```bash
   algorithm run k-means-plus-plus --example
   ```

14. **Silhouette Score** (`silhouette-score`): How well each point sits in its cluster: a way to choose k.

   ```bash
   algorithm run silhouette-score --example
   ```

15. **DBSCAN** (`dbscan`): Clusters as dense regions, of any shape, with noise left out.

   ```bash
   algorithm run dbscan --example
   ```

16. **Perceptron** (`perceptron`): One neuron: a weighted sum and a threshold, nudged after every mistake.

   ```bash
   algorithm run perceptron --example
   ```

17. **Activation Functions** (`activation-functions`): The bends that let a network be more than one big linear function.

   ```bash
   algorithm run activation-functions --example
   ```

18. **Neural Network Learning XOR** (`xor-network`): XOR is not linearly separable: one hidden layer and backpropagation learn it.

   ```bash
   algorithm run xor-network --example
   ```


## How an operating system works

Processes and the CPU, then memory, disks, deadlock and threads: the decisions an OS makes every millisecond.

- Level: intermediate, for operating systems course.
- Walk first: [Data structures: from arrays to trees](#data-structures-from-arrays-to-trees).
- In the terminal: `algorithm learn how-an-os-works`.

1. **First Come, First Served** (`fcfs-scheduling`): Processes run in arrival order: one long job makes everyone behind it wait.

   ```bash
   algorithm run fcfs-scheduling --example
   ```

2. **Shortest Job First** (`sjf-scheduling`): Shortest job first gives the lowest average waiting time, if the lengths are known.

   ```bash
   algorithm run sjf-scheduling --example
   ```

3. **Round Robin** (`round-robin`): Each process gets a slice of time in turn: try quantum 1 and quantum 10 and compare the waiting times.

   ```bash
   algorithm run round-robin --text 'name arrival burst priority
   P1 0 5 1
   P2 1 3 2
   P3 2 8 1
   P4 3 6 3' --option quantum=1
   ```

4. **Multilevel Feedback Queue** (`mlfq-scheduling`): Jobs that use their whole slice drop to a lower queue: interactive ones stay responsive.

   ```bash
   algorithm run mlfq-scheduling --example
   ```

5. **First Fit** (`first-fit`): Place a process in the first hole big enough: holes appear between allocations.

   ```bash
   algorithm run first-fit --example
   ```

6. **Address Translation (paging and TLB)** (`address-translation`): A virtual address becomes a page number and an offset, then a physical frame.

   ```bash
   algorithm run address-translation --example
   ```

7. **FIFO Page Replacement** (`fifo-page-replacement`): Evict the oldest page: with 3 frames this string faults 9 times.

   ```bash
   algorithm run fifo-page-replacement 1 2 3 4 1 2 5 1 2 3 4 5 --option frames=3
   ```

8. **FIFO Page Replacement** (`fifo-page-replacement`): Belady's anomaly: with 4 frames the same string faults 10 times, more memory and more faults.

   ```bash
   algorithm run fifo-page-replacement 1 2 3 4 1 2 5 1 2 3 4 5 --option frames=4
   ```

9. **LRU Page Replacement** (`lru-page-replacement`): Evict the page used least recently: no anomaly.

   ```bash
   algorithm run lru-page-replacement --example
   ```

10. **Optimal Page Replacement** (`optimal-page-replacement`): Evict the page used furthest in the future: the best possible, and impossible in practice.

   ```bash
   algorithm run optimal-page-replacement --example
   ```

11. **FCFS Disk Scheduling** (`fcfs-disk`): Serve disk requests in order: the head zigzags across the platter.

   ```bash
   algorithm run fcfs-disk --example
   ```

12. **SCAN (Elevator)** (`scan-disk`): Sweep one way, then back, like an elevator: far less head movement.

   ```bash
   algorithm run scan-disk --example
   ```

13. **Banker's Algorithm** (`bankers-algorithm`): Grant a request only if every process can still finish: no deadlock.

   ```bash
   algorithm run bankers-algorithm --example
   ```

14. **Race Condition** (`race-condition`): Two threads add to a counter without a lock: some updates are lost.

   ```bash
   algorithm run race-condition --example
   ```

15. **Producer-Consumer** (`producer-consumer`): A bounded buffer: producers wait when it is full, consumers when it is empty.

   ```bash
   algorithm run producer-consumer --example
   ```

16. **Dining Philosophers** (`dining-philosophers`): Everyone takes their left fork and waits for the right: deadlock, then the fix.

   ```bash
   algorithm run dining-philosophers --example
   ```

17. **CPU Cache Simulator** (`cache-simulator`): Addresses map to cache lines: watch hits and misses as the program walks memory.

   ```bash
   algorithm run cache-simulator --example
   ```


## How systems talk and agree

From one lossy link to a cluster that agrees: retransmission, congestion, clocks, quorums, consensus, and the infrastructure that keeps it up.

- Level: intermediate, for backend developers.
- Walk first: [How an operating system works](#how-an-operating-system-works).
- In the terminal: `algorithm learn how-systems-talk-and-agree`.

1. **Stop-and-Wait ARQ** (`stop-and-wait`): Send one frame, wait for its ack, resend on a timeout: reliable, but one frame per round trip.

   ```bash
   algorithm run stop-and-wait --example
   ```

2. **Go-Back-N ARQ** (`go-back-n`): A window of frames in flight: faster, but one loss resends everything after it.

   ```bash
   algorithm run go-back-n --example
   ```

3. **TCP Congestion Control** (`tcp-congestion-control`): Slow start doubles the window, a loss halves it: TCP's sawtooth.

   ```bash
   algorithm run tcp-congestion-control --example
   ```

4. **DNS Resolution** (`dns-resolution`): A name walks from the root to the authoritative server, then sits in caches.

   ```bash
   algorithm run dns-resolution --example
   ```

5. **Lamport Clocks** (`lamport-clock`): No shared clock: order events by counters that every message carries.

   ```bash
   algorithm run lamport-clock --example
   ```

6. **Vector Clocks** (`vector-clock`): One counter per node tells concurrent events apart from ordered ones.

   ```bash
   algorithm run vector-clock --example
   ```

7. **Quorum Reads and Writes** (`quorum-consistency`): With R + W > N every read meets the latest write.

   ```bash
   algorithm run quorum-consistency --example
   ```

8. **Two-Phase Commit** (`two-phase-commit`): Everyone votes, then everyone commits; one no aborts all, and a crashed coordinator blocks them.

   ```bash
   algorithm run two-phase-commit --example
   ```

9. **Raft Consensus** (`raft`): Elect a leader, replicate its log to a majority: at most one leader per term.

   ```bash
   algorithm run raft --example
   ```

10. **Round-Robin Load Balancing** (`round-robin-load-balancer`): Spread requests across servers; then compare least connections.

   ```bash
   algorithm run round-robin-load-balancer --example
   ```

11. **Circuit Breaker** (`circuit-breaker`): Stop calling a failing dependency, then test it again after a timeout.

   ```bash
   algorithm run circuit-breaker --example
   ```


## System design from estimates to architectures

Size a system, give it IDs, keep it up, then design real ones: a URL shortener, a key-value store, a feed, payments, and the data patterns behind them.

- Level: intermediate, for engineers preparing a design or an interview.
- Walk first: [How systems talk and agree](#how-systems-talk-and-agree).
- Read: [docs/guides/system-design.md](../docs/guides/system-design.md).
- In the terminal: `algorithm learn system-design`.

1. **Back-of-the-Envelope Estimation** (`back-of-the-envelope`): Start with numbers: 10 million users reading 20 times a day is about 2,300 reads a second, peaking near 7,000.

   ```bash
   algorithm run back-of-the-envelope --example
   ```

2. **Latency Numbers Every Programmer Should Know** (`latency-numbers`): Memory is 100 times a cache hit, a datacenter round trip 5,000 times: keep the hot path in memory.

   ```bash
   algorithm run latency-numbers --example
   ```

3. **Availability in Series and Parallel** (`availability-nines`): Each dependency in series costs availability; each replica in parallel adds a nine.

   ```bash
   algorithm run availability-nines --example
   ```

4. **Snowflake IDs** (`snowflake-id`): IDs that sort by time with no coordination, and what a clock going back does.

   ```bash
   algorithm run snowflake-id --example
   ```

5. **Design a URL Shortener** (`url-shortener`): A counter or a hash for codes, and a cache that serves most redirects.

   ```bash
   algorithm run url-shortener --example
   ```

6. **Leader-Follower Replication** (`leader-follower-replication`): Followers lag; a user can miss their own write unless reads follow the leader for a moment.

   ```bash
   algorithm run leader-follower-replication --example
   ```

7. **Range vs Hash vs Directory Sharding** (`sharding-strategies`): Range sharding piles hot keys on one shard; hashing spreads them.

   ```bash
   algorithm run sharding-strategies --example
   ```

8. **Design a Key-Value Store (Dynamo)** (`key-value-store`): A ring, N replicas, R and W quorums, and sloppy quorums while a node is down.

   ```bash
   algorithm run key-value-store --example
   ```

9. **Design a News Feed** (`news-feed`): Push or pull, and the celebrity whose post would be millions of writes.

   ```bash
   algorithm run news-feed --example
   ```

10. **CAP and PACELC** (`cap-pacelc`): During a partition, refuse writes or diverge; without one, pay latency or risk staleness.

   ```bash
   algorithm run cap-pacelc --example
   ```

11. **Transactional Outbox** (`transactional-outbox`): Two writes without a transaction lose events; one transaction with an outbox does not.

   ```bash
   algorithm run transactional-outbox --example
   ```

12. **The Saga Pattern** (`saga-pattern`): A failure undoes the finished steps in reverse order.

   ```bash
   algorithm run saga-pattern --example
   ```

13. **Design a Payment System** (`payment-system`): Idempotency keys stop double charges; the ledger always sums to zero.

   ```bash
   algorithm run payment-system --example
   ```

14. **Cache Stampede** (`cache-stampede`): A hot key expires: with the lock fix, one database read instead of dozens.

   ```bash
   algorithm run cache-stampede --example
   ```
