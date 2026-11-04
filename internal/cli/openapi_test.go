package cli

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// TestOpenAPIDocumentsEveryRoute fails when api/openapi.yaml misses a route
// that serve.go registers, or documents one it does not.
func TestOpenAPIDocumentsEveryRoute(t *testing.T) {
	spec, err := os.ReadFile("../../api/openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	code, err := os.ReadFile("serve.go")
	if err != nil {
		t.Fatal(err)
	}
	documented := map[string]bool{}
	for _, m := range regexp.MustCompile(`(?m)^  (/api/[a-z]+)`).FindAllStringSubmatch(string(spec), -1) {
		documented[m[1]] = true
	}
	served := map[string]bool{}
	for _, m := range regexp.MustCompile(`\{"(/api/[a-z]+)/?", serve`).FindAllStringSubmatch(string(code), -1) {
		served[m[1]] = true
	}
	for p := range served {
		if !documented[p] {
			t.Errorf("%s is served but not in api/openapi.yaml", p)
		}
	}
	for p := range documented {
		if !served[p] {
			t.Errorf("%s is in api/openapi.yaml but not served", p)
		}
	}
	if len(served) == 0 || !strings.Contains(string(spec), "openapi: 3") {
		t.Fatal("found no routes; did serve.go or the spec change shape?")
	}
}
