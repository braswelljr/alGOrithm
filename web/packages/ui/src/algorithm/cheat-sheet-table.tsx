import { LevelBadge } from '@braswelljr/algorithm-ui/algorithm/level-badge';
import { Card } from '@braswelljr/algorithm-ui/components/card';
import { cn } from '@braswelljr/algorithm-ui/lib/utils';

export type CheatSheetRow = {
  id: string;
  name: string;
  href?: string;
  level?: string;
  time: string;
  space: string;
  keyIdea?: string;
  whenToUse?: string;
};

/**
 * A category on one page: every entry's cost and key idea, in a real table
 * with a caption and column headers, styled to print on paper (the site's
 * print stylesheet drops everything around it).
 *
 *   <CheatSheetTable caption="Sorting" rows={rows} />
 */
export function CheatSheetTable({
  caption,
  rows,
  className
}: {
  caption: string;
  rows: CheatSheetRow[];
  className?: string;
}) {
  return (
    <Card className={cn('overflow-x-auto py-0 print:overflow-visible print:shadow-none print:ring-0', className)}>
      <table className="w-full border-collapse text-sm">
        <caption className="px-4 pt-4 pb-2 text-left text-base font-semibold text-fg print:px-0">{caption}</caption>
        <thead>
          <tr className="border-b border-border text-left text-xs text-muted uppercase">
            <th scope="col" className="px-4 py-2 font-medium">Entry</th>
            <th scope="col" className="px-4 py-2 font-medium">Time</th>
            <th scope="col" className="px-4 py-2 font-medium">Space</th>
            <th scope="col" className="px-4 py-2 font-medium">Key idea</th>
            <th scope="col" className="px-4 py-2 font-medium print:hidden">Level</th>
          </tr>
        </thead>
        <tbody>
          {rows.map((r) => (
            <tr
              key={r.id}
              className="border-b border-border align-top last:border-0 print:break-inside-avoid"
            >
              <th scope="row" className="px-4 py-2 text-left font-medium text-fg">
                {r.href ? (
                  <a href={r.href} className="underline-offset-4 hover:underline">{r.name}</a>
                ) : (
                  r.name
                )}
              </th>
              <td className="px-4 py-2 font-mono text-xs whitespace-nowrap text-fg">{r.time}</td>
              <td className="px-4 py-2 font-mono text-xs whitespace-nowrap text-fg">{r.space}</td>
              <td className="px-4 py-2 text-pretty text-muted">{r.keyIdea ?? ''}</td>
              <td className="px-4 py-2 print:hidden">{r.level ? <LevelBadge level={r.level} /> : null}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </Card>
  );
}
