// Package makefile is a small make for the editor's terminal: it reads a
// Makefile and brings targets up to date by running their recipes in the
// shell, as GNU make does for the common case.
//
// It knows rules (`targets: prerequisites`, with recipes indented by a tab or
// after a `;`), .PHONY, variables (`=`, `:=`, `::=`, `?=`, `+=`, from the
// command line and from the environment), the automatic variables $@, $<, $^
// and $?, `$(wildcard ...)`, the @, - and + recipe prefixes, and -f, -C, -n
// and -s. It does not know pattern rules, conditionals, include, or the other
// functions; a line it cannot read is an error with its line number, never a
// silent skip.
package makefile

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/braswelljr/gnux/core"
)

// Eval runs one recipe line in a shell and returns its status.
type Eval func(line string, s core.Streams) int

// Run is the make command: make [-f FILE] [-C DIR] [-n] [-s] [VAR=value]... [TARGET]...
func Run(args []string, s core.Streams, eval Eval) int {
	m := &maker{streams: s, eval: eval, cmdline: map[string]string{}}
	var file, dir string
	var goals []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		next := func() (string, bool) {
			if i+1 >= len(args) {
				fmt.Fprintf(s.Err, "make: option requires an argument -- '%s'\n", strings.TrimLeft(a, "-"))
				return "", false
			}
			i++
			return args[i], true
		}
		switch {
		case a == "-f" || a == "--file":
			v, ok := next()
			if !ok {
				return 2
			}
			file = v
		case a == "-C" || a == "--directory":
			v, ok := next()
			if !ok {
				return 2
			}
			dir = v
		case a == "-n" || a == "--dry-run" || a == "--just-print":
			m.dry = true
		case a == "-s" || a == "--silent" || a == "--quiet":
			m.silent = true
		case a == "-h" || a == "--help":
			fmt.Fprint(s.Out, usage)
			return 0
		case strings.HasPrefix(a, "-j"), a == "-B", a == "--always-make":
			if a == "-B" || a == "--always-make" {
				m.always = true
			}
		case strings.HasPrefix(a, "-"):
			fmt.Fprintf(s.Err, "make: unknown option %s\n", a)
			return 2
		case strings.Contains(a, "="):
			k, v, _ := strings.Cut(a, "=")
			m.cmdline[strings.TrimSpace(k)] = v
		default:
			goals = append(goals, a)
		}
	}

	if dir != "" {
		back, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(s.Err, "make: %v\n", err)
			return 2
		}
		if err := os.Chdir(dir); err != nil {
			fmt.Fprintf(s.Err, "make: *** %v.  Stop.\n", err)
			return 2
		}
		wd, _ := os.Getwd()
		fmt.Fprintf(s.Out, "make: Entering directory '%s'\n", wd)
		defer func() {
			fmt.Fprintf(s.Out, "make: Leaving directory '%s'\n", wd)
			_ = os.Chdir(back)
		}()
	}

	if file == "" {
		// Exact names, as make looks for them, on a case-insensitive disk too.
		var names []string
		if entries, err := os.ReadDir("."); err == nil {
			for _, e := range entries {
				names = append(names, e.Name())
			}
		}
		for _, name := range []string{"GNUmakefile", "makefile", "Makefile"} {
			if slices.Contains(names, name) {
				file = name
				break
			}
		}
		if file == "" {
			if len(goals) == 0 {
				fmt.Fprintln(s.Err, "make: *** No targets specified and no makefile found.  Stop.")
			} else {
				fmt.Fprintf(s.Err, "make: *** No rule to make target '%s'.  Stop.\n", goals[0])
			}
			return 2
		}
	}
	f, err := os.Open(file)
	if err != nil {
		fmt.Fprintf(s.Err, "make: %v\n", err)
		return 2
	}
	err = m.parse(file, f)
	_ = f.Close()
	if err != nil {
		fmt.Fprintf(s.Err, "make: %v\n", err)
		return 2
	}

	if len(goals) == 0 {
		if m.first == "" {
			fmt.Fprintln(s.Err, "make: *** No targets.  Stop.")
			return 2
		}
		goals = []string{m.first}
	}
	for _, g := range goals {
		remade, err := m.build(g, nil)
		if errors.Is(err, errFailed) {
			return 2
		}
		if err != nil {
			fmt.Fprintf(s.Err, "make: *** %v\n", err)
			return 2
		}
		if !remade {
			if r := m.rules[g]; (r != nil && len(r.recipe) == 0) || m.phony[g] {
				fmt.Fprintf(s.Out, "make: Nothing to be done for '%s'.\n", g)
			} else {
				fmt.Fprintf(s.Out, "make: '%s' is up to date.\n", g)
			}
		}
	}
	return 0
}

const usage = `Usage: make [options] [VAR=value]... [target]...
  -f FILE   read FILE as the makefile
  -C DIR    change to DIR first
  -n        print the recipes without running them
  -s        do not echo the recipes
  -B        remake every target
`

type line struct {
	text string
	at   int
}

type rule struct {
	prereqs []string
	recipe  []line
	at      int
}

