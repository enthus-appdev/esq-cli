package main

import (
	"os"

	"github.com/enthus-appdev/esq-cli/internal/cmd"
)

var (
	version = "dev"
	commit  = "none"
)

func main() {
	exitCode := cmd.Execute(version, commit)
	os.Exit(exitCode)
}
