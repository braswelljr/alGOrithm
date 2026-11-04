/** What the "Real-world code" section shows for an entry. */

export interface RealWorldUse {
  title: string;
  where: string;
  example: string;
}

/** The distinct example programs an entry's uses link, in order. */
export function examplePrograms(uses: readonly RealWorldUse[]): string[] {
  const seen = new Set<string>();
  for (const u of uses) if (u.example) seen.add(u.example);
  return [...seen];
}

/** The command that runs an example program from the repository root. */
export function runCommand(path: string): string {
  return `go run ./${path.replace(/^\.\//, '').replace(/\/$/, '')}`;
}

/** The program's directory name, used as its title: "leaderboard". */
export function programName(path: string): string {
  return path.replace(/\/$/, '').split('/').pop() ?? path;
}
