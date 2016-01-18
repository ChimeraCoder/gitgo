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

// findGitDir is like normalizeBasename but does not require
// a repository with a valid file descriptor to operate on.
// If pwd is non-nil, it will use the provided file. Otherwise,
// it will default to the current working directory.
func findGitDir(pwd *os.File) (dir *os.File, err error) {
	if pwd == nil {
		pwd, err = os.Open(".")
		if err != nil {
			return nil, err
		}
	}

	candidate := pwd
	candidateName := candidate.Name()
	if filepath.Base(candidateName) != ".git" {
		candidateName = path.Join(candidateName, ".git")
	}
	for {
		candidate, err = os.Open(candidateName)
		if err == nil {
			return candidate, nil
			break
		}
		if !os.IsNotExist(err) {
			return nil, err
		}

		// This should not be the main condition of the for loop
		// just in case the filesystem root directory contains
		// a .git subdirectory
		// TODO check for mountpoint
		if candidateName == "/.git" {
			return nil, fmt.Errorf("not a git repository (or any parent up to root /")
		}
		candidateName, err = filepath.Abs(path.Join(candidateName, "..", "..", ".git"))
		candidate.Close()
	}
	return nil, fmt.Errorf("could not find the git repository")
}
