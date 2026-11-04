/**
 * An in-memory POSIX-style filesystem that satisfies the contract Go's js/wasm
 * runtime expects from the Node-style `fs` and `process` globals (see
 * `$GOROOT/src/syscall/fs_js.go`).
 *
 * The core is synchronous and throws {@link FsError}; {@link MemFS.nodeFs}
 * adapts it to the callback style Go calls. It supports directories, regular
 * files, symbolic links, hard links, permissions for a single user, timestamps
 * and open files that outlive their last name.
 *
 * This file is plain erasable TypeScript with no imports so the same source
 * runs in the browser bundle and under `node --test`.
 */

export const S_IFMT = 0o170000;
export const S_IFDIR = 0o040000;
export const S_IFCHR = 0o020000;
export const S_IFREG = 0o100000;
export const S_IFLNK = 0o120000;

/** Open flags, with Linux's values. Go only needs them to be distinct bits. */
export const constants = {
  O_RDONLY: 0,
  O_WRONLY: 1,
  O_RDWR: 2,
  O_CREAT: 0o100,
  O_EXCL: 0o200,
  O_TRUNC: 0o1000,
  O_APPEND: 0o2000,
  O_DIRECTORY: 0o200000
} as const;

const O_ACCMODE = 3;

/** Symbolic links followed in one lookup before it fails with ELOOP. */
const MAX_SYMLINKS = 40;

/** A uid/gid of -1 (as a uint32) means "leave unchanged" in chown. */
const ID_UNCHANGED = 0xffffffff;

/**
 * The error codes this filesystem raises. Every one must be a key of Go's
 * `errnoByCode` table: Go panics on a code it does not know.
 */
export type ErrorCode =
  | 'EACCES'
  | 'EBADF'
  | 'EBUSY'
  | 'EEXIST'
  | 'EINVAL'
  | 'EIO'
  | 'EISDIR'
  | 'ELOOP'
  | 'ENOENT'
  | 'ENOSPC'
  | 'ENOTDIR'
  | 'ENOTEMPTY'
  | 'EPERM';

const MESSAGES: Record<ErrorCode, string> = {
  EACCES: 'permission denied',
  EBADF: 'bad file descriptor',
  EBUSY: 'resource busy or locked',
  EEXIST: 'file already exists',
  EINVAL: 'invalid argument',
  EIO: 'i/o error',
  EISDIR: 'illegal operation on a directory',
  ELOOP: 'too many symbolic links encountered',
  ENOENT: 'no such file or directory',
  ENOSPC: 'no space left on device',
  ENOTDIR: 'not a directory',
  ENOTEMPTY: 'directory not empty',
  EPERM: 'operation not permitted'
};

/** The error every operation throws, shaped like Node's system errors. */
export class FsError extends Error {
  code: ErrorCode;
  syscall: string;
  path?: string;

  constructor(code: ErrorCode, syscall: string, path?: string) {
    super(`${code}: ${MESSAGES[code]}, ${syscall}${path === undefined ? '' : ` '${path}'`}`);
    this.name = 'FsError';
    this.code = code;
    this.syscall = syscall;
    this.path = path;
  }
}

interface NodeBase {
  ino: number;
  /** Permission bits only (0o7777); the type bits come from `kind`. */
  mode: number;
  uid: number;
  gid: number;
  nlink: number;
  atimeMs: number;
  mtimeMs: number;
  ctimeMs: number;
}

interface FileNode extends NodeBase {
  kind: 'file';
  /** Backing store; only the first `size` bytes are the file's contents. */
  data: Uint8Array;
  size: number;
  /** Open descriptors, so the bytes of an unlinked file are freed on close. */
  opens: number;
}

interface DirNode extends NodeBase {
  kind: 'dir';
  entries: Map<string, Inode>;
  parent: DirNode;
  name: string;
}

interface LinkNode extends NodeBase {
  kind: 'symlink';
  target: string;
}

type Inode = FileNode | DirNode | LinkNode;

interface OpenFile {
  node: FileNode | DirNode;
  flags: number;
  pos: number;
}

/** Where a path lookup ended. `node` is undefined when the last name is free. */
interface Lookup {
  parent: DirNode;
  name: string;
  node: Inode | undefined;
  /** The path's last component was a literal `.` or `..`. */
  dot: boolean;
  /** The path ended in a slash, so it can only name a directory. */
  slash: boolean;
}

/** The object Go receives from stat, lstat and fstat. */
export interface Stats {
  dev: number;
  ino: number;
  mode: number;
  nlink: number;
  uid: number;
  gid: number;
  rdev: number;
  size: number;
  blksize: number;
  blocks: number;
  atimeMs: number;
  mtimeMs: number;
  ctimeMs: number;
  isDirectory(): boolean;
  isFile(): boolean;
  isSymbolicLink(): boolean;
}

