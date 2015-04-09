package main

import (
	"fmt"
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
		fmt.Println(result)

	default:
		log.Fatalf("no such command: %s", module)
	}
}
