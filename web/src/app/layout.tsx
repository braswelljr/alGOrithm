import type { ReactNode } from 'react';
import type { Metadata, Viewport } from 'next';
import { NuqsAdapter } from 'nuqs/adapters/next/app';
import { COLORS } from '@/components/brand/marks';
import { JsonLd } from '@/components/shared/json-ld';
import { author, basePath, siteDescription, siteName, siteTagline, siteUrl } from '@/config/site';
import {
  appleIcons,
  browserconfig,
  darkTabIcons,
  favicon,
  faviconSvg,
  safariPinnedTab,
  startupImages,
  tabIcons,
  tileColor,
  type RasterIcon
} from '@/lib/brand';
import { fontVariables } from '@/lib/fonts';
import { siteLd } from '@/lib/structured-data';
import { Provider } from '@/providers/provider';
import '@/styles/main.css';

/** PNG icon entries, with their sizes, and a media query when they are for dark chrome. */
function pngs(list: RasterIcon[], media?: string) {
  return list.map((i) => ({
    url: i.href,
    sizes: `${i.width}x${i.height}`,
    type: 'image/png',
    ...(media ? { media } : {})
  }));
}

export const metadata: Metadata = {
  metadataBase: new URL(`${siteUrl}${basePath}/`),
  title: { default: `${siteName}: ${siteTagline}`, template: `%s · ${siteName}` },
  description: siteDescription,
  applicationName: siteName,
  keywords: [
    'algorithms',
    'data structures',
    'algorithm visualizer',
    'learn algorithms',
    'sorting algorithms',
    'graph algorithms',
    'pathfinding',
    'dynamic programming',
    'big o notation',
    'operating systems',
    'cpu scheduling',
    'cryptography',
    'machine learning',
    'go',
    'golang',
    'webassembly'
  ],
  authors: [{ name: author.name, url: author.url }],
  creator: author.name,
  publisher: author.name,
  category: 'education',
  openGraph: {
    type: 'website',
    siteName,
    title: `${siteName}: ${siteTagline}`,
    description: siteDescription,
    locale: 'en_US',
    images: [{ url: '/og.png', width: 1200, height: 630, alt: `${siteName}: ${siteTagline}` }]
  },
  twitter: {
    card: 'summary_large_image',
    title: `${siteName}: ${siteTagline}`,
    description: siteDescription,
    images: ['/og.png']
  },
  // Every brand file is in public/ (logos/, icons/, favicon.ico), written by
  // `pnpm --dir web icons`, and named once in src/lib/brand.ts.
  icons: {
    icon: [
      { url: favicon.href, sizes: favicon.sizes.map((n) => `${n}x${n}`).join(' ') },
      { url: faviconSvg.href, type: 'image/svg+xml' },
      ...pngs(tabIcons),
      ...pngs(darkTabIcons, '(prefers-color-scheme: dark)')
    ],
    apple: pngs(appleIcons),
    other: [{ rel: 'mask-icon', url: safariPinnedTab.href, color: tileColor }]
  },
  appleWebApp: {
    capable: true,
    title: siteName,
    statusBarStyle: 'default',
    startupImage: startupImages.map((i) => ({ url: i.href, media: i.media }))
  },
  formatDetection: { telephone: false },
  other: {
    'msapplication-TileColor': tileColor,
    'msapplication-config': browserconfig.href
  },
  robots: { index: true, follow: true }
};

export const viewport: Viewport = {
  themeColor: [
    { media: '(prefers-color-scheme: light)', color: COLORS.lightBg },
    { media: '(prefers-color-scheme: dark)', color: COLORS.darkBg }
  ],
  colorScheme: 'light dark'
};

// Marks the document as embedded (?embed=1) before it paints, so the header
// and footer are hidden by CSS without a flash and without the page reading
// its query (see SiteLayout).
const embedScript =
  "try{if(new URLSearchParams(location.search).get('embed')==='1')document.documentElement.dataset.embed='1'}catch(e){}";

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html
      lang="en"
      dir="ltr"
      className={fontVariables}
      suppressHydrationWarning
    >
      <head>
        <script dangerouslySetInnerHTML={{ __html: embedScript }} />
      </head>
      <body>
        <JsonLd data={siteLd()} />
        <a
          href="#main"
          className="focus:bg-card sr-only focus:not-sr-only focus:absolute focus:top-2 focus:left-2 focus:rounded focus:p-2"
        >
          Skip to content
        </a>
        <NuqsAdapter>
          <Provider>{children}</Provider>
        </NuqsAdapter>
      </body>
    </html>
  );
}
