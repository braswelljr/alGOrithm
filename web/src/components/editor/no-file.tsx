'use client';

/**
 * What the editor shows with no file open: a word on what to do, the home
 * folder's files and a few of the repository's to open with one click, and a
 * new file. The lists are read from the terminal's filesystem, so they show
 * what is there now.
 */
import { LuFileCode, LuFilePlus } from 'react-icons/lu';
import { Button } from '@/components/ui/button';
import {
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle
} from '@/components/ui/empty';
import { Heading } from '@/components/ui/heading';
import { listDir } from '@/editor/files';
import type { MemFS } from '@/editor/memfs';
import { FileIcon } from './file-icon';

/** A few of the repository's files, for a first look. */
const PICKS = [
  'algorithms/sorting/quick.go',
  'algorithms/searching/binary.go',
  'algorithms/pathfinding/astar.go',
  'structures/heaps/heaps.go',
  'os/scheduling/algorithms.go',
  'security/hashing/consistent.go'
];

export function NoFile({
  fs,
  home,
  library,
  onOpen,
  onNew
}: {
  fs: MemFS;
  home: string;
  library: string;
  onOpen: (path: string) => void;
  onNew: () => void;
}) {
  const mine = listDir(fs, home).filter((e) => !e.dir);
  const picks = PICKS.map((p) => `${library}/${p}`).filter((p) => fs.existsSync(p));

  const list = (title: string, paths: string[], base: string) =>
    paths.length ? (
      <section
        aria-label={title}
        className="w-full"
      >
        <Heading
          level={3}
          className="text-muted mb-1 text-[11px] font-semibold tracking-widest uppercase"
        >
          {title}
        </Heading>
        <ul className="grid gap-0.5">
          {paths.map((p) => (
            <li key={p}>
              <Button
                variant="plain"
                size="sm"
                onClick={() => onOpen(p)}
                className="w-full justify-start font-mono text-[13px]"
              >
                <FileIcon path={p} />
                <span className="truncate">{p.slice(base.length + 1)}</span>
              </Button>
            </li>
          ))}
        </ul>
      </section>
    ) : null;

  return (
    <div className="h-full overflow-auto">
      <Empty className="min-h-full">
        <EmptyHeader>
          <EmptyMedia variant="icon">
            <LuFileCode aria-hidden />
          </EmptyMedia>
          <EmptyTitle>No file is open</EmptyTitle>
          <EmptyDescription>
            Open one below or from the explorer, type code FILE in the terminal, or start a new one.
          </EmptyDescription>
        </EmptyHeader>
        <EmptyContent className="max-w-md">
          <Button
            variant="secondary"
            size="sm"
            onClick={onNew}
          >
            <LuFilePlus aria-hidden />
            New file
          </Button>
          <div className="grid w-full gap-4 text-start sm:grid-cols-2">
            {list(
              'Your files',
              mine.map((e) => e.path),
              home
            )}
            {list('From the library', picks, library)}
          </div>
        </EmptyContent>
      </Empty>
    </div>
  );
}
