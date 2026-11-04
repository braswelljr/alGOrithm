package core

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/structures/trees"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// treesEntries is every entry of the trees category.
func treesEntries() []Entry {
	return []Entry{
		buildTreeFromTraversalsEntry(),
		expressionTreeEntry(),
		invertTreeEntry(),
		lowestCommonAncestorEntry(),
		morrisTraversalEntry(),
		pathSumEntry(),
		serializeTreeEntry(),
		treeHeightDiameterEntry(),
		treeInorderEntry(),
		treeLevelOrderEntry(),
		treePostorderEntry(),
		treePreorderEntry(),
		validateBstEntry(),
	}
}

// treeSetup is the snippet's setup for an entry that parses a tree.
func treeSetup(text string) string {
	return fmt.Sprintf("t, err := trees.Parse(%q)\n\tif err != nil {\n\t\tpanic(err)\n\t}\n\t", text)
}

// visitSteps turns a traced walk into one step per visited node, the node
// "visit" and the ones before it "visited".
func visitSteps(t *trees.Tree, events []trace.Event, values []int) []Step {
	base := t.Snapshot()
	steps := []Step{{Op: Op{Name: "start"}, Result: fmt.Sprintf("%d nodes", len(base.Nodes)), Snapshot: base}}
	seen := map[int]bool{}
	byID := map[int]int{}
	for i, n := range base.Nodes {
		byID[n.ID] = i
	}
	n := 0
	for _, e := range events {
		if e.Kind != trace.Visit {
			continue
		}
		id := e.Values[0]
		s := trace.Snapshot{Layout: base.Layout, Edges: base.Edges, Nodes: append([]trace.Node(nil), base.Nodes...)}
		for j := range s.Nodes {
			if seen[s.Nodes[j].ID] {
				s.Nodes[j].Role = "visited"
			}
		}
		s.Nodes[byID[id]].Role = "visit"
		seen[id] = true
		res := fmt.Sprintf("visit %s", base.Nodes[byID[id]].Label)
		if n < len(values) {
			res = fmt.Sprintf("visit %d (output so far: %s)", values[n], joinInts(values[:n+1]))
		}
		n++
		steps = append(steps, Step{Op: Op{Name: "visit"}, Result: res, Snapshot: s})
	}
	return steps
}

func parseTree(in Input) (*trees.Tree, error) {
	t, err := trees.Parse(in.Text)
	if err != nil {
		return nil, wrap(err)
	}
	if t.Root == nil {
		return nil, inputError("give a tree in level order, like \"1 2 3 null 4\"")
	}
	return t, nil
}

const treesSample = "4 2 6 1 3 5 7"

func treesWalk(id, name, fn, summary, space string, traced func(*trees.Tree, trace.Recorder) []int) Entry {
	file := "structures/trees/traversal.go"
	e := entry(id, name, file, fn)
	e.Summary, e.Time, e.Space, e.Input, e.Traced = summary, "O(n)", space, Text, true
	e.Example = &Input{Text: "1 2 3 4 5 null 6"}
	e.Run = func(in Input, r trace.Recorder) (Output, error) {
		t, err := parseTree(in)
		if err != nil {
			return Output{}, err
		}
		var log trace.Log
		vals := traced(t, &log)
		for _, ev := range log.Events {
			r.Record(ev)
		}
		return Output{Values: vals, Steps: visitSteps(t, log.Events, vals), Lines: []string{"order: " + joinInts(vals)}}, nil
	}
	e.Call = func(in Input) (string, string) { return treeSetup(in.Text), "trees." + fn + "(t)" }
	return e
}

func treesSingle(id, name, file, fn, summary, time, space, example string, run func(t *trees.Tree, in Input) (Output, error), call func(in Input) string) Entry {
	e := entry(id, name, file, fn)
	e.Summary, e.Time, e.Space, e.Input = summary, time, space, Text
	e.Example = &Input{Text: example, Options: map[string]int{}}
	e.Run = func(in Input, _ trace.Recorder) (Output, error) {
		t, err := parseTree(in)
		if err != nil {
			return Output{}, err
		}
		out, err := run(t, in)
		if err != nil {
			return Output{}, err
		}
		if out.Steps == nil {
			out.Steps = []Step{{Op: Op{Name: "tree"}, Result: trees.Serialize(t), Snapshot: t.Snapshot()}}
		}
		return out, nil
	}
	e.Call = func(in Input) (string, string) { return treeSetup(in.Text), call(in) }
	return e
}

