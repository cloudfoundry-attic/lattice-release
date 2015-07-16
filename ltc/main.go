package main

import (
	"os"

	"github.com/cloudfoundry-incubator/lattice/ltc/setup_cli"
)

func main() {
	defer os.Stdout.Write([]byte("\n"))

	if err := setup_cli.NewCliApp().Run(os.Args); err != nil {
		os.Exit(1)
	}
}
