/**
 * A file's icon in the explorer and the tabs, as a code editor shows them: the
 * language's mark in its own color, and a plain glyph for the rest (archives,
 * images, licenses, text). The icon is decoration; the name beside it says
 * what the file is.
 */
import type { IconType } from 'react-icons';
import {
  LuBookOpen,
  LuDatabase,
  LuFile,
  LuFileArchive,
  LuFileCode,
  LuFileDiff,
  LuFileImage,
  LuFileLock,
  LuFileText,
  LuFolder,
  LuFolderOpen,
  LuScale
} from 'react-icons/lu';
import {
  SiCss,
  SiDocker,
  SiDotenv,
  SiEditorconfig,
  SiGit,
  SiGnubash,
  SiGo,
  SiHtml5,
  SiJavascript,
  SiJson,
  SiLua,
  SiMake,
  SiMarkdown,
  SiPython,
  SiRuby,
  SiRust,
  SiSvg,
  SiToml,
  SiTypescript,
  SiXml,
  SiYaml
} from 'react-icons/si';
import { languageOf, type Language } from '@/editor/files';
import { cn } from '@/lib/utils';

type Look = { icon: IconType; color: string };

const BY_LANGUAGE: Partial<Record<Language, Look>> = {
  go: { icon: SiGo, color: 'text-[#00ADD8]' },
  shell: { icon: SiGnubash, color: 'text-emerald-500' },
  make: { icon: SiMake, color: 'text-orange-500' },
  markdown: { icon: SiMarkdown, color: 'text-sky-500' },
  json: { icon: SiJson, color: 'text-amber-500' },
  yaml: { icon: SiYaml, color: 'text-rose-500' },
  toml: { icon: SiToml, color: 'text-stone-500' },
  dockerfile: { icon: SiDocker, color: 'text-[#2496ED]' },
  javascript: { icon: SiJavascript, color: 'text-yellow-500' },
  typescript: { icon: SiTypescript, color: 'text-[#3178C6]' },
  css: { icon: SiCss, color: 'text-indigo-500' },
  html: { icon: SiHtml5, color: 'text-orange-600' },
  xml: { icon: SiXml, color: 'text-orange-400' },
  python: { icon: SiPython, color: 'text-[#3776AB]' },
  rust: { icon: SiRust, color: 'text-orange-700' },
  sql: { icon: LuDatabase, color: 'text-cyan-600' },
  diff: { icon: LuFileDiff, color: 'text-emerald-600' },
  protobuf: { icon: LuFileCode, color: 'text-teal-500' },
  lua: { icon: SiLua, color: 'text-blue-700' },
  ruby: { icon: SiRuby, color: 'text-red-600' }
};

const ARCHIVE = /\.(gz|zst|zip|tar|tgz|bz2|xz|7z)$/;
const IMAGE = /\.(png|jpe?g|gif|webp|ico|bmp|avif)$/;

/** How a file looks, by its name first, then its language. */
export function fileLook(path: string): Look {
  const name = path.slice(path.lastIndexOf('/') + 1).toLowerCase();
  if (name === 'go.mod' || name === 'go.sum') return { icon: SiGo, color: 'text-fuchsia-500' };
  if (name.startsWith('readme')) return { icon: LuBookOpen, color: 'text-sky-500' };
  if (name.startsWith('license') || name.startsWith('copying'))
    return { icon: LuScale, color: 'text-amber-600' };
  if (name === '.gitignore' || name === '.gitattributes')
    return { icon: SiGit, color: 'text-[#F05032]' };
  if (name === '.env' || name.startsWith('.env.'))
    return { icon: SiDotenv, color: 'text-yellow-600' };
  if (name === '.editorconfig') return { icon: SiEditorconfig, color: 'text-muted' };
  if (name.endsWith('.svg')) return { icon: SiSvg, color: 'text-amber-500' };
  if (name.endsWith('.lock')) return { icon: LuFileLock, color: 'text-muted' };
  if (ARCHIVE.test(name)) return { icon: LuFileArchive, color: 'text-violet-500' };
  if (IMAGE.test(name)) return { icon: LuFileImage, color: 'text-pink-500' };
  if (name.endsWith('_test.go')) return { icon: SiGo, color: 'text-emerald-500' };
  const look = BY_LANGUAGE[languageOf(path)];
  if (look) return look;
  return name.endsWith('.txt') || name.endsWith('.log')
    ? { icon: LuFileText, color: 'text-muted' }
    : { icon: LuFile, color: 'text-muted' };
}

export function FileIcon({ path, className }: { path: string; className?: string }) {
  const { icon: Icon, color } = fileLook(path);
  return (
    <Icon
      aria-hidden
      className={cn('size-4 shrink-0', color, className)}
    />
  );
}

export function FolderIcon({ open, className }: { open: boolean; className?: string }) {
  const Icon = open ? LuFolderOpen : LuFolder;
  return (
    <Icon
      aria-hidden
      className={cn('text-warning size-4 shrink-0', className)}
    />
  );
}
