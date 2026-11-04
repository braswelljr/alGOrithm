'use client';

/**
 * The editor's explorer: the visitor's home folder and the repository's
 * packages, as a tree of folders that open and close, each file with its
 * language's icon. A right click (or the menu key) offers what a code editor's
 * does: new file and folder, rename (F2), duplicate, delete (Delete, after
 * asking), copy the path, download, and open the folder in the terminal.
 *
 * It reads the terminal's own filesystem, so a file made in the terminal
 * appears here (the tree is read again after every command), and one made
 * here is there for `catx`, `go run` and `make`.
 */
import { useMemo, useState } from 'react';
import {
  LuChevronRight,
  LuChevronsDownUp,
  LuCopy,
  LuDownload,
  LuFilePlus,
  LuFiles,
  LuFolderPlus,
  LuPencil,
  LuRefreshCw,
  LuSquareTerminal,
  LuTrash2
} from 'react-icons/lu';
import { toast } from 'sonner';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle
} from '@/components/ui/alert-dialog';
import { Button } from '@/components/ui/button';
import {
  ContextMenu,
  ContextMenuContent,
  ContextMenuItem,
  ContextMenuSeparator,
  ContextMenuShortcut,
  ContextMenuTrigger
} from '@/components/ui/context-menu';
import { listDir, type Entry } from '@/editor/files';
import type { MemFS } from '@/editor/memfs';
import { cn } from '@/lib/utils';
import { FileIcon, FolderIcon } from './file-icon';

type Pending = { kind: 'file' | 'folder' | 'rename'; in: string; from?: string };
type Target = { path: string; dir: boolean; root: boolean };

function parentOf(path: string) {
  return path.slice(0, path.lastIndexOf('/')) || '/';
}

function baseName(path: string) {
  return path.slice(path.lastIndexOf('/') + 1);
}

