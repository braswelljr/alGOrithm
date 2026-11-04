//go:build js && wasm

// Command terminal is the WebAssembly build behind the editor's terminal: a
// gnux shell with every gnux tool and the commands in package terminal. Once
// started it installs an API on the JavaScript global object and waits.
//
//	algorithmTerminal.commands   string[] - every command the shell knows
//	algorithmTerminal.home       string - the home directory
//	algorithmTerminal.library    string - where the repository's packages are
//	algorithmTerminal.run(line, id?)  Promise<number> - run one line in terminal id (0 if
//	                                  left out; made on first use), resolve its status
//	algorithmTerminal.cwd(id?)        string - that terminal's working directory
//	algorithmTerminal.close(id)       forget a terminal (its variables and directory)
//	algorithmTerminal.setenv(name, value)
//	algorithmTerminal.format(path, text) {text, error} - the file formatted (package tooling)
//	algorithmTerminal.lint(path, text)   string - the problems in the file, as JSON
//	algorithmTerminal.themes     string[] - gnux's terminal themes
//	algorithmTerminal.onOutput = (fd, chunk, id) => ...  1 stdout, 2 stderr
//	algorithmTerminal.onOpen = (path) => ...             `code FILE` asked to open path
//
// Each terminal is a shell of its own, with its own variables and directory;
// the environment is shared. Lines run one at a time across all of them.
//
// Files live in the Node-style `fs` the page supplies (an in-memory one), so
// nothing touches the visitor's disk; only `go run` sends anything out, to the
// Go Playground, and it says so. When ready it calls algorithmTerminalReady.
//
// Build with `make web/wasm`.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"syscall"
	"syscall/js"

	"github.com/braswelljr/gnux/core"
	"github.com/braswelljr/gnux/core/shell"

	"github.com/braswelljr/alGOrithm/web/runtime/terminal"
	"github.com/braswelljr/alGOrithm/web/runtime/tooling"
)

// outputBudget caps what one line may print: the browser runs Go on its only
// thread, so an endless writer would otherwise freeze the page.
const outputBudget = 1 << 20

type output struct {
	api       js.Value
	id        int
	remaining int
	truncated bool
}

type writer struct {
	out *output
	fd  int
}

// IsTerminal says the scrollback is a terminal, so tools draw gnux's view.
func (writer) IsTerminal() bool { return true }

func (w writer) Write(p []byte) (int, error) {
	if w.out.remaining <= 0 {
		w.out.truncated = true
		return 0, syscall.EPIPE
	}
	n := min(len(p), w.out.remaining)
	w.out.remaining -= n
	w.out.emit(w.fd, p[:n])
	if n < len(p) {
		w.out.truncated = true
		return n, syscall.EPIPE
	}
	return n, nil
}

func (o *output) emit(fd int, p []byte) {
	fn := o.api.Get("onOutput")
	if fn.Type() != js.TypeFunction {
		return
	}
	chunk := js.Global().Get("Uint8Array").New(len(p))
	js.CopyBytesToJS(chunk, p)
	fn.Invoke(fd, chunk, o.id)
}

func strs(list []string) js.Value {
	out := make([]any, len(list))
	for i, s := range list {
		out[i] = s
	}
	return js.ValueOf(out)
}

// session is one terminal: a shell, where its output goes, and its directory.
type session struct {
	sh  *shell.Shell
	out *output
	wd  string
}

// idOf reads an optional terminal id argument.
func idOf(args []js.Value, i int) int {
	if len(args) > i && args[i].Type() == js.TypeNumber {
		return args[i].Int()
	}
	return 0
}

