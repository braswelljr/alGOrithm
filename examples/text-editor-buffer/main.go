// Command text-editor-buffer edits a line the way a text editor does with a
// gap buffer: typing at the cursor fills a gap of free space, so inserts and
// backspaces there are O(1). Emacs keeps its buffers this way.
package main

import (
	"fmt"

	"github.com/braswelljr/alGOrithm/pkg/structures/buffers"
)

func main() {
	buf := buffers.NewGapBuffer("func main() {}")
	fmt.Printf("start:      %q\n", buf.String())
	buf.Move(13) // cursor between the braces
	buf.Insert(" println(\"hi\") ")
	fmt.Printf("typed:      %q\n", buf.String())
	buf.Backspace(1)
	fmt.Printf("backspace:  %q\n", buf.String())
	buf.Move(5)
	buf.Delete(4)
	buf.Insert("start")
	fmt.Printf("rename:     %q (cursor at %d)\n", buf.String(), buf.Cursor())
}
