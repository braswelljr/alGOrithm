/**
 * The editor's filesystem before the terminal starts: the system folders,
 * owned by root as on a real machine, and the visitor's home, theirs. The
 * terminal then writes the repository and the starter files into it
 * (web/runtime/terminal Seed). After gnux's playground seed.
 */
import { MemFS } from './memfs';

export const HOME = '/home/guest';

/** The environment the terminal starts with. */
export const ENV: Record<string, string> = {
  HOME,
  USER: 'guest',
  LOGNAME: 'guest',
  SHELL: '/bin/gnux',
  TMPDIR: '/tmp',
  LANG: 'C.UTF-8',
  TERM: 'xterm-256color',
  COLUMNS: '100'
};

/** A filesystem with /home/guest ready and the working directory in it. */
export function createEditorFS(options: ConstructorParameters<typeof MemFS>[0] = {}): MemFS {
  const fs = new MemFS(options);
  fs.asRoot(() => {
    fs.mkdirSync('/home', 0o755);
    fs.mkdirSync('/tmp', 0o777);
    fs.chmodSync('/tmp', 0o1777);
    fs.mkdirSync(HOME, 0o755);
    fs.chownSync(HOME, fs.uid, fs.gid);
  });
  fs.chdir(HOME);
  return fs;
}
