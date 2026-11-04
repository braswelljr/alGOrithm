package filesystems

import "fmt"

// JournalResult is a Journaling run: the writes done before the crash,
// what recovery did, and whether the file system is consistent after it.
type JournalResult struct {
	Writes     []string `json:"writes"`
	Recovery   []string `json:"recovery"`
	Appended   bool     `json:"appended"`   // the append survived
	Consistent bool     `json:"consistent"` // bitmap, inode and data agree
}

// Journaling appends one block to a file, which writes three things: the
// data block, the inode (its new size and block pointer), and the free
// bitmap (the block is now used); and crashes after crashAfter writes (or
// never, if crashAfter is past the end).
//
// Without a journal, a crash between the writes can leave an inode pointing
// at a block the bitmap calls free (it will be handed to another file), or
// a block marked used that no file owns (leaked): fsck has to scan the
// whole disk to find out. With metadata journaling (ext4's ordered mode) the
// data block goes first, then the inode and bitmap go to the journal with a
// commit record, and only then to their homes (the checkpoint). Recovery
// replays a committed transaction and discards an uncommitted one, so the
// append either fully happened or not at all.
//
// Time: O(1). Space: O(1).
func Journaling(crashAfter int, journaled bool) (JournalResult, error) {
	if crashAfter < 0 || crashAfter > 10 {
		return JournalResult{}, inputErr("crash after 0-10 writes")
	}
	var plan []string
	if journaled {
		plan = []string{"data block (in place)", "journal: begin", "journal: new inode", "journal: new bitmap", "journal: COMMIT", "checkpoint: inode", "checkpoint: bitmap", "journal: free the transaction"}
	} else {
		plan = []string{"data block", "inode", "bitmap"}
	}
	done := map[string]bool{}
	var res JournalResult
	for i, w := range plan {
		if i == crashAfter {
			res.Writes = append(res.Writes, fmt.Sprintf("CRASH before %q", w))
			break
		}
		done[w] = true
		res.Writes = append(res.Writes, "write "+w)
	}
	var inode, bitmap bool
	if journaled {
		inode, bitmap = done["checkpoint: inode"], done["checkpoint: bitmap"]
		switch {
		case done["journal: COMMIT"] && !done["journal: free the transaction"]:
			res.Recovery = append(res.Recovery, "the journal holds a committed transaction: replay it (write the inode and bitmap again)")
			inode, bitmap = true, true
		case done["journal: begin"] && !done["journal: COMMIT"]:
			res.Recovery = append(res.Recovery, "the journal holds a transaction with no commit record: discard it")
		default:
			res.Recovery = append(res.Recovery, "the journal is empty: nothing to do")
		}
	} else {
		inode, bitmap = done["inode"], done["bitmap"]
		res.Recovery = append(res.Recovery, "no journal: fsck must scan every inode and the bitmap to find problems")
	}
	res.Appended = inode && bitmap
	res.Consistent = inode == bitmap
	switch {
	case inode && !bitmap:
		res.Recovery = append(res.Recovery, "INCONSISTENT: the inode points at a block the bitmap says is free; another file could get it")
	case bitmap && !inode:
		res.Recovery = append(res.Recovery, "INCONSISTENT: a block is marked used but no inode points at it (a leak)")
	case res.Appended:
		res.Recovery = append(res.Recovery, "consistent: the append happened")
	default:
		res.Recovery = append(res.Recovery, "consistent: the append did not happen (the data block is just unused space)")
	}
	return res, nil
}