export function Explorer({
  fs,
  roots,
  version,
  active,
  onOpen,
  onCreate,
  onRename,
  onDelete,
  onDuplicate,
  onTerminal,
  onRefresh
}: {
  fs: MemFS;
  /** The trees shown, each with its label: the home folder, the library. */
  roots: { path: string; label: string }[];
  /** Bumped when the files may have changed, to read them again. */
  version: number;
  /** The file open in the editor, marked in the tree. */
  active?: string;
  onOpen: (path: string) => void;
  onCreate: (path: string, folder: boolean) => void;
  onRename: (from: string, to: string) => void;
  onDelete: (path: string) => void;
  onDuplicate: (path: string) => void;
  /** Opens a terminal in a folder. */
  onTerminal: (dir: string) => void;
  onRefresh: () => void;
}) {
  const [open, setOpen] = useState<Set<string>>(() => new Set(roots.map((r) => r.path)));
  const [selected, setSelected] = useState(roots[0]?.path ?? '/');
  const [pending, setPending] = useState<Pending | null>(null);
  const [name, setName] = useState('');
  const [menuFor, setMenuFor] = useState<Target | null>(null);
  const [doomed, setDoomed] = useState<Target | null>(null);

  const isRoot = (path: string) => roots.some((r) => r.path === path);

  const toggle = (path: string) =>
    setOpen((s) => {
      const next = new Set(s);
      if (next.has(path)) next.delete(path);
      else next.add(path);
      return next;
    });

  // The folder a new file goes in: the selected folder, or the selected file's.
  const target = useMemo(() => {
    try {
      return fs.statSync(selected).isDirectory() ? selected : parentOf(selected);
    } catch {
      return roots[0]?.path ?? '/';
    }
    // version: the selection's kind may have changed.
  }, [fs, selected, version, roots]);

  const create = (kind: 'file' | 'folder', inDir: string) => {
    setOpen((s) => new Set(s).add(inDir));
    setPending({ kind, in: inDir });
    setName('');
  };
  const rename = (path: string) => {
    setPending({ kind: 'rename', in: '', from: path });
    setName(baseName(path));
  };

  const finish = () => {
    const n = name.trim();
    if (pending && n && !n.includes('/')) {
      if (pending.kind === 'rename' && pending.from) {
        if (n !== baseName(pending.from)) onRename(pending.from, `${parentOf(pending.from)}/${n}`);
      } else {
        onCreate(`${pending.in}/${n}`, pending.kind === 'folder');
        setOpen((s) => new Set(s).add(pending.in));
      }
    }
    setPending(null);
    setName('');
  };

  /** A path from the root it is under, as Copy relative path gives it. */
  const relative = (path: string) => {
    const root = roots.find((r) => path === r.path || path.startsWith(`${r.path}/`));
    return root ? path.slice(root.path.length + 1) || '.' : path;
  };

  const copy = (text: string) => {
    navigator.clipboard.writeText(text).then(
      () => toast.success('Copied', { description: text }),
      () => toast.error('Could not copy')
    );
  };

  const download = (path: string) => {
    try {
      const bytes = fs.readFileSync(path);
      const url = URL.createObjectURL(new Blob([new Uint8Array(bytes)]));
      const a = document.createElement('a');
      a.href = url;
      a.download = baseName(path);
      a.click();
      setTimeout(() => URL.revokeObjectURL(url), 1000);
    } catch {
      toast.error(`Could not read ${baseName(path)}`);
    }
  };

  const onRowKey = (e: React.KeyboardEvent, entry: Target) => {
    if (e.key === 'F2' && !entry.root) {
      e.preventDefault();
      rename(entry.path);
    } else if ((e.key === 'Delete' || (e.key === 'Backspace' && e.metaKey)) && !entry.root) {
      e.preventDefault();
      setDoomed(entry);
    }
  };

  const nameField = (indent: number) => (
    // ui-ok: the explorer's inline name field, as a code editor's tree has
    <input
      autoFocus
      value={name}
      onChange={(e) => setName(e.target.value)}
      onKeyDown={(e) => {
        if (e.key === 'Enter') finish();
        if (e.key === 'Escape') {
          setPending(null);
          setName('');
        }
      }}
      onBlur={finish}
      aria-label={pending?.kind === 'rename' ? 'New name' : `Name of the new ${pending?.kind}`}
      // card-ok: the explorer's inline name field, an input not a card
      className="border-primary bg-bg text-fg my-0.5 w-[calc(100%-1rem)] rounded-sm border px-1 font-mono text-[13px] outline-none"
      style={{ marginInlineStart: indent }}
    />
  );

  const row = (e: Entry, depth: number): React.ReactNode => {
    const indent = 8 + depth * 12;
    const isOpen = open.has(e.path);
    if (pending?.kind === 'rename' && pending.from === e.path) {
      return <li key={e.path}>{nameField(indent)}</li>;
    }
    const entry = { path: e.path, dir: e.dir, root: false };
    return (
      <li key={e.path}>
        <div
          data-path={e.path}
          data-dir={e.dir || undefined}
          className={cn(
            'group/row flex items-center gap-1 rounded-sm pe-1 text-[13px]',
            e.path === active
              ? 'bg-primary/20 text-fg'
              : e.path === selected
                ? 'bg-fg/8'
                : 'hover:bg-fg/5'
          )}
        >
          {/* ui-ok: a tree row is the explorer's item itself, as in a code editor */}
          <button
            type="button"
            onClick={() => {
              setSelected(e.path);
              if (e.dir) toggle(e.path);
              else onOpen(e.path);
            }}
            onKeyDown={(k) => onRowKey(k, entry)}
            aria-expanded={e.dir ? isOpen : undefined}
            className="flex min-w-0 flex-1 items-center gap-1.5 py-0.5 text-start outline-none focus-visible:underline"
            style={{ paddingInlineStart: indent }}
          >
            {e.dir ? (
              <>
                <LuChevronRight
                  aria-hidden
                  className={cn('size-3.5 shrink-0 transition-transform', isOpen && 'rotate-90')}
                />
                <FolderIcon open={isOpen} />
              </>
            ) : (
              <FileIcon
                path={e.path}
                className="ms-[18px]"
              />
            )}
            <span className="truncate">{e.name}</span>
          </button>
          <span className="hidden shrink-0 gap-0.5 group-focus-within/row:flex group-hover/row:flex">
            <Button
              variant="plain"
              size="icon-xs"
              aria-label={`Rename ${e.name}`}
              onClick={() => rename(e.path)}
            >
              <LuPencil aria-hidden />
            </Button>
            <Button
              variant="plain"
              size="icon-xs"
              aria-label={`Delete ${e.name}`}
              onClick={() => setDoomed(entry)}
            >
              <LuTrash2 aria-hidden />
            </Button>
          </span>
        </div>
        {e.dir && isOpen ? (
          <ul>
            {pending && pending.kind !== 'rename' && pending.in === e.path ? (
              <li>{nameField(indent + 12)}</li>
            ) : null}
            {listDir(fs, e.path).map((c) => row(c, depth + 1))}
          </ul>
        ) : null}
      </li>
    );
  };

  // The right-clicked row, or the home folder for the empty space below the tree.
  const menu = menuFor ?? { path: roots[0]?.path ?? '/', dir: true, root: true };
  const menuDir = menu.dir ? menu.path : parentOf(menu.path);

  return (
    <nav
      aria-label="Files"
      data-tour="editor-explorer"
      className="flex h-full min-h-0 flex-col"
    >
      <div className="flex shrink-0 items-center justify-between gap-1 px-2 py-1.5">
        <span className="text-muted text-[11px] font-semibold tracking-widest uppercase">
          Explorer
        </span>
        <span className="flex gap-0.5">
          <Button
            variant="plain"
            size="icon-xs"
            aria-label="New file"
            title={`New file in ${target}`}
            onClick={() => create('file', target)}
          >
            <LuFilePlus aria-hidden />
          </Button>
          <Button
            variant="plain"
            size="icon-xs"
            aria-label="New folder"
            title={`New folder in ${target}`}
            onClick={() => create('folder', target)}
          >
            <LuFolderPlus aria-hidden />
          </Button>
          <Button
            variant="plain"
            size="icon-xs"
            aria-label="Refresh"
            title="Read the files again"
            onClick={onRefresh}
          >
            <LuRefreshCw aria-hidden />
          </Button>
          <Button
            variant="plain"
            size="icon-xs"
            aria-label="Collapse folders"
            title="Collapse folders"
            onClick={() => setOpen(new Set())}
          >
            <LuChevronsDownUp aria-hidden />
          </Button>
        </span>
      </div>
      <ContextMenu>
        <ContextMenuTrigger
          className="min-h-0 flex-1 overflow-auto pb-4"
          onContextMenu={(ev) => {
            const el = (ev.target as HTMLElement).closest<HTMLElement>('[data-path]');
            const path = el?.dataset.path;
            if (!path) setMenuFor(null);
            else {
              setMenuFor({ path, dir: el.dataset.dir !== undefined, root: isRoot(path) });
              setSelected(path);
            }
          }}
        >
          {roots.map((r) => {
            const isOpen = open.has(r.path);
            return (
              <section
                key={r.path}
                aria-label={r.label}
              >
                {/* ui-ok: a tree's root row, as a code editor's explorer has */}
                <button
                  type="button"
                  data-path={r.path}
                  data-dir
                  onClick={() => {
                    toggle(r.path);
                    setSelected(r.path);
                  }}
                  aria-expanded={isOpen}
                  className="text-fg flex w-full items-center gap-1 px-1 py-1 text-[11px] font-bold tracking-wider uppercase outline-none focus-visible:underline"
                >
                  <LuChevronRight
                    aria-hidden
                    className={cn('size-3.5 transition-transform', isOpen && 'rotate-90')}
                  />
                  {r.label}
                </button>
                {isOpen ? (
                  <ul>
                    {pending && pending.kind !== 'rename' && pending.in === r.path ? (
                      <li>{nameField(20)}</li>
                    ) : null}
                    {listDir(fs, r.path).map((e) => row(e, 1))}
                  </ul>
                ) : null}
              </section>
            );
          })}
        </ContextMenuTrigger>
        <ContextMenuContent className="min-w-56">
          {!menu.dir ? (
            <ContextMenuItem onClick={() => onOpen(menu.path)}>Open</ContextMenuItem>
          ) : null}
          <ContextMenuItem onClick={() => create('file', menuDir)}>
            <LuFilePlus aria-hidden />
            New file...
          </ContextMenuItem>
          <ContextMenuItem onClick={() => create('folder', menuDir)}>
            <LuFolderPlus aria-hidden />
            New folder...
          </ContextMenuItem>
          <ContextMenuItem onClick={() => onTerminal(menuDir)}>
            <LuSquareTerminal aria-hidden />
            Open in a terminal
          </ContextMenuItem>
          <ContextMenuSeparator />
          <ContextMenuItem onClick={() => copy(menu.path)}>
            <LuCopy aria-hidden />
            Copy path
          </ContextMenuItem>
          <ContextMenuItem onClick={() => copy(relative(menu.path))}>
            <LuCopy aria-hidden />
            Copy relative path
          </ContextMenuItem>
          {!menu.dir ? (
            <>
              <ContextMenuItem onClick={() => onDuplicate(menu.path)}>
                <LuFiles aria-hidden />
                Duplicate
              </ContextMenuItem>
              <ContextMenuItem onClick={() => download(menu.path)}>
                <LuDownload aria-hidden />
                Download
              </ContextMenuItem>
            </>
          ) : null}
          {!menu.root ? (
            <>
              <ContextMenuSeparator />
              <ContextMenuItem onClick={() => rename(menu.path)}>
                <LuPencil aria-hidden />
                Rename...
                <ContextMenuShortcut>F2</ContextMenuShortcut>
              </ContextMenuItem>
              <ContextMenuItem
                variant="destructive-secondary"
                onClick={() => setDoomed(menu)}
              >
                <LuTrash2 aria-hidden />
                Delete
                <ContextMenuShortcut>Del</ContextMenuShortcut>
              </ContextMenuItem>
            </>
          ) : null}
        </ContextMenuContent>
      </ContextMenu>

      <AlertDialog
        open={doomed !== null}
        onOpenChange={(o) => {
          if (!o) setDoomed(null);
        }}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>
              Delete {doomed ? baseName(doomed.path) : ''}
              {doomed?.dir ? ' and everything in it' : ''}?
            </AlertDialogTitle>
            <AlertDialogDescription>
              It is removed from this browser&apos;s copy of the files. The starting files come back
              with Forget my files.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              variant="destructive"
              onClick={() => {
                if (doomed) onDelete(doomed.path);
                setDoomed(null);
              }}
            >
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </nav>
  );
}
