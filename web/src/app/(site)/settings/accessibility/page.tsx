import type { Metadata } from 'next';
import { AccessibilitySettings } from '@/components/settings/accessibility-settings';
import { pageMetadata } from '@/lib/metadata';

export const metadata: Metadata = {
  ...pageMetadata({
    title: 'Accessibility',
    description: 'Accessibility settings: kept in this browser.',
    path: '/settings/accessibility/'
  }),
  robots: { index: false }
};

export default function AccessibilityPage() {
  return <AccessibilitySettings />;
}
