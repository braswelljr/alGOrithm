'use client';

import { useEffect } from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarRail,
  SidebarSeparator,
  useSidebar
} from '@braswelljr/algorithm-ui/components/sidebar';
import { SidebarBrand } from '@/components/shared/sidebar-brand';
import { useNavigation } from '@/context/use-navigation';
import type { NavigationItem } from '@/types/navigation';
import { cn } from '@/utils/utils';

/**
 * The site's navigation: the navy rail, the logo at the top, the open page
 * picked out in the accent, and the design pages apart in the footer. The
 * links come from config/navigation.ts, the one menu.
 */
export function AppSidebar() {
  const { isMobile, setOpenMobile } = useSidebar();
  const pathname = usePathname();
  const { groups, design } = useNavigation();

  // A phone's drawer closes once the reader has gone somewhere.
  useEffect(() => {
    if (isMobile) setOpenMobile(false);
  }, [pathname, isMobile, setOpenMobile]);

  return (
    <Sidebar
      id="site-sidebar"
      collapsible="icon"
      className="h-dvh border-x-0! in-data-[embed=1]:hidden"
    >
      <SidebarHeader>
        <SidebarBrand />
      </SidebarHeader>

      <SidebarContent>
        {/* One labelled block per section, always in the same order, so a
            page is found in the same place on every visit. */}
        <nav
          aria-label="Main"
          className="flex flex-col gap-2"
        >
          {groups.map((group) => (
            <NavGroup
              key={group.label}
              label={group.label}
              items={group.items}
            />
          ))}
        </nav>
      </SidebarContent>

      {design.length > 0 ? (
        <>
          <SidebarSeparator className="bg-white/15" />
          <SidebarFooter>
            <nav aria-label="Design">
              <NavGroup
                label="Design"
                items={design}
                className="p-0"
              />
            </nav>
          </SidebarFooter>
        </>
      ) : null}

      <SidebarRail />
    </Sidebar>
  );
}

export function NavGroup({
  label,
  items,
  className
}: {
  label: string;
  items: NavigationItem[];
  className?: string;
}) {
  return (
    <SidebarGroup className={className}>
      <SidebarGroupLabel className="text-white/60">{label}</SidebarGroupLabel>
      <SidebarMenu className="gap-1.5">
        {items.map((item) => (
          <SidebarMenuItem key={item.href}>
            <SidebarMenuButton
              tooltip={item.label}
              isActive={item.isActive}
              render={
                <Link
                  href={item.href}
                  aria-current={item.isActive ? 'page' : undefined}
                />
              }
              className={cn('text-white hover:text-white')}
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
  );
}
