package filesystems

import (
	"fmt"
	"sort"
	"strings"
)

// INode is one inode in a Namespace run.
type INode struct {
	Number  int            `json:"number"`
	Kind    string         `json:"kind"` // "dir", "file" or "symlink"
	Links   int            `json:"links"`
	Entries map[string]int `json:"entries,omitempty"` // a directory's names
	Data    string         `json:"data,omitempty"`    // a file's text, or a symlink's target path
}

// NamespaceResult is the log of a Namespace or Resolve run and the inodes
// still allocated at the end.
type NamespaceResult struct {
	Log    []string `json:"log"`
	Inodes []INode  `json:"inodes"`
}

type namespace struct {
	inodes map[int]*INode
	next   int
	log    []string
}

func newNamespace() *namespace {
	root := &INode{Number: 2, Kind: "dir", Links: 2, Entries: map[string]int{".": 2, "..": 2}}
	return &namespace{inodes: map[int]*INode{2: root}, next: 3}
}

// resolve walks path from the root, following symbolic links (at most 8, or
// ELOOP), and logs each lookup when verbose. With parent it stops at the
// last component and returns the directory and the final name.
func (ns *namespace) resolve(path string, parent, verbose bool, depth int) (*INode, string, error) {
	if depth > 8 {
		return nil, "", inputErr("%s: too many levels of symbolic links (ELOOP)", path)
	}
	parts := strings.FieldsFunc(path, func(r rune) bool { return r == '/' })
	cur := ns.inodes[2]
	if verbose {
		ns.log = append(ns.log, "  start at the root directory, inode 2")
	}
	for i, name := range parts {
		last := i == len(parts)-1
		if last && parent {
			return cur, name, nil
		}
		if cur.Kind != "dir" {
			return nil, "", inputErr("%s: not a directory (ENOTDIR)", path)
		}
		num, ok := cur.Entries[name]
		if !ok {
			return nil, "", inputErr("%s: no such file or directory (ENOENT)", path)
		}
		n := ns.inodes[num]
		if verbose {
			ns.log = append(ns.log, fmt.Sprintf("  look up %q in directory inode %d: inode %d (%s)", name, cur.Number, num, n.Kind))
		}
		if n.Kind == "symlink" {
			if verbose {
				ns.log = append(ns.log, fmt.Sprintf("  inode %d is a symbolic link to %q: resolve that instead", num, n.Data))
			}
			target, _, err := ns.resolve(n.Data, false, verbose, depth+1)
			if err != nil {
				return nil, "", err
			}
			n = target
		}
		cur = n
	}
	if parent {
		return nil, "", inputErr("%s: a path with a name is needed", path)
	}
	return cur, "", nil
}

func (ns *namespace) add(kind, data string) *INode {
	n := &INode{Number: ns.next, Kind: kind, Data: data}
	if kind == "dir" {
		n.Entries = map[string]int{}
	}
	ns.inodes[ns.next] = n
	ns.next++
	return n
}

