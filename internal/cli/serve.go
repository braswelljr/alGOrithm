package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	algorithm "github.com/braswelljr/alGOrithm"
	"github.com/braswelljr/alGOrithm/internal/core"
	"github.com/braswelljr/alGOrithm/internal/version"
)

// Handler is the registry as a JSON API:
//
//	GET  /api/catalog            the categories and every entry
//	GET  /api/entries/{id}       one entry, with its source and doc
//	POST /api/run/{id}           body: an Input; ?limit=N caps the events
//	POST /api/compare            body: {"ids": [...], "input": Input}
//	POST /api/snippet/{id}       body: an Input; the Go program and the command
func Handler() http.Handler {
	mux := http.NewServeMux()
	for _, rt := range routes {
		mux.HandleFunc(rt.path, rt.handle)
	}
	return mux
}

// routes are the API's endpoints; a new one is a row here and a handler below.
var routes = []struct {
	path   string
	handle http.HandlerFunc
}{
	{"/api/catalog", serveCatalog},
	{"/api/entries/", serveEntry},
	{"/api/run/", serveRun},
	{"/api/compare", serveCompare},
	{"/api/snippet/", serveSnippet},
}

func serveCatalog(w http.ResponseWriter, _ *http.Request) {
	reply(w, http.StatusOK, map[string]any{"version": version.Get(), "categories": core.Categories, "entries": core.Entries()})
}

func serveEntry(w http.ResponseWriter, r *http.Request) {
	e, ok := entryAt(w, r, "/api/entries/")
	if !ok {
		return
	}
	code, _ := algorithm.File(e.File)
	reply(w, http.StatusOK, map[string]any{"entry": e, "code": code, "doc": algorithm.Doc(e.File, e.Func)})
}

func serveRun(w http.ResponseWriter, r *http.Request) {
	e, ok := entryAt(w, r, "/api/run/")
	if !ok {
		return
	}
	var in core.Input
	if err := decode(r, &in); err != nil {
		fail(w, err)
		return
	}
	limit, err := limitOf(r)
	if err != nil {
		fail(w, err)
		return
	}
	res, err := core.Execute(e.ID, in, limit)
	if err != nil {
		fail(w, err)
		return
	}
	reply(w, http.StatusOK, res)
}

func serveCompare(w http.ResponseWriter, r *http.Request) {
	var body struct {
		IDs   []string   `json:"ids"`
		Input core.Input `json:"input"`
	}
	if err := decode(r, &body); err != nil {
		fail(w, err)
		return
	}
	res, err := core.Compare(body.IDs, body.Input)
	if err != nil {
		fail(w, err)
		return
	}
	reply(w, http.StatusOK, res)
}

func serveSnippet(w http.ResponseWriter, r *http.Request) {
	e, ok := entryAt(w, r, "/api/snippet/")
	if !ok {
		return
	}
	in := *e.Example
	if r.ContentLength != 0 {
		if err := decode(r, &in); err != nil {
			fail(w, err)
			return
		}
	}
	reply(w, http.StatusOK, core.MakeSnippet(e, in))
}

// reply writes v as JSON with the given status.
func reply(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// fail replies 400 for rejected input and 500 for anything else.
func fail(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	if errors.Is(err, core.ErrInput) {
		status = http.StatusBadRequest
	}
	reply(w, status, map[string]string{"error": err.Error()})
}

// decode reads a JSON body of at most 8 MiB into v, rejecting unknown fields.
func decode(r *http.Request, v any) error {
	dec := json.NewDecoder(http.MaxBytesReader(nil, r.Body, 8<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return fmt.Errorf("%w: the body is not valid JSON input: %w", core.ErrInput, err)
	}
	return nil
}

// entryAt looks up the entry named by the path after prefix, replying 404
// when there is none.
func entryAt(w http.ResponseWriter, r *http.Request, prefix string) (core.Entry, bool) {
	id := r.URL.Path[len(prefix):]
	e, ok := core.Lookup(id)
	if !ok {
		reply(w, http.StatusNotFound, map[string]string{"error": "no entry " + id})
	}
	return e, ok
}

// limitOf reads ?limit=N, the events to keep (0, the default, keeps
// trace.DefaultLimit); a value that is not a number is rejected.
func limitOf(r *http.Request) (int, error) {
	l := r.URL.Query().Get("limit")
	if l == "" {
		return 0, nil
	}
	n, err := strconv.Atoi(l)
	if err != nil {
		return 0, fmt.Errorf("%w: limit %q is not a whole number", core.ErrInput, l)
	}
	return n, nil
}

func runServe(args []string, s Streams) error {
	o := newOptions("serve")
	if _, err := o.parse(args); err != nil {
		return err
	}
	srv := &http.Server{Addr: o.addr, Handler: Handler(), ReadHeaderTimeout: 5 * time.Second}
	fmt.Fprintf(s.Err, "%s: serving the API on http://%s/api/catalog\n", version.AppName, o.addr)
	return srv.ListenAndServe()
}
