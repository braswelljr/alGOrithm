package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/braswelljr/alGOrithm/internal/core"
)

// runQuiz asks what an entry does next, one question at a time, reading the
// answers (a choice's number) from the input stream. With --json it prints
// the questions instead, for another program to ask.
func runQuiz(args []string, s Streams) error {
	o := newOptions("quiz")
	rest, err := o.parse(args)
	if err != nil {
		return err
	}
	e, err := lookupArg(rest)
	if err != nil {
		return err
	}
	if len(rest) == 1 && o.values == "" && o.file == "" && o.generate == "" && o.text == "" && o.ops == "" {
		o.example = true // no data given: quiz on the entry's example
	}
	in, err := o.input(e, rest[1:], s)
	if err != nil {
		return err
	}
	qs, err := core.Quiz(e.ID, in, o.count, o.seed)
	if err != nil {
		return err
	}
	if o.asJSON {
		return writeJSON(s.Out, qs)
	}
	if len(qs) == 0 {
		fmt.Fprintf(s.Out, "There is nothing to ask about %s on this input.\n", e.Name)
		return nil
	}
	return askAll(qs, e, s)
}

// askAll asks each question, keeping a score, a streak and a row of dots
// for progress. It stops early, with the score so far, when the input ends.
func askAll(qs []core.Question, e core.Entry, s Streams) error {
	p := newPaint(s.Out)
	fmt.Fprintf(s.Out, "%s %d questions. Answer with the number of a choice.\n", p.bold(e.Name+":"), len(qs))
	sc := bufio.NewScanner(s.In)
	var g quizGame
	for i, q := range qs {
		fmt.Fprintf(s.Out, "\n%s  %s\n", g.dots(len(qs), p), p.dim(fmt.Sprintf("question %d of %d, at %s %d", i+1, len(qs), q.Unit, q.Step)))
		if q.Context != "" {
			fmt.Fprintln(s.Out, p.dim(q.Context))
		}
		fmt.Fprintln(s.Out, p.bold(q.Prompt))
		for j, c := range q.Choices {
			fmt.Fprintf(s.Out, "  %d) %s\n", j+1, c)
		}
		pick, ok := readChoice(sc, len(q.Choices), s)
		if !ok {
			break
		}
		if g.answer(pick == q.Answer) {
			fmt.Fprintf(s.Out, "%s %s\n", p.green("+ Right."), q.Explanation)
		} else {
			fmt.Fprintf(s.Out, "%s it is %d) %s. %s\n", p.red("x Not quite:"), q.Answer+1, q.Choices[q.Answer], q.Explanation)
		}
		fmt.Fprintln(s.Out, g.line(p))
	}
	fmt.Fprintf(s.Out, "\n%s  Score: %d of %d, %d points, best streak %d.\n", g.dots(len(qs), p), g.right, len(g.marks), g.points, g.best)
	return sc.Err()
}

// quizGame keeps the score the way the site's game does, without its time
// bonus: 100 points times a streak multiplier for each right answer.
type quizGame struct {
	marks                       []bool
	right, streak, best, points int
}

func (g *quizGame) answer(right bool) bool {
	g.marks = append(g.marks, right)
	if !right {
		g.streak = 0
		return false
	}
	g.right++
	g.streak++
	g.best = max(g.best, g.streak)
	g.points += 100 * streakMultiplier(g.streak)
	return true
}

// streakMultiplier is x1 for streaks 1-3, x2 for 4-6, x3 for 7-9, x4 after,
// as in the site's game (web/packages/visualizer/src/core/game.ts).
func streakMultiplier(streak int) int { return min(4, 1+max(0, streak-1)/3) }

// line is the score line after an answer.
func (g *quizGame) line(p paint) string {
	streak := fmt.Sprintf("streak %d", g.streak)
	if g.streak > 0 {
		streak += fmt.Sprintf(" (x%d)", streakMultiplier(g.streak))
	}
	return p.dim(fmt.Sprintf("score %d  %s  %d of %d right", g.points, streak, g.right, len(g.marks)))
}

// dots is the progress: + for right, x for wrong, . for still to come.
func (g *quizGame) dots(n int, p paint) string {
	var b strings.Builder
	for i := 0; i < n; i++ {
		switch {
		case i >= len(g.marks):
			b.WriteString(p.dim("."))
		case g.marks[i]:
			b.WriteString(p.green("+"))
		default:
			b.WriteString(p.red("x"))
		}
	}
	return "[" + b.String() + "]"
}

// paint adds color only when the output is a terminal, so piped output and
// tests stay plain.
type paint struct{ on bool }

func newPaint(w io.Writer) paint {
	f, ok := w.(*os.File)
	if !ok || os.Getenv("NO_COLOR") != "" {
		return paint{}
	}
	info, err := f.Stat()
	return paint{on: err == nil && info.Mode()&os.ModeCharDevice != 0}
}

func (p paint) wrap(code, s string) string {
	if !p.on {
		return s
	}
	return "\x1b[" + code + "m" + s + "\x1b[0m"
}

func (p paint) bold(s string) string  { return p.wrap("1", s) }
func (p paint) dim(s string) string   { return p.wrap("2", s) }
func (p paint) green(s string) string { return p.wrap("32", s) }
func (p paint) red(s string) string   { return p.wrap("31", s) }

// readChoice reads lines until one is a choice's number (1 to n), and
// returns it counted from 0; false at the end of the input.
func readChoice(sc *bufio.Scanner, n int, s Streams) (int, bool) {
	for {
		fmt.Fprint(s.Out, "> ")
		if !sc.Scan() {
			fmt.Fprintln(s.Out)
			return 0, false
		}
		k, err := strconv.Atoi(strings.TrimSpace(sc.Text()))
		if err == nil && k >= 1 && k <= n {
			return k - 1, true
		}
		fmt.Fprintf(s.Out, "Type a number from 1 to %d.\n", n)
	}
}
