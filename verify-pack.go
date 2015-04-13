package gitgo

import (
	"fmt"
	"path"
	"path/filepath"
)

func GetIdxPath(dotGitRootPath string) (idxFilePath string, err error) {
	files, err := filepath.Glob(path.Join(dotGitRootPath, "objects/pack", "*.idx"))
	idxFilePath = files[0]
	return
}

func GetIdxPath(dotGitRootPath string) (idxFilePath string, err error) {
	files, err := filepath.Glob(path.Join(dotGitRootPath, "objects/pack", "*.idx"))
	idxFilePath = files[0]
	return
}

// func main() {
// 	// fmt.Println("i got...")
// 	fmt.Println(getIdxPath("test_data/dot_git"))
// }

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