export interface MemFSOptions {
  uid?: number;
  gid?: number;
  umask?: number;
  /** Total bytes of file data allowed before writes fail with ENOSPC. */
  maxBytes?: number;
  /** Total files, directories and links allowed before creation fails with ENOSPC. */
  maxInodes?: number;
  /** Receives whole lines written to fd 1 or 2 by the Go runtime itself (panics, println). */
  onConsole?: (fd: number, line: string) => void;
}

type Callback<T = unknown> = (err: Error | null, value?: T) => void;

/** The callback-style surface installed as `globalThis.fs`. */
export interface NodeFs {
  constants: typeof constants;
  writeSync(fd: number, buf: Uint8Array): number;
  open(path: string, flags: number, mode: number, cb: Callback<number>): void;
  close(fd: number, cb: Callback): void;
  read(
    fd: number,
    buf: Uint8Array,
    offset: number,
    length: number,
    position: number | null,
    cb: Callback<number>
  ): void;
  write(
    fd: number,
    buf: Uint8Array,
    offset: number,
    length: number,
    position: number | null,
    cb: Callback<number>
  ): void;
  fstat(fd: number, cb: Callback<Stats>): void;
  stat(path: string, cb: Callback<Stats>): void;
  lstat(path: string, cb: Callback<Stats>): void;
  mkdir(path: string, mode: number, cb: Callback): void;
  readdir(path: string, cb: Callback<string[]>): void;
  rmdir(path: string, cb: Callback): void;
  unlink(path: string, cb: Callback): void;
  rename(from: string, to: string, cb: Callback): void;
  chmod(path: string, mode: number, cb: Callback): void;
  fchmod(fd: number, mode: number, cb: Callback): void;
  chown(path: string, uid: number, gid: number, cb: Callback): void;
  fchown(fd: number, uid: number, gid: number, cb: Callback): void;
  lchown(path: string, uid: number, gid: number, cb: Callback): void;
  utimes(path: string, atime: number, mtime: number, cb: Callback): void;
  symlink(target: string, path: string, cb: Callback): void;
  readlink(path: string, cb: Callback<string>): void;
  link(existing: string, path: string, cb: Callback): void;
  truncate(path: string, length: number, cb: Callback): void;
  ftruncate(fd: number, length: number, cb: Callback): void;
  fsync(fd: number, cb: Callback): void;
}

/** The subset of Node's `process` Go reads, installed as `globalThis.process`. */
export interface NodeProcess {
  pid: number;
  ppid: number;
  env: Record<string, string>;
  cwd(): string;
  chdir(path: string): void;
  umask(mask?: number): number;
  getuid(): number;
  getgid(): number;
  geteuid(): number;
  getegid(): number;
  getgroups(): number[];
}

const encoder = new TextEncoder();
const decoder = new TextDecoder();

export class MemFS {
  readonly uid: number;
  readonly gid: number;

  private readonly root: DirNode;
  private cwdNode: DirNode;
  private mask: number;
  private nextIno = 2;
  private nextFd = 3;
  private readonly fds = new Map<number, OpenFile>();
  private usedBytes = 0;
  private inodes = 1;
  private readonly maxBytes: number;
  private readonly maxInodes: number;
  private readonly onConsole: (fd: number, line: string) => void;
  private readonly consoleBuf: [string, string] = ['', ''];
  /** Set while seeding, to create root-owned files and skip permission checks. */
  private privileged = false;

  constructor(options: MemFSOptions = {}) {
    this.uid = options.uid ?? 1000;
    this.gid = options.gid ?? 1000;
    this.mask = options.umask ?? 0o022;
    this.maxBytes = options.maxBytes ?? 64 * 1024 * 1024;
    this.maxInodes = options.maxInodes ?? 100_000;
    this.onConsole =
      options.onConsole ?? ((fd, line) => (fd === 2 ? console.error(line) : console.log(line)));

    const now = Date.now();
    const root = {
      kind: 'dir',
      ino: 1,
      mode: 0o755,
      uid: 0,
      gid: 0,
      nlink: 2,
      atimeMs: now,
      mtimeMs: now,
      ctimeMs: now,
      entries: new Map<string, Inode>(),
      name: ''
    } as DirNode;
    root.parent = root;
    this.root = root;
    this.cwdNode = root;
  }

  // ---------------------------------------------------------------------------
  // Process state
  // ---------------------------------------------------------------------------

  /** The absolute path of the working directory. Throws ENOENT if it was removed. */
  cwd(): string {
    if (this.cwdNode.nlink === 0) throw new FsError('ENOENT', 'uv_cwd');
    return this.pathOf(this.cwdNode);
  }

  chdir(path: string): void {
    const { node } = this.lookup(path, true, 'chdir');
    if (!node) throw new FsError('ENOENT', 'chdir', path);
    if (node.kind !== 'dir') throw new FsError('ENOTDIR', 'chdir', path);
    this.need(node, 1, 'chdir', path);
    this.cwdNode = node;
  }

