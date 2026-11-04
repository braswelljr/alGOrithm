/** The swap race's arithmetic. */

/** The fewest swaps that sort `values`: its length less its cycles. */
export function fewestSwaps(values: number[]): number {
  const order = values.map((v, i) => [v, i] as const).sort((a, b) => a[0] - b[0]);
  const seen = new Array<boolean>(values.length).fill(false);
  let swaps = 0;
  for (let i = 0; i < values.length; i++) {
    if (seen[i] || order[i][1] === i) continue;
    let j = i;
    let cycle = 0;
    while (!seen[j]) {
      seen[j] = true;
      j = order[j][1];
      cycle++;
    }
    swaps += cycle - 1;
  }
  return swaps;
}
