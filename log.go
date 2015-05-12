package gitgo

import (
	"fmt"
	"reflect"
)

func Log(name SHA, basedir string) ([]Commit, error) {
	repo := Repository{Basedir: basedir}
	err := repo.normalizeBasename()
	if err != nil {
		return nil, err
	}
	as, err := repo.allAncestors(name)
	if err != nil {
		return nil, err
	}

	obj, err := repo.Object(name)
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

func (r *Repository) allAncestors(name SHA) ([]Commit, error) {
	basedir := r.Basedir
	obj, err := r.Object(name)
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
		obj, err := r.Object(commit.Parents[0])
		if err != nil {
			return nil, err
		}
		parent, ok := obj.(Commit)
		if !ok {
			fmt.Println(reflect.TypeOf(obj))
			return nil, fmt.Errorf("receved non-commit object parent: %s (%s)", commit.Parents[0], obj.Type())
		}

		parents = append(parents, parent)
		ancestors, err := r.allAncestors(SHA(commit.Parents[0]))
		if err != nil {
			return parents, err
		}
		parents = append(parents, ancestors...)
	}
	return parents, nil
}
