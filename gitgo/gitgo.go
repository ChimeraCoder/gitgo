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
	if len(args) <= 2 {
		fmt.Println("gitgo: must provide either cat-file or log subcommand")
		return
	}
	module := args[1]
	switch module {
	case "cat-file":
		if len(args) < 3 {
			fmt.Println("must specify hash with `cat-file`")
			os.Exit(1)
		}
		hash := args[2]
		result, err := gitgo.CatFile(gitgo.SHA(hash))
		if err != nil {
			log.Fatal(err)
		}
		io.Copy(os.Stdout, result)

	case "log":
		if len(args) < 3 {
			fmt.Println("must specify commit name with `log`")
			os.Exit(1)
		}
		hash := gitgo.SHA(args[2])
		commits, err := gitgo.Log(hash, "")
		if err != nil {
			log.Fatal(err)
		}
		b := bytes.NewBuffer(nil)
		for _, commit := range commits {
			fmt.Fprintf(b, "commit %s\nAuthor: %s\nDate:   %s\n\n    %s\n", commit.Name, commit.Author, commit.AuthorDate.Format(gitgo.RFC2822), bytes.Replace(commit.Message, []byte("\n"), []byte("\n    "), -1))
		}
		io.Copy(os.Stdout, b)
	default:
		log.Fatalf("no such command: %s", module)
	}
}
