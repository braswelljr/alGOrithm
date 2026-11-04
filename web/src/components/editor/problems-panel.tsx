'use client';

/**
 * The problems panel, as a code editor's: every open file's problems from the
 * terminal's checks (web/runtime/tooling), errors first, grouped by file. A
 * problem opens its file at its line. Severity is said in words and by icon,
 * never by color alone.
 */
import { LuCircleCheck, LuCircleX, LuInfo, LuTriangleAlert } from 'react-icons/lu';
import { Heading } from '@/components/ui/heading';
import type { Diagnostic } from '@/editor/terminal-runtime';
import { cn } from '@/lib/utils';
import { FileIcon } from './file-icon';

const RANK = { error: 0, warning: 1, info: 2 } as const;

export function countProblems(problems: Record<string, Diagnostic[]>) {
  let errors = 0;
  let warnings = 0;
  let infos = 0;
  for (const list of Object.values(problems))
    for (const d of list) {
      if (d.severity === 'error') errors++;
      else if (d.severity === 'warning') warnings++;
      else infos++;
    }
  return { errors, warnings, infos };
}

export function SeverityIcon({
  severity,
  className
}: {
  severity: Diagnostic['severity'];
  className?: string;
}) {
  const Icon = severity === 'error' ? LuCircleX : severity === 'warning' ? LuTriangleAlert : LuInfo;
  return (
    <Icon
      aria-hidden
      className={cn(
        'size-3.5 shrink-0',
        severity === 'error'
          ? 'text-danger'
          : severity === 'warning'
            ? 'text-warning'
            : 'text-sky-500',
        className
      )}
    />
  );
}

export function ProblemsPanel({
  problems,
  home,
  onGo
}: {
  problems: Record<string, Diagnostic[]>;
  home: string;
  onGo: (path: string, line: number, col: number) => void;
}) {
  const files = Object.entries(problems)
    .filter(([, list]) => list.length > 0)
    .map(
      ([path, list]) =>
        [
          path,
          [...list].sort((a, b) => RANK[a.severity] - RANK[b.severity] || a.line - b.line)
        ] as const
    )
    .sort(([a], [b]) => a.localeCompare(b));

  if (files.length === 0) {
    return (
      <p className="text-muted flex h-full items-center justify-center gap-2 text-sm">
        <LuCircleCheck
          aria-hidden
          className="text-success size-4"
        />
        No problems in the open files.
      </p>
    );
  }
  return (
    <div className="h-full min-h-0 overflow-auto overscroll-contain py-1 font-mono text-[12.5px]">
      {files.map(([path, list]) => (
        <section
          key={path}
          aria-label={path}
        >
          <Heading
            level={3}
            className="text-fg flex items-center gap-1.5 px-3 py-1 font-sans text-xs font-semibold"
          >
            <FileIcon path={path} />
            {path.slice(path.lastIndexOf('/') + 1)}
            <span className="text-muted truncate font-normal">
              {path.slice(0, path.lastIndexOf('/')).replace(home, '~')}
            </span>
            <span className="bg-subtle text-fg ms-1 rounded-full px-1.5 text-[11px]">
              {list.length}
            </span>
          </Heading>
          <ul>
            {list.map((d, i) => (
              <li key={`${d.line}:${d.col}:${i}`}>
                {/* ui-ok: a problem row is a list item to open, as in a code editor */}
                <button
                  type="button"
                  onClick={() => onGo(path, d.line, d.col)}
                  className="hover:bg-fg/5 focus-visible:bg-fg/8 flex w-full items-start gap-2 py-0.5 ps-7 pe-3 text-start outline-none"
                >
                  <SeverityIcon
                    severity={d.severity}
                    className="mt-0.5"
                  />
                  <span className="sr-only">{d.severity}:</span>
                  <span className="min-w-0 flex-1">{d.message}</span>
                  <span className="text-muted shrink-0">
                    {d.source} [{d.line}:{d.col}]
                  </span>
                </button>
              </li>
            ))}
          </ul>
        </section>
      ))}
    </div>
  );
}