  /** Sets the file-creation mask when given one, and returns the previous mask. */
  umask(mask?: number): number {
    const old = this.mask;
    if (mask !== undefined) this.mask = mask & 0o777;
    return old;
  }

  /** Lexically resolves `path` against the working directory, like `path.resolve`. */
  resolvePath(...segments: string[]): string {
    let joined = '';
    for (const s of segments) {
      if (s === '') continue;
      joined = s.startsWith('/') ? s : joined === '' ? s : `${joined}/${s}`;
    }
    if (!joined.startsWith('/')) {
      let base = '/';
      try {
        base = this.cwd();
      } catch {
        // A removed working directory has no path; fall back to the root.
      }
      joined = `${base}/${joined}`;
    }
    const out: string[] = [];
    for (const part of joined.split('/')) {
      if (part === '' || part === '.') continue;
      if (part === '..') out.pop();
      else out.push(part);
    }
    return `/${out.join('/')}`;
  }

  /** Runs `fn` as the superuser: no permission checks, and new files belong to root. */
  asRoot<T>(fn: () => T): T {
    const was = this.privileged;
    this.privileged = true;
    try {
      return fn();
    } finally {
      this.privileged = was;
    }
  }

  // ---------------------------------------------------------------------------
  // Path-based operations
  // ---------------------------------------------------------------------------

  statSync(path: string): Stats {
    const { node } = this.lookup(path, true, 'stat');
    if (!node) throw new FsError('ENOENT', 'stat', path);
    return this.stats(node);
  }

  lstatSync(path: string): Stats {
    const { node } = this.lookup(path, false, 'lstat');
    if (!node) throw new FsError('ENOENT', 'lstat', path);
    return this.stats(node);
  }

  existsSync(path: string): boolean {
    try {
      return this.lookup(path, false, 'lstat').node !== undefined;
    } catch {
      return false;
    }
  }

  mkdirSync(path: string, mode = 0o777): void {
    const at = this.lookup(path, false, 'mkdir');
    if (at.node) throw new FsError('EEXIST', 'mkdir', path);
    this.need(at.parent, 3, 'mkdir', path);
    const dir = this.newNode('mkdir', path, mode & ~this.mask & 0o7777, {
      kind: 'dir',
      entries: new Map<string, Inode>(),
      parent: at.parent,
      name: at.name
    }) as DirNode;
    dir.nlink = 2;
    at.parent.entries.set(at.name, dir);
    at.parent.nlink++;
    this.touchDir(at.parent);
  }

  /** `mkdir -p`: creates every missing directory along `path`. */
  mkdirpSync(path: string, mode = 0o777): void {
    const abs = this.resolvePath(path);
    let cur = '';
    for (const part of abs.split('/').filter(Boolean)) {
      cur += `/${part}`;
      if (!this.existsSync(cur)) this.mkdirSync(cur, mode);
    }
  }

  readdirSync(path: string): string[] {
    const { node } = this.lookup(path, true, 'scandir');
    if (!node) throw new FsError('ENOENT', 'scandir', path);
    if (node.kind !== 'dir') throw new FsError('ENOTDIR', 'scandir', path);
    this.need(node, 4, 'scandir', path);
    node.atimeMs = Date.now();
    return [...node.entries.keys()];
  }

  rmdirSync(path: string): void {
    const at = this.lookup(path, false, 'rmdir');
    const node = at.node;
    if (!node) throw new FsError('ENOENT', 'rmdir', path);
    if (node.kind !== 'dir') throw new FsError('ENOTDIR', 'rmdir', path);
    if (at.dot)
      throw new FsError(
        path.replace(/\/+$/, '').endsWith('..') ? 'ENOTEMPTY' : 'EINVAL',
        'rmdir',
        path
      );
    if (node === this.root) throw new FsError('EBUSY', 'rmdir', path);
    if (node.entries.size > 0) throw new FsError('ENOTEMPTY', 'rmdir', path);
    this.need(at.parent, 3, 'rmdir', path);
    at.parent.entries.delete(at.name);
    at.parent.nlink--;
    node.nlink = 0;
    this.inodes--;
    this.touchDir(at.parent);
  }

  unlinkSync(path: string): void {
    const at = this.lookup(path, false, 'unlink');
    const node = at.node;
    if (!node) throw new FsError('ENOENT', 'unlink', path);
    // Linux's answer; Go's os.Remove relies on it to fall through to rmdir.
    if (node.kind === 'dir') throw new FsError('EISDIR', 'unlink', path);
    this.need(at.parent, 3, 'unlink', path);
    at.parent.entries.delete(at.name);
    this.touchDir(at.parent);
    this.dropLink(node);
  }

