'use client';

/**
 * Choosing one file, the way gsl-special-courses' upload dialog does: an empty
 * drop area to drop a file on or click, and once a file is chosen, the file as
 * an attachment row with Replace and Remove. The browser's own drag and drop
 * and file picker, so no library; the file never leaves the page.
 */
import { useRef, useState, type ReactNode } from 'react';
import { LuFileText, LuRefreshCw, LuTrash2, LuUpload } from 'react-icons/lu';
import {
  Attachment,
  AttachmentAction,
  AttachmentActions,
  AttachmentContent,
  AttachmentDescription,
  AttachmentMedia,
  AttachmentTitle
} from '@/components/ui/attachment';
import { cn } from '@/lib/utils';

/** A size for a person: 2.1 KB, 3.4 MB. */
export function formatBytes(n: number) {
  if (n < 1024) return `${n} B`;
  if (n < 1048576) return `${(n / 1024).toFixed(1)} KB`;
  return `${(n / 1048576).toFixed(1)} MB`;
}

export function FileDropzone({
  id,
  accept,
  file,
  error,
  hint,
  onChoose,
  onClear
}: {
  id: string;
  /** The input's accept list, e.g. ".csv,.json". */
  accept: string;
  /** The file in hand, if any: shown as the attachment row. */
  file: { name: string; size: number; kind?: string } | null;
  /** Why the last file was refused, shown on the row in place of its size. */
  error?: string | null;
  /** The line under the drop area's call to action. */
  hint?: ReactNode;
  onChoose: (file: File) => void;
  onClear: () => void;
}) {
  const input = useRef<HTMLInputElement>(null);
  const [over, setOver] = useState(false);
  const pick = () => input.current?.click();

  return (
    <div className="flex w-full flex-col gap-2">
      {/* ui-ok: the browser's file picker, hidden; the drop area and Replace open it */}
      <input
        ref={input}
        id={id}
        type="file"
        accept={accept}
        className="sr-only"
        onChange={(e) => {
          const f = e.target.files?.[0];
          // Cleared, so choosing the same file again still reads it.
          e.target.value = '';
          if (f) onChoose(f);
        }}
      />
      {file ? (
        <Attachment state={error ? 'error' : 'done'}>
          <AttachmentMedia>
            <LuFileText aria-hidden />
          </AttachmentMedia>
          <AttachmentContent>
            <AttachmentTitle title={file.name}>{file.name}</AttachmentTitle>
            <AttachmentDescription>
              {error ?? [file.kind, formatBytes(file.size)].filter(Boolean).join(' · ')}
            </AttachmentDescription>
          </AttachmentContent>
          <AttachmentActions>
            <AttachmentAction onClick={pick}>
              <LuRefreshCw aria-hidden />
              Replace
            </AttachmentAction>
            <AttachmentAction
              size="icon-sm"
              onClick={onClear}
              aria-label={`Remove ${file.name}`}
            >
              <LuTrash2 aria-hidden />
            </AttachmentAction>
          </AttachmentActions>
        </Attachment>
      ) : (
        // ui-ok: the drop area is the picker's own button, drawn as a surface
        <button
          type="button"
          onClick={pick}
          onDragOver={(e) => {
            if (!e.dataTransfer.types.includes('Files')) return;
            e.preventDefault();
            e.dataTransfer.dropEffect = 'copy';
            setOver(true);
          }}
          onDragLeave={() => setOver(false)}
          onDrop={(e) => {
            e.preventDefault();
            setOver(false);
            const f = e.dataTransfer.files[0];
            if (f) onChoose(f);
          }}
          aria-describedby={hint ? `${id}-hint` : undefined}
          className={cn(
            'bg-subtle hover:bg-subtle/70 focus-visible:ring-accent flex h-32 w-full cursor-pointer flex-col items-center justify-center gap-1.5 rounded-xl px-4 text-center transition-colors outline-none focus-visible:ring-2',
            over && 'bg-primary/15 ring-primary ring-2'
          )}
        >
          <span className="flex items-center gap-2 text-sm font-medium">
            <LuUpload
              aria-hidden
              className="size-4"
            />
            {over ? 'Drop the file here' : 'Choose a file, or drop one here'}
          </span>
          {hint ? (
            <span
              id={`${id}-hint`}
              className="text-muted text-sm"
            >
              {hint}
            </span>
          ) : null}
        </button>
      )}
    </div>
  );
}
