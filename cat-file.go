package gitgo

import (
	"compress/zlib"
	"io/ioutil"
	"os"
	"path"
)

func CatFile(inputSha string) (result string, err error) {

	filename := path.Join(".git", "objects", inputSha[:2], inputSha[2:])

	f, err := os.Open(filename)
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
