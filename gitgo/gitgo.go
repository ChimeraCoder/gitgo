package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/ChimeraCoder/gitgo"
)

func main() {
	args := os.Args
	module := args[1]
	switch module {
	case "cat-file":
		hash := args[2]
		result, err := gitgo.CatFile(gitgo.SHA(hash))
		if err != nil {
			log.Fatal(err)
		}
		io.Copy(os.Stdout, result)

	case "log":
		hash := gitgo.SHA(args[2])
		commits, err := gitgo.Log(hash, "")
		if err != nil {
			log.Fatal(err)
		}
		b := bytes.NewBuffer(nil)
		for _, commit := range commits {
			fmt.Fprintf(b, "commit: %s\n%s%s\n\n\n", commit.Name, commit.Author, commit.Message)
		}
		io.Copy(os.Stdout, b)
	default:
		log.Fatalf("no such command: %s", module)
	}
}
