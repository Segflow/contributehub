package codechange

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"sort"
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

// FileApplyChanges applies the changes to file filename. It does not edit the file, the expected file content is returned.
func FileApplyChanges(filename string, changes []CodeChange) ([]byte, error) {
	sort.SliceStable(changes, func(i, j int) bool {
		return changes[i].Offset < changes[j].Offset
	})

	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

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

// FileApplyChangesInplace is like FileApplyChanges but edits the file
func FileApplyChangesInplace(filename string, changes []CodeChange) error {
	content, err := FileApplyChanges(filename, changes)
	if err != nil {
		return fmt.Errorf("error applying changes to file %q: %v", filename, err)
	}

	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error overwriting the file %q: %v", filename, err)
	}

	_, err = f.Write(content)
	if err != nil {
		return fmt.Errorf("error writing result to file %q: %v", filename, err)
	}

	return nil
}
