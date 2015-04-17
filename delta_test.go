package gitgo

import (
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

func Test_Delta(t *testing.T) {
	fin, err := os.Open("test_data/zlib.c")
	if err != nil {
		t.Error(err)
		return
	}

	deltaf, err := os.Open("test_data/zlib-delta")
	if err != nil {
		t.Error(err)
		return
	}

	expectedf, err := os.Open("test_data/zlib-changed.c")
	if err != nil {
		t.Error(err)
		return
	}

	restored := Delta(fin, deltaf)
	if !readersEqual(expectedf, restored) {
		t.Errorf("delta application failed")
	}
}

// readersEqual checks that two readers have the same contents
func readersEqual(r1, r2 io.Reader) bool {
	bts, err := ioutil.ReadAll(r1)
	if err != nil {
		panic(err)
	}
	bts2, err := ioutil.ReadAll(r2)
	if err != nil {
		panic(err)
	}
	if !reflect.DeepEqual(bts, bts2) {
		return false
	}
	return true
}
