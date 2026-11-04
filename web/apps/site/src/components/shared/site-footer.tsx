import Link from 'next/link';
import { Button } from '@braswelljr/algorithm-ui/components/button';
import { LogoColor } from '@/components/brand/logo';
import { author, licenseUrl, npmUrl, repoUrl } from '@/config/site';

const linkClass =
  'rounded-sm underline-offset-4 outline-none hover:text-fg hover:underline focus-visible:ring-2 focus-visible:ring-accent';

const pages = [
  { href: '/learn/', label: 'Learn' },
  { href: '/docs/', label: 'Docs' },
  { href: '/visualize/', label: 'Visualize' },
  { href: '/big-o/', label: 'Big O' },
  { href: '/system-design/', label: 'System design' }
];

const external = [
  { href: repoUrl, label: 'GitHub' },
  { href: npmUrl, label: 'npm' },
  { href: licenseUrl, label: 'License' }
];

/** The foot of every page: where to go next, where the project lives, and whose it is. Hidden in embed mode. */
export function SiteFooter() {
  return (
    <footer className="border-t border-border px-5 py-8 text-sm text-muted in-data-[embed=1]:hidden lg:px-6">
      <div className="flex flex-col gap-6 sm:flex-row sm:items-start sm:justify-between">
        <div className="flex flex-col gap-2">
          <LogoColor
            title="alGOrithm"
            className="size-8"
          />
          <p className="max-w-md text-xs text-pretty">
            Every algorithm here is the Go package from the repository, compiled to WebAssembly.
            Your data stays in this browser.
          </p>
          <p className="text-xs">
            Built by{' '}
            <a
              href={author.url}
              rel="author"
              className={`${linkClass} font-medium text-fg underline decoration-accent`}
            >
              {author.name}
            </a>
          </p>
        </div>
        <nav
          aria-label="Footer"
          className="-mx-2.5 flex flex-wrap gap-1"
        >
          {pages.map((l) => (
            <Button
              key={l.href}
              variant="ghost"
              size="sm"
              render={<Link href={l.href} />}
              className="text-muted"
            >
              {l.label}
            </Button>
          ))}
          {external.map((l) => (
            <Button
              key={l.href}
              variant="ghost"
              size="sm"
              render={
                <a
                  href={l.href}
                  rel="noreferrer"
                />
              }
              className="text-muted"
            >
              {l.label}
            </Button>
          ))}
        </nav>
      </div>
    </footer>
  );
}
