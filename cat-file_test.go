package gitgo

import (
	"strings"
	"testing"
)

func Test_CatFile(t *testing.T) {
	const inputSha = "97eed02ebe122df8fdd853c1215d8775f3d9f1a1"
	const expected = "commit 190\x00" + `tree 9de6c72106b169990a83ce7090c7cad84b6b506b
author aditya <dev@chimeracoder.net> 1428075900 -0400
committer aditya <dev@chimeracoder.net> 1428075900 -0400

First commit. Create .gitignore`
	result, err := CatFile(inputSha)
	if err != nil {
		t.Error(err)
		return
	}

	if strings.Trim(result, "\n\r") != strings.Trim(expected, "\n\r") {

		t.Errorf("Expected and result don't match:\n%s \n\n\nresult: \n%s", expected, result)
	}
}
