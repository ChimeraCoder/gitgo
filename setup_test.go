package gitgo

import (
	"log"
	"os"
	"path"
)

var RepoDir *os.File

func init() {

	_, err := os.Stat(path.Join("test_data", ".git"))
	if err != nil {
		if !os.IsNotExist(err) {
			log.Fatal(err)
		}
		err := os.Symlink(path.Join("dot_git"), path.Join("test_data", ".git"))
		if err != nil {
			log.Fatal(err)
		}
	}
}

func init() {
	var err error
	RepoDir, err = os.Open(path.Join("test_data", ".git"))
	if err != nil {
		panic(err)
	}

}
