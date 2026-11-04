package core

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path"
	"regexp"
	"strings"
	"testing"
)

// TestLayout keeps the registry one file per domain: no file is a numbered,
// "more" or "split" leftover, and each category's <category>Entries function
// lives in its domain's file, <domain>.go or <domain>_<category>.go (the
// structures domain calls its own category "linear").
func TestLayout(t *testing.T) {
	bad := regexp.MustCompile(`more|[0-9]\.go$|_split`)
	files, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		if bad.MatchString(f.Name()) {
			t.Errorf("%s: name entries by domain and category, not as more, numbered or split files", f.Name())
		}
	}
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, ".", func(fi os.FileInfo) bool { return !strings.HasSuffix(fi.Name(), "_test.go") }, 0)
	if err != nil {
		t.Fatal(err)
	}
	fileOf := map[string]string{}
	for name, f := range pkgs["core"].Files {
		for _, d := range f.Decls {
			if fd, ok := d.(*ast.FuncDecl); ok && fd.Recv == nil {
				fileOf[fd.Name.Name] = path.Base(name)
			}
		}
	}
	for _, c := range Categories {
		domain := strings.SplitN(strings.TrimPrefix(c.Import, Module+"/pkg/"), "/", 2)[0]
		own := c.ID
		if c.ID == domain {
			own = "linear"
		}
		got, ok := fileOf[c.ID+"Entries"]
		if !ok {
			t.Errorf("category %s has no %sEntries function", c.ID, c.ID)
			continue
		}
		if got != domain+".go" && got != domain+"_"+own+".go" {
			t.Errorf("%sEntries is in %s, not %s.go or %s_%s.go", c.ID, got, domain, domain, own)
		}
	}
}
