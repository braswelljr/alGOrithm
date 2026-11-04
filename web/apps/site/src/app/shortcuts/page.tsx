import type { Metadata } from 'next';
import { PageHeading } from '@/components/shared/page-header';
import { ShortcutTrainer } from '@/components/shortcuts/shortcut-trainer';
import { pageMetadata } from '@/lib/metadata';

export const metadata: Metadata = pageMetadata({
  title: 'Keyboard shortcuts',
  description:
    'Every keyboard shortcut on alGOrithm, and a trainer that asks for each one until your fingers know it.',
  path: '/shortcuts/'
});

export default function ShortcutsPage() {
  return (
    <>
      <PageHeading title="Keyboard shortcuts">
        The whole site is a few keys away. Practice them here: the trainer names an action, you
        press its keys, and the keyboard lights up as you go.
      </PageHeading>
      <ShortcutTrainer />
    </>
  );
}
