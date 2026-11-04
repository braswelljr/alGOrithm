// Command dev is the development loop: it builds the algorithm command and the
// WebAssembly, starts the visualizer's dev server, and then watches the Go
// files. On every change it rebuilds both, runs the tests of the packages that
// changed, and bumps web/public/dev-build.json, which the open site
// polls to reload itself with the new wasm. Run it with `make dev`.
//
//	go run ./tools/dev [-web=false] [-test=false] [-interval 500ms]
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"syscall"
	"time"
)

// wasmSite is where the site serves the wasm and dev-build.json from.
const wasmSite = "web/public"

// skip lists the directories the watcher never looks in.
var skip = map[string]bool{".git": true, "web": true, "bin": true, "node_modules": true, ".agents": true, ".claude": true, ".idea": true, "docs": true}

func main() {
	web := flag.Bool("web", true, "start the website's dev server (needs pnpm and web/node_modules)")
	test := flag.Bool("test", true, "run the tests of the packages that changed")
	interval := flag.Duration("interval", 500*time.Millisecond, "how often to look for changes")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	build := 0
	rebuild := func(changed []string) {
		build++
		start := time.Now()
		err := buildAll()
		status := "ok"
		if err != nil {
			status = "failed"
			logf("build %d failed:\n%v", build, err)
		}
		writeBuildFile(build, err)
		if err == nil && *test && len(changed) > 0 {
			runTests(ctx, changed)
		}
		logf("build %d %s in %v", build, status, time.Since(start).Round(time.Millisecond))
	}

	rebuild(nil)
	if *web {
		if _, err := os.Stat("web/node_modules"); err != nil {
			logf("web/node_modules is missing; run `make web/install` first, or pass -web=false")
		} else if err := startSite(ctx); err != nil {
			logf("could not start the site: %v", err)
		}
	}

	logf("watching the Go files (Ctrl-C to stop)")
	last := snapshot()
	tick := time.NewTicker(*interval)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			logf("stopped")
			return
		case <-tick.C:
		}
		now := snapshot()
		changed := diff(last, now)
		if len(changed) == 0 {
			continue
		}
		// Wait for the editor to finish writing: until nothing changes for one interval.
		for {
			time.Sleep(*interval)
			again := snapshot()
			if len(diff(now, again)) == 0 {
				break
			}
			changed = append(changed, diff(now, again)...)
			now = again
		}
		last = now
		logf("changed: %s", strings.Join(short(changed), ", "))
		rebuild(changed)
	}
}

func logf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "dev %s  %s\n", time.Now().Format("15:04:05"), fmt.Sprintf(format, args...))
}

// snapshot maps each watched file to its size and modification time.
func snapshot() map[string]string {
	files := map[string]string{}
	_ = filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil //nolint:nilerr // a file that vanished mid-walk is simply not watched
		}
		if d.IsDir() {
			if path != "." && (skip[d.Name()] || strings.HasPrefix(d.Name(), ".")) {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(path, ".go") || path == "go.mod" {
			if info, err := d.Info(); err == nil {
				files[path] = fmt.Sprintf("%d %d", info.Size(), info.ModTime().UnixNano())
			}
		}
		return nil
	})
	return files
}

func diff(a, b map[string]string) []string {
	var out []string
	for p, v := range b {
		if a[p] != v {
			out = append(out, p)
		}
	}
	for p := range a {
		if _, ok := b[p]; !ok {
			out = append(out, p)
		}
	}
	sort.Strings(out)
	return out
}

func short(paths []string) []string {
	if len(paths) > 6 {
		return append(paths[:6:6], fmt.Sprintf("and %d more", len(paths)-6))
	}
	return paths
}

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %s: %w\n%s", name, strings.Join(args, " "), err, out)
	}
	return nil
}

// buildAll builds the command into bin/ and the wasm into the site's public
// folder, with wasm_exec.js beside it.
func buildAll() error {
	if err := run("go", "build", "-o", "bin/algorithm", "./cmd/algorithm"); err != nil {
		return err
	}
	if err := os.MkdirAll(wasmSite, 0o755); err != nil {
		return err
	}
	cmd := exec.Command("go", "build", "-trimpath", "-o", filepath.Join(wasmSite, "algorithm.wasm"), "./cmd/algorithm-wasm")
	cmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("wasm: %w\n%s", err, out)
	}
	execJS := filepath.Join(runtime.GOROOT(), "lib", "wasm", "wasm_exec.js")
	if _, err := os.Stat(execJS); err != nil {
		execJS = filepath.Join(runtime.GOROOT(), "misc", "wasm", "wasm_exec.js") // Go before 1.24
	}
	return copyFile(execJS, filepath.Join(wasmSite, "wasm_exec.js"))
}

func copyFile(from, to string) error {
	b, err := os.ReadFile(from)
	if err != nil {
		return err
	}
	return os.WriteFile(to, b, 0o644)
}

// writeBuildFile tells the open site that a new build is there.
func writeBuildFile(n int, err error) {
	v := map[string]any{"build": n, "ok": err == nil, "at": time.Now().Format(time.RFC3339)}
	if err != nil {
		v["error"] = err.Error()
	}
	b, _ := json.Marshal(v)
	_ = os.MkdirAll(wasmSite, 0o755)
	_ = os.WriteFile(filepath.Join(wasmSite, "dev-build.json"), b, 0o644)
}

// runTests runs the tests of the packages whose files changed, and of the
// registry, which runs every entry.
func runTests(ctx context.Context, changed []string) {
	pkgs := map[string]bool{"./internal/core": true}
	for _, f := range changed {
		if strings.HasSuffix(f, ".go") {
			pkgs["./"+filepath.ToSlash(filepath.Dir(f))] = true
		}
	}
	args := []string{"test", "-count=1"}
	for p := range pkgs {
		if p != "./." && !strings.HasPrefix(p, "./cmd/algorithm-wasm") {
			args = append(args, p)
		}
	}
	cmd := exec.CommandContext(ctx, "go", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		logf("tests failed:\n%s", out)
		return
	}
	logf("tests passed: %s", strings.Join(args[2:], " "))
}

// startSite runs the site's dev server in its own process group, stopped with
// the watcher.
func startSite(ctx context.Context) error {
	cmd := exec.Command("pnpm", "--dir", "web", "dev")
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	ownGroup(cmd)
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() {
		<-ctx.Done()
		stopGroup(cmd)
	}()
	logf("site dev server started (pid %d)", cmd.Process.Pid)
	return nil
}
