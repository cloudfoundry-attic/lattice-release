package main

import (
	"os"

	"github.com/cloudfoundry-incubator/lattice/ltc/setup_cli"
)

func main() {
	defer os.Stdout.Write([]byte("\n"))
	var badFlags string
	cliApp := setup_cli.NewCliApp()

	if len(os.Args) > 1 {
		flags := GetCommandFlags(cliApp, os.Args[1])
		badFlags = MatchArgAndFlags(flags, os.Args[2:])
		if badFlags != "" {
			badFlags = badFlags + "\n\n"
		}
	}

	InjectHelpTemplate(badFlags)

	if len(os.Args) == 1 || os.Args[1] == "help" || os.Args[1] == "h" || RequestHelp(os.Args[1:]) {
		cliApp.Run(os.Args)
	} else {
		CallCoreCommand(os.Args[0:], cliApp)
	}
}
