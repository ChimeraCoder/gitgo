package gitgo

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
)

type Repository struct {
	Basedir   os.File
	packfiles []*packfile
}

func (r *Repository) Object(input SHA) (obj GitObject, err error) {
	err = r.normalizeBasename()
	if r.packfiles == nil {
		packfiles, err := r.listPackfiles()
		if err != nil {
			return nil, err
		}
		r.packfiles = packfiles
	}
	basedir := &r.Basedir
	if path.Base(basedir.Name()) != ".git" {
		basedirName := basedir.Name()
		basedir.Close()
		basedir, err = os.Open(path.Join(basedirName, ".git"))
		if err != nil {
			return nil, err
		}
	}
	obj, err = newObject(input, basedir, r.packfiles)
	return obj, err
}

func (r *Repository) normalizeBasename() error {
	var err error
	candidate := &r.Basedir
	if candidate.Name() == "" {
		candidate, err = os.Open(".")
		if err != nil {
			return err
		}
	}
	candidateName := candidate.Name()
	if filepath.Base(candidateName) != ".git" {
		candidateName = path.Join(candidateName, ".git")
	}
	for {
		candidate, err = os.Open(candidateName)
		if err == nil {
			r.Basedir = *candidate
			break
		}
		if !os.IsNotExist(err) {
			return err
		}

		// This should not be the main condition of the for loop
		// just in case the filesystem root directory contains
		// a .git subdirectory
		// TODO check for mountpoint
		if candidateName == "/.git" {
			return fmt.Errorf("not a git repository (or any parent up to root /")
		}
		candidateName, err = filepath.Abs(path.Join(candidateName, "..", "..", ".git"))
		candidate.Close()
	}
	return nil
}
