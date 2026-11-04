/** Lines of text a run returns: an explanation, an order, a comparison. */
export function LinesView({ lines, upTo }: { lines: string[]; upTo?: number }) {
  const shown = upTo === undefined ? lines : lines.slice(0, upTo);
  return (
    <ol className="av-lines">
      {shown.map((l, i) => (
        <li key={i}>{l}</li>
      ))}
    </ol>
  );
}
