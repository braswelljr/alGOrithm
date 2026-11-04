package ztools

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRoundTrip compresses a file and decompresses it back with both
// formats: the same bytes, the source removed unless -k.
func TestRoundTrip(t *testing.T) {
	text := []byte(strings.Repeat("the quick brown fox jumps over the lazy dog\n", 200))
	for _, c := range []struct{ pack, unpack, ext string }{{"gzip", "gunzip", ".gz"}, {"zstd", "unzstd", ".zst"}} {
		dir := t.TempDir()
		file := filepath.Join(dir, "fox.txt")
		if err := os.WriteFile(file, text, 0o644); err != nil {
			t.Fatal(err)
		}
		var out, errs bytes.Buffer
		if s := Run(c.pack, []string{file}, nil, &out, &errs); s != 0 {
			t.Fatalf("%s: status %d, %s", c.pack, s, errs.String())
		}
		if _, err := os.Stat(file); !os.IsNotExist(err) {
			t.Errorf("%s left the source", c.pack)
		}
		packed, _ := os.ReadFile(file + c.ext)
		if len(packed) >= len(text) {
			t.Errorf("%s: %d bytes from %d", c.pack, len(packed), len(text))
		}
		if s := Run(c.unpack, []string{"-k", file + c.ext}, nil, &out, &errs); s != 0 {
			t.Fatalf("%s: status %d, %s", c.unpack, s, errs.String())
		}
		back, _ := os.ReadFile(file)
		if !bytes.Equal(back, text) {
			t.Errorf("%s then %s changed the text", c.pack, c.unpack)
		}
		if _, err := os.Stat(file + c.ext); err != nil {
			t.Errorf("%s -k removed the archive", c.unpack)
		}
	}
}

// TestStreams compresses standard input to standard output and back.
func TestStreams(t *testing.T) {
	var packed, back, errs bytes.Buffer
	if s := Run("zstd", nil, strings.NewReader("hello"), &packed, &errs); s != 0 {
		t.Fatal(errs.String())
	}
	if s := Run("zstd", []string{"-d"}, &packed, &back, &errs); s != 0 || back.String() != "hello" {
		t.Errorf("zstd round trip: %q (%s)", back.String(), errs.String())
	}
}

func TestMistakes(t *testing.T) {
	var out, errs bytes.Buffer
	dir := t.TempDir()
	if s := Run("gunzip", []string{filepath.Join(dir, "notes.txt")}, nil, &out, &errs); s != 1 {
		t.Errorf("gunzip of a file without .gz: status %d", s)
	}
	if s := Run("gzip", []string{"-z"}, nil, &out, &errs); s != 2 {
		t.Errorf("an unknown option: status %d", s)
	}
	if s := Run("gzip", []string{filepath.Join(dir, "missing")}, nil, &out, &errs); s != 1 {
		t.Errorf("a missing file: status %d", s)
	}
}
