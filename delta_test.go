package gitgo

import (
	"bytes"
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

func Test_parseVarInt(t *testing.T) {
	type pair struct {
		b []byte
		i int
	}
	inputs := []pair{
		pair{[]byte{145, 46}, 5905},
		pair{[]byte{137, 49}, 6281},
	}
	for _, p := range inputs {
		input := p.b
		expected := p.i
		result, err := parseVarInt(bytes.NewBuffer(input))
		if err != nil {
			t.Error(err)
			return
		}
		if result != expected {
			t.Errorf("Expected %d and received %d", expected, result)
		}
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