  renameSync(from: string, to: string): void {
    const src = this.lookup(from, false, 'rename');
    const node = src.node;
    if (!node) throw new FsError('ENOENT', 'rename', from);
    const dst = this.lookup(to, false, 'rename');
    if (src.dot || dst.dot) throw new FsError('EINVAL', 'rename', from);
    if (node === this.root) throw new FsError('EBUSY', 'rename', from);
    if (dst.slash && node.kind !== 'dir') throw new FsError('ENOTDIR', 'rename', to);
    const target = dst.node;
    if (target === node) return; // Same file under two names: POSIX says do nothing.

    if (node.kind === 'dir') {
      // A directory cannot be moved into itself or its own subtree.
      for (let d = dst.parent; ; d = d.parent) {
        if (d === node) throw new FsError('EINVAL', 'rename', from);
        if (d === this.root) break;
      }
      if (target) {
        if (target.kind !== 'dir') throw new FsError('ENOTDIR', 'rename', to);
        if (target.entries.size > 0) throw new FsError('ENOTEMPTY', 'rename', to);
      }
    } else if (target?.kind === 'dir') {
      throw new FsError('EISDIR', 'rename', to);
    }
    this.need(src.parent, 3, 'rename', from);
    this.need(dst.parent, 3, 'rename', to);

    if (target) {
      dst.parent.entries.delete(dst.name);
      if (target.kind === 'dir') {
        dst.parent.nlink--;
        target.nlink = 0;
        this.inodes--;
      } else {
        this.dropLink(target);
      }
    }
    src.parent.entries.delete(src.name);
    dst.parent.entries.set(dst.name, node);
    if (node.kind === 'dir') {
      src.parent.nlink--;
      dst.parent.nlink++;
      node.parent = dst.parent;
      node.name = dst.name;
    }
    node.ctimeMs = Date.now();
    this.touchDir(src.parent);
    this.touchDir(dst.parent);
  }

  linkSync(existing: string, path: string): void {
    const src = this.lookup(existing, false, 'link');
    const node = src.node;
    if (!node) throw new FsError('ENOENT', 'link', existing);
    if (node.kind === 'dir') throw new FsError('EPERM', 'link', existing);
    const at = this.lookup(path, false, 'link');
    if (at.node) throw new FsError('EEXIST', 'link', path);
    if (at.slash) throw new FsError('ENOENT', 'link', path);
    this.need(at.parent, 3, 'link', path);
    at.parent.entries.set(at.name, node);
    node.nlink++;
    node.ctimeMs = Date.now();
    this.touchDir(at.parent);
  }

  symlinkSync(target: string, path: string): void {
    if (target === '') throw new FsError('ENOENT', 'symlink', path);
    const at = this.lookup(path, false, 'symlink');
    if (at.node) throw new FsError('EEXIST', 'symlink', path);
    if (at.slash) throw new FsError('ENOENT', 'symlink', path);
    this.need(at.parent, 3, 'symlink', path);
    const link = this.newNode('symlink', path, 0o777, { kind: 'symlink', target });
    at.parent.entries.set(at.name, link);
    this.touchDir(at.parent);
  }

  readlinkSync(path: string): string {
    const { node } = this.lookup(path, false, 'readlink');
    if (!node) throw new FsError('ENOENT', 'readlink', path);
    if (node.kind !== 'symlink') throw new FsError('EINVAL', 'readlink', path);
    return node.target;
  }

  chmodSync(path: string, mode: number): void {
    const { node } = this.lookup(path, true, 'chmod');
    if (!node) throw new FsError('ENOENT', 'chmod', path);
    this.setMode(node, mode, 'chmod', path);
  }

  chownSync(path: string, uid: number, gid: number): void {
    const { node } = this.lookup(path, true, 'chown');
    if (!node) throw new FsError('ENOENT', 'chown', path);
    this.setOwner(node, uid, gid, 'chown', path);
  }

  lchownSync(path: string, uid: number, gid: number): void {
    const { node } = this.lookup(path, false, 'lchown');
    if (!node) throw new FsError('ENOENT', 'lchown', path);
    this.setOwner(node, uid, gid, 'lchown', path);
  }

  /** Sets the access and modification times, given in seconds since the epoch. */
  utimesSync(path: string, atime: number, mtime: number): void {
    const { node } = this.lookup(path, true, 'utime');
    if (!node) throw new FsError('ENOENT', 'utime', path);
    if (!this.owns(node)) throw new FsError('EPERM', 'utime', path);
    node.atimeMs = Math.floor(atime * 1000);
    node.mtimeMs = Math.floor(mtime * 1000);
    node.ctimeMs = Date.now();
  }

  truncateSync(path: string, length: number): void {
    const { node } = this.lookup(path, true, 'truncate');
    if (!node) throw new FsError('ENOENT', 'truncate', path);
    if (node.kind === 'dir') throw new FsError('EISDIR', 'truncate', path);
    if (node.kind !== 'file') throw new FsError('EINVAL', 'truncate', path);
    this.need(node, 2, 'truncate', path);
    this.resize(node, length, 'truncate');
  }

  // ---------------------------------------------------------------------------
  // Descriptor-based operations
  // ---------------------------------------------------------------------------

