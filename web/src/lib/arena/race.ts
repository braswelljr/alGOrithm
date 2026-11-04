/**
 * The path race's clock, apart from drawing so it can be tested. A car waits
 * for its thinking time on the start cell, then enters each next cell of its
 * path in that cell's cost times a step; so a dear cell (a 9) slows everyone,
 * and a pathfinder that weighs the costs finds the faster road.
 */

/** When a car reaches each cell of its path, in milliseconds from the green light. */
export function arrivals(
  path: number[],
  cost: number[],
  stepMs: number,
  thinkMs: number
): number[] {
  const times: number[] = [];
  let t = thinkMs;
  path.forEach((cell, k) => {
    if (k > 0) t += Math.max(1, cost[cell] ?? 1) * stepMs;
    times.push(t);
  });
  return times;
}

/** Where a car is at `t`: the cells it is between and how far along, 0 to 1. */
export function positionAt(path: number[], times: number[], t: number) {
  if (path.length < 1) return null;
  if (t <= times[0]) return { from: path[0], to: path[0], along: 0, done: false };
  for (let k = 1; k < path.length; k++) {
    if (t < times[k]) {
      return {
        from: path[k - 1],
        to: path[k],
        along: (t - times[k - 1]) / (times[k] - times[k - 1]),
        done: false
      };
    }
  }
  const last = path[path.length - 1];
  return { from: last, to: last, along: 1, done: true };
}

/** The places, 1 first, for finishing times (Infinity for a car that did not finish). */
export function places(times: number[]): number[] {
  return times.map((t) => 1 + times.filter((u) => u < t).length);
}

/** Points for a place among `cars`, and a bonus for each rival beaten. */
export function racePoints(place: number, cars: number): number {
  const base = [400, 250, 150][place - 1] ?? 80;
  return base + 20 * (cars - place);
}
