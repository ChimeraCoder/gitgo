package gitgo

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"testing"
)

func Test_PatchDelta(t *testing.T) {
	suite := [][3]string{
		[3]string{"test_data/test-delta.c", "test_data/test-delta-new.delta", "test_data/test-delta-new.c"},
		[3]string{"test_data/zlib.c", "test_data/zlib-delta", "test_data/zlib-changed.c"},
	}

	for _, trpl := range suite {

		fin, err := os.Open(trpl[0])
		if err != nil {
			t.Error(err)
			continue
		}
		defer fin.Close()

		deltaf, err := os.Open(trpl[1])
		if err != nil {
			t.Error(err)
			continue
		}
		defer deltaf.Close()

		expectedf, err := os.Open(trpl[2])
		if err != nil {
			t.Error(err)
			continue
		}
		defer expectedf.Close()

		restored, err := patchDelta(fin, deltaf)
		if err != nil {
			t.Error(err)
			continue
		}

		bts1, err := ioutil.ReadAll(expectedf)
		if err != nil {
			t.Error(err)
			continue
		}

		bts2, err := ioutil.ReadAll(restored)
		if err != nil {
			t.Error(err)
			continue
		}

		if len(bts1) != len(bts2) {
			t.Errorf("Expected %d bytes and received %d", len(bts1), len(bts2))
			log.Printf("%q", bts2)
			continue
		}

		if !reflect.DeepEqual(bts1, bts2) {
			t.Errorf("delta application failed")
			continue
		}
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
	log.Printf("1: %s", bts)
	log.Printf("2: %s", bts2)
	if !reflect.DeepEqual(bts, bts2) {
		return false
	}
	return true
}
