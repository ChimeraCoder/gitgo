package gitgo

import (
	"path"
	"reflect"
	"testing"
)

func Test_GetIdxPath(t *testing.T) {
	var testDirPath = "test_data/dot_git/"
	result, err := GetIdxPath(testDirPath)

	if err != nil {
		t.Error(err)
		return
	}

	expected := path.Join(
		testDirPath,
		"objects/pack",
		"pack-d310969c4ba0ebfe725685fa577a1eec5ecb15b2.idx",
	)

	if !reflect.DeepEqual(expected, result) {
		t.Errorf("Expected and result don't match:\n%+v\n%+v", expected, result)
	}
}

func Test_VerifyPack(t *testing.T) {
	const expected = `fe89ee30bbcdfdf376beae530cc53f967012f31c commit 267 184 12
3ead3116d0378089f5ce61086354aac43e736b01 commit 243 170 196
1d833eb5b6c5369c0cb7a4a3e20ded237490145f commit 262 180 366
a7f92c920ce85f07a33f948aa4fa2548b270024f commit 250 172 546
97eed02ebe122df8fdd853c1215d8775f3d9f1a1 commit 190 132 718
d22fc8a57073fdecae2001d00aff921440d3aabd tree   121 115 850
df891299372c34b57e41cfc50a0113e2afac3210 tree   25 37 965 1 d22fc8a57073fdecae2001d00aff921440d3aabd
af6e4fe91a8f9a0f3c03cbec9e1d2aac47345d67 blob   18 23 1002
6b32b1ac731898894c403f6b621bdda167ab8d7c blob   1645 700 1025
7147f43ae01c9f04a78d6e80544ed84def06e958 blob   1824 697 1725
05d3cc770bd3524cc25d47e083d8942ad25033f0 blob   16 28 2422 1 7147f43ae01c9f04a78d6e80544ed84def06e958
c3b8133617bbdb72e237b0f163fade7fbf1f0c18 blob   381 317 2450 2 05d3cc770bd3524cc25d47e083d8942ad25033f0
8264d7bcc297e15c452a7aef3a2e40934762b7e3 tree   25 38 2767 1 d22fc8a57073fdecae2001d00aff921440d3aabd
254671773e8cd91e07e36546c9a2d9c27e8dfeec tree   121 115 2805
ba74813270ff557c4a5d1be0562a141bbee4d3e6 blob   16 28 2920 1 6b32b1ac731898894c403f6b621bdda167ab8d7c
b45377f6daf59a4cec9e8de64f5df1533a7994cd blob   10 21 2948 1 7147f43ae01c9f04a78d6e80544ed84def06e958
9de6c72106b169990a83ce7090c7cad84b6b506b tree   38 49 2969
non delta: 11 objects
chain length = 1: 5 objects
chain length = 2: 1 object
.git/objects/pack/pack-d310969c4ba0ebfe725685fa577a1eec5ecb15b2.pack: ok

`

	result, err := VerifyPack("test_data/dot_git/objects/pack/pack-d310969c4ba0ebfe725685fa577a1eec5ecb15b2.idx")
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(expected, result) {
		t.Errorf("Expected and result don't match:\n%+v\n%+v", expected, result)
	}

}
