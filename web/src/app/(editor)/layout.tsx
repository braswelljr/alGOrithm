import type { ReactNode } from 'react';
import { PreferencesScope } from '@/components/settings/preferences-scope';
import { PageTitleProvider } from '@/context/use-page-title';

/**
 * The editor takes the whole screen: no site sidebar or top bar around it. On a
 * wide screen it is fixed, so the page itself never scrolls (focusing the
 * terminal or the editor used to scroll it up) and each panel scrolls inside
 * itself; on a phone the parts stack down a page that scrolls.
 */
export default function EditorRouteLayout({ children }: { children: ReactNode }) {
  return (
    <PageTitleProvider>
      <PreferencesScope>
        <main
          id="main"
          tabIndex={-1}
          className="outline-none md:fixed md:inset-0 md:overflow-hidden"
        >
          {children}
        </main>
      </PreferencesScope>
    </PageTitleProvider>
  );
}
