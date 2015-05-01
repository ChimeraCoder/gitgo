package gitgo

import "testing"

func Test_Log(t *testing.T) {
	const input SHA = "1d833eb5b6c5369c0cb7a4a3e20ded237490145f"
	expected := []SHA{"a7f92c920ce85f07a33f948aa4fa2548b270024f", "97eed02ebe122df8fdd853c1215d8775f3d9f1a1"}
	parents, err := Log(input, RepoDir)
	if err != nil {
		t.Error(err)
	}

	if len(parents) != len(expected) {
		t.Errorf("received %d parents (expected %d)", len(parents), len(expected))
	}
	for i, parent := range parents {
		if parent.Name != expected[i] {
			t.Errorf("received incorrect parents: \nexpected: %+v\nreceived: %+v", expected, parents)
		}
	}
}
