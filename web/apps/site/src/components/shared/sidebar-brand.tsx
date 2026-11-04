import Link from 'next/link';
import { LogoColor } from '@/components/brand/logo';
import { siteName, siteTagline } from '@/config/site';

/**
 * The identity at the top of the sidebar, and the way home: the logo with the
 * name and tagline beside it while the sidebar is open, the logo alone when it
 * collapses to icons. Named for what it is and where it goes.
 */
export function SidebarBrand() {
  return (
    <Link
      href="/"
      aria-label={`${siteName}: go to the catalog`}
      className="mt-2 flex items-center gap-2.5 overflow-hidden rounded-md px-1 outline-none focus-visible:ring-2 focus-visible:ring-sidebar-ring"
    >
      <LogoColor className="size-9 shrink-0 group-data-[collapsible=icon]:size-8" />
      <span className="flex min-w-0 flex-col leading-tight group-data-[collapsible=icon]:hidden">
        <span className="truncate text-sm font-bold text-white">{siteName}</span>
        <span className="truncate text-xs text-white/70">{siteTagline}</span>
      </span>
    </Link>
  );
}
