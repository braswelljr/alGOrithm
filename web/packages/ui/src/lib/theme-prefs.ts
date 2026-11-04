import { accents, defaultAccent, defaultNeutral, defaultViz, neutrals, vizPalettes } from '@braswelljr/algorithm-ui/lib/colors';

/** The reader's color choices, beside light/dark (which next-themes keeps). */
export type ThemePrefs = { neutral: string; accent: string; viz: string };

export const prefsKey = 'algorithm:theme';

export const defaultPrefs: ThemePrefs = {
  neutral: defaultNeutral,
  accent: defaultAccent,
  viz: defaultViz
};

/** Reads the saved choices, dropping any that are not known. */
export function readPrefs(raw: string | null): ThemePrefs {
  let v: unknown = null;
  try {
    v = raw ? JSON.parse(raw) : null;
  } catch {
    v = null;
  }
  const o = (v && typeof v === 'object' ? v : {}) as Record<string, unknown>;
  const pick = (value: unknown, ids: readonly { id: string }[], def: string) =>
    typeof value === 'string' && ids.some((x) => x.id === value) ? value : def;
  return {
    neutral: pick(o.neutral, neutrals, defaultNeutral),
    accent: pick(o.accent, accents, defaultAccent),
    viz: pick(o.viz, vizPalettes, defaultViz)
  };
}

/** Sets the choices on <html> as data attributes (the defaults as none). */
export function applyPrefs(p: ThemePrefs, root: HTMLElement = document.documentElement) {
  const set = (name: string, value: string, def: string) => {
    if (value === def) root.removeAttribute(`data-${name}`);
    else root.setAttribute(`data-${name}`, value);
  };
  set('neutral', p.neutral, defaultNeutral);
  set('accent', p.accent, defaultAccent);
  set('viz', p.viz, defaultViz);
}

/**
 * A script for <head> that applies the saved choices before the first paint,
 * so a reader who picked rose never sees indigo flash. It is a string, run
 * inline; it only reads localStorage.
 */
export const prefsScript = `try{var p=JSON.parse(localStorage.getItem('${prefsKey}')||'{}');var r=document.documentElement;${[
  ['neutral', defaultNeutral, neutrals.map((n) => n.id)],
  ['accent', defaultAccent, accents.map((a) => a.id)],
  ['viz', defaultViz, vizPalettes.map((v) => v.id)]
]
  .map(
    ([k, d, ids]) =>
      `if(${JSON.stringify(ids)}.indexOf(p.${k})>=0&&p.${k}!=='${d}')r.setAttribute('data-${k}',p.${k});`
  )
  .join('')}}catch(e){}`;
