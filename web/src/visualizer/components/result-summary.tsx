import type { Entry, Input, Output } from '../core/schema';

/** The run's answer in a sentence, for every kind of entry. */
export function summarize(
  entry: Pick<Entry, 'category' | 'id'> | undefined,
  input: Input,
  o: Output
): string {
  if (o.timeline.length) {
    const end = Math.max(...o.timeline.map((s) => s.end));
    return `Done at time ${end}.`;
  }
  if (o.track) {
    let moved = 0;
    for (let i = 1; i < o.track.path.length; i++)
      moved += Math.abs(o.track.path[i] - o.track.path[i - 1]);
    return `The head moved ${moved} tracks over ${Math.max(o.track.path.length - 1, 0)} seeks.`;
  }
  if (o.grid) {
    return o.path.length
      ? `A path of ${o.path.length - 1} steps${o.number !== undefined ? ` costing ${o.number}` : ''}.`
      : 'No path from the start to the end.';
  }
  if (o.matches) {
    return o.matches.length
      ? `${o.matches.length} match${o.matches.length === 1 ? '' : 'es'}, at ${o.matches.join(', ')}.`
      : 'No match.';
  }
  if (o.table) {
    if (o.text !== undefined && o.text !== '')
      return `"${o.text}" (length ${o.number ?? o.text.length}).`;
    if (o.values.length && entry?.id !== 'longest-increasing-subsequence') {
      return `${o.number} coins: ${o.values.join(', ')}.`;
    }
    if (o.values.length) return `${o.values.join(', ')} (length ${o.number}).`;
    if (o.number !== undefined)
      return o.number < 0 ? 'It cannot be made.' : `The answer is ${o.text ?? o.number}.`;
  }
  if (o.index !== undefined) {
    return o.index < 0
      ? `${input.target} is not there.`
      : `${input.target} is at position ${o.index} of the ${entry?.category === 'searching' || entry?.category === 'search' ? 'sorted ' : ''}list.`;
  }
  if (o.steps.length) {
    const last = o.steps[o.steps.length - 1];
    return `${o.steps.length - 1} operation${o.steps.length === 2 ? '' : 's'}; the last: ${last.error ?? last.result ?? last.op.name}.`;
  }
  if (o.lines.length) return `${o.lines.length} moves.`;
  if (o.values.length)
    return `Sorted: ${o.values.slice(0, 30).join(', ')}${o.values.length > 30 ? ', ...' : ''}.`;
  if (o.bool !== undefined) return o.bool ? 'Yes.' : 'No.';
  if (o.number !== undefined) return `The answer is ${o.text ?? o.number}.`;
  if (o.text !== undefined) return o.text;
  return 'Done.';
}

export function ResultSummary(props: {
  entry?: Pick<Entry, 'category' | 'id'>;
  input: Input;
  output: Output;
}) {
  return (
    <p>
      <strong>Result: </strong>
      {summarize(props.entry, props.input, props.output)}
    </p>
  );
}
