// Package cli is the algorithm command: run any entry of the registry on your
// own data, see every step, compare entries, read their source, and serve them
// to other programs (HTTP and MCP).
package cli

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"

	"github.com/braswelljr/alGOrithm/internal/version"
)

// Streams are the command's input and outputs. Only cmd/algorithm passes the
// process's own.
type Streams struct {
	In       io.Reader
	Out, Err io.Writer
}

// Exit statuses.
const (
	ExitOK      = 0
	ExitFailure = 1 // the run failed, or its input was rejected
	ExitUsage   = 2 // the command line was wrong
)

type usageError string

func (e usageError) Error() string { return string(e) }

func usagef(format string, args ...any) error { return usageError(fmt.Sprintf(format, args...)) }

// command is one subcommand: its name, a line for the help, and what it does.
type command struct {
	name, help string
	run        func(args []string, s Streams) error
}

func commands() []command {
	return []command{
		{"list", "list [CATEGORY]      every entry, or one category's (--tag T, --kind K, --level L to filter)", runList},
		{"info", "info ID              what an entry does, its cost, input and options", runInfo},
		{"tags", "tags                 every tag, what it means, and how many entries carry it", runTags},
		{"learn", "learn [PATH]         learning paths: ordered entries to run, with what to notice", runLearn},
		{"run", "run ID [VALUE...]    run an entry on your data and print the result", runRun},
		{"compare", "compare [ID...]      run several entries on the same data (default: every sort)", runCompare},
		{"bench", "bench ID             measure an entry at growing sizes and fit its growth (--sizes, --input)", runBench},
		{"quiz", "quiz ID [VALUE...]   predict what an entry does next, checked against the run (--count, --seed, --json)", runQuiz},
		{"source", "source ID            print the entry's Go source and doc comment", runSource},
		{"snippet", "snippet ID [...]     print the Go program and the command for a run", runSnippet},
		{"data", "data                 print the values the data options give", runData},
		{"serve", "serve [--addr A]     serve the registry as a JSON API over HTTP", runServe},
		{"mcp", "mcp                  serve the registry as MCP tools on standard input and output", runMCP},
	}
}

// Main runs the command line (argv[0] included) and returns the exit status.
func Main(argv []string, s Streams) int {
	err := dispatch(argv[1:], s)
	var ue usageError
	switch {
	case err == nil, errors.Is(err, flag.ErrHelp):
		return ExitOK
	case errors.As(err, &ue):
		fmt.Fprintf(s.Err, "%s: %s\nTry '%s help' for more information.\n", version.AppName, err, version.AppName)
		return ExitUsage
	}
	fmt.Fprintf(s.Err, "%s: %s\n", version.AppName, err)
	return ExitFailure
}

func dispatch(args []string, s Streams) error {
	if len(args) == 0 {
		return help(s)
	}
	switch args[0] {
	case "help", "-h", "--help":
		return help(s)
	case "version", "--version":
		_, err := fmt.Fprintf(s.Out, "%s %s\n", version.AppName, version.Get())
		return err
	}
	for _, c := range commands() {
		if c.name == args[0] {
			return c.run(args[1:], s)
		}
	}
	return usagef("unknown command %q", args[0])
}

func help(s Streams) error {
	fmt.Fprintf(s.Out, "Usage: %s COMMAND [ARGS]\n\nRun the alGOrithm algorithms and data structures on your own data.\n\nCommands:\n", version.AppName)
	for _, c := range commands() {
		fmt.Fprintf(s.Out, "  %s\n", c.help)
	}
	fmt.Fprintf(s.Out, "  version              print the version\n\n%s", flagsHelp)
	return nil
}

func writeJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
