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
	expected := GitObject{
		Type:      "commit",
		Tree:      "9de6c72106b169990a83ce7090c7cad84b6b506b",
		Parents:   nil,
		Author:    "aditya <dev@chimeracoder.net> 1428075900 -0400",
		Committer: "aditya <dev@chimeracoder.net> 1428075900 -0400",
		Message:   "First commit. Create .gitignore",
		Size:      "190",
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

	expected := GitObject{
		Type:      "commit",
		Tree:      "d22fc8a57073fdecae2001d00aff921440d3aabd",
		Parents:   []string{"1d833eb5b6c5369c0cb7a4a3e20ded237490145f"},
		Author:    "aditya <dev@chimeracoder.net> 1428349896 -0400",
		Committer: "aditya <dev@chimeracoder.net> 1428349896 -0400",
		Message:   "Remove extraneous logging statements\n",
		Size:      "243",
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
