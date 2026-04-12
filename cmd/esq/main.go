package main

import (
	"os"

	"github.com/enthus-appdev/esq-cli/internal/cmd"
)

var version = "dev"

func main() {
	exitCode := cmd.Execute(version)
	os.Exit(exitCode)
}