func treesRoles(t *trees.Tree, role string, values ...int) trace.Snapshot {
	s := t.Snapshot()
	want := map[int]bool{}
	for _, v := range values {
		if n := t.Find(v); n != nil {
			want[n.ID] = true
		}
	}
	for i := range s.Nodes {
		if want[s.Nodes[i].ID] {
			s.Nodes[i].Role = role
		}
	}
	return s
}

func treeLevelOrderEntry() Entry {
	level := entry("tree-level-order", "Level-Order Traversal", "structures/trees/traversal.go", "LevelOrder")
	level.Summary, level.Time, level.Space, level.Input, level.Traced = "Visit the tree level by level, left to right, with a queue: breadth-first search on a tree.", "O(n)", "O(w) for the widest level", Text, true
	level.Example = &Input{Text: "1 2 3 4 5 null 6"}
	level.Run = func(in Input, r trace.Recorder) (Output, error) {
		t, err := parseTree(in)
		if err != nil {
			return Output{}, err
		}
		var log trace.Log
		levels := trees.TraceLevelOrder(t, &log)
		var flat []int
		var lines []string
		for i, l := range levels {
			flat = append(flat, l...)
			lines = append(lines, fmt.Sprintf("level %d: %s", i, joinInts(l)))
		}
		for _, ev := range log.Events {
			r.Record(ev)
		}
		return Output{Values: flat, Lines: lines, Steps: visitSteps(t, log.Events, flat)}, nil
	}
	level.Call = func(in Input) (string, string) { return treeSetup(in.Text), "trees.LevelOrder(t)" }
	return level
}

func morrisTraversalEntry() Entry {
	morris := entry("morris-traversal", "Morris Inorder Traversal", "structures/trees/traversal.go", "Morris")
	morris.Summary, morris.Time, morris.Space, morris.Input, morris.Traced = "Inorder with no stack: thread each node's predecessor back to it, follow the threads, then remove them (Morris, 1979).", "O(n)", "O(1)", Text, true
	morris.Example = &Input{Text: treesSample}
	morris.Run = func(in Input, r trace.Recorder) (Output, error) {
		t, err := parseTree(in)
		if err != nil {
			return Output{}, err
		}
		base := t.Snapshot()
		var log trace.Log
		vals := trees.TraceMorris(t, &log)
		steps := []Step{{Op: Op{Name: "start"}, Result: "no stack: threads instead", Snapshot: base}}
		threads := map[[2]int]bool{}
		out := 0
		for _, ev := range log.Events {
			r.Record(ev)
			s := trace.Snapshot{Layout: base.Layout, Nodes: append([]trace.Node(nil), base.Nodes...), Edges: append([]trace.Edge(nil), base.Edges...)}
			res := ""
			switch ev.Kind {
			case trace.Link:
				threads[[2]int{ev.Values[0], ev.Values[1]}] = true
				res = "thread the predecessor back to its successor"
			case trace.Unlink:
				delete(threads, [2]int{ev.Values[0], ev.Values[1]})
				res = "remove the thread: the left subtree is done"
			case trace.Visit:
				res = fmt.Sprintf("output %d", vals[out])
				out++
				for j := range s.Nodes {
					if s.Nodes[j].ID == ev.Values[0] {
						s.Nodes[j].Role = "visit"
					}
				}
			default: // Morris reports only links, unlinks and visits
			}
			ths := make([][2]int, 0, len(threads))
			for th := range threads {
				ths = append(ths, th)
			}
			sort.Slice(ths, func(i, j int) bool { return ths[i][0] < ths[j][0] || ths[i][0] == ths[j][0] && ths[i][1] < ths[j][1] })
			for _, th := range ths { // sorted, so every run draws the same snapshot
				s.Edges = append(s.Edges, trace.Edge{From: th[0], To: th[1], Label: "thread"})
			}
			steps = append(steps, Step{Op: Op{Name: ev.Kind.String()}, Result: res, Snapshot: s})
		}
		return Output{Values: vals, Steps: steps, Lines: []string{"inorder: " + joinInts(vals)}}, nil
	}
	morris.Call = func(in Input) (string, string) { return treeSetup(in.Text), "trees.Morris(t)" }
	return morris
}

func treeHeightDiameterEntry() Entry {
	measure := treesSingle("tree-height-diameter", "Height, Diameter and Balance", "structures/trees/measure.go", "Diameter",
		"The tree's height, its diameter (the longest path between any two nodes, in edges) and whether it is balanced.", "O(n)", "O(h)", "1 2 3 4 5 null null 6 null null 7",
		func(t *trees.Tree, _ Input) (Output, error) {
			b := trees.IsBalanced(t)
			h := trees.Height(t)
			return Output{Number: &h, Bool: &b, Summary: map[string]float64{"height": float64(h), "diameter": float64(trees.Diameter(t)), "nodes": float64(t.Size())}}, nil
		}, func(Input) string { return "trees.Height(t), trees.Diameter(t), trees.IsBalanced(t)" })
	return measure
}

