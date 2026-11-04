/**
 * Boots the real terminal.wasm in Node against the in-memory filesystem,
 * through the same bootTerminal the editor page uses: the files it seeds,
 * gnux's tools, a script, make, the compression commands, `code FILE`,
 * terminals of their own, and the format and lint checks.
 */
import assert from 'node:assert/strict';
import { existsSync, readFileSync } from 'node:fs';
import { test } from 'node:test';
import { fileURLToPath } from 'node:url';
import vm from 'node:vm';
import { createEditorFS, ENV } from '../../src/editor/seed.ts';
import { bootTerminal, type TerminalSession } from '../../src/editor/terminal-runtime.ts';

const publicDir = fileURLToPath(new URL('../../public/', import.meta.url));
const built = existsSync(`${publicDir}terminal.wasm`);

let session: TerminalSession | undefined;
let out = ['', ''];

async function boot(): Promise<TerminalSession> {
  if (session) return session;
  session = await bootTerminal({
    fs: createEditorFS(),
    env: { ...ENV, GNUX_PRETTY: 'never' },
    loadGoRuntime: async () => {
      vm.runInThisContext(readFileSync(`${publicDir}wasm_exec.js`, 'utf8'), {
        filename: 'wasm_exec.js'
      });
    },
    instantiate: async (imports) =>
      WebAssembly.instantiate(
        await WebAssembly.compile(readFileSync(`${publicDir}terminal.wasm`)),
        imports
      ),
    timeoutMs: 120_000
  });
  const decoder = [new TextDecoder(), new TextDecoder()];
  session.onOutput((fd, chunk) => {
    out[fd - 1] += decoder[fd - 1].decode(chunk, { stream: true });
  });
  return session;
}

async function run(line: string) {
  const s = await boot();
  out = ['', ''];
  const status = await s.run(line);
  return { status, stdout: out[0], stderr: out[1] };
}

test(
  'the terminal boots in the browser runtime',
  { skip: !built && 'make web/wasm first' },
  async (t) => {
    const s = await boot();

    await t.test('it seeds the repository and the starters', () => {
      assert.ok(
        s.fs.existsSync(`${s.library}/algorithms/sorting/quick.go`),
        'quick.go is under the library'
      );
      const main = new TextDecoder().decode(s.fs.readFileSync(`${s.home}/main.go`));
      assert.match(main, /package main/);
      assert.equal(s.cwd(), s.home);
    });

    await t.test('it knows gnux and the editor commands', () => {
      for (const name of ['lsx', 'grepx', 'sedx', 'go', 'bash', 'zstd']) {
        assert.ok(s.commands.includes(name), `${name} is a command`);
      }
    });

    await t.test('gnux tools run over the seeded files', async () => {
      const r = await run(`lsx ${s.library}/algorithms | grepx -c sorting`);
      assert.equal(r.status, 0, r.stderr);
      assert.equal(r.stdout.trim(), '1');
    });

    await t.test('a script runs with bash', async () => {
      const r = await run('bash hello.sh');
      assert.equal(r.status, 0, r.stderr);
      assert.match(r.stdout, /Hello from the terminal/);
      assert.match(r.stdout, /line 3/);
    });

    await t.test('files compress and come back', async () => {
      const r = await run(
        'echox round trip > note.txt && zstd note.txt && unzstd note.txt.zst && catx note.txt'
      );
      assert.equal(r.status, 0, r.stderr);
      assert.equal(r.stdout.trim(), 'round trip');
    });

    await t.test('code FILE asks the editor to open it', async () => {
      let opened = '';
      s.onOpen((p) => (opened = p));
      const r = await run('code main.go');
      assert.equal(r.status, 0, r.stderr);
      assert.equal(opened, `${s.home}/main.go`);
    });

    await t.test('go explains where programs run', async () => {
      const r = await run('go version');
      assert.equal(r.status, 0);
      assert.match(r.stdout, /Go Playground/);
    });

    await t.test('make runs the starter Makefile', async () => {
      const r = await run('make hello');
      assert.equal(r.status, 0, r.stderr);
      assert.match(r.stdout, /Hello from the terminal/);
    });

    await t.test('each terminal keeps its own folder and output', async () => {
      const got: string[] = [];
      const stop = s.listen(7, (_fd, chunk) => got.push(new TextDecoder().decode(chunk)));
      assert.equal(await s.run('cd /tmp && pwdx', 7), 0);
      stop();
      assert.equal(s.cwd(7), '/tmp');
      assert.equal(s.cwd(0), s.home, 'terminal 0 stayed home');
      assert.match(got.join(''), /\/tmp/);
      s.close(7);
      assert.equal(s.cwd(7), s.home, 'a closed terminal starts afresh');
    });

    await t.test('Go files are formatted and checked', () => {
      const f = s.format('/x/main.go', 'package main\nfunc main(){println( 1 )}\n');
      assert.equal(f.error, undefined);
      assert.equal(f.text, 'package main\n\nfunc main() { println(1) }\n');
      assert.match(s.format('/x/main.go', 'package main\nfunc {').error ?? '', /syntax error/);
      const problems = s.lint(
        '/x/main.go',
        'package main\n\nimport "os"\n\nfunc main() {\n\tx := 1\n}\n'
      );
      const messages = problems.map((d) => d.message).join('\n');
      assert.match(messages, /"os" imported and not used/);
      assert.match(messages, /declared and not used: x/);
      assert.ok(problems.every((d) => d.line >= 1));
      assert.ok(problems.some((d) => d.severity === 'error'));
    });

    await t.test('Makefiles and JSON are checked', () => {
      const make = s.lint('/x/Makefile', 'all:\n    echo spaces\n');
      assert.equal(make[0]?.source, 'make');
      assert.equal(make[0]?.line, 2);
      assert.equal(s.lint('/x/a.json', '{"a": }\n')[0]?.severity, 'error');
    });

    s.dispose();
  }
);
