import type { Metadata } from 'next';
import { OpenFirstTab } from '@/components/settings/open-first-tab';
import { pageMetadata } from '@/lib/metadata';

export const metadata: Metadata = {
  ...pageMetadata({
    title: 'Settings',
    description: 'How the site looks and reads in this browser.',
    path: '/settings/'
  }),
  robots: { index: false }
};

export default function SettingsIndex() {
  return <OpenFirstTab />;
}
