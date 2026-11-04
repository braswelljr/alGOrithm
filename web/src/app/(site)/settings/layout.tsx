import { SettingsTabs } from '@/components/settings/settings-tabs';

export default function SettingsLayout({ children }: { children: React.ReactNode }) {
  return <SettingsTabs>{children}</SettingsTabs>;
}
