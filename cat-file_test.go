package gitgo

import (
	"log"
	"reflect"
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

func Test_parseObj(t *testing.T) {
	expected := GitObject{
		Type: "commit",
		Tree: "9de6c72106b169990a83ce7090c7cad84b6b506b",
		//Parents:   nil,
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

		numFields := reflect.ValueOf(expected).NumField()
		for i := 0; i < numFields; i++ {
			if !reflect.DeepEqual(reflect.ValueOf(expected).Field(i), reflect.ValueOf(result).Field(i)) {
				log.Printf("Different value for %s", reflect.TypeOf(expected).Field(i).Name)
				log.Printf("Different value for %s", reflect.TypeOf(expected).Field(i))
			}
		}

	}
	log.Print(reflect.DeepEqual(expected.Parents, result.Parents))
}