type variable struct {
	value     string
	recursive bool
}

type maker struct {
	streams core.Streams
	eval    Eval
	file    string
	cmdline map[string]string
	vars    map[string]variable
	rules   map[string]*rule
	phony   map[string]bool
	first   string
	made    map[string]bool
	dry     bool
	silent  bool
	always  bool
}

func (m *maker) parse(name string, r io.Reader) error {
	m.file = name
	m.vars = map[string]variable{}
	m.rules = map[string]*rule{}
	m.phony = map[string]bool{}
	m.made = map[string]bool{}

	var current []*rule
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 64*1024), 4*1024*1024)
	n := 0
	for sc.Scan() {
		n++
		text, at := sc.Text(), n
		for strings.HasSuffix(text, "\\") && sc.Scan() {
			n++
			text = strings.TrimSuffix(text, "\\") + " " + strings.TrimLeft(sc.Text(), " \t")
		}
		if strings.HasPrefix(text, "\t") {
			if len(current) == 0 {
				if strings.TrimSpace(text) == "" || strings.HasPrefix(strings.TrimSpace(text), "#") {
					continue
				}
				return fmt.Errorf("%s:%d: *** recipe commences before first target.  Stop", name, at)
			}
			for _, r := range current {
				r.recipe = append(r.recipe, line{text[1:], at})
			}
			continue
		}
		if i := strings.Index(text, "#"); i >= 0 {
			text = text[:i]
		}
		trimmed := strings.TrimSpace(text)
		if trimmed == "" {
			continue
		}
		if k, op, v, ok := assignment(trimmed); ok {
			m.assign(m.expand(k, nil), op, v)
			current = nil
			continue
		}
		colon := strings.Index(trimmed, ":")
		if colon < 0 {
			return fmt.Errorf("%s:%d: *** missing separator.  Stop", name, at)
		}
		targets := strings.Fields(m.expand(trimmed[:colon], nil))
		rest := strings.TrimPrefix(trimmed[colon+1:], ":")
		var inline string
		if semi := strings.Index(rest, ";"); semi >= 0 {
			rest, inline = rest[:semi], strings.TrimSpace(rest[semi+1:])
		}
		prereqs := strings.Fields(m.expand(rest, nil))
		current = nil
		for _, t := range targets {
			if t == ".PHONY" {
				for _, p := range prereqs {
					m.phony[p] = true
				}
				continue
			}
			if strings.HasPrefix(t, ".") {
				continue // .SUFFIXES, .DEFAULT_GOAL and the like
			}
			if strings.Contains(t, "%") {
				return fmt.Errorf("%s:%d: *** pattern rules are not supported here.  Stop", name, at)
			}
			r := m.rules[t]
			if r == nil {
				r = &rule{at: at}
				m.rules[t] = r
				if m.first == "" {
					m.first = t
				}
			}
			r.prereqs = append(r.prereqs, prereqs...)
			if inline != "" {
				r.recipe = append(r.recipe, line{inline, at})
			}
			current = append(current, r)
		}
	}
	return sc.Err()
}

// assignment splits `NAME op value`, where op is =, :=, ::=, ?= or +=.
func assignment(s string) (name, op, value string, ok bool) {
	eq := strings.Index(s, "=")
	if eq <= 0 {
		return "", "", "", false
	}
	op = "="
	k := s[:eq]
	switch {
	case strings.HasSuffix(k, "::"):
		op, k = "::=", k[:len(k)-2]
	case strings.HasSuffix(k, ":"):
		op, k = ":=", k[:len(k)-1]
	case strings.HasSuffix(k, "?"):
		op, k = "?=", k[:len(k)-1]
	case strings.HasSuffix(k, "+"):
		op, k = "+=", k[:len(k)-1]
	}
	if strings.ContainsAny(k, ":") {
		return "", "", "", false // a rule whose prerequisites hold an =
	}
	k = strings.TrimSpace(k)
	if k == "" || strings.ContainsAny(k, " \t") {
		return "", "", "", false
	}
	return k, op, strings.TrimSpace(s[eq+1:]), true
}

func (m *maker) assign(name, op, value string) {
	if _, set := m.cmdline[name]; set {
		return // the command line wins
	}
	switch op {
	case "=":
		m.vars[name] = variable{value, true}
	case ":=", "::=":
		m.vars[name] = variable{m.expand(value, nil), false}
	case "?=":
		if _, ok := m.vars[name]; !ok && os.Getenv(name) == "" {
			m.vars[name] = variable{value, true}
		}
	case "+=":
		v, ok := m.vars[name]
		if !ok {
			m.vars[name] = variable{value, true}
			return
		}
		if !v.recursive {
			value = m.expand(value, nil)
		}
		if v.value != "" {
			value = v.value + " " + value
		}
		m.vars[name] = variable{value, v.recursive}
	}
}

func (m *maker) lookup(name string, auto map[string]string, depth int) string {
	if v, ok := auto[name]; ok {
		return v
	}
	if v, ok := m.cmdline[name]; ok {
		return v
	}
	if v, ok := m.vars[name]; ok {
		if v.recursive && depth < 32 {
			return m.expandDepth(v.value, auto, depth+1)
		}
		return v.value
	}
	return os.Getenv(name)
}

