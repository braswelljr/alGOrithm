// Package dynamic holds dynamic-programming algorithms, one file each: edit
// distance, longest common subsequence, coin change and longest increasing
// subsequence. Each fills a table of answers to smaller problems and reads its
// answer from the table; TraceX reports each cell written (Set, the index of
// the cell row by row), the cells read to compute it (Compare), and the cells
// of the answer (Mark "path"), which is what the visualizer draws:
//
//	d, table := dynamic.EditDistance("kitten", "sitting") // 3
package dynamic
