package main

import (
	"os"

	"github.com/dpemmons/mcprinter/cmd"
)

func main() {
	if err := cmd.NewRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
