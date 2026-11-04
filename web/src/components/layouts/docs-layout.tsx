'use client';

/**
 * The docs' own layout, a main layout of its own beside the site layout. It has
 * the same sidebar and top bar as the site, so moving between them feels like one
 * site, but its sidebar lists the docs: Overview, Tags, then every domain's
 * categories and their entries. On a phone the sidebar is the same drawer.
 */
import { useEffect, useState, type ReactNode } from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { BookOpen, Hash } from 'lucide-react';
import { LuArrowLeft, LuChevronDown } from 'react-icons/lu';
import { DashboardNavbar } from '@/components/dashboard/dashboard-navbar';
import { PreferencesScope } from '@/components/settings/preferences-scope';
import { DevReload } from '@/components/shared/dev-reload';
import { SidebarBrand } from '@/components/shared/sidebar-brand';
import { SiteFooter } from '@/components/shared/site-footer';
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible';
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarInset,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarMenuSub,
  SidebarMenuSubButton,
  SidebarMenuSubItem,
  SidebarProvider,
  SidebarRail,
  SidebarSeparator,
  useSidebar
} from '@/components/ui/sidebar';
import { NavigationProvider } from '@/context/use-navigation';
import { PageTitleProvider } from '@/context/use-page-title';
import { docsHref } from '@/lib/entries';

/** The list the docs sidebar shows: domains, their categories, their entries. */
export type DocsNav = {
  topic: string;
  categories: { id: string; name: string; entries: { id: string; name: string }[] }[];
}[];

/**
 * One category, folded. Controlled, so that going to an entry in a folded
 * category opens it, while the reader can still fold any category by hand.
 */
function DocsCategory({
  category,
  current
}: {
  category: DocsNav[number]['categories'][number];
  current?: string;
}) {
  const holdsCurrent = category.entries.some((e) => e.id === current);
  const [open, setOpen] = useState(holdsCurrent);
  // Arriving at an entry inside opens its category (set while rendering, as
  // React advises for state that follows a prop, rather than in an effect).
  const [held, setHeld] = useState(holdsCurrent);
  if (held !== holdsCurrent) {
    setHeld(holdsCurrent);
    if (holdsCurrent) setOpen(true);
  }

  return (
    <Collapsible
      open={open}
      onOpenChange={setOpen}
      render={<SidebarMenuItem />}
    >
      {/* The category holding the open entry is a keycap too, so the way to
          the page shows when the category is folded. */}
      <CollapsibleTrigger
        render={
          <SidebarMenuButton
            isActive={holdsCurrent}
            className="group/category"
          />
        }
      >
        <span className="truncate">{category.name}</span>
        <LuChevronDown
          aria-hidden
          className="ms-auto size-3.5 transition-transform group-data-panel-open/category:rotate-180 motion-reduce:transition-none"
        />
      </CollapsibleTrigger>
      <CollapsibleContent>
        <SidebarMenuSub className="border-white/15">
          {category.entries.map((e) => (
            <SidebarMenuSubItem key={e.id}>
              <SidebarMenuSubButton
                isActive={e.id === current}
                render={
                  <Link
                    href={docsHref(e.id)}
                    aria-current={e.id === current ? 'page' : undefined}
                  />
                }
              >
                <span className="truncate">{e.name}</span>
              </SidebarMenuSubButton>
            </SidebarMenuSubItem>
          ))}
        </SidebarMenuSub>
      </CollapsibleContent>
    </Collapsible>
  );
}

function DocsSidebar({ nav }: { nav: DocsNav }) {
  const { isMobile, setOpenMobile } = useSidebar();
  const pathname = usePathname() ?? '';
  const current = /^\/docs\/([^/]+)\/?$/.exec(pathname)?.[1];

  // A phone's drawer closes once the reader has gone somewhere.
  useEffect(() => {
    if (isMobile) setOpenMobile(false);
  }, [pathname, isMobile, setOpenMobile]);

  const top = [
    { href: '/docs/', label: 'Overview', icon: BookOpen, active: /^\/docs\/?$/.test(pathname) },
    { href: '/docs/tags/', label: 'Tags', icon: Hash, active: pathname.startsWith('/docs/tags') }
  ];

  return (
    <Sidebar
      id="site-sidebar"
      className="h-dvh border-x-0! in-data-[embed=1]:hidden"
    >
      <SidebarHeader>
        <SidebarBrand />
      </SidebarHeader>

      <SidebarContent>
        <nav
          aria-label="Docs"
          data-tour="sidebar"
          className="flex flex-col gap-2"
        >
          <SidebarGroup>
            <SidebarGroupLabel className="text-white/60">Docs</SidebarGroupLabel>
            <SidebarMenu className="gap-1.5">
              {top.map((item) => (
                <SidebarMenuItem key={item.href}>
                  <SidebarMenuButton
                    isActive={item.active}
                    render={
                      <Link
                        href={item.href}
                        aria-current={item.active ? 'page' : undefined}
                      />
                    }
                  >
                    <item.icon
                      aria-hidden
                      className="size-4 shrink-0"
                    />
                    <span className="truncate">{item.label}</span>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              ))}
            </SidebarMenu>
          </SidebarGroup>
          {nav.map((group) => (
            <SidebarGroup key={group.topic}>
              <SidebarGroupLabel className="text-white/60">{group.topic}</SidebarGroupLabel>
              <SidebarMenu className="gap-0.5">
                {group.categories.map((c) => (
                  <DocsCategory
                    key={c.id}
                    category={c}
                    current={current}
                  />
                ))}
              </SidebarMenu>
            </SidebarGroup>
          ))}
        </nav>
      </SidebarContent>

      <SidebarSeparator className="bg-white/15" />
      <SidebarFooter>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton render={<Link href="/" />}>
              <LuArrowLeft
                aria-hidden
                className="size-4 shrink-0"
              />
              <span className="truncate">Back to the site</span>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarFooter>

      <SidebarRail />
    </Sidebar>
  );
}

/** The docs layout: the same frame as SiteLayout, with the docs list in the sidebar. */
export function DocsLayout({ nav, children }: { nav: DocsNav; children: ReactNode }) {
  return (
    <NavigationProvider>
      {process.env.NODE_ENV === 'development' ? <DevReload /> : null}
      <PageTitleProvider>
        <PreferencesScope>
          <SidebarProvider>
            <DocsSidebar nav={nav} />
            <SidebarInset>
              <DashboardNavbar />
              <main
                id="main"
                tabIndex={-1}
                className="flex w-full min-w-0 flex-1 flex-col p-5 outline-none in-data-[embed=1]:p-3 lg:p-6"
              >
                {children}
              </main>
              <SiteFooter />
            </SidebarInset>
          </SidebarProvider>
        </PreferencesScope>
      </PageTitleProvider>
    </NavigationProvider>
  );
}