// run applies one command.
func (ns *namespace) run(f []string) error {
	arg := func(i int) string {
		if i < len(f) {
			return f[i]
		}
		return ""
	}
	switch f[0] {
	case "mkdir", "create", "symlink":
		path := arg(1)
		if f[0] == "symlink" {
			path = arg(2)
		}
		dir, name, err := ns.resolve(path, true, false, 0)
		if err != nil {
			return err
		}
		if _, exists := dir.Entries[name]; exists {
			return inputErr("%s: file exists (EEXIST)", path)
		}
		var n *INode
		switch f[0] {
		case "mkdir":
			n = ns.add("dir", "")
			n.Entries["."], n.Entries[".."] = n.Number, dir.Number
			n.Links = 2
			dir.Links++ // the new ".." points at it
		case "create":
			n = ns.add("file", strings.Join(f[2:], " "))
			n.Links = 1
		default:
			n = ns.add("symlink", arg(1))
			n.Links = 1
		}
		dir.Entries[name] = n.Number
		ns.log = append(ns.log, fmt.Sprintf("%s: inode %d (%s), entry %q in directory %d", strings.Join(f, " "), n.Number, n.Kind, name, dir.Number))
	case "link":
		target, _, err := ns.resolve(arg(1), false, false, 0)
		if err != nil {
			return err
		}
		if target.Kind == "dir" {
			return inputErr("%s: hard links to directories are not allowed (EPERM)", arg(1))
		}
		dir, name, err := ns.resolve(arg(2), true, false, 0)
		if err != nil {
			return err
		}
		dir.Entries[name] = target.Number
		target.Links++
		ns.log = append(ns.log, fmt.Sprintf("link %s %s: a second name for inode %d, link count now %d", arg(1), arg(2), target.Number, target.Links))
	case "rm":
		dir, name, err := ns.resolve(arg(1), true, false, 0)
		if err != nil {
			return err
		}
		num, ok := dir.Entries[name]
		if !ok {
			return inputErr("%s: no such file (ENOENT)", arg(1))
		}
		n := ns.inodes[num]
		if n.Kind == "dir" {
			return inputErr("%s: is a directory", arg(1))
		}
		delete(dir.Entries, name)
		n.Links--
		msg := fmt.Sprintf("rm %s: entry removed, inode %d link count now %d", arg(1), num, n.Links)
		if n.Links == 0 {
			delete(ns.inodes, num)
			msg += ": no names left, so the inode and its blocks are freed"
		}
		ns.log = append(ns.log, msg)
	case "cat", "resolve":
		ns.log = append(ns.log, strings.Join(f, " "))
		n, _, err := ns.resolve(arg(1), false, f[0] == "resolve", 0)
		if err != nil {
			ns.log = append(ns.log, "  "+strings.TrimPrefix(err.Error(), ErrInput.Error()+": "))
			return nil
		}
		if f[0] == "cat" {
			if n.Kind != "file" {
				ns.log = append(ns.log, "  is a directory")
				return nil
			}
			ns.log = append(ns.log, fmt.Sprintf("  %q (inode %d)", n.Data, n.Number))
		} else {
			ns.log = append(ns.log, fmt.Sprintf("  resolved to inode %d (%s)", n.Number, n.Kind))
		}
	default:
		return inputErr("unknown command %q (mkdir, create, link, symlink, rm, cat, resolve)", f[0])
	}
	return nil
}

func (ns *namespace) result() NamespaceResult {
	var nums []int
	for n := range ns.inodes {
		nums = append(nums, n)
	}
	sort.Ints(nums)
	res := NamespaceResult{Log: ns.log}
	for _, n := range nums {
		res.Inodes = append(res.Inodes, *ns.inodes[n])
	}
	return res
}

// Namespace runs a script of file-system commands, one a line:
//
//	mkdir PATH          make a directory
//	create PATH TEXT    make a file holding TEXT
//	link OLD NEW        a hard link: a second name for the same inode
//	symlink TARGET NEW  a symbolic link: a small file holding a path
//	rm PATH             remove a name; the inode goes when no names are left
//	cat PATH            read a file, following symbolic links
//
// A hard link keeps the file alive after the original name is removed; a
// symbolic link only remembers the path, so it dangles. The root directory
// is inode 2, as on ext4.
//
// Time: O(n * d) for n commands on paths d deep. Space: O(inodes).
func Namespace(script string) (NamespaceResult, error) {
	ns := newNamespace()
	calls, nums := lines(script)
	for k, f := range calls {
		if err := ns.run(f); err != nil {
			return ns.result(), fmt.Errorf("line %d: %w", nums[k], err)
		}
	}
	return ns.result(), nil
}

// Resolve sets up the namespace with script (as in Namespace) and then
// resolves path, logging every directory lookup and symbolic link it
// follows: the work the kernel does on every open. Each component costs a
// directory read, which is why kernels cache them.
//
// Time: O(n * d) for n commands and a path d deep. Space: O(inodes).
func Resolve(script, path string) (NamespaceResult, error) {
	ns := newNamespace()
	calls, nums := lines(script)
	for k, f := range calls {
		if err := ns.run(f); err != nil {
			return ns.result(), fmt.Errorf("line %d: %w", nums[k], err)
		}
	}
	ns.log = append(ns.log, "")
	if err := ns.run([]string{"resolve", path}); err != nil {
		return ns.result(), err
	}
	return ns.result(), nil
}
