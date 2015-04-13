// package gitgo
package main

import (
	"compress/zlib"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
)

func GetIdxPath(dotGitRootPath string) (idxFilePath string, err error) {
	files, err := filepath.Glob(path.Join(dotGitRootPath, "objects/pack", "*.idx"))
	idxFilePath = files[0]
	return
}

func VerifyPack(idxFilePath string) (verifyPackResult string, err error) {
	f, err := os.Open(idxFilePath)
	if err != nil {
		return
	}
	defer f.Close()

	r, err := zlib.NewReader(f)
	if err != nil {
		return
	}

	bts, err := ioutil.ReadAll(r)
	if err != nil {
		return
	}

	return string(bts), nil
}

func main() {
	path, _ := GetIdxPath("test_data/dot_git")
	fmt.Println("idx path is: ")
	fmt.Println(path)

	fmt.Println("Content is: ")
	idxFileContent, _ := VerifyPack(path)
	fmt.Println(idxFileContent)
}
