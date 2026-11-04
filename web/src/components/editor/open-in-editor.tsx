'use client';

/**
 * Opens the editor with an entry to try: its Go snippet (on the data shown, or
 * its example) as ~/<entry>/main.go, and the entry's own file in the library
 * beside it, to read, change and run.
 */
import { useRouter } from 'next/navigation';
import { LuSquareTerminal } from 'react-icons/lu';
import { Button } from '@/components/ui/button';
import { handOff } from '@/editor/handoff';

export function OpenInEditor({
  id,
  program,
  file,
  size = 'sm',
  variant = 'ghost'
}: {
  /** The entry's id: the program goes in a folder of that name. */
  id: string;
  /** The Go program to start from. */
  program: string;
  /** The entry's source file, relative to pkg/. */
  file?: string;
  size?: 'sm' | 'default' | 'md';
  variant?: 'ghost' | 'secondary';
}) {
  const router = useRouter();
  return (
    <Button
      variant={variant}
      size={size}
      onClick={() => {
        handOff({ name: id, main: program, file });
        router.push('/editor/');
      }}
    >
      <LuSquareTerminal aria-hidden />
      Open in the editor
    </Button>
  );
}