func validateBstEntry() Entry {
	bst := treesSingle("validate-bst", "Validate a Binary Search Tree", "structures/trees/measure.go", "IsBST",
		"Whether every node is greater than its whole left subtree and less than its whole right one, by passing down the allowed range.", "O(n)", "O(h)", "10 5 15 null 12",
		func(t *trees.Tree, _ Input) (Output, error) {
			ok := trees.IsBST(t)
			line := "a binary search tree"
			if !ok {
				line = "not a binary search tree (check whole subtrees, not just children)"
			}
			return Output{Bool: &ok, Lines: []string{line, "inorder: " + joinInts(trees.Inorder(t))}}, nil
		}, func(Input) string { return "trees.IsBST(t)" })
	return bst
}

func lowestCommonAncestorEntry() Entry {
	lca := treesSingle("lowest-common-ancestor", "Lowest Common Ancestor", "structures/trees/lca.go", "LowestCommonAncestor",
		"The deepest node with both a and b below it (or being one of them), from asking each subtree whether it holds a or b.", "O(n)", "O(h)", "3 5 1 6 2 0 8 null null 7 4",
		func(t *trees.Tree, in Input) (Output, error) {
			a, b := in.opt("a", 7), in.opt("b", 4)
			n, ok := trees.LowestCommonAncestor(t, a, b)
			if !ok {
				return Output{}, inputError(fmt.Sprintf("%d or %d is not in the tree", a, b))
			}
			s := treesRoles(t, "visited", a, b)
			for i := range s.Nodes {
				if s.Nodes[i].ID == n.ID {
					s.Nodes[i].Role = "visit"
				}
			}
			return Output{Number: ptr(n.Value), Lines: []string{fmt.Sprintf("the lowest common ancestor of %d and %d is %d", a, b, n.Value)},
				Steps: []Step{{Op: Op{Name: "lca"}, Result: fmt.Sprintf("%d", n.Value), Snapshot: s}}}, nil
		}, func(in Input) string {
			return fmt.Sprintf("trees.LowestCommonAncestor(t, %d, %d)", in.opt("a", 7), in.opt("b", 4))
		})
	lca.Options = []Option{{"a", "first value", 7, -1 << 30, 1 << 30}, {"b", "second value", 4, -1 << 30, 1 << 30}}
	lca.Example.Options = map[string]int{"a": 7, "b": 4}
	return lca
}

func serializeTreeEntry() Entry {
	ser := treesSingle("serialize-tree", "Serialize and Deserialize", "structures/trees/build.go", "Serialize",
		"Write a tree as its level order with null for missing children, and read it back: the same tree, node for node.", "O(n)", "O(n)", "[1, 2, 3, null, null, 4, 5]",
		func(t *trees.Tree, _ Input) (Output, error) {
			s := trees.Serialize(t)
			back, _ := trees.Parse(s)
			return Output{Text: s, Bool: ptr(trees.Serialize(back) == s), Lines: []string{"serialized: " + s, "deserialized and serialized again: " + trees.Serialize(back)}}, nil
		}, func(Input) string { return "trees.Serialize(t)" })
	return ser
}

func pathSumEntry() Entry {
	path := treesSingle("path-sum", "Path Sum", "structures/trees/build.go", "PathSum",
		"Every root-to-leaf path whose values add up to a target, by depth-first search keeping the current path.", "O(n * h)", "O(h)", "5 4 8 11 null 13 4 7 2 null null 5 1",
		func(t *trees.Tree, in Input) (Output, error) {
			paths := trees.PathSum(t, in.Target)
			var lines []string
			var on []int
			for _, p := range paths {
				lines = append(lines, joinInts(p))
				on = append(on, p...)
			}
			if len(lines) == 0 {
				lines = []string{fmt.Sprintf("no root-to-leaf path sums to %d", in.Target)}
			}
			return Output{Number: ptr(len(paths)), Lines: lines, Steps: []Step{{Op: Op{Name: "paths"}, Result: fmt.Sprintf("%d path(s)", len(paths)), Snapshot: treesRoles(t, "visited", on...)}}}, nil
		}, func(in Input) string { return fmt.Sprintf("trees.PathSum(t, %d)", in.Target) })
	path.Params = []string{"target"}
	path.Example.Target = 22
	return path
}