  openSync(path: string, flags: number = constants.O_RDONLY, mode = 0o666): number {
    const access = flags & O_ACCMODE;
    const creating = (flags & constants.O_CREAT) !== 0;
    const exclusive = creating && (flags & constants.O_EXCL) !== 0;
    const wantsWrite = access !== constants.O_RDONLY;
    const truncating = (flags & constants.O_TRUNC) !== 0;

    // O_CREAT|O_EXCL must not follow a symbolic link in the last component.
    const at = this.lookup(path, !exclusive, 'open');
    let node = at.node;
    if (node && exclusive) throw new FsError('EEXIST', 'open', path);

    if (!node) {
      if (!creating) throw new FsError('ENOENT', 'open', path);
      if (at.slash) throw new FsError('EISDIR', 'open', path);
      this.need(at.parent, 3, 'open', path);
      node = this.newNode('open', path, mode & ~this.mask & 0o7777, {
        kind: 'file',
        data: new Uint8Array(0),
        size: 0,
        opens: 0
      });
      at.parent.entries.set(at.name, node);
      this.touchDir(at.parent);
    } else {
      if (node.kind === 'symlink') throw new FsError('ELOOP', 'open', path);
      if (node.kind === 'dir') {
        if (wantsWrite || truncating) throw new FsError('EISDIR', 'open', path);
      } else if ((flags & constants.O_DIRECTORY) !== 0) {
        throw new FsError('ENOTDIR', 'open', path);
      }
      if (access !== constants.O_WRONLY) this.need(node, 4, 'open', path);
      if (wantsWrite || truncating) this.need(node, 2, 'open', path);
      if (truncating && node.kind === 'file') this.resize(node, 0, 'open');
    }

    const file = node as FileNode | DirNode;
    if (file.kind === 'file') file.opens++;
    const fd = this.nextFd++;
    this.fds.set(fd, { node: file, flags, pos: 0 });
    return fd;
  }

  closeSync(fd: number): void {
    if (fd >= 0 && fd <= 2) return;
    const file = this.fds.get(fd);
    if (!file) throw new FsError('EBADF', 'close');
    this.fds.delete(fd);
    if (file.node.kind === 'file') {
      file.node.opens--;
      this.release(file.node);
    }
  }

  /**
   * Reads up to `length` bytes into `buf` at `offset`. A null `position` reads
   * from, and advances, the descriptor's own offset; a number reads from there
   * and leaves the offset alone. Returns 0 at end of file.
   */
  readSync(
    fd: number,
    buf: Uint8Array,
    offset: number,
    length: number,
    position: number | null = null
  ): number {
    if (fd === 0) return 0; // Standard input is always at end of file.
    const file = this.fds.get(fd);
    if (!file) throw new FsError('EBADF', 'read');
    if (file.node.kind === 'dir') throw new FsError('EISDIR', 'read');
    if ((file.flags & O_ACCMODE) === constants.O_WRONLY) throw new FsError('EBADF', 'read');
    const node = file.node;
    const start = position === null || position === undefined ? file.pos : Number(position);
    if (start < 0) throw new FsError('EINVAL', 'read');
    const n = Math.max(0, Math.min(length, buf.length - offset, node.size - start));
    if (n > 0) buf.set(node.data.subarray(start, start + n), offset);
    if (position === null || position === undefined) file.pos = start + n;
    node.atimeMs = Date.now();
    return n;
  }

  /**
   * Writes `length` bytes of `buf` from `offset`. A null `position` writes at,
   * and advances, the descriptor's own offset. With O_APPEND every write lands
   * at the end of the file, whatever position was asked for.
   *
   * Descriptors 1 and 2 are not part of the virtual filesystem: gnux reports
   * its output through its own callback, so only the Go runtime writes here.
   */
  writeSync(
    fd: number,
    buf: Uint8Array,
    offset = 0,
    length = buf.length - offset,
    position: number | null = null
  ): number {
    if (fd === 1 || fd === 2) {
      this.console(fd, buf.subarray(offset, offset + length));
      return length;
    }
    const file = this.fds.get(fd);
    if (!file) throw new FsError('EBADF', 'write');
    if (file.node.kind === 'dir' || (file.flags & O_ACCMODE) === constants.O_RDONLY) {
      throw new FsError('EBADF', 'write');
    }
    const node = file.node;
    const appending = (file.flags & constants.O_APPEND) !== 0;
    const sequential = position === null || position === undefined;
    const start = appending ? node.size : sequential ? file.pos : Number(position);
    if (start < 0) throw new FsError('EINVAL', 'write');
    const end = start + length;
    if (end > node.size) this.resize(node, end, 'write');
    node.data.set(buf.subarray(offset, offset + length), start);
    if (sequential || appending) file.pos = end;
    if (length > 0) node.mtimeMs = node.ctimeMs = Date.now();
    return length;
  }