func main() {
	api := js.Global().Get("Object").New()
	_ = os.Setenv("HOME", terminal.Home)
	_ = os.Setenv("GOPATH", terminal.GoPath)
	if err := terminal.Seed(); err != nil {
		fmt.Fprintln(os.Stderr, "terminal: seeding the files:", err)
	}
	_ = os.Chdir(terminal.Home)

	open := func(path string) {
		if fn := api.Get("onOpen"); fn.Type() == js.TypeFunction {
			fn.Invoke(path)
		}
	}
	exec := terminal.Exec(open)
	sessions := map[int]*session{}
	get := func(id int) *session {
		if s, ok := sessions[id]; ok {
			return s
		}
		out := &output{api: api, id: id}
		sh := shell.New(
			// Interactive: a syntax error must not end the session.
			shell.Options{Interactive: true, Pretty: true, Color: true},
			exec,
			core.Streams{In: strings.NewReader(""), Out: writer{out, 1}, Err: writer{out, 2}},
		)
		sh.LoopLimit = 10000
		s := &session{sh: sh, out: out, wd: terminal.Home}
		sessions[id] = s
		return s
	}

	api.Set("commands", strs(terminal.Commands()))
	api.Set("home", terminal.Home)
	api.Set("library", terminal.Library)
	api.Set("themes", strs(core.ThemeNames()))
	api.Set("setenv", js.FuncOf(func(_ js.Value, args []js.Value) any {
		if len(args) < 2 {
			return false
		}
		if args[1].String() == "" {
			return os.Unsetenv(args[0].String()) == nil
		}
		return os.Setenv(args[0].String(), args[1].String()) == nil
	}))

	// One line at a time: the process has one working directory, so each
	// terminal's is put back before its line runs and kept after.
	var mu sync.Mutex
	api.Set("cwd", js.FuncOf(func(_ js.Value, args []js.Value) any {
		mu.Lock()
		defer mu.Unlock()
		return get(idOf(args, 0)).wd
	}))
	api.Set("close", js.FuncOf(func(_ js.Value, args []js.Value) any {
		mu.Lock()
		defer mu.Unlock()
		delete(sessions, idOf(args, 0))
		return nil
	}))
	api.Set("run", js.FuncOf(func(_ js.Value, args []js.Value) any {
		line := ""
		if len(args) > 0 {
			line = args[0].String()
		}
		id := idOf(args, 1)
		// Filesystem and network calls complete on the JavaScript event loop,
		// so the work runs on a goroutine while this returns a Promise.
		executor := js.FuncOf(func(_ js.Value, settle []js.Value) any {
			resolve := settle[0]
			go func() {
				mu.Lock()
				defer mu.Unlock()
				s := get(id)
				if err := os.Chdir(s.wd); err != nil {
					_ = os.Chdir(terminal.Home)
				}
				s.out.remaining, s.out.truncated = outputBudget, false
				status := s.sh.Eval(line)
				if s.out.truncated {
					s.out.emit(2, fmt.Appendf(nil, "terminal: output truncated after %d bytes\n", outputBudget))
				}
				if wd, err := os.Getwd(); err == nil {
					s.wd = wd
				}
				resolve.Invoke(status)
			}()
			return nil
		})
		defer executor.Release()
		return js.Global().Get("Promise").New(executor)
	}))

	// Formatting and checking read no files, so they answer at once.
	api.Set("format", js.FuncOf(func(_ js.Value, args []js.Value) any {
		if len(args) < 2 {
			return nil
		}
		text, err := tooling.Format(args[0].String(), args[1].String())
		res := map[string]any{"text": text}
		if err != nil {
			res["error"] = err.Error()
		}
		return js.ValueOf(res)
	}))
	api.Set("lint", js.FuncOf(func(_ js.Value, args []js.Value) any {
		if len(args) < 2 {
			return "[]"
		}
		b, err := json.Marshal(tooling.Lint(args[0].String(), args[1].String()))
		if err != nil {
			return "[]"
		}
		return string(b)
	}))

	js.Global().Set("algorithmTerminal", api)
	if ready := js.Global().Get("algorithmTerminalReady"); ready.Type() == js.TypeFunction {
		ready.Invoke(api)
	}
	select {}
}