func (m *maker) expand(s string, auto map[string]string) string {
	return m.expandDepth(s, auto, 0)
}

// expandDepth replaces $(NAME), ${NAME}, $X and $$ in s.
func (m *maker) expandDepth(s string, auto map[string]string, depth int) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c != '$' || i+1 >= len(s) {
			b.WriteByte(c)
			continue
		}
		i++
		switch open := s[i]; open {
		case '$':
			b.WriteByte('$')
		case '(', '{':
			closeCh := byte(')')
			if open == '{' {
				closeCh = '}'
			}
			end, level := -1, 0
			for j := i + 1; j < len(s); j++ {
				switch s[j] {
				case open:
					level++
				case closeCh:
					if level == 0 {
						end = j
					}
					level--
				}
				if end >= 0 {
					break
				}
			}
			if end < 0 {
				b.WriteString(s[i-1:])
				return b.String()
			}
			inner := m.expandDepth(s[i+1:end], auto, depth)
			b.WriteString(m.reference(inner, auto, depth))
			i = end
		default:
			b.WriteString(m.lookup(string(open), auto, depth))
		}
	}
	return b.String()
}

// reference is the value of one $(...): a variable, or a function it knows.
func (m *maker) reference(inner string, auto map[string]string, depth int) string {
	if fn, arg, ok := strings.Cut(inner, " "); ok && fn == "wildcard" {
		var out []string
		for _, pat := range strings.Fields(arg) {
			got, _ := filepath.Glob(pat)
			out = append(out, got...)
		}
		return strings.Join(out, " ")
	}
	return m.lookup(strings.TrimSpace(inner), auto, depth)
}

func mtime(path string) (time.Time, bool) {
	st, err := os.Stat(path)
	if err != nil {
		return time.Time{}, false
	}
	return st.ModTime(), true
}

// stop is an error worded as GNU make words it.
type stop string

func (s stop) Error() string { return string(s) }

// errFailed is a recipe that failed; its message is already printed.
var errFailed = errors.New("failed")

// build brings target up to date and says whether it ran anything for it.
func (m *maker) build(target string, chain []string) (bool, error) {
	if slices.Contains(chain, target) {
		fmt.Fprintf(m.streams.Err, "make: Circular %s <- %s dependency dropped.\n", chain[len(chain)-1], target)
		return false, nil
	}
	if done, ok := m.made[target]; ok {
		return done, nil
	}
	r := m.rules[target]
	if r == nil {
		if _, ok := mtime(target); ok {
			m.made[target] = false
			return false, nil
		}
		if len(chain) == 0 {
			return false, stop(fmt.Sprintf("No rule to make target '%s'.  Stop.", target))
		}
		return false, stop(fmt.Sprintf("No rule to make target '%s', needed by '%s'.  Stop.", target, chain[len(chain)-1]))
	}

	chain = append(chain, target)
	stale := m.always || m.phony[target]
	when, exists := mtime(target)
	if !exists {
		stale = true
	}
	var newer []string
	for _, p := range r.prereqs {
		remade, err := m.build(p, chain)
		if err != nil {
			return false, err
		}
		pt, ok := mtime(p)
		if remade || !ok || !exists || pt.After(when) {
			stale = true
			newer = append(newer, p)
		}
	}
	if !stale || len(r.recipe) == 0 {
		m.made[target] = false
		return false, nil
	}

	first := ""
	if len(r.prereqs) > 0 {
		first = r.prereqs[0]
	}
	auto := map[string]string{
		"@": target,
		"<": first,
		"^": strings.Join(dedupe(r.prereqs), " "),
		"?": strings.Join(dedupe(newer), " "),
	}
	for _, l := range r.recipe {
		text := m.expand(l.text, auto)
		quiet, ignore := m.silent, false
		for len(text) > 0 && strings.ContainsRune("@-+ \t", rune(text[0])) {
			switch text[0] {
			case '@':
				quiet = true
			case '-':
				ignore = true
			}
			text = text[1:]
		}
		if strings.TrimSpace(text) == "" {
			continue
		}
		if !quiet || m.dry {
			fmt.Fprintln(m.streams.Out, text)
		}
		if m.dry {
			continue
		}
		back, _ := os.Getwd()
		status := m.eval(text, m.streams)
		if back != "" {
			_ = os.Chdir(back)
		}
		if status != 0 {
			if ignore {
				fmt.Fprintf(m.streams.Err, "make: [%s:%d: %s] Error %d (ignored)\n", m.file, l.at, target, status)
				continue
			}
			fmt.Fprintf(m.streams.Err, "make: *** [%s:%d: %s] Error %d\n", m.file, l.at, target, status)
			return false, errFailed
		}
	}
	m.made[target] = true
	return true, nil
}

func dedupe(list []string) []string {
	seen := map[string]bool{}
	out := list[:0:0]
	for _, s := range list {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}
