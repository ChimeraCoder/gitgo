package gitgo

import "fmt"

func Log(name SHA, basedir string) ([]SHA, error) {
	obj, err := NewObject(name, basedir)
	if err != nil {
		return nil, fmt.Errorf("commit not found")
	}

	var commit Commit
	switch obj := obj.(type) {
	case *packObject:
		commit, err = obj.Commit(basedir)
		if err != nil {
			return nil, err
		}
	case Commit:
		commit = obj
	default:
		return nil, fmt.Errorf("not a commit")
	}

	parents := []SHA{}
	if len(commit.Parents) > 0 {
		// By default, git-log uses the first parent in merges
		parents = append(parents, commit.Parents[0])
		ancestors, err := Log(SHA(commit.Parents[0]), basedir)
		if err != nil {
			return parents, err
		}
		parents = append(parents, ancestors...)
	}
	return parents, nil
}
