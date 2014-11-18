package test_helpers

import (
	"flag"

	"github.com/codegangsta/cli"
)

func ContextFromArgsAndCommand(args []string, command cli.Command) *cli.Context {
	flagSet := FlagsetFromCommandAndArgs(command, args)
	return cli.NewContext(&cli.App{}, flagSet, flagSet)
}

func FlagsetFromCommandAndArgs(command cli.Command, args []string) *flag.FlagSet {
	flagSet := &flag.FlagSet{}

	for _, flag := range command.Flags {
		flag.Apply(flagSet)
	}
	flagSet.Parse(args)
	return flagSet
}
