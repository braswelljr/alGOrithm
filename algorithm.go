// Package algorithm is the root of alGOrithm, a collection of algorithms and
// data structures in Go to read, import, run on your own data, and watch step
// by step. The code is in one package per topic, grouped by domain under pkg/:
//
//	algorithms/   sorting, searching, pathfinding, graphs, dynamic, greedy,
//	              backtracking, strmatch, text, recursion, compression
//	math/         numbers, numerical, geometry, bitwise, randomized
//	security/     ciphers, cryptography, hashing
//	structures/   stacks, queues and lists, with arrays, sets, trees,
//	              collections and the other structure packages under it
//	os/           scheduling, memory, disk, deadlock, concurrency, cache
//	ml/           clustering, regression, classification, neural
//	systems/      networking, databases, distributed
//	trace/        the events and snapshots every traced algorithm reports
//	dataset/      real data in: text, CSV, JSON, and seeded generators
//
// This package itself only carries the source of those packages (Source,
// File, Doc, PackageDoc), for the command and the visualizer, which show it;
// importing a topic package does not import this one.
package algorithm

import (
	"embed"
	"io/fs"
	"path"
	"strings"
	"sync"
)

//go:embed pkg
var embedded embed.FS

// source is the embedded tree with the "pkg/" prefix dropped, so paths read
// "sorting/quick.go" whatever the layout of the module.
var source = mustSub(embedded, "pkg")

// Source is the Go source of every topic package, test files included, at
// paths like "algorithms/sorting/quick.go".
var Source = source

//go:embed examples/*/main.go
var examples embed.FS

// Example returns the source of a program in examples/, named by its
// directory ("examples/leaderboard" or "leaderboard").
func Example(dir string) (string, bool) {
	dir = strings.TrimPrefix(strings.TrimSuffix(dir, "/"), "examples/")
	if dir == "" || strings.Contains(dir, "/") {
		return "", false
	}
	b, err := fs.ReadFile(examples, "examples/"+dir+"/main.go")
	return string(b), err == nil
}

// File returns the source of one file ("algorithms/sorting/quick.go").
func File(path string) (string, bool) {
	b, err := fs.ReadFile(source, path)
	return string(b), err == nil
}

// Doc returns the doc comment of the top-level func or type called name in
// the file at path, or of a method when name is "Type.Method", without the
// comment markers, or "" when there is none.
func Doc(path, name string) string {
	code, ok := File(path)
	if !ok {
		return ""
	}
	if typ, method, isMethod := strings.Cut(name, "."); isMethod {
		return docBefore(code, func(line string) bool {
			recv, rest, ok := strings.Cut(strings.TrimPrefix(line, "func ("), ") ")
			return ok && strings.HasPrefix(line, "func (") && strings.Contains(recv, typ) &&
				(strings.HasPrefix(rest, method+"(") || strings.HasPrefix(rest, method+"["))
		})
	}
	return docBefore(code, func(line string) bool {
		return strings.HasPrefix(line, "func "+name+"(") || strings.HasPrefix(line, "func "+name+"[") ||
			strings.HasPrefix(line, "type "+name+" ") || strings.HasPrefix(line, "type "+name+"[")
	})
}

// PackageDoc returns the package comment of a topic package, named by its
// directory ("algorithms/sorting") or by its name alone ("sorting").
func PackageDoc(pkg string) string {
	dir := Dir(pkg)
	entries, err := fs.ReadDir(source, dir)
	if err != nil {
		return ""
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		code, _ := File(dir + "/" + e.Name())
		if d := docBefore(code, func(line string) bool { return strings.HasPrefix(line, "package ") }); d != "" {
			return d
		}
	}
	return ""
}

// docBefore returns the // comment block directly above the first line that
// matches.
func docBefore(code string, match func(string) bool) string {
	lines := strings.Split(code, "\n")
	for i, l := range lines {
		if !match(l) {
			continue
		}
		j := i
		for j > 0 && strings.HasPrefix(lines[j-1], "//") {
			j--
		}
		var b strings.Builder
		for _, c := range lines[j:i] {
			c = strings.TrimPrefix(c, "//")
			c = strings.TrimPrefix(c, " ")
			if strings.HasPrefix(c, "go:") {
				continue
			}
			b.WriteString(c)
			b.WriteByte('\n')
		}
		return strings.TrimSpace(b.String())
	}
	return ""
}

// Dir returns the directory of the package named pkg ("sorting" gives
// "algorithms/sorting"), pkg itself when it is already a directory, or "" when
// there is no such package. Package names are unique across pkg/.
func Dir(pkg string) string {
	if strings.Contains(pkg, "/") {
		return pkg
	}
	return dirs()[pkg]
}

// dirs maps every package name under pkg/ to its directory, built once.
var dirs = sync.OnceValue(func() map[string]string {
	m := map[string]string{}
	_ = fs.WalkDir(source, ".", func(p string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.HasSuffix(p, ".go") {
			m[path.Base(path.Dir(p))] = path.Dir(p)
		}
		return nil
	})
	return m
})

// mustSub is fs.Sub for a directory the embed directive guarantees exists.
func mustSub(fsys fs.FS, dir string) fs.FS {
	sub, err := fs.Sub(fsys, dir)
	if err != nil {
		panic(err)
	}
	return sub
}
