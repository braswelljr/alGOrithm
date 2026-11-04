/** What each input kind means, in words, for the docs. */
import type { Entry } from '@/visualizer/core';

export const inputWords: Record<Entry['input'], string> = {
  ints: 'a list of whole numbers',
  'sorted-ints': 'a list of whole numbers (sorted first) and a target to find',
  'any-ints': 'a list of whole numbers, in any order, and a target to find',
  text: 'a text',
  'two-texts': 'two words',
  number: 'one whole number',
  'two-numbers': 'two whole numbers',
  ops: 'starting values and a list of operations',
  grid: 'a grid of walls, open cells and costs, with a start and an end',
  pattern: 'a text and a pattern to find in it',
  amount: 'a list of coin values and an amount',
  processes: 'processes, one per line: name, arrival time, burst time and an optional priority',
  pages: 'a reference string of page numbers',
  tracks: 'the disk tracks requested',
  partitions: 'memory partition sizes and process sizes',
  banker: "the Banker's algorithm matrices: available, maximum and allocated resources",
  simulation: 'its options only',
  graph: 'the edges of a graph, one a line, with optional weights',
  points: 'points in the plane, one "x y" a line'
};
