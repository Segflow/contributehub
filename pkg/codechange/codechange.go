package codechange

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"sort"

	"github.com/davecgh/go-spew/spew"
)

type CodeChange struct {
	Filename string
	Line     int
	Column   int
	Offset   int

	// String to add to the current position
	Add []byte

	// Number of characters to delete from the current position
	Delete int
}

func FileApplyChanges(filename string, changes []CodeChange) ([]byte, error) {
	sort.SliceStable(changes, func(i, j int) bool {
		return changes[i].Offset < changes[j].Offset
	})

	spew.Dump(changes)
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(f)

	buf := bytes.NewBuffer(nil)

	var lastOffset int

	for _, change := range changes {
		readCount := change.Offset - lastOffset
		b := make([]byte, readCount)
		count, err := io.ReadFull(reader, b)
		if count != readCount {
			return nil, err
		}

		buf.Write(b)
		buf.Write(change.Add)

		lastOffset = readCount
	}

	// Copy the result
	io.Copy(buf, reader)

	return buf.Bytes(), nil
}
