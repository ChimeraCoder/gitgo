package gitgo

import (
	"log"
	"os"
	"path"
)

var RepoDir = path.Join("test_data", ".git")

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
