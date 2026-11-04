import type { Metadata } from 'next';
import { PageHeading } from '@/components/shared/page-header';
import { ColorsExplorer } from '@/components/colors/colors-explorer';
import { pageMetadata } from '@/lib/metadata';

export const metadata: Metadata = pageMetadata({
  title: 'Colors',
  description:
    'The alGOrithm color system: OKLCH neutrals and accents, a colorblind-safe visualization palette, and every pair checked against WCAG AA.',
  path: '/colors/'
});

export default function ColorsPage() {
  return (
    <>
      <PageHeading title="Colors">
        Every color on the site is a token in OKLCH, in light and dark, with three neutrals, six
        accents and three palettes for the visualizations. Every pair you read is checked to meet
        WCAG AA, and no role is ever shown by color alone.
      </PageHeading>
      <ColorsExplorer />
    </>
  );
}
