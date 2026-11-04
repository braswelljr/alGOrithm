import { ogImage } from '@/lib/og';
import { screenshots } from '@/lib/screenshots';

// A static PNG at build time (out/screenshots/wide.png), drawn like the social card.
export const dynamic = 'force-static';

export function GET() {
  const s = screenshots.find((x) => x.name === 'wide')!;
  return ogImage({
    eyebrow: 'Learn · Quiz · Visualize · Big O',
    title: 'Algorithms and data structures, step by step',
    subtitle: 'The real Go code, compiled to WebAssembly, running on your data in the browser.',
    size: { width: s.width, height: s.height }
  });
}
