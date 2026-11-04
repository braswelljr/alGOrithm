/**
 * The two store screenshots in the web app manifest, drawn at build time by
 * the social card renderer (src/app/screenshots/<name>.png/route.tsx).
 */
export const screenshots = [
  {
    name: 'wide',
    href: '/screenshots/wide.png',
    width: 1280,
    height: 720,
    formFactor: 'wide' as const,
    label: 'alGOrithm: run algorithms on your data and watch every step'
  },
  {
    name: 'narrow',
    href: '/screenshots/narrow.png',
    width: 720,
    height: 1280,
    formFactor: 'narrow' as const,
    label: 'alGOrithm on a phone: learning paths, quizzes and the visualizer'
  }
];
