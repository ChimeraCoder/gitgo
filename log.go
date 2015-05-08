package gitgo

import "fmt"

func Log(name SHA, basedir string) ([]Commit, error) {
	as, err := allAncestors(name, basedir)
	if err != nil {
		return nil, err
	}

	obj, err := NewObject(name, basedir)
	if err != nil {
		return nil, fmt.Errorf("commit not found: %s", err)
	}
	result := make([]Commit, len(as)+1)
	result[0] = obj.(Commit)
	for i, o := range as {
		result[i+1] = o
	}
	return result, nil
}

func allAncestors(name SHA, basedir string) ([]Commit, error) {
	obj, err := NewObject(name, basedir)
	if err != nil {
		return nil, fmt.Errorf("commit not found: %s", err)
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
			return nil, fmt.Errorf("receved non-commit object parent: %s", commit.Parents[0])
		}

		parents = append(parents, parent)
		ancestors, err := allAncestors(SHA(commit.Parents[0]), basedir)
		if err != nil {
			return parents, err
		}
		parents = append(parents, ancestors...)
	}
	return parents, nil
}