func invertTreeEntry() Entry {
	invert := treesSingle("invert-tree", "Invert a Tree", "structures/trees/build.go", "Invert",
		"Mirror the tree by swapping every node's children.", "O(n)", "O(h)", treesSample,
		func(t *trees.Tree, _ Input) (Output, error) {
			before := t.Snapshot()
			trees.Invert(t)
			return Output{Text: trees.Serialize(t), Steps: []Step{
				{Op: Op{Name: "before"}, Result: "the tree", Snapshot: before},
				{Op: Op{Name: "invert"}, Result: "every node's children swapped: " + trees.Serialize(t), Snapshot: t.Snapshot()},
			}}, nil
		}, func(Input) string { return "trees.Serialize(trees.Invert(t))" })
	return invert
}

func buildTreeFromTraversalsEntry() Entry {
	build := entry("build-tree-from-traversals", "Build a Tree from Two Traversals", "structures/trees/build.go", "BuildFromPreIn")
	build.Summary, build.Time, build.Space, build.Input = "Rebuild a tree from its preorder and inorder values: the first preorder value is the root, and it splits the inorder list in two.", "O(n)", "O(n)", TwoTexts
	build.Example = &Input{Text: "3 9 20 15 7", Other: "9 3 15 20 7"}
	build.Run = func(in Input, _ trace.Recorder) (Output, error) {
		pre, err1 := treeInts(in.Text)
		ino, err2 := treeInts(in.Other)
		if err1 != nil || err2 != nil {
			return Output{}, inputError("give the preorder and the inorder values as whole numbers")
		}
		t, err := trees.BuildFromPreIn(pre, ino)
		if err != nil {
			return Output{}, wrap(err)
		}
		return Output{Text: trees.Serialize(t), Lines: []string{"level order: " + trees.Serialize(t), "postorder: " + joinInts(trees.Postorder(t))},
			Steps: []Step{{Op: Op{Name: "build"}, Result: trees.Serialize(t), Snapshot: t.Snapshot()}}}, nil
	}
	build.Call = func(in Input) (string, string) {
		pre, _ := treeInts(in.Text)
		ino, _ := treeInts(in.Other)
		return fmt.Sprintf("t, err := trees.BuildFromPreIn(%s, %s)\n\tif err != nil {\n\t\tpanic(err)\n\t}\n\t", ints(pre), ints(ino)), "trees.Serialize(t)"
	}
	return build
}

func expressionTreeEntry() Entry {
	expr := entry("expression-tree", "Expression Tree", "structures/trees/expression.go", "ParseExpr")
	expr.Summary, expr.Time, expr.Space, expr.Input = "Parse arithmetic into a tree with the shunting-yard algorithm, evaluate it in postorder, and print it in prefix and postfix.", "O(n)", "O(n)", Text
	expr.Example = &Input{Text: "(3 + 4) * 2 - 10 / 5"}
	expr.Run = func(in Input, _ trace.Recorder) (Output, error) {
		if len(in.Text) > 10_000 {
			return Output{}, inputError("at most 10000 characters")
		}
		e, err := trees.ParseExpr(in.Text)
		if err != nil {
			return Output{}, wrap(err)
		}
		v, err := e.Eval()
		if err != nil {
			return Output{}, wrap(err)
		}
		return Output{Number: &v, Lines: []string{"prefix:  " + e.Prefix(), "postfix: " + e.Postfix(), "infix:   " + e.Infix(), fmt.Sprintf("value:   %d", v)},
			Steps: []Step{{Op: Op{Name: "parse"}, Result: e.Infix(), Snapshot: e.Snapshot()}}}, nil
	}
	expr.Call = func(in Input) (string, string) {
		return fmt.Sprintf("e, err := trees.ParseExpr(%q)\n\tif err != nil {\n\t\tpanic(err)\n\t}\n\t", in.Text), "e.Eval()"
	}
	return expr
}

func treePreorderEntry() Entry {
	return treesWalk("tree-preorder", "Preorder Traversal", "Preorder", "Visit the root, then the left subtree, then the right: the order to copy a tree.", "O(h)", trees.TracePreorder)
}

func treeInorderEntry() Entry {
	return treesWalk("tree-inorder", "Inorder Traversal", "Inorder", "Visit the left subtree, then the root, then the right: ascending order in a binary search tree.", "O(h)", trees.TraceInorder)
}

func treePostorderEntry() Entry {
	return treesWalk("tree-postorder", "Postorder Traversal", "Postorder", "Visit both subtrees, then the root: the order to delete a tree or evaluate an expression.", "O(h)", trees.TracePostorder)
}

func treeInts(s string) ([]int, error) {
	var out []int
	for _, f := range strings.FieldsFunc(s, func(r rune) bool { return r == ' ' || r == ',' || r == '\n' }) {
		v, err := strconv.Atoi(f)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, nil
}
