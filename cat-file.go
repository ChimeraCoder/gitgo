package gitgo

import (
	"compress/zlib"

	"io"
	"os"
	"path"
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

// readObjectFile returns an io.Reader that reads the object file
// corresponding to the given SHA
func readObjectFile(input SHA, basedir string) (result io.Reader, err error) {
	filename := path.Join(basedir, "objects", string(input[:2]), string(input[2:]))

	f, err := os.Open(filename)
	if err != nil {
		return
	}
	defer f.Close()

	return zlib.NewReader(f)
}
