import type { ReactNode } from 'react';
import { SiteLayout } from '@/components/layouts/site-layout';

/** Every page but the docs sits in the site layout: sidebar, top bar, page. */
export default function SiteRouteLayout({ children }: { children: ReactNode }) {
  return <SiteLayout>{children}</SiteLayout>;
}
