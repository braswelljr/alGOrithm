import { cn } from '@/lib/utils';
import type { BitRow } from '../core/schema';

/** Tint, outline and legend word per role; the field name is always shown
 * as text above it, so no role is told by color alone. */
const roles: Record<string, { cell: string; label: string }> = {
  sign: { cell: 'border-sign bg-sign/15', label: 'text-sign' },
  exponent: { cell: 'border-exponent bg-exponent/15', label: 'text-exponent' },
  fraction: { cell: 'border-fraction bg-fraction/15', label: 'text-fraction' },
  parity: { cell: 'border-parity bg-parity/15', label: 'text-parity' },
  error: {
    cell: 'border-danger bg-danger/20 ring-2 ring-danger',
    label: 'text-danger'
  },
  data: { cell: 'border-accent bg-accent/10', label: 'text-accent' },
  flag: { cell: 'border-border bg-subtle', label: 'text-muted' },
  segment: { cell: 'border-border bg-subtle', label: 'text-muted' },
  byte: { cell: 'border-border bg-card', label: 'text-muted' },
  value: { cell: 'border-border bg-card', label: 'text-muted' }
};

/**
 * Rows of bits split into labeled fields: a float's sign, exponent and
 * fraction, a Hamming codeword with its parity bits, the bytes of a
 * character. Rows appear one at a time up to `upTo`; a 1 is bold and filled,
 * a 0 plain, and each newly shown row's bits turn over once (not under
 * reduced motion).
 */
export function BitsView({ rows, upTo }: { rows: BitRow[]; upTo: number }) {
  return (
    <div className="flex flex-col gap-4">
      {rows.map((row, ri) => {
        const shown = ri < upTo;
        const fresh = ri === upTo - 1;
        return (
          <section
            key={ri}
            aria-label={`${row.label}: ${row.fields.map((f) => `${f.name} ${f.bits}`).join(', ')}${row.note ? `. ${row.note}` : ''}`}
            className={cn('transition-opacity duration-200', shown ? 'opacity-100' : 'opacity-30')}
          >
            <div className="mb-1.5 flex flex-wrap items-baseline gap-x-3">
              <h4 className="text-fg m-0 font-mono text-sm font-semibold">{row.label}</h4>
              {row.note ? <span className="text-muted text-xs">{row.note}</span> : null}
            </div>
            {row.fields.length < 1 ? (
              <p className="text-muted m-0 text-xs italic">no bits</p>
            ) : null}
            <div className="flex flex-wrap gap-x-3 gap-y-2">
              {row.fields.map((f, fi) => {
                const r = roles[f.role] ?? roles.value;
                return (
                  <div
                    key={fi}
                    className="flex flex-col gap-1"
                  >
                    <span className={cn('font-mono text-[0.6875rem] font-medium', r.label)}>
                      {f.name}
                    </span>
                    <div className="flex gap-0.5 [perspective:400px]">
                      {[...f.bits].map((b, bi) => (
                        <span
                          key={bi}
                          className={cn(
                            'grid h-8 w-6 place-items-center rounded-md border font-mono text-sm tabular-nums',
                            r.cell,
                            b === '1' ? 'text-fg font-bold' : 'text-muted',
                            fresh ? 'animate-bit-flip motion-reduce:animate-none' : ''
                          )}
                          style={
                            fresh
                              ? {
                                  animationDelay: `${Math.min(bi, 24) * 12}ms`,
                                  animationFillMode: 'backwards'
                                }
                              : undefined
                          }
                        >
                          {shown ? b : '·'}
                        </span>
                      ))}
                    </div>
                  </div>
                );
              })}
            </div>
          </section>
        );
      })}
    </div>
  );
}
