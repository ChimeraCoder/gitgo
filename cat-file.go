package gitgo

import (
	"compress/zlib"

	"io"
	"os"
	"path"
)

type KeyType string

type SHA string

const (
	TreeKey      KeyType = "tree"
	ParentKey            = "parent"
	AuthorKey            = "author"
	CommitterKey         = "committer"
)

// readObjectFile returns an io.Reader that reads the object file
// corresponding to the given SHA
func readObjectFile(input SHA) (result io.Reader, err error) {

	filename := path.Join(".git", "objects", string(input[:2]), string(input[2:]))

	f, err := os.Open(filename)
	if err != nil {
		return
	}
	defer f.Close()

	return zlib.NewReader(f)
}
