package gitgo

import "fmt"

func Log(name SHA, basedir string) ([]Commit, error) {
	obj, err := NewObject(name, basedir)
	if err != nil {
		return nil, fmt.Errorf("commit not found")
	}

	var commit Commit
	switch obj := obj.(type) {
	case *packObject:
		// TODO check that this case is logically possible
		commit, err = obj.Commit(basedir)
		if err != nil {
			return nil, err
		}
	case Commit:
		commit = obj
	default:
		return nil, fmt.Errorf("not a commit")
	}

	parents := []Commit{}
	if len(commit.Parents) > 0 {
		// By default, git-log uses the first parent in merges
		obj, err := NewObject(commit.Parents[0], basedir)
		if err != nil {
			return nil, err
		}
		parent, ok := obj.(Commit)
		if !ok {
			return nil, fmt.Errorf("receved non-commit object parent: %", commit.Parents[0])
		}

		parents = append(parents, parent)
		ancestors, err := Log(SHA(commit.Parents[0]), basedir)
		if err != nil {
			return parents, err
		}
		parents = append(parents, ancestors...)
	}
	return parents, nil
}
