// Package tooling formats and checks the editor's files, as a code editor's
// format-on-save and problems panel do: Go with gofmt and the type checker,
// JSON, Makefiles and shell scripts with what each allows, and every text
// file for trailing spaces and a final newline. It reads no files and sends
// nothing anywhere; the editor passes the text in.
package tooling

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/scanner"
	"go/token"
	"go/types"
	"path"
	"strings"

	"github.com/braswelljr/gnux/core/shell"
)

// Severity says how much a diagnostic matters.
type Severity string

// The severities, as the editor shows them.
const (
	Error   Severity = "error"
	Warning Severity = "warning"
	Info    Severity = "info"
)

// Diagnostic is one problem found in a file. Lines and columns count from 1;
// an end of zero means the problem covers its line from Col on.
type Diagnostic struct {
	Line     int      `json:"line"`
	Col      int      `json:"col"`
	EndLine  int      `json:"endLine,omitempty"`
	EndCol   int      `json:"endCol,omitempty"`
	Severity Severity `json:"severity"`
	Message  string   `json:"message"`
	// Source names the check: gofmt, go, json, make, shell, style.
	Source string `json:"source"`
}

// Kind is the kind of file a name is, as far as formatting and checks go.
type Kind string

// The kinds of file with checks of their own.
const (
	Go    Kind = "go"
	JSON  Kind = "json"
	Make  Kind = "make"
	Shell Kind = "shell"
	Text  Kind = "text"
)

// KindOf tells a file's kind from its name.
func KindOf(name string) Kind {
	base := strings.ToLower(path.Base(name))
	switch {
	case strings.HasSuffix(base, ".go"):
		return Go
	case strings.HasSuffix(base, ".json"):
		return JSON
	case base == "makefile" || base == "gnumakefile" || strings.HasSuffix(base, ".mk"):
		return Make
	case strings.HasSuffix(base, ".sh"), strings.HasSuffix(base, ".bash"), strings.HasSuffix(base, ".gnux"):
		return Shell
	}
	return Text
}

// ErrSyntax is wrapped by Format when a file cannot be formatted because it
// does not parse; Lint says where.
var ErrSyntax = errors.New("syntax error")

// Format returns the file's text formatted: gofmt for Go, two-space indents
// for JSON, and for every kind no trailing spaces and one final newline.
func Format(name, src string) (string, error) {
	switch KindOf(name) {
	case Go:
		out, err := format.Source([]byte(src))
		if err != nil {
			return src, fmt.Errorf("%w: %v", ErrSyntax, err)
		}
		return string(out), nil
	case JSON:
		if strings.TrimSpace(src) == "" {
			return "", nil
		}
		var b bytes.Buffer
		if err := json.Indent(&b, []byte(src), "", "  "); err != nil {
			return src, fmt.Errorf("%w: %v", ErrSyntax, err)
		}
		return b.String() + "\n", nil
	}
	return tidy(src), nil
}

// tidy drops trailing spaces and ends the text with exactly one newline.
func tidy(src string) string {
	lines := strings.Split(src, "\n")
	for i, l := range lines {
		lines[i] = strings.TrimRight(l, " \t\r")
	}
	out := strings.TrimRight(strings.Join(lines, "\n"), "\n")
	if out == "" {
		return ""
	}
	return out + "\n"
}

// Lint returns the problems in a file, in order of position.
func Lint(name, src string) []Diagnostic {
	var out []Diagnostic
	switch KindOf(name) {
	case Go:
		out = lintGo(name, src)
	case JSON:
		out = lintJSON(src)
	case Make:
		out = lintMake(src)
	case Shell:
		out = lintShell(src)
	}
	return append(out, style(name, src)...)
}

func lintGo(name, src string) []Diagnostic {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, name, src, parser.AllErrors|parser.ParseComments)
	if err != nil {
		var list scanner.ErrorList
		if errors.As(err, &list) {
			out := make([]Diagnostic, 0, len(list))
			for _, e := range list {
				out = append(out, Diagnostic{Line: e.Pos.Line, Col: e.Pos.Column, Severity: Error, Message: e.Msg, Source: "go"})
			}
			return out
		}
		return []Diagnostic{{Line: 1, Col: 1, Severity: Error, Message: err.Error(), Source: "go"}}
	}

	// The names the file's imports go by, which a stand-in package (named
	// after its path's last element by the checker) may not answer to.
	imported := map[string]bool{}
	for _, spec := range f.Imports {
		imported[importName(strings.Trim(spec.Path.Value, "`\""))] = true
		if spec.Name != nil {
			imported[spec.Name.Name] = true
		}
	}
	var out []Diagnostic
	// Type-check the file alone. Imports are not available in the browser, so
	// each one is a stand-in package: the checker then stays quiet about what
	// is used from it, and still finds undefined names, unused variables and
	// imports, and mistakes in the file's own types.
	conf := types.Config{
		Importer: stubImporter{},
		Error: func(err error) {
			var te types.Error
			if !errors.As(err, &te) || strings.Contains(te.Msg, "could not import") ||
				imported[strings.TrimPrefix(te.Msg, "undefined: ")] {
				return
			}
			p := fset.Position(te.Pos)
			sev := Error
			if te.Soft {
				sev = Warning
			}
			out = append(out, Diagnostic{Line: p.Line, Col: p.Column, Severity: sev, Message: te.Msg, Source: "go"})
		},
	}
	_, _ = conf.Check(f.Name.Name, fset, []*ast.File{f}, nil)
	out = append(out, unusedImports(fset, f)...)

	if formatted, err := format.Source([]byte(src)); err == nil && string(formatted) != src {
		line := firstDifference(src, string(formatted))
		out = append(out, Diagnostic{Line: line, Col: 1, Severity: Warning, Message: "not formatted with gofmt (format the file to fix)", Source: "gofmt"})
	}
	return out
}

