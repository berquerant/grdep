package main

import (
	"os"

	"github.com/berquerant/grdep/cmd/subcmd"
)

func main() {
	if err := subcmd.Execute(); err != nil {
		os.Exit(1)
	}
}
