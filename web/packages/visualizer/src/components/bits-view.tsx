import type { BitRow } from '../core/schema';

/** Tint, outline and legend word per role; the field name is always shown
 * as text above it, so no role is told by color alone. */
const roles: Record<string, { cell: string; label: string }> = {
  sign: { cell: 'av:border-sign av:bg-sign/15', label: 'av:text-sign' },
  exponent: { cell: 'av:border-exponent av:bg-exponent/15', label: 'av:text-exponent' },
  fraction: { cell: 'av:border-fraction av:bg-fraction/15', label: 'av:text-fraction' },
  parity: { cell: 'av:border-parity av:bg-parity/15', label: 'av:text-parity' },
  error: {
    cell: 'av:border-danger av:bg-danger/20 av:ring-2 av:ring-danger',
    label: 'av:text-danger'
  },
  data: { cell: 'av:border-accent av:bg-accent/10', label: 'av:text-accent' },
  flag: { cell: 'av:border-border av:bg-subtle', label: 'av:text-muted' },
  segment: { cell: 'av:border-border av:bg-subtle', label: 'av:text-muted' },
  byte: { cell: 'av:border-border av:bg-card', label: 'av:text-muted' },
  value: { cell: 'av:border-border av:bg-card', label: 'av:text-muted' }
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
    <div className="av:flex av:flex-col av:gap-4">
      {rows.map((row, ri) => {
        const shown = ri < upTo;
        const fresh = ri === upTo - 1;
        return (
          <section
            key={ri}
            aria-label={`${row.label}: ${row.fields.map((f) => `${f.name} ${f.bits}`).join(', ')}${row.note ? `. ${row.note}` : ''}`}
            className={`av:transition-opacity av:duration-200 ${shown ? 'av:opacity-100' : 'av:opacity-30'}`}
          >
            <div className="av:mb-1.5 av:flex av:flex-wrap av:items-baseline av:gap-x-3">
              <h4 className="av:m-0 av:font-mono av:text-sm av:font-semibold av:text-fg">
                {row.label}
              </h4>
              {row.note ? <span className="av:text-xs av:text-muted">{row.note}</span> : null}
            </div>
            {row.fields.length === 0 ? (
              <p className="av:m-0 av:text-xs av:italic av:text-muted">no bits</p>
            ) : null}
            <div className="av:flex av:flex-wrap av:gap-x-3 av:gap-y-2">
              {row.fields.map((f, fi) => {
                const r = roles[f.role] ?? roles.value;
                return (
                  <div
                    key={fi}
                    className="av:flex av:flex-col av:gap-1"
                  >
                    <span className={`av:font-mono av:text-[0.6875rem] av:font-medium ${r.label}`}>
                      {f.name}
                    </span>
                    <div className="av:flex av:gap-0.5 av:[perspective:400px]">
                      {[...f.bits].map((b, bi) => (
                        <span
                          key={bi}
                          className={`av:grid av:h-8 av:w-6 av:place-items-center av:rounded-md av:border av:font-mono av:text-sm av:tabular-nums ${r.cell} ${
                            b === '1' ? 'av:font-bold av:text-fg' : 'av:text-muted'
                          } ${fresh ? 'av:animate-bit-flip av:motion-reduce:animate-none' : ''}`}
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
