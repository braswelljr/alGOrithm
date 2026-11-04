import type { Metadata } from 'next';
import Image from 'next/image';
import { Badge } from '@braswelljr/algorithm-ui/components/badge';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader
} from '@braswelljr/algorithm-ui/components/card';
import { Frame, FramePanel } from '@braswelljr/algorithm-ui/components/frame';
import { Heading } from '@braswelljr/algorithm-ui/components/heading';
import {
  ACTIVE_CONCEPT,
  COLORS,
  CONCEPTS,
  markSvg,
  mono,
  PALETTE,
  type ConceptId,
  type Palette
} from '@/components/brand/marks';
import { PageHeading } from '@/components/shared/page-header';
import { brandPreviews } from '@/lib/brand';
import { pageMetadata } from '@/lib/metadata';

export const metadata: Metadata = {
  ...pageMetadata({
    title: 'Brand',
    description:
      'The alGOrithm logo: three concepts at every size, in color and one color, light and dark.',
    path: '/brand/'
  }),
  // A page for choosing the logo, not for search results.
  robots: { index: false, follow: true }
};

const sizes = [16, 32, 48, 180, 512];

/** A concept as an image: a data URL from the one source in marks.ts. */
function Mark({
  id,
  palette,
  size,
  label
}: {
  id: ConceptId;
  palette: Palette;
  size: number;
  label: string;
}) {
  const src = `data:image/svg+xml,${encodeURIComponent(markSvg(id, palette))}`;
  return (
    <Image
      src={src}
      alt={label}
      width={size}
      height={size}
      unoptimized
    />
  );
}

/**
 * The real 16 px PNG (from scripts/icons.mjs, hinted variant included), shown
 * 4x with nearest-neighbor scaling, so each pixel's edge can be judged.
 */
function Zoom({ id }: { id: ConceptId }) {
  return (
    <Frame className="flex-row gap-1">
      {brandPreviews
        .filter((p) => p.concept === id)
        .map((p) => (
          <FramePanel
            key={p.href}
            className="flex flex-col items-center gap-2 p-4"
            style={{ background: p.scheme === 'dark' ? COLORS.darkBg : COLORS.lightBg }}
          >
            <Image
              src={p.href}
              alt={`${CONCEPTS[id].name} rendered at 16 px on ${p.scheme}, shown at 4x`}
              width={64}
              height={64}
              unoptimized
              className="[image-rendering:pixelated]"
            />
            <figcaption
              className="font-mono text-xs"
              style={{ color: p.scheme === 'dark' ? COLORS.paper : COLORS.ink }}
            >
              16 px, 4x
            </figcaption>
          </FramePanel>
        ))}
    </Frame>
  );
}

/** One row of sizes on one ground. */
function Sizes({
  id,
  palette,
  ground,
  scheme
}: {
  id: ConceptId;
  palette: Palette;
  ground: string;
  scheme: string;
}) {
  return (
    <FramePanel
      className="flex items-end gap-6 overflow-x-auto p-4"
      style={{ background: ground }}
    >
      {sizes.map((s) => (
        <figure
          key={s}
          className="flex flex-col items-center gap-2"
        >
          <Mark
            id={id}
            palette={palette}
            size={s}
            label={`${CONCEPTS[id].name}, ${s}px, ${scheme}`}
          />
          <figcaption
            className="font-mono text-xs"
            style={{ color: ground === COLORS.darkBg ? COLORS.paper : COLORS.ink }}
          >
            {s}px
          </figcaption>
        </figure>
      ))}
    </FramePanel>
  );
}

export default function BrandPage() {
  return (
    <>
      <PageHeading title="The logo">
        Three original concepts for the mark, icon only, at 16, 32, 48, 180 and 512 pixels, with the
        real 16 px rendering at 4x, in full color and one color, on light and dark. The site uses
        the one marked in use; switching is one line, ACTIVE_CONCEPT in components/brand/marks.ts,
        then <code className="font-mono">pnpm --dir web/apps/site icons</code>.
      </PageHeading>

      <div className="flex flex-col gap-8">
        {(Object.keys(CONCEPTS) as ConceptId[]).map((id, i) => {
          const c = CONCEPTS[id];
          const active = id === ACTIVE_CONCEPT;
          return (
            <Card
              key={id}
              variant={active ? 'selected' : 'default'}
              render={<section aria-labelledby={`concept-${id}`} />}
            >
              <CardHeader>
                <div className="flex flex-wrap items-baseline gap-3">
                  <Heading
                    level={2}
                    id={`concept-${id}`}
                  >
                    ({String.fromCharCode(97 + i)}) {c.name}
                  </Heading>
                  <code className="font-mono text-sm text-muted">{id}</code>
                  {active ? <Badge variant="info">In use</Badge> : null}
                </div>
                <CardDescription className="max-w-3xl">{c.idea}</CardDescription>
              </CardHeader>
              <CardContent className="grid min-w-0 gap-4 xl:grid-cols-[auto_minmax(0,1fr)]">
                <Frame className="flex-row gap-1 self-start">
                  <FramePanel
                    className="p-6"
                    style={{ background: COLORS.lightBg }}
                  >
                    <Mark
                      id={id}
                      palette={PALETTE.light}
                      size={180}
                      label={`${c.name}, full color on light`}
                    />
                  </FramePanel>
                  <FramePanel
                    className="p-6"
                    style={{ background: COLORS.darkBg }}
                  >
                    <Mark
                      id={id}
                      palette={PALETTE.dark}
                      size={180}
                      label={`${c.name}, full color on dark`}
                    />
                  </FramePanel>
                </Frame>
                <div className="flex min-w-0 flex-col gap-3">
                  <Zoom id={id} />
                  <Frame>
                    <Sizes
                      id={id}
                      palette={PALETTE.light}
                      ground={COLORS.lightBg}
                      scheme="color on light"
                    />
                    <Sizes
                      id={id}
                      palette={PALETTE.dark}
                      ground={COLORS.darkBg}
                      scheme="color on dark"
                    />
                    <Sizes
                      id={id}
                      palette={mono(COLORS.ink)}
                      ground="#ffffff"
                      scheme="one color on light"
                    />
                    <Sizes
                      id={id}
                      palette={mono(COLORS.paper)}
                      ground={COLORS.darkBg}
                      scheme="one color on dark"
                    />
                  </Frame>
                </div>
              </CardContent>
            </Card>
          );
        })}
      </div>
    </>
  );
}
