package gitgo

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
)

type Repository struct {
	Basedir   string
	packfiles []*packfile
}

func (r *Repository) Object(input SHA) (obj GitObject, err error) {
	err = r.normalizeBasename()
	packfiles, err := listPackfiles(r.Basedir)
	if err != nil {
		return nil, err
	}
	r.packfiles = packfiles
	basedir := r.Basedir
	if path.Base(basedir) != ".git" {
		basedir = path.Join(basedir, ".git")
	}
	obj, err = newObject(input, basedir, r.packfiles)
	if err != nil {
		panic(err)
	}
	return obj, err
}

func (r *Repository) normalizeBasename() error {
	candidate := r.Basedir
	if candidate == "" {
		candidate = "."
	}
	if filepath.Base(candidate) != ".git" {
		candidate = path.Join(candidate, ".git")
	}
	for {
		_, err := os.Stat(candidate)
		if err == nil {
			r.Basedir = candidate
			break
		}
		if !os.IsNotExist(err) {
			return err
		}

		// This should not be the main condition of the for loop
		// just in case the filesystem root directory contains
		// a .git subdirectory
		// TODO check for mountpoint
		if candidate == "/.git" {
			return fmt.Errorf("not a git repository (or any parent up to root /")
		}
		candidate, err = filepath.Abs(path.Join(candidate, "..", "..", ".git"))
		if err != nil {
			panic(err)
		}
	}
	return nil
}
