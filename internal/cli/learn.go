package cli

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/braswelljr/alGOrithm/internal/core"
)

// runLearn lists the learning paths, or prints one path's steps with the
// command that runs each.
func runLearn(args []string, s Streams) error {
	o := newOptions("learn")
	rest, err := o.parse(args)
	if err != nil {
		return err
	}
	switch len(rest) {
	case 0:
		return listPaths(s, o.asJSON)
	case 1:
		p, ok := core.PathByID(rest[0])
		if !ok {
			return usagef("no learning path %q; run 'algorithm learn' to list them", rest[0])
		}
		if o.asJSON {
			return writeJSON(s.Out, p.View())
		}
		return printPath(s, p)
	}
	return usagef("learn takes at most one path id")
}

func listPaths(s Streams, asJSON bool) error {
	paths := core.Paths()
	if asJSON {
		out := make([]core.PathView, len(paths))
		for i, p := range paths {
			out[i] = p.View()
		}
		return writeJSON(s.Out, out)
	}
	w := tabwriter.NewWriter(s.Out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tLEVEL\tSTEPS\tTITLE")
	for _, p := range paths {
		fmt.Fprintf(w, "%s\t%s\t%d\t%s\n", p.ID, p.Level, len(p.Steps), p.Title)
	}
	if err := w.Flush(); err != nil {
		return err
	}
	_, err := fmt.Fprintln(s.Out, "\nRun 'algorithm learn ID' to see a path's steps.")
	return err
}

func printPath(s Streams, p core.Path) error {
	fmt.Fprintf(s.Out, "%s (%s; %s)\n%s\n", p.Title, p.Level, p.Audience, p.Summary)
	if len(p.Prerequisites) > 0 {
		fmt.Fprintf(s.Out, "First: %s\n", strings.Join(p.Prerequisites, ", "))
	}
	if p.Guide != "" {
		fmt.Fprintf(s.Out, "Read: %s\n", p.Guide)
	}
	for i, st := range p.View().Steps {
		fmt.Fprintf(s.Out, "\n%2d. %s\n    %s\n    $ %s\n", i+1, st.Name, st.Note, st.Command)
	}
	return nil
}