  fstatSync(fd: number): Stats {
    if (fd >= 0 && fd <= 2) return this.ttyStats(fd);
    const file = this.fds.get(fd);
    if (!file) throw new FsError('EBADF', 'fstat');
    return this.stats(file.node);
  }

  fchmodSync(fd: number, mode: number): void {
    const file = this.fds.get(fd);
    if (!file) throw new FsError('EBADF', 'fchmod');
    this.setMode(file.node, mode, 'fchmod');
  }

  fchownSync(fd: number, uid: number, gid: number): void {
    const file = this.fds.get(fd);
    if (!file) throw new FsError('EBADF', 'fchown');
    this.setOwner(file.node, uid, gid, 'fchown');
  }

  ftruncateSync(fd: number, length: number): void {
    const file = this.fds.get(fd);
    if (!file) throw new FsError('EBADF', 'ftruncate');
    if (file.node.kind !== 'file' || (file.flags & O_ACCMODE) === constants.O_RDONLY) {
      throw new FsError('EINVAL', 'ftruncate');
    }
    this.resize(file.node, length, 'ftruncate');
  }

  fsyncSync(fd: number): void {
    if (fd >= 0 && fd <= 2) return;
    if (!this.fds.has(fd)) throw new FsError('EBADF', 'fsync');
  }

  // ---------------------------------------------------------------------------
  // Whole-file conveniences (for seeding, the UI and tests)
  // ---------------------------------------------------------------------------

  readFileSync(path: string): Uint8Array {
    const fd = this.openSync(path, constants.O_RDONLY);
    try {
      const { size } = this.fstatSync(fd);
      const buf = new Uint8Array(size);
      this.readSync(fd, buf, 0, size, 0);
      return buf;
    } finally {
      this.closeSync(fd);
    }
  }

  readTextSync(path: string): string {
    return decoder.decode(this.readFileSync(path));
  }

  writeFileSync(path: string, data: string | Uint8Array, mode = 0o666): void {
    const bytes = typeof data === 'string' ? encoder.encode(data) : data;
    const fd = this.openSync(
      path,
      constants.O_WRONLY | constants.O_CREAT | constants.O_TRUNC,
      mode
    );
    try {
      this.writeSync(fd, bytes, 0, bytes.length, 0);
    } finally {
      this.closeSync(fd);
    }
  }

  // ---------------------------------------------------------------------------
  // The Node-style objects Go talks to
  // ---------------------------------------------------------------------------

  /**
   * The callback-style `fs` object. Callbacks are invoked synchronously, which
   * Go's runtime supports (its own fallback `fs.write` does the same) and which
   * avoids a trip through the event loop for every system call.
   */
  nodeFs(): NodeFs {
    // Wraps a synchronous operation as `(...args, callback)`. The callback is
    // called outside the try block so an exception inside it is never mistaken
    // for a filesystem error and reported twice.
    const wrap = <A extends unknown[], R>(fn: (...args: A) => R) => {
      return (...args: [...A, Callback<R>]): void => {
        const cb = args.pop() as Callback<R>;
        let value: R;
        try {
          value = fn(...(args as unknown as A));
        } catch (err) {
          cb(this.toNodeError(err));
          return;
        }
        cb(null, value);
      };
    };

    return {
      constants,
      writeSync: (fd, buf) => this.writeSync(fd, buf),
      open: wrap((path: string, flags: number, mode: number) => this.openSync(path, flags, mode)),
      close: wrap((fd: number) => this.closeSync(fd)),
      read: wrap(
        (fd: number, buf: Uint8Array, offset: number, length: number, position: number | null) =>
          this.readSync(fd, buf, offset, length, position)
      ),
      write: wrap(
        (fd: number, buf: Uint8Array, offset: number, length: number, position: number | null) =>
          this.writeSync(fd, buf, offset, length, position)
      ),
      fstat: wrap((fd: number) => this.fstatSync(fd)),
      stat: wrap((path: string) => this.statSync(path)),
      lstat: wrap((path: string) => this.lstatSync(path)),
      mkdir: wrap((path: string, mode: number) => this.mkdirSync(path, mode)),
      readdir: wrap((path: string) => this.readdirSync(path)),
      rmdir: wrap((path: string) => this.rmdirSync(path)),
      unlink: wrap((path: string) => this.unlinkSync(path)),
      rename: wrap((from: string, to: string) => this.renameSync(from, to)),
      chmod: wrap((path: string, mode: number) => this.chmodSync(path, mode)),
      fchmod: wrap((fd: number, mode: number) => this.fchmodSync(fd, mode)),
      chown: wrap((path: string, uid: number, gid: number) => this.chownSync(path, uid, gid)),
      fchown: wrap((fd: number, uid: number, gid: number) => this.fchownSync(fd, uid, gid)),
      lchown: wrap((path: string, uid: number, gid: number) => this.lchownSync(path, uid, gid)),
      utimes: wrap((path: string, atime: number, mtime: number) =>
        this.utimesSync(path, atime, mtime)
      ),
      symlink: wrap((target: string, path: string) => this.symlinkSync(target, path)),
      readlink: wrap((path: string) => this.readlinkSync(path)),
      link: wrap((existing: string, path: string) => this.linkSync(existing, path)),
      truncate: wrap((path: string, length: number) => this.truncateSync(path, length)),
      ftruncate: wrap((fd: number, length: number) => this.ftruncateSync(fd, length)),
      fsync: wrap((fd: number) => this.fsyncSync(fd))
    };
  }

