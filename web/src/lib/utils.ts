import { clsx, type ClassValue } from 'clsx';
import { extendTailwindMerge } from 'tailwind-merge';

// tailwind-merge only knows Tailwind's built-in scales, so a custom size from
// the @theme block (--text-xsm in styles/globals.css) is not recognised as a
// font size. It gets bucketed as a text *colour* instead, stops conflicting
// with text-sm, and both survive the merge, leaving the winner to stylesheet
// order, which is why `text-xsm` on a Button silently lost to the variant's
// `text-sm`. Declaring it here makes the two conflict again so the caller wins.
const twMerge = extendTailwindMerge({
  extend: {
    classGroups: {
      'font-size': [{ text: ['xsm'] }]
    }
  }
});

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function createSlug(str: string) {
  const slug = String(str);
  return slug
    .toLowerCase()
    .trim()
    .replace(/[^\w\s-]/g, '') // 1. Remove non-word chars (special chars like ?, :, /, &)
    .replace(/\s+/g, '-') // 2. Replace spaces with -
    .replace(/-+/g, '-'); // 3. Replace multiple dashes with single -
}
