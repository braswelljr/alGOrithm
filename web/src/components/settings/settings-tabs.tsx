'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { PageContent } from '@/components/shared/page-content';
import { PageHeader } from '@/components/shared/page-header';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Separator } from '@/components/ui/separator';
import { Tabs, TabsList, TabsTab } from '@/components/ui/tabs';
import { SETTINGS_SECTIONS } from '@/config/settings';

/**
 * Settings, as sections that are routes, the way gsl-certificates lays them
 * out. The open section is read from the address rather than held here, so
 * arriving on /settings/accessibility/ from a link opens Accessibility at once.
 */
export function SettingsTabs({ children }: { children: React.ReactNode }) {
  const pathname = usePathname() ?? '';
  const segment = pathname.replace(/^\/settings\/?/, '').split('/')[0];
  const active = SETTINGS_SECTIONS.find((s) => s.to === segment);
  return (
    <PageContent>
      <PageHeader
        title="Settings"
        hideTitle
        description="How the site looks and reads in this browser. Every choice applies at once and is kept here, nowhere else."
      />
      {/* One section is no choice: no strip and no rule. */}
      {SETTINGS_SECTIONS.length > 1 ? (
        <Tabs value={active?.to ?? SETTINGS_SECTIONS[0]?.to}>
          <ScrollArea className="h-auto shrink-0">
            <TabsList
              indicatorClassName="bg-primary!"
              aria-label="Settings sections"
              className="w-max"
            >
              {SETTINGS_SECTIONS.map((s) => (
                <TabsTab
                  key={s.to}
                  value={s.to}
                  nativeButton={false}
                  render={<Link href={`/settings/${s.to}/`} />}
                >
                  {s.label}
                </TabsTab>
              ))}
            </TabsList>
          </ScrollArea>
          <Separator />
        </Tabs>
      ) : null}
      {active ? <p className="text-muted text-sm">{active.description}</p> : null}
      {children}
    </PageContent>
  );
}
