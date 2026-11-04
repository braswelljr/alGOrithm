// Package buffers holds the structures text editors keep a document in, so
// typing in the middle of a long file does not copy the whole file: a
// GapBuffer (Emacs), which keeps free space at the cursor, and a Rope (many
// modern editors), a tree of short strings that splits and joins without
// copying the text around them.
package buffers
