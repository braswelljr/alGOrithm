import { readFileSync } from 'node:fs';
import path from 'node:path';
import { PALETTE, toneCss } from '@/components/brand/marks';
import { LOGO } from '@/lib/brand';

/**
 * The mark from its file, public/logos/mark.svg, read at build time, with its
 * colors set to the dark palette for the card's dark ground, as a data URL
 * Satori can draw.
 */
const markFile = readFileSync(path.join(process.cwd(), LOGO.mark.file), 'utf8').replace(
  /<style>.*?<\/style>/s,
  `<style>${toneCss(PALETTE.dark)}</style>`
);
const LOGO_SRC = `data:image/svg+xml;base64,${Buffer.from(markFile).toString('base64')}`;

/** The site's colors, written out: Satori cannot read the stylesheet's variables. */
const COLORS = {
  ground: '#12141c',
  accent: '#8f93ff',
  violet: '#4b4fd8',
  text: '#f4f5f8',
  muted: '#a4a8b8'
};

type OgCardProps = {
  title: string;
  description?: string;
  /** A small line above the title: the section, the category and cost. */
  eyebrow: string;
  /** Shown bottom right: the host. */
  site: string;
};

/**
 * The Open Graph card, for next/og. It is rendered by Satori, not a browser:
 * inline styles only, flex layout, no CSS variables, so the colors are
 * written out.
 */
export function OgCard({ title, description, eyebrow, site }: OgCardProps) {
  return (
    <div
      style={{
        width: '100%',
        height: '100%',
        display: 'flex',
        flexDirection: 'column',
        justifyContent: 'space-between',
        padding: 72,
        background: COLORS.ground,
        backgroundImage: `radial-gradient(circle at 85% 0%, ${COLORS.violet}66, transparent 55%), radial-gradient(circle at 10% 110%, ${COLORS.accent}33, transparent 50%)`,
        color: COLORS.text,
        fontFamily: 'Source Sans 3'
      }}
    >
      <div style={{ display: 'flex', alignItems: 'center', gap: 20 }}>
        <img
          alt=""
          width={72}
          height={72}
          src={LOGO_SRC}
        />
        <div style={{ display: 'flex', fontSize: 30, color: COLORS.muted }}>{eyebrow}</div>
      </div>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
        <div
          style={{
            display: 'flex',
            fontSize: 76,
            fontWeight: 700,
            fontFamily: 'Roboto Serif',
            letterSpacing: -2,
            lineHeight: 1.05,
            color: '#ffffff'
          }}
        >
          {title}
        </div>
        {description ? (
          <div style={{ display: 'flex', fontSize: 30, color: COLORS.muted, lineHeight: 1.4 }}>
            {description.length > 150 ? `${description.slice(0, 147)}...` : description}
          </div>
        ) : null}
      </div>
      <div
        style={{ display: 'flex', justifyContent: 'space-between', fontSize: 24, color: '#7c8194' }}
      >
        <div style={{ display: 'flex' }}>$ algorithm run quick-sort --trace</div>
        <div style={{ display: 'flex' }}>{site} · by braswelljr</div>
      </div>
    </div>
  );
}
