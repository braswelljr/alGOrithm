'use client';

import { useEffect } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { SETTINGS_FIRST, SETTINGS_SECTIONS } from '@/config/settings';

/** /settings/ opens on the first tab. A static export has no server to
 *  redirect, so the page replaces itself, and lists the tabs until it has. */
export function OpenFirstTab() {
  const router = useRouter();
  useEffect(() => router.replace(`/settings/${SETTINGS_FIRST}/`), [router]);
  return (
    <ul className="text-sm">
      {SETTINGS_SECTIONS.map((s) => (
        <li key={s.to}>
          <Link
            href={`/settings/${s.to}/`}
            className="text-accent underline-offset-4 hover:underline"
          >
            {s.label}
          </Link>
        </li>
      ))}
    </ul>
  );
}
