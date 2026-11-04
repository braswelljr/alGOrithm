package networking

import (
	"fmt"
	"strings"
)

// zones is a small made-up DNS hierarchy: the root knows the top-level
// domains, each of those knows its domains' name servers, and those know
// the addresses. The addresses are illustrative, not real.
var zones = map[string]map[string]string{
	".":             {"com": "a.gtld-servers.net", "org": "a0.org.afilias-nst.info", "dev": "ns-tld1.charlestonroadregistry.com", "edu": "a.edu-servers.net"},
	"com":           {"example.com": "a.iana-servers.net", "shop.com": "ns1.shop.com"},
	"org":           {"wikipedia.org": "ns0.wikimedia.org"},
	"dev":           {"go.dev": "ns1.google.com"},
	"edu":           {"mit.edu": "ns1.mit.edu"},
	"example.com":   {"example.com": "93.184.215.14", "www.example.com": "93.184.215.14", "mail.example.com": "93.184.215.20"},
	"shop.com":      {"www.shop.com": "203.0.113.7"},
	"wikipedia.org": {"www.wikipedia.org": "198.51.100.224", "en.wikipedia.org": "198.51.100.224"},
	"go.dev":        {"go.dev": "192.0.2.21", "pkg.go.dev": "192.0.2.22"},
	"mit.edu":       {"www.mit.edu": "192.0.2.80"},
}

// Resolver looks names up in the simulated hierarchy, caching what it
// learns (answers and which server serves a zone) for the next lookup.
// Iterative: the resolver asks the root, which refers it to the top-level
// server, which refers it to the domain's server, which answers. Recursive:
// each server asks the next itself and passes the answer back, so the root
// does work for everyone (which is why real root servers refuse to recurse).
// Time to live, record types other than addresses, and failures of servers
// are left out.
type Resolver struct {
	Recursive bool
	Timeline  Timeline
	tick      int
	cache     map[string]string // name or zone -> address or "zone"
}

// NewResolver returns a resolver with an empty cache.
func NewResolver(recursive bool) *Resolver {
	return &Resolver{Recursive: recursive, cache: map[string]string{}}
}

// Lookup is the outcome of one name's resolution.
type Lookup struct {
	Name     string   `json:"name"`
	Answer   string   `json:"answer"` // an address, or "NXDOMAIN"
	Messages int      `json:"messages"`
	Steps    []string `json:"steps"`
}

// Resolve finds name's address. It returns an error only for a malformed
// name; a name that does not exist is answered NXDOMAIN.
//
// Time: O(labels) messages. Space: O(labels).
func (r *Resolver) Resolve(name string) (Lookup, error) {
	name = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(name), "."))
	labels := strings.Split(name, ".")
	if name == "" || len(labels) < 2 || len(name) > 253 {
		return Lookup{}, fmt.Errorf("%w: %q is not a full domain name", ErrInput, name)
	}
	lk := Lookup{Name: name}
	msg := func(from, to, what string) {
		lk.Messages++
		lk.Steps = append(lk.Steps, fmt.Sprintf("%s -> %s: %s", from, to, what))
		r.Timeline.add(r.tick, from, what+" -> "+to, "running")
		r.tick++
	}
	msg("client", "resolver", name+"?")
	if a, ok := r.cache[name]; ok {
		msg("resolver", "client", a+" (cached)")
		lk.Answer = a
		return lk, nil
	}
	tld := labels[len(labels)-1]
	domain := strings.Join(labels[len(labels)-2:], ".")
	// the chain of zones from the root down, and what each says
	chain := []string{".", tld, domain}
	server := func(zone string) string {
		if zone == "." {
			return "root"
		}
		return "." + zone
	}
	answer := "NXDOMAIN"
	start := 0
	if _, ok := r.cache[domain]; ok {
		start = 2 // the resolver already knows the domain's server
	} else if _, ok := r.cache[tld]; ok {
		start = 1
	}
	if r.Recursive {
		prev := "resolver"
		depth := start
		for depth < len(chain) {
			z := chain[depth]
			msg(prev, server(z), name+"?")
			prev = server(z)
			child := ""
			if depth+1 < len(chain) {
				child = chain[depth+1]
			}
			if _, ok := zones[z][child]; child != "" && ok {
				depth++
				continue
			}
			if a, ok := zones[z][name]; ok && child == "" {
				answer = a
			}
			break
		}
		// the answer travels back up the chain
		for d := depth; d > start; d-- {
			msg(server(chain[d]), server(chain[d-1]), answer)
		}
		msg(server(chain[start]), "resolver", answer) // the resolver learns only the answer
	} else {
		for d := start; d < len(chain); d++ {
			z := chain[d]
			msg("resolver", server(z), name+"?")
			if d+1 < len(chain) {
				ns, ok := zones[z][chain[d+1]]
				if !ok {
					msg(server(z), "resolver", "NXDOMAIN")
					break
				}
				msg(server(z), "resolver", "ask "+ns+" for "+chain[d+1])
				r.cache[chain[d+1]] = "zone"
				continue
			}
			if a, ok := zones[z][name]; ok {
				answer = a
			}
			msg(server(z), "resolver", answer)
		}
	}
	if answer != "NXDOMAIN" {
		r.cache[name] = answer
	}
	lk.Answer = answer
	msg("resolver", "client", answer)
	return lk, nil
}
