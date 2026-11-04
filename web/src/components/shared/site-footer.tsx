import Link from 'next/link';
import { LogoColor } from '@/components/brand/logo';
import { Button } from '@/components/ui/button';
import { author, licenseUrl, npmUrl, repoUrl } from '@/config/site';
import { cn } from '@/lib/utils';

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
    <footer className="border-border text-muted border-t px-5 py-8 text-sm in-data-[embed=1]:hidden lg:px-6">
      <div className="flex flex-col gap-6">
        <div className="flex flex-col gap-2">
          <LogoColor
            title="alGOrithm"
            className="size-8"
          />
          <p className="text-sm text-pretty">
            Every algorithm here is the Go package from the repository, compiled to WebAssembly.
            Your data stays in this browser.
          </p>
          <p className="text-sm">
            Built by{' '}
            <a
              href={author.url}
              rel="author"
              className={cn(linkClass, 'text-fg decoration-accent font-medium underline')}
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
