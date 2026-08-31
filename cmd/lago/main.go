package main

import (
	"os"

	"github.com/getlago/lago-cli/internal/cli"
)

func main() {
	os.Exit(cli.Execute(os.Stdin, os.Stdout, os.Stderr))
}
