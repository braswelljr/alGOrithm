import type { ComponentType } from 'react';

/** The menu's sections, in the order they are shown. Design sits apart, in
 *  the sidebar's footer. */
export const NAV_SECTIONS = ['Overview', 'Run', 'Concepts', 'Reference', 'Design'] as const;

export type NavSectionLabel = (typeof NAV_SECTIONS)[number];

/** One page in the menu. */
export type NavItem = {
  label: string;
  /** The page's path, with the trailing slash the static export serves. */
  href: string;
  icon: ComponentType<{ className?: string; 'aria-hidden'?: boolean }>;
};

/** A labelled block of the menu. Every page has a fixed place. */
export type NavSection = { label: NavSectionLabel; items: NavItem[] };

/** A page as the sidebar shows it on one address. */
export type NavigationItem = NavItem & { isActive: boolean };

/** A section as the sidebar shows it on one address. */
export type NavigationGroup = { label: NavSectionLabel; items: NavigationItem[] };
