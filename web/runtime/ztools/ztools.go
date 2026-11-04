// Package ztools is the terminal's compression commands, on klauspost's
// compress packages: gzip and gunzip, zstd and unzstd. Each works the way the
// GNU and zstd tools do for the common case: compress FILE to FILE.gz (or
// .zst) and remove FILE, -k to keep it, -c to write to standard output, -d to
// decompress; with no file, standard input to standard output.
package ztools

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/klauspost/compress/gzip"
	"github.com/klauspost/compress/zstd"
)

// Names are the commands this package provides.
var Names = []string{"gzip", "gunzip", "zstd", "unzstd"}

type format struct {
	ext        string
	compress   func(w io.Writer) (io.WriteCloser, error)
	decompress func(r io.Reader) (io.Reader, func(), error)
}

var formats = map[string]format{
	"gzip": {
		ext:      ".gz",
		compress: func(w io.Writer) (io.WriteCloser, error) { return gzip.NewWriter(w), nil },
		decompress: func(r io.Reader) (io.Reader, func(), error) {
			z, err := gzip.NewReader(r)
			if err != nil {
				return nil, nil, err
			}
			return z, func() { _ = z.Close() }, nil
		},
	},
	"zstd": {
		ext: ".zst",
		compress: func(w io.Writer) (io.WriteCloser, error) {
			return zstd.NewWriter(w)
		},
		decompress: func(r io.Reader) (io.Reader, func(), error) {
			z, err := zstd.NewReader(r)
			if err != nil {
				return nil, nil, err
			}
			return z, z.Close, nil
		},
	},
}

// Run runs one of Names with its arguments and returns the exit status: 0
// success, 1 a file that could not be read or written, 2 a usage mistake.
func Run(name string, args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	kind, decompress := name, false
	switch name {
	case "gunzip":
		kind, decompress = "gzip", true
	case "unzstd":
		kind, decompress = "zstd", true
	case "gzip", "zstd":
	default:
		fmt.Fprintf(stderr, "%s: unknown command\n", name)
		return 2
	}
	f := formats[kind]
	keep, toStdout := false, false
	var files []string
	for _, a := range args {
		switch {
		case a == "-d" || a == "--decompress":
			decompress = true
		case a == "-k" || a == "--keep":
			keep = true
		case a == "-c" || a == "--stdout":
			toStdout = true
		case a == "-h" || a == "--help":
			fmt.Fprintf(stdout, "Usage: %s [-d] [-k] [-c] [FILE]...\nCompress or decompress (%s) each FILE, or standard input.\n", name, f.ext)
			return 0
		case strings.HasPrefix(a, "-") && a != "-":
			fmt.Fprintf(stderr, "%s: invalid option -- '%s'\n", name, strings.TrimLeft(a, "-"))
			return 2
		default:
			files = append(files, a)
		}
	}
	if len(files) == 0 {
		return stream(name, f, decompress, stdin, stdout, stderr)
	}
	status := 0
	for _, file := range files {
		if err := one(f, decompress, keep, toStdout, file, stdout); err != nil {
			fmt.Fprintf(stderr, "%s: %v\n", name, err)
			status = 1
		}
	}
	return status
}

func stream(name string, f format, decompress bool, in io.Reader, out, errw io.Writer) int {
	if err := convert(f, decompress, in, out); err != nil {
		fmt.Fprintf(errw, "%s: %v\n", name, err)
		return 1
	}
	return 0
}

func convert(f format, decompress bool, in io.Reader, out io.Writer) error {
	if decompress {
		r, done, err := f.decompress(in)
		if err != nil {
			return err
		}
		defer done()
		_, err = io.Copy(out, r)
		return err
	}
	w, err := f.compress(out)
	if err != nil {
		return err
	}
	if _, err := io.Copy(w, in); err != nil {
		_ = w.Close()
		return err
	}
	return w.Close()
}

// one converts a file: FILE to FILE.ext (or back), removing the source
// unless keep, or to stdout with -c.
func one(f format, decompress, keep, toStdout bool, file string, stdout io.Writer) error {
	target := file + f.ext
	if decompress {
		if !strings.HasSuffix(file, f.ext) {
			return fmt.Errorf("%s: unknown suffix -- ignored", file)
		}
		target = strings.TrimSuffix(file, f.ext)
	} else if strings.HasSuffix(file, f.ext) && !toStdout {
		return fmt.Errorf("%s already has %s suffix -- unchanged", file, f.ext)
	}
	in, err := os.Open(file)
	if err != nil {
		return err
	}
	defer in.Close()
	if toStdout {
		return convert(f, decompress, in, stdout)
	}
	if _, err := os.Stat(target); err == nil {
		return fmt.Errorf("%s already exists", target)
	}
	out, err := os.Create(target)
	if err != nil {
		return err
	}
	if err := convert(f, decompress, in, out); err != nil {
		_ = out.Close()
		_ = os.Remove(target)
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	if !keep {
		return os.Remove(file)
	}
	return nil
}
