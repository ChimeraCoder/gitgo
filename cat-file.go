package gitgo

import (
	"bytes"
	"io"
)

type keyType string

// SHA represents the SHA-1 hash used by git
type SHA string

const (
	treeKey      keyType = "tree"
	parentKey            = "parent"
	authorKey            = "author"
	committerKey         = "committer"
)

// CatFile implements git cat-file for the command-line
// tool. Currently it supports only the -t fiag
func CatFile(name SHA) (io.Reader, error) {
	obj, err := NewObject(name, "")
	if err != nil {
		return nil, err
	}
	return bytes.NewReader([]byte(obj.Type())), err
}
