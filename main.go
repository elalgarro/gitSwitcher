package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func sanityCheck() (inRepo bool) {
	_, err := exec.Command("git", "rev-parse").Output()
	//above command errors if we are not in a git repo.
	return err == nil
}

func main() {
	if !sanityCheck() {
		os.Stderr.WriteString("must be in Git repo")
		os.Exit(1)
		return
	}
	for _, arg := range os.Args {
		fmt.Println(arg)
	}
	err := gatherBranches()

	if err != nil {
		os.Stderr.Write([]byte(err.Error()))
		os.Exit(1)
	}
}

// print branches

func gatherBranches() error {
	val, err := exec.Command("git", "branch").Output()
	if err != nil {
		return err
	}
	branches := strings.Split(string(val), "\n")

	for _, branch := range branches {
		fmt.Printf("Branch: %s", branch)
	}
	return nil
}

func intitalOptions(opts []string) {}