  /** The `process` object. `cwd` and `chdir` throw errors carrying a `code`, as Go expects. */
  nodeProcess(): NodeProcess {
    return {
      pid: 1,
      ppid: 0,
      env: {},
      cwd: () => this.cwd(),
      chdir: (path) => this.chdir(path),
      umask: (mask) => this.umask(mask),
      getuid: () => this.uid,
      getgid: () => this.gid,
      geteuid: () => this.uid,
      getegid: () => this.gid,
      getgroups: () => [this.gid]
    };
  }

  /** The `path` object; Go only calls `resolve`. */
  nodePath(): { resolve(...segments: string[]): string } {
    return { resolve: (...segments) => this.resolvePath(...segments) };
  }

  // ---------------------------------------------------------------------------
  // Internals
  // ---------------------------------------------------------------------------

  /**
   * Walks `path` and reports where it ends. Symbolic links in every component
   * but the last are always followed; the last one only when `follow` is set or
   * the path ends in a slash. A missing last component is not an error (the
   * caller may be about to create it); a missing earlier one is ENOENT.
   */
  private lookup(path: string, follow: boolean, syscall: string): Lookup {
    if (path === '') throw new FsError('ENOENT', syscall, path);
    let start = this.cwdNode;
    if (path.startsWith('/')) start = this.root;
    else if (start.nlink === 0) throw new FsError('ENOENT', syscall, path);

    const slash = path.endsWith('/');
    // Components still to walk, last first, so a link's target can be pushed on top.
    const todo = path
      .split('/')
      .filter((p) => p !== '')
      .reverse();
    const at: Lookup = { parent: start.parent, name: start.name, node: start, dot: false, slash };
    let links = 0;

    while (todo.length > 0) {
      const part = todo.pop() as string;
      const last = todo.length === 0;
      const dir = at.node;
      if (!dir) throw new FsError('ENOENT', syscall, path);
      if (dir.kind !== 'dir') throw new FsError('ENOTDIR', syscall, path);
      this.need(dir, 1, syscall, path);

      if (part === '.' || part === '..') {
        const to = part === '.' ? dir : dir.parent;
        at.parent = to.parent;
        at.name = to.name;
        at.node = to;
        at.dot = true;
        continue;
      }
      at.dot = false;

      const child = dir.entries.get(part);
      if (child?.kind === 'symlink' && (!last || follow || slash)) {
        if (++links > MAX_SYMLINKS) throw new FsError('ELOOP', syscall, path);
        if (child.target.startsWith('/')) {
          at.parent = this.root;
          at.name = '';
          at.node = this.root;
        }
        // Otherwise the target is relative to `dir`, which `at` still names.
        const parts = child.target.split('/').filter((p) => p !== '');
        for (let i = parts.length - 1; i >= 0; i--) todo.push(parts[i]);
        continue;
      }
      at.parent = dir;
      at.name = part;
      at.node = child;
    }

    if (slash && at.node && at.node.kind !== 'dir') throw new FsError('ENOTDIR', syscall, path);
    return at;
  }

  private pathOf(dir: DirNode): string {
    const parts: string[] = [];
    for (let d = dir; d !== this.root; d = d.parent) parts.push(d.name);
    return `/${parts.reverse().join('/')}`;
  }

  private owns(node: Inode): boolean {
    return this.privileged || node.uid === this.uid;
  }

  /** Throws EACCES unless the user has all of the `want` permission bits (r=4, w=2, x=1). */
  private need(node: Inode, want: number, syscall: string, path?: string): void {
    if (this.privileged) return;
    const shift = node.uid === this.uid ? 6 : node.gid === this.gid ? 3 : 0;
    if (((node.mode >> shift) & want) !== want) throw new FsError('EACCES', syscall, path);
  }

  private newNode<T extends { kind: Inode['kind'] }>(
    syscall: string,
    path: string,
    mode: number,
    rest: T
  ): T & NodeBase {
    if (this.inodes >= this.maxInodes) throw new FsError('ENOSPC', syscall, path);
    this.inodes++;
    const now = Date.now();
    return {
      ino: this.nextIno++,
      mode,
      uid: this.privileged ? 0 : this.uid,
      gid: this.privileged ? 0 : this.gid,
      nlink: 1,
      atimeMs: now,
      mtimeMs: now,
      ctimeMs: now,
      ...rest
    };
  }

  private touchDir(dir: DirNode): void {
    dir.mtimeMs = dir.ctimeMs = Date.now();
  }

