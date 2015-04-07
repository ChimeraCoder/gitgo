package gitgo

import (
	"reflect"
	"strings"
	"testing"
)

func Test_CatFile(t *testing.T) {
	const inputSha = SHA("97eed02ebe122df8fdd853c1215d8775f3d9f1a1")
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

func Test_parseObjInitialCommit(t *testing.T) {
	expected := Commit{
		_type:     "commit",
		Tree:      "9de6c72106b169990a83ce7090c7cad84b6b506b",
		Parents:   nil,
		Author:    "aditya <dev@chimeracoder.net> 1428075900 -0400",
		Committer: "aditya <dev@chimeracoder.net> 1428075900 -0400",
		Message:   "First commit. Create .gitignore",
		size:      "190",
	}

	const input = "commit 190\x00" + `tree 9de6c72106b169990a83ce7090c7cad84b6b506b
author aditya <dev@chimeracoder.net> 1428075900 -0400
committer aditya <dev@chimeracoder.net> 1428075900 -0400

First commit. Create .gitignore`
	result, err := parseObj(input)
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(expected, result) {
		t.Errorf("Expected and result don't match:\n%+v\n%+v", expected, result)
	}
}

func Test_parseObj(t *testing.T) {
	const inputSha = SHA("3ead3116d0378089f5ce61086354aac43e736b01")

	expected := Commit{
		_type:     "commit",
		Tree:      "d22fc8a57073fdecae2001d00aff921440d3aabd",
		Parents:   []string{"1d833eb5b6c5369c0cb7a4a3e20ded237490145f"},
		Author:    "aditya <dev@chimeracoder.net> 1428349896 -0400",
		Committer: "aditya <dev@chimeracoder.net> 1428349896 -0400",
		Message:   "Remove extraneous logging statements\n",
		size:      "243",
	}

	str, err := CatFile(inputSha)
	if err != nil {
		t.Error(err)
	}

	result, err := parseObj(str)
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(expected, result) {
		t.Errorf("Expected and result don't match:\n%+v\n%+v", expected, result)
	}
}

func Test_ParseTree(t *testing.T) {
	const inputSha = SHA("1efecd717188441397c07f267cf468fdf04d4796")
	expected := Tree{
		_type: "tree",
		size:  "156",
		Blobs: []objectMeta{
			objectMeta{SHA("af6e4fe91a8f9a0f3c03cbec9e1d2aac47345d67"), "100644", ".gitignore"},
			objectMeta{SHA("f45d37d9add8f21eb84678f6d2c66377c4dd0c5e"), "100644", "cat-file.go"},
			objectMeta{SHA("2c225b962d6666011c69ca5c2c67204959f8ba32"), "100644", "cat-file_test.go"},
		},
		Trees: []objectMeta{
			objectMeta{SHA("d564d0bc3dd917926892c55e3706cc116d5b165e"), "040000", "examples"},
		},
	}
	result, err := NewObject(inputSha)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("Expected and result don't match:\n\n%+v\n\n%+v", expected, result)
	}

}

func Test_ParseBlob(t *testing.T) {
	const inputSha = SHA("af6e4fe91a8f9a0f3c03cbec9e1d2aac47345d67")
	expected := Blob{
		_type:    "blob",
		size:     "18",
		Contents: "*.swp\n*.swo\n*.swn\n",
	}
	result, err := NewObject(inputSha)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("Expected and result don't match:\n\n%+v\n\n%+v", expected, result)
	}

}
