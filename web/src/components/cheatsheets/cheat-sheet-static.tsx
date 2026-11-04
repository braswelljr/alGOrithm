import Link from 'next/link';
import type { CheatRow } from '@/components/cheatsheets/cheat-sheet-table';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow
} from '@/components/ui/table';
import { docsHref } from '@/lib/entries';

/**
 * The cheat sheet as a plain table, rendered on the server: what the exported
 * page carries for a reader without JavaScript and for search engines, until
 * the interactive table (which reads its state from the address) takes over.
 */
export function CheatSheetStatic({ rows }: { rows: CheatRow[] }) {
  return (
    <Table className="min-w-[52rem]">
      <TableHeader>
        <TableRow>
          <TableHead>Entry</TableHead>
          <TableHead>Domain</TableHead>
          <TableHead>Category</TableHead>
          <TableHead>Level</TableHead>
          <TableHead>Time</TableHead>
          <TableHead>Space</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {rows.map((r) => (
          <TableRow key={r.id}>
            <TableCell>
              <Link
                href={docsHref(r.id)}
                className="font-medium"
              >
                {r.name}
              </Link>
              <p className="text-muted text-sm">{r.keyIdea}</p>
            </TableCell>
            <TableCell>{r.topic}</TableCell>
            <TableCell>{r.category}</TableCell>
            <TableCell className="capitalize">{r.level}</TableCell>
            <TableCell className="font-mono">{r.time}</TableCell>
            <TableCell className="font-mono">{r.space}</TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}