  /** Removes one name from a file or link, freeing it once nothing refers to it. */
  private dropLink(node: FileNode | LinkNode): void {
    node.nlink--;
    node.ctimeMs = Date.now();
    if (node.nlink > 0) return;
    this.inodes--;
    if (node.kind === 'file') this.release(node);
  }

  /** Frees a file's bytes when it has neither a name nor an open descriptor. */
  private release(node: FileNode): void {
    if (node.nlink > 0 || node.opens > 0) return;
    this.usedBytes -= node.size;
    node.size = 0;
    node.data = new Uint8Array(0);
  }

  private resize(node: FileNode, size: number, syscall: string): void {
    if (!Number.isFinite(size) || size < 0) throw new FsError('EINVAL', syscall);
    size = Math.floor(size);
    if (size > node.size && this.usedBytes + (size - node.size) > this.maxBytes) {
      throw new FsError('ENOSPC', syscall);
    }
    if (size > node.data.length) {
      // Grow geometrically so a stream of small appends stays linear overall.
      const grown = new Uint8Array(
        Math.max(size, Math.min(node.data.length * 2, size + (1 << 20)), 64)
      );
      grown.set(node.data.subarray(0, node.size));
      node.data = grown;
    } else if (size < node.size) {
      // Bytes past the new end must read back as zeros if the file grows again.
      node.data.fill(0, size, node.size);
    }
    this.usedBytes += size - node.size;
    node.size = size;
    node.mtimeMs = node.ctimeMs = Date.now();
  }

  private setMode(node: Inode, mode: number, syscall: string, path?: string): void {
    if (!this.owns(node)) throw new FsError('EPERM', syscall, path);
    node.mode = mode & 0o7777;
    node.ctimeMs = Date.now();
  }

  private setOwner(node: Inode, uid: number, gid: number, syscall: string, path?: string): void {
    const newUid = uid === ID_UNCHANGED || uid === -1 ? node.uid : uid;
    const newGid = gid === ID_UNCHANGED || gid === -1 ? node.gid : gid;
    // An ordinary user can neither give a file away nor take one over.
    if (
      !this.privileged &&
      (!this.owns(node) || newUid !== node.uid || (newGid !== this.gid && newGid !== node.gid))
    ) {
      throw new FsError('EPERM', syscall, path);
    }
    node.uid = newUid;
    node.gid = newGid;
    node.ctimeMs = Date.now();
  }

  private stats(node: Inode): Stats {
    const type = node.kind === 'dir' ? S_IFDIR : node.kind === 'file' ? S_IFREG : S_IFLNK;
    const size =
      node.kind === 'file'
        ? node.size
        : node.kind === 'dir'
          ? 4096
          : encoder.encode(node.target).length;
    return makeStats({
      dev: 1,
      ino: node.ino,
      mode: type | node.mode,
      nlink: node.nlink,
      uid: node.uid,
      gid: node.gid,
      rdev: 0,
      size,
      blksize: 4096,
      blocks: Math.ceil(size / 512),
      atimeMs: node.atimeMs,
      mtimeMs: node.mtimeMs,
      ctimeMs: node.ctimeMs
    });
  }

  /** Standard streams look like a character device, as a terminal would. */
  private ttyStats(fd: number): Stats {
    const now = Date.now();
    return makeStats({
      dev: 2,
      ino: fd,
      mode: S_IFCHR | 0o620,
      nlink: 1,
      uid: this.uid,
      gid: this.gid,
      rdev: 0x8800 + fd,
      size: 0,
      blksize: 1024,
      blocks: 0,
      atimeMs: now,
      mtimeMs: now,
      ctimeMs: now
    });
  }

  /** Buffers runtime output and hands it on a line at a time. */
  private console(fd: number, bytes: Uint8Array): void {
    const i = fd === 2 ? 1 : 0;
    this.consoleBuf[i] += decoder.decode(bytes);
    const nl = this.consoleBuf[i].lastIndexOf('\n');
    if (nl === -1) return;
    const lines = this.consoleBuf[i].slice(0, nl);
    this.consoleBuf[i] = this.consoleBuf[i].slice(nl + 1);
    this.onConsole(fd, lines);
  }

  /** Go panics on an error without a known `code`, so anything unexpected becomes EIO. */
  private toNodeError(err: unknown): Error {
    if (err instanceof FsError) return err;
    console.error('memfs: unexpected error', err);
    const wrapped = new FsError('EIO', 'memfs');
    if (err instanceof Error) wrapped.message += `: ${err.message}`;
    return wrapped;
  }
}

function makeStats(fields: Omit<Stats, 'isDirectory' | 'isFile' | 'isSymbolicLink'>): Stats {
  const type = fields.mode & S_IFMT;
  return {
    ...fields,
    isDirectory: () => type === S_IFDIR,
    isFile: () => type === S_IFREG,
    isSymbolicLink: () => type === S_IFLNK
  };
}
