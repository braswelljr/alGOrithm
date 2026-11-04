import type { Metadata } from 'next';
import { PageHeading } from '@/components/shared/page-header';
import { Gallery } from '@/components/ui-gallery/gallery';
import { pageMetadata } from '@/lib/metadata';

export const metadata: Metadata = pageMetadata({
  title: 'Components',
  description:
    'The alGOrithm component library: every component in light and dark, and the radius, elevation, motion and type tokens behind them.',
  path: '/ui/'
});

export default function UiPage() {
  return (
    <>
      <PageHeading title="Components">
        Every component the site is built from (@braswelljr/algorithm-ui), in light and dark side
        by side, and the tokens behind them. Build with these; if something is missing, add it to
        the library.
      </PageHeading>
      <Gallery />
    </>
  );
}
