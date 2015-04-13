package gitgo

import (
	"compress/zlib"

	"io/ioutil"
	"os"
	"path"
)

func getIdxPath(dotGitRootPath string) (idxFilePath string, err error) {
	files, _ := path.filepath.Glob(path.Join(idxFilePath, "*.idx"))
	return files[0]
}

// type KeyType string
// type SHA string
//
// const (
// 	TreeKey      KeyType = "tree"
// 	ParentKey            = "parent"
// 	AuthorKey            = "author"
// 	CommitterKey         = "committer"
// )

// func CatFile(input SHA) (result string, err error) {
//
// 	filename := path.Join(".git", "objects", string(input[:2]), string(input[2:]))
//
// 	f, err := os.Open(filename)
// 	if err != nil {
// 		return
// 	}
// 	defer f.Close()
//
// 	r, err := zlib.NewReader(f)
// 	if err != nil {
// 		return
// 	}
//
// 	bts, err := ioutil.ReadAll(r)
// 	if err != nil {
// 		return
// 	}
//
// 	return string(bts), nil
// }
