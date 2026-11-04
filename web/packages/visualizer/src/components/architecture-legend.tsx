/** The marks a system design's requests are drawn with, and their legend,
 * apart from the React Flow view so a page can show them without loading it. */
import { STATUS_WORDS } from '../core/replay/architecture';
import type { RequestStatus } from '../core/schema';

/** The shape each request status is drawn as, so status never depends on
 * color alone: a filled dot for ok and hit (hit adds a ring), a hollow dot
 * for a miss, a square for queued, a cross for an error, a triangle for a
 * dropped request. */
export function StatusMark({ status, size = 10 }: { status: RequestStatus; size?: number }) {
  const r = size / 2;
  const color =
    status === 'error' || status === 'dropped'
      ? 'var(--av-danger, var(--av-accent))'
      : status === 'miss' || status === 'queued'
        ? 'var(--av-muted)'
        : 'var(--av-accent)';
  switch (status) {
    case 'hit':
      return (
        <g>
          <circle
            r={r + 2}
            fill="none"
            stroke={color}
            strokeWidth="1.5"
          />
          <circle
            r={r - 1}
            fill={color}
          />
        </g>
      );
    case 'miss':
      return (
        <circle
          r={r - 1}
          fill="var(--av-bg, white)"
          stroke={color}
          strokeWidth="2"
        />
      );
    case 'queued':
      return (
        <rect
          x={-r + 1}
          y={-r + 1}
          width={size - 2}
          height={size - 2}
          fill={color}
        />
      );
    case 'error':
      return (
        <path
          d={`M${-r} ${-r}L${r} ${r}M${r} ${-r}L${-r} ${r}`}
          stroke={color}
          strokeWidth="2.5"
        />
      );
    case 'dropped':
      return (
        <path
          d={`M0 ${-r}L${r} ${r}L${-r} ${r}Z`}
          fill={color}
        />
      );
    default:
      return (
        <circle
          r={r - 1}
          fill={color}
        />
      );
  }
}

/** What each mark means, in words beside its shape. */
export function ArchitectureLegend() {
  const statuses: RequestStatus[] = ['ok', 'hit', 'miss', 'queued', 'error', 'dropped'];
  return (
    <ul
      className="av:m-0 av:flex av:list-none av:flex-wrap av:gap-3 av:p-0 av:text-xs av:text-muted"
      aria-label="Request marks"
    >
      {statuses.map((s) => (
        <li
          key={s}
          className="av:flex av:items-center av:gap-1.5"
        >
          <svg
            width="16"
            height="16"
            viewBox="-8 -8 16 16"
            aria-hidden
          >
            <StatusMark status={s} />
          </svg>
          {STATUS_WORDS[s]}
        </li>
      ))}
    </ul>
  );
}
