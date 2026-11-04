import type { MetadataRoute } from 'next';
import { COLORS } from '@/components/brand/marks';
import { appIcons, maskableIcons, monochromeIcons, shortcuts, type RasterIcon } from '@/lib/brand';
import { screenshots } from '@/lib/screenshots';
import { basePath, siteDescription } from '@/config/site';

export const dynamic = 'force-static';

const at = (href: string) => `${basePath}${href}`;

function icons(list: RasterIcon[], purpose: 'any' | 'maskable' | 'monochrome') {
  return list.map((i) => ({
    src: at(i.href),
    sizes: `${i.width}x${i.height}`,
    type: 'image/png',
    purpose
  }));
}

/** The installable app, written as /manifest.webmanifest. */
export default function manifest(): MetadataRoute.Manifest {
  return {
    id: `${basePath}/`,
    name: 'alGOrithm: algorithms, data structures and CS, visualized',
    short_name: 'alGOrithm',
    description: siteDescription,
    start_url: `${basePath}/`,
    scope: `${basePath}/`,
    display: 'standalone',
    display_override: ['window-controls-overlay', 'standalone'],
    orientation: 'any',
    theme_color: COLORS.accent,
    background_color: COLORS.lightBg,
    categories: ['education', 'developer', 'productivity'],
    lang: 'en',
    dir: 'ltr',
    icons: [
      ...icons(appIcons, 'any'),
      ...icons(maskableIcons, 'maskable'),
      ...icons(monochromeIcons, 'monochrome')
    ],
    shortcuts: shortcuts.map((s) => ({
      name: s.name,
      short_name: s.name,
      description: s.description,
      url: at(s.url),
      icons: [{ src: at(s.icon.href), sizes: '96x96', type: 'image/png' }]
    })),
    screenshots: screenshots.map((s) => ({
      src: at(s.href),
      sizes: `${s.width}x${s.height}`,
      type: 'image/png',
      form_factor: s.formFactor,
      label: s.label
    }))
  };
}
