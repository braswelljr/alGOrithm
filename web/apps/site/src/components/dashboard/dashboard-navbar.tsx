'use client';

import { useEffect, useRef, type ComponentProps } from 'react';
import Link from 'next/link';
import { ThemePicker } from '@braswelljr/algorithm-ui/algorithm/theme-picker';
import { Button } from '@braswelljr/algorithm-ui/components/button';
import { SidebarTrigger } from '@braswelljr/algorithm-ui/components/sidebar';
import { GitHubIcon } from '@/components/brand/github-icon';
import { LogoColor } from '@/components/brand/logo';
import { Breadcrumbs } from '@/components/shared/breadcrumbs';
import { CommandSearchButton } from '@/components/shared/command-search-button';
import { repoUrl, siteName } from '@/config/site';
import { useNavigation } from '@/context/use-navigation';
import { cn } from '@/utils/utils';

/**
 * The top bar: the sidebar toggle, the page's name, search, the theme and the
 * repository, with the breadcrumb trail beneath it.
 *
 * The page's name is its menu entry's label. It is a line of the chrome
 * rather than a heading: each page carries its own `h1`, and a second one here
 * would give the document two.
 */
export function DashboardNavbar({ className, ...props }: ComponentProps<'header'>) {
  const { active } = useNavigation();
  const title = active?.label ?? siteName;

  // The bar's height, published as `--header-height` so a rail (the `rail`
  // utility) can stick directly under it. Measured, because the trail below
  // comes and goes.
  const headerRef = useRef<HTMLElement>(null);
  useEffect(() => {
    const header = headerRef.current;
    if (!header || typeof ResizeObserver === 'undefined') return;
    const root = document.documentElement;
    const publish = () => root.style.setProperty('--header-height', `${header.offsetHeight}px`);
    const observer = new ResizeObserver(publish);
    observer.observe(header);
    publish();
    return () => {
      observer.disconnect();
      root.style.removeProperty('--header-height');
    };
  }, []);

  return (
    <header
      ref={headerRef}
      id="site-navbar"
      {...props}
      className={cn(
        'sticky top-0 z-40 divide-y divide-border border-b border-border bg-bg/90 backdrop-blur-sm in-data-[embed=1]:hidden',
        className
      )}
    >
      <div className="flex h-12 shrink-0 items-center justify-between gap-2">
        <div className="ms-5 flex min-w-0 items-center gap-3 lg:ms-6">
          <SidebarTrigger
            aria-controls="site-sidebar"
            className="text-fg"
          />
          {/* On a phone the logo stands in for the hidden sidebar's brand. */}
          <Link
            href="/"
            aria-label={`${siteName}: go to the catalog`}
            className="shrink-0 rounded-sm outline-none focus-visible:ring-2 focus-visible:ring-accent md:hidden"
          >
            <LogoColor className="size-7" />
          </Link>
          <p className="line-clamp-1 min-w-0 font-semibold uppercase">{title}</p>
        </div>

        <div
          data-slot="header-actions"
          className="me-5 flex items-center gap-2 lg:me-6"
        >
          {/* Before the rest, because it is the way to everything else. */}
          <CommandSearchButton />
          <ThemePicker />
          <Button
            variant="ghost"
            size="icon-sm"
            render={<a href={repoUrl} />}
            className="text-muted hover:text-fg"
            aria-label="Source code on GitHub"
          >
            <GitHubIcon className="size-5" />
          </Button>
        </div>
      </div>

      {/* The trail shows only below the top level, where there is somewhere
          to go back to. */}
      <Breadcrumbs className="hidden px-6 py-2 lg:flex" />
    </header>
  );
}
