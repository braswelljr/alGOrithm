package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"strings"

	algorithm "github.com/braswelljr/alGOrithm"
	"github.com/braswelljr/alGOrithm/internal/core"
	"github.com/braswelljr/alGOrithm/internal/version"
)

// runMCP serves the registry over the Model Context Protocol (JSON-RPC 2.0,
// one message a line on standard input and output), so an AI assistant can
// list, run, compare and read the algorithms as tools.
func runMCP(args []string, s Streams) error {
	if len(args) > 0 {
		return usagef("mcp takes no arguments")
	}
	in := bufio.NewScanner(s.In)
	in.Buffer(make([]byte, 1<<20), 8<<20)
	enc := json.NewEncoder(s.Out)
	for in.Scan() {
		var req struct {
			ID     json.RawMessage `json:"id"`
			Method string          `json:"method"`
			Params json.RawMessage `json:"params"`
		}
		if err := json.Unmarshal(in.Bytes(), &req); err != nil {
			continue
		}
		if req.ID == nil {
			continue // a notification: nothing to answer
		}
		result, rpcErr := mcpCall(req.Method, req.Params)
		resp := map[string]any{"jsonrpc": "2.0", "id": req.ID}
		if rpcErr != nil {
			resp["error"] = rpcErr
		} else {
			resp["result"] = result
		}
		if err := enc.Encode(resp); err != nil {
			return err
		}
	}
	return in.Err()
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

var mcpTools = []map[string]any{
	{"name": "list_entries", "description": "List every algorithm and data structure: id, name, category, cost, input kind.",
		"inputSchema": map[string]any{"type": "object", "properties": map[string]any{"category": map[string]any{"type": "string"}}}},
	{"name": "run_entry", "description": "Run an entry on an input and return its output, counts of comparisons and swaps, and (up to limit) its step-by-step events.",
		"inputSchema": map[string]any{"type": "object", "required": []string{"id"}, "properties": map[string]any{
			"id": map[string]any{"type": "string"}, "input": map[string]any{"type": "object"}, "limit": map[string]any{"type": "integer"}}}},
	{"name": "compare_entries", "description": "Run several entries on the same input and return their counts and times.",
		"inputSchema": map[string]any{"type": "object", "required": []string{"ids", "input"}, "properties": map[string]any{
			"ids": map[string]any{"type": "array", "items": map[string]any{"type": "string"}}, "input": map[string]any{"type": "object"}}}},
	{"name": "list_paths", "description": "List the learning paths (ordered entries to run, with what to notice at each step and the command that runs it), or one path by id.",
		"inputSchema": map[string]any{"type": "object", "properties": map[string]any{"id": map[string]any{"type": "string"}}}},
	{"name": "read_source", "description": "Return an entry's Go source file and the doc comment of its function.",
		"inputSchema": map[string]any{"type": "object", "required": []string{"id"}, "properties": map[string]any{"id": map[string]any{"type": "string"}}}},
}

func mcpCall(method string, params json.RawMessage) (any, *rpcError) {
	switch method {
	case "initialize":
		return map[string]any{"protocolVersion": "2024-11-05", "capabilities": map[string]any{"tools": map[string]any{}},
			"serverInfo": map[string]any{"name": version.AppName, "version": version.Get()}}, nil
	case "ping":
		return map[string]any{}, nil
	case "tools/list":
		return map[string]any{"tools": mcpTools}, nil
	case "tools/call":
		var p struct {
			Name      string          `json:"name"`
			Arguments json.RawMessage `json:"arguments"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &rpcError{-32602, "invalid params"}
		}
		out, err := mcpTool(p.Name, p.Arguments)
		if err != nil {
			return map[string]any{"isError": true, "content": []map[string]any{{"type": "text", "text": err.Error()}}}, nil
		}
		b, _ := json.MarshalIndent(out, "", "  ")
		return map[string]any{"content": []map[string]any{{"type": "text", "text": string(b)}}}, nil
	}
	return nil, &rpcError{-32601, "method not found: " + method}
}

func mcpTool(name string, raw json.RawMessage) (any, error) {
	var a struct {
		ID       string     `json:"id"`
		IDs      []string   `json:"ids"`
		Category string     `json:"category"`
		Input    core.Input `json:"input"`
		Limit    int        `json:"limit"`
	}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &a); err != nil {
			return nil, fmt.Errorf("invalid arguments: %w", err)
		}
	}
	switch name {
	case "list_paths":
		if a.ID != "" {
			p, ok := core.PathByID(a.ID)
			if !ok {
				return nil, fmt.Errorf("no learning path %q", a.ID)
			}
			return p.View(), nil
		}
		out := make([]core.PathView, 0)
		for _, p := range core.Paths() {
			out = append(out, p.View())
		}
		return out, nil
	case "list_entries":
		var out []map[string]string
		for _, e := range core.Entries() {
			if a.Category == "" || strings.EqualFold(e.Category, a.Category) {
				out = append(out, map[string]string{"id": e.ID, "name": e.Name, "category": e.Category, "time": e.Time, "input": string(e.Input), "summary": e.Summary})
			}
		}
		return out, nil
	case "run_entry":
		if a.Limit == 0 {
			a.Limit = 500
		}
		return core.Execute(a.ID, a.Input, a.Limit)
	case "compare_entries":
		return core.Compare(a.IDs, a.Input)
	case "read_source":
		e, ok := core.Lookup(a.ID)
		if !ok {
			return nil, fmt.Errorf("no entry %q", a.ID)
		}
		code, _ := algorithm.File(e.File)
		return map[string]string{"file": e.File, "doc": algorithm.Doc(e.File, e.Func), "code": code}, nil
	}
	return nil, fmt.Errorf("no tool %q", name)
}
