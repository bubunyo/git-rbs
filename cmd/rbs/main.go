package main

import (
	"os"

	rbs "github.com/bubunyo/git-rbs"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "worktree" {
		rbs.RunWorktree()
		return
	}
	rbs.Run()
}
