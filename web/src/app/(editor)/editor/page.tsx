import type { Metadata } from 'next';
import { EditorApp } from '@/components/editor/editor-app';
import { Heading } from '@/components/ui/heading';
import { pageMetadata } from '@/lib/metadata';

export const metadata: Metadata = pageMetadata({
  title: 'Editor',
  description:
    'A code editor in the browser with every algorithm and data structure of alGOrithm: edit them, write your own Go and shell scripts, and run them in a terminal with the gnux tools.',
  path: '/editor/'
});

export default function EditorPage() {
  return (
    <>
      {/* The page's heading, for screen readers: the editor's title bar shows it. */}
      <Heading
        level={1}
        className="sr-only"
      >
        Editor
      </Heading>
      <EditorApp />
    </>
  );
}
