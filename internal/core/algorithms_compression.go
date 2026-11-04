package core

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/algorithms/compression"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// compressionEntries is every entry of the compression category.
func compressionEntries() []Entry {
	return []Entry{
		burrowsWheelerEntry(),
		lz77Entry(),
		lzwEntry(),
		moveToFrontEntry(),
		runLengthEncodingEntry(),
	}
}

// textBytes checks a compression entry's text.
func textBytes(in Input, limit int) ([]byte, error) {
	if in.Text == "" {
		return nil, inputError("give the text to compress")
	}
	if len(in.Text) > limit {
		return nil, inputError(fmt.Sprintf("at most %d characters", limit))
	}
	return []byte(in.Text), nil
}

// rowSnapshot draws text as a row of characters.
func rowSnapshot(labels []string) trace.Snapshot {
	s := trace.Snapshot{Layout: "row"}
	for i, l := range labels {
		s.Nodes = append(s.Nodes, trace.Node{ID: i, Label: l, Parent: -1})
	}
	return s
}

func compressionTextEntry(id, name, file, fn, summary, time, space, ex string, run func(data []byte, in Input) (Output, error)) Entry {
	e := entry(id, name, "algorithms/compression/"+file, fn)
	e.Summary, e.Time, e.Space, e.Input, e.Example, e.Limits = summary, time, space, Text, &Input{Text: ex}, "at most 2000 characters"
	e.Run = func(in Input, _ trace.Recorder) (Output, error) {
		data, err := textBytes(in, 2000)
		if err != nil {
			return Output{}, err
		}
		return run(data, in)
	}
	e.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("compression.%s([]byte(%q))", fn, in.Text)
	}
	return e
}

func runLengthEncodingEntry() Entry {
	rle := compressionTextEntry("run-length-encoding", "Run-Length Encoding", "runlength.go", "RunLength", "Write each run of equal characters as the character and a count: great on long runs, twice the size with none.", "O(n)", "O(runs)", "WWWWWWWWWWWWBWWWWWWWWWWWWBBBWWWWWWWWWWWWWW",
		func(data []byte, _ Input) (Output, error) {
			runs := compression.RunLength(data)
			var parts []string
			out := Output{}
			for _, r := range runs {
				parts = append(parts, fmt.Sprintf("%d%c", r.Count, r.Value))
			}
			back, _ := compression.RunLengthDecode(runs)
			out.Text = strings.Join(parts, "")
			out.Lines = []string{fmt.Sprintf("%d runs: %s", len(runs), strings.Join(parts, " ")), "decodes to " + string(back)}
			out.Summary = map[string]float64{"runs": float64(len(runs)), "ratio": compression.Ratio(len(data), 2*len(runs))}
			return out, nil
		})
	return rle
}

func lz77Entry() Entry {
	lz := compressionTextEntry("lz77", "LZ77", "lz77.go", "LZ77", "Replace repeats with (offset, length) pointers back into a sliding window of what came before, plus the next character; the idea under gzip and PNG.", "O(n * window)", "O(tokens)", "abracadabra abracadabra",
		func(data []byte, in Input) (Output, error) {
			w := in.opt("window", 64)
			if w < 1 || w > 4096 {
				return Output{}, inputError("the window is 1 to 4096")
			}
			tokens := compression.LZ77(data, w)
			out := Output{}
			var sofar []byte
			for _, t := range tokens {
				part := []byte{t.Next}
				if t.Length > 0 {
					start := len(sofar) - t.Offset
					part = nil
					for k := 0; k < t.Length; k++ {
						part = append(part, sofar[start+k%t.Offset])
					}
					part = append(part, t.Next)
				}
				sofar = append(sofar, part...)
				label := fmt.Sprintf("(%d, %d, %q)", t.Offset, t.Length, t.Next)
				out.Lines = append(out.Lines, fmt.Sprintf("%s adds %q", label, part))
				out.Steps = append(out.Steps, Step{Op: Op{Name: "token", Value: label}, Result: string(part), Snapshot: rowSnapshot(chars(sofar))})
			}
			out.Number = ptr(len(tokens))
			out.Summary = map[string]float64{"tokens": float64(len(tokens)), "ratio": compression.Ratio(len(data), 3*len(tokens))}
			return out, nil
		})
	lz.Options = []Option{{"window", "window size", 64, 1, 4096}}
	lz.Example.Options = map[string]int{"window": 64}
	lz.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("compression.LZ77([]byte(%q), %d)", in.Text, in.opt("window", 64))
	}
	return lz
}

