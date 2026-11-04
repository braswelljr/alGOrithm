import { Suspense, type ReactNode } from 'react';
import { SidebarInset, SidebarProvider } from '@braswelljr/algorithm-ui/components/sidebar';
import { AppSidebar } from '@/components/dashboard/app-sidebar';
import { DashboardNavbar } from '@/components/dashboard/dashboard-navbar';
import { DevReload } from '@/components/shared/dev-reload';
import { SiteFooter } from '@/components/shared/site-footer';
import { NavigationProvider } from '@/context/use-navigation';

/**
 * The dashboard shell around every page: the collapsible sidebar (a drawer on
 * a phone), the top bar with the page's name and trail, and the page beneath.
 *
 * In embed mode (?embed=1, a page shown inside someone else's iframe) an
 * inline script in the root layout marks the document before it paints and
 * the in-data-[embed=1] variants hide the chrome, so nothing here reads the
 * query: a component that did would make Next.js render every page in the
 * browser only, leaving the exported HTML, and search engines, empty. Only the
 * page goes inside Suspense, so a page that reads its query still keeps the
 * chrome in the HTML.
 */
export function SiteShell({ children }: { children: ReactNode }) {
  return (
    <NavigationProvider>
      {/* Development only: in a production build this and its zod import drop out. */}
      {process.env.NODE_ENV === 'development' ? <DevReload /> : null}
      <SidebarProvider>
        <AppSidebar />
        <SidebarInset>
          <DashboardNavbar />
          <main
            id="main"
            tabIndex={-1}
            // The one place a page's outer padding is set: pages add none of
            // their own and run to the page's edge rather than a fixed width.
            className="flex w-full min-w-0 flex-1 flex-col p-5 outline-none in-data-[embed=1]:p-3 lg:p-6"
          >
            <Suspense>{children}</Suspense>
          </main>
          <SiteFooter />
        </SidebarInset>
      </SidebarProvider>
    </NavigationProvider>
  );
}
