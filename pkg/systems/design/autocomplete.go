package design

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// Term is a search term and how often it was searched.
type Term struct {
	Text  string
	Count int
}

// ParseTerms reads one term a line, "text count", the count optional (1).
func ParseTerms(text string) ([]Term, error) {
	var out []Term
	for i, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		f := strings.Fields(line)
		n := 1
		if len(f) > 1 {
			v, err := strconv.Atoi(f[len(f)-1])
			if err == nil {
				n, f = v, f[:len(f)-1]
			}
		}
		if err := check(n > 0, "line %d: the count must be positive", i+1); err != nil {
			return nil, err
		}
		out = append(out, Term{strings.ToLower(strings.Join(f, " ")), n})
	}
	return out, check(len(out) > 0, "no terms")
}

// Autocomplete models search-as-you-type: offline, an aggregator builds a
// table from every prefix to its TopK most searched terms (in production, a
// trie whose nodes cache their top k, rebuilt from the query logs), so each
// keystroke is one lookup instead of a scan of every term under the prefix.
// It returns the suggestions for each prefix of typed, and how many terms a
// scan at read time would have looked at. It leaves out typo tolerance and
// personalization.
//
// Time: O(total term length x TopK) to build, O(1) per keystroke. Space: O(total term length x TopK).
func Autocomplete(terms []Term, typed string, topK int) (Report, error) {
	if err := check(len(terms) > 0 && topK >= 1 && topK <= 20 && typed != "", "terms, a typed query and a top k of 1 to 20"); err != nil {
		return Report{}, err
	}
	merged := map[string]int{}
	for _, t := range terms {
		merged[t.Text] += t.Count
	}
	all := make([]Term, 0, len(merged))
	for t, n := range merged {
		all = append(all, Term{t, n})
	}
	sort.Slice(all, func(i, j int) bool {
		if all[i].Count != all[j].Count {
			return all[i].Count > all[j].Count
		}
		return all[i].Text < all[j].Text
	})
	table := map[string][]Term{}
	under := map[string]int{}
	for _, t := range all {
		for i := 1; i <= len(t.Text); i++ {
			p := t.Text[:i]
			under[p]++
			if len(table[p]) < topK {
				table[p] = append(table[p], t)
			}
		}
	}
	a := arch([]Component{
		{ID: "client", Kind: Client, Label: "Search box"},
		{ID: "api", Kind: Service, Label: "Suggest service"},
		{ID: "top", Kind: Cache, Label: "Prefix -> top k", Notes: "built offline"},
		{ID: "logs", Kind: Queue, Label: "Query logs"},
		{ID: "agg", Kind: Worker, Label: "Aggregator", Notes: "rebuilds the table"},
	}, "client>api:keystroke", "api>top:lookup", "client>logs:searched", "logs>agg", "agg>top:rebuild")
	var r Report
	r.Arch = a
	typed = strings.ToLower(typed)
	scanned := 0
	for i := 1; i <= len(typed); i++ {
		p := typed[:i]
		sugg := make([]string, len(table[p]))
		for k, t := range table[p] {
			sugg[k] = fmt.Sprintf("%s (%d)", t.Text, t.Count)
		}
		scanned += under[p]
		shown := strings.Join(sugg, ", ")
		if len(sugg) == 0 {
			shown = "no suggestions"
		}
		r.say("%q: %s", p, shown)
		if a.recording() {
			f := a.frame(fmt.Sprintf("typed %q: %d suggestions in one lookup", p, len(sugg)))
			a.send(f, "client", "api", i, OK)
			a.send(f, "api", "top", i, map[bool]Status{true: Hit, false: Miss}[len(sugg) > 0])
		}
	}
	f := a.frame("the aggregator rebuilds the table from the logs")
	a.send(f, "client", "logs", 0, Queued)
	a.send(f, "logs", "agg", 0, OK)
	a.send(f, "agg", "top", 0, OK)
	r.put("prefixes stored", float64(len(table)))
	r.put("lookups", float64(len(typed)))
	r.put("terms a scan would read", float64(scanned))
	return r, nil
}