func lzwEntry() Entry {
	lzw := compressionTextEntry("lzw", "LZW", "lzw.go", "LZW", "Emit the code of the longest phrase already in the dictionary and add that phrase plus the next character as a new code; GIF's compressor.", "O(n)", "O(n)", "TOBEORNOTTOBEORTOBEORNOT",
		func(data []byte, _ Input) (Output, error) {
			codes, added := compression.LZW(data)
			out := Output{Values: codes}
			var dict []string
			for i, p := range added {
				dict = append(dict, fmt.Sprintf("%d=%s", 256+i, p))
				out.Steps = append(out.Steps, Step{Op: Op{Name: "add", Value: p}, Result: strconv.Itoa(256 + i), Snapshot: rowSnapshot(dict)})
				out.Lines = append(out.Lines, fmt.Sprintf("code %d = %q", 256+i, p))
			}
			out.Summary = map[string]float64{"codes": float64(len(codes)), "dictionary entries": float64(len(added)), "ratio (12-bit codes)": compression.Ratio(8*len(data), 12*len(codes))}
			return out, nil
		})
	return lzw
}

func burrowsWheelerEntry() Entry {
	bwt := compressionTextEntry("burrows-wheeler", "Burrows-Wheeler Transform", "bwt.go", "BurrowsWheeler", "Sort every rotation of the text and keep the last column: repeated contexts bunch equal characters together, and the transform can be undone.", "O(n^2 log n) here, O(n) with suffix arrays", "O(n)", "banana",
		func(data []byte, _ Input) (Output, error) {
			if len(data) > 300 {
				return Output{}, inputError("at most 300 characters, so the rotations fit")
			}
			s := string(data)
			last, idx := compression.BurrowsWheeler(s)
			back, _ := compression.InverseBurrowsWheeler(last, idx)
			out := Output{Text: last, Number: ptr(idx)}
			for i, r := range compression.Rotations(s) {
				line := r
				if i == idx {
					line += "   <- the text, row " + strconv.Itoa(i)
				}
				out.Lines = append(out.Lines, line)
			}
			out.Lines = append(out.Lines, fmt.Sprintf("last column %q, row %d; inverting gives %q", last, idx, back))
			return out, nil
		})
	bwt.Limits = "at most 300 characters"
	bwt.Call = func(in Input) (string, string) { return "", fmt.Sprintf("compression.BurrowsWheeler(%q)", in.Text) }
	return bwt
}

func moveToFrontEntry() Entry {
	mtf := compressionTextEntry("move-to-front", "Move-to-Front", "mtf.go", "MoveToFront", "Replace each byte with its place in a list and move it to the front, so recently seen bytes become small numbers.", "O(256 n)", "O(n)", "nnbaaa",
		func(data []byte, _ Input) (Output, error) {
			ranks := compression.MoveToFront(data)
			out := Output{}
			zeros := 0
			for i, r := range ranks {
				out.Values = append(out.Values, int(r))
				if r == 0 {
					zeros++
				}
				if i < 50 {
					out.Lines = append(out.Lines, fmt.Sprintf("%q -> %d", data[i], r))
				}
			}
			out.Summary = map[string]float64{"zeros": float64(zeros)}
			return out, nil
		})
	return mtf
}
