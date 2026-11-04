import type { Metadata } from 'next';
import { AppearanceSettings } from '@/components/settings/appearance-settings';
import { pageMetadata } from '@/lib/metadata';

export const metadata: Metadata = {
  ...pageMetadata({
    title: 'Appearance',
    description: 'Appearance settings: kept in this browser.',
    path: '/settings/appearance/'
  }),
  robots: { index: false }
};

export default function AppearancePage() {
  return <AppearanceSettings />;
}