// unusedImports finds imports the file never refers to. The type checker
// cannot say so for an import it could not load, so this reads the syntax: an
// import's name is its own, or the last element of its path (without a major
// version suffix such as /v2).
func unusedImports(fset *token.FileSet, f *ast.File) []Diagnostic {
	used := map[string]bool{}
	ast.Inspect(f, func(n ast.Node) bool {
		if sel, ok := n.(*ast.SelectorExpr); ok {
			if id, ok := sel.X.(*ast.Ident); ok {
				used[id.Name] = true
			}
		}
		return true
	})
	var out []Diagnostic
	for _, spec := range f.Imports {
		p := strings.Trim(spec.Path.Value, "`\"")
		name := importName(p)
		if spec.Name != nil {
			name = spec.Name.Name
		}
		if name == "_" || name == "." || used[name] {
			continue
		}
		pos := fset.Position(spec.Pos())
		out = append(out, Diagnostic{Line: pos.Line, Col: pos.Column, Severity: Error, Message: fmt.Sprintf("%q imported and not used", p), Source: "go"})
	}
	return out
}

func importName(p string) string {
	parts := strings.Split(p, "/")
	last := parts[len(parts)-1]
	if len(parts) > 1 && len(last) > 1 && last[0] == 'v' && strings.Trim(last[1:], "0123456789") == "" {
		last = parts[len(parts)-2]
	}
	if i := strings.IndexAny(last, ".-"); i > 0 {
		last = last[:i]
	}
	return last
}

// stubImporter answers every import with an empty, incomplete package, which
// the type checker treats as unknown rather than wrong.
type stubImporter struct{}

func (stubImporter) Import(p string) (*types.Package, error) {
	return nil, fmt.Errorf("could not import %s", p)
}

func firstDifference(a, b string) int {
	al, bl := strings.Split(a, "\n"), strings.Split(b, "\n")
	for i := range min(len(al), len(bl)) {
		if al[i] != bl[i] {
			return i + 1
		}
	}
	return min(len(al), len(bl))
}

func lintJSON(src string) []Diagnostic {
	if strings.TrimSpace(src) == "" {
		return nil
	}
	var v any
	err := json.Unmarshal([]byte(src), &v)
	if err == nil {
		return nil
	}
	line, col := 1, 1
	var se *json.SyntaxError
	if errors.As(err, &se) {
		line, col = position(src, int(se.Offset))
	}
	return []Diagnostic{{Line: line, Col: col, Severity: Error, Message: err.Error(), Source: "json"}}
}

// position turns a byte offset into a line and column.
func position(src string, offset int) (line, col int) {
	offset = max(0, min(offset, len(src)))
	before := src[:offset]
	line = strings.Count(before, "\n") + 1
	col = offset - strings.LastIndex(before, "\n")
	return line, col
}

// lintMake finds what make itself refuses: a recipe indented with spaces.
func lintMake(src string) []Diagnostic {
	var out []Diagnostic
	inRule := false
	for i, l := range strings.Split(src, "\n") {
		trimmed := strings.TrimSpace(l)
		switch {
		case strings.HasPrefix(l, "\t"):
			if !inRule && trimmed != "" && !strings.HasPrefix(trimmed, "#") {
				out = append(out, Diagnostic{Line: i + 1, Col: 1, Severity: Error, Message: "recipe line before any rule", Source: "make"})
			}
		case trimmed == "" || strings.HasPrefix(trimmed, "#"):
		case strings.HasPrefix(l, " ") && inRule:
			out = append(out, Diagnostic{Line: i + 1, Col: 1, Severity: Error, Message: "a recipe line starts with a tab, not spaces (make: missing separator)", Source: "make"})
		default:
			inRule = isRule(trimmed)
		}
	}
	return out
}

// isRule says whether a Makefile line opens a rule: `targets: prerequisites`,
// not an assignment (`=`, `:=`, `?=`, `+=`, `::=`).
func isRule(line string) bool {
	colon := strings.Index(line, ":")
	if colon < 0 {
		return false
	}
	if eq := strings.Index(line, "="); eq >= 0 && eq < colon+2 {
		return false
	}
	return true
}

// lintShell reports a script that is not finished: an if, loop, case, quote
// or brace left open. gnux's shell does not export its parser, so this is what
// can be checked without running the script.
func lintShell(src string) []Diagnostic {
	if strings.TrimSpace(src) == "" || !shell.Incomplete(src) {
		return nil
	}
	lines := strings.Count(strings.TrimRight(src, "\n"), "\n") + 1
	return []Diagnostic{{Line: lines, Col: 1, Severity: Error, Message: "the script is not finished: an if, loop, case, quote or brace is not closed", Source: "shell"}}
}

// style finds trailing spaces and a missing final newline, in every file.
func style(name, src string) []Diagnostic {
	if src == "" {
		return nil
	}
	var out []Diagnostic
	lines := strings.Split(src, "\n")
	for i, l := range lines {
		t := strings.TrimRight(l, " \t\r")
		if len(t) < len(l) {
			out = append(out, Diagnostic{Line: i + 1, Col: len(t) + 1, EndLine: i + 1, EndCol: len(l) + 1, Severity: Info, Message: "trailing whitespace", Source: "style"})
		}
	}
	if !strings.HasSuffix(src, "\n") && KindOf(name) != JSON {
		out = append(out, Diagnostic{Line: len(lines), Col: len(lines[len(lines)-1]) + 1, Severity: Info, Message: "no newline at the end of the file", Source: "style"})
	}
	return out
}
