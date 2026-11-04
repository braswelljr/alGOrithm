/**
 * A learner's progress, kept in this browser only (docs/rules/the-users-data-never-leaves-the-browser.md):
 * the placement test's level and the learning paths marked done.
 */
export type Placement = { level: 'beginner' | 'intermediate' | 'advanced'; at: string };

const placementKey = 'algorithm:placement';
const doneKey = 'algorithm:paths-done';

function read(key: string): unknown {
  try {
    return JSON.parse(localStorage.getItem(key) ?? 'null');
  } catch {
    return null;
  }
}

function write(key: string, value: unknown) {
  try {
    localStorage.setItem(key, JSON.stringify(value));
  } catch {
    // Blocked storage: progress lasts this visit.
  }
}

export function readPlacement(): Placement | null {
  const v = read(placementKey) as Partial<Placement> | null;
  return v && (v.level === 'beginner' || v.level === 'intermediate' || v.level === 'advanced')
    ? { level: v.level, at: typeof v.at === 'string' ? v.at : '' }
    : null;
}

export function savePlacement(level: Placement['level']) {
  write(placementKey, { level, at: new Date().toISOString() });
}

export function readDone(): string[] {
  const v = read(doneKey);
  return Array.isArray(v) ? v.filter((x): x is string => typeof x === 'string') : [];
}

export function markDone(id: string, done = true) {
  const now = readDone().filter((x) => x !== id);
  write(doneKey, done ? [...now, id] : now);
}
