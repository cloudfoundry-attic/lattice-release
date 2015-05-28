package setup_cli

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/codegangsta/cli"
)

func InjectHelpTemplate(badFlags string) {
	cli.CommandHelpTemplate = fmt.Sprintf(`%sNAME:
   {{join .Names ", "}} - {{.Usage}}
{{with .ShortName}}
ALIAS:
   {{.Aliases}}
{{end}}
USAGE:
   {{.Description}}{{with .Flags}}
OPTIONS:
{{range .}}   {{.}}
{{end}}{{else}}
{{end}}`, badFlags)
}

func MatchArgAndFlags(flags []string, args []string) string {
	var badFlag string
	var lastPassed bool
	multipleFlagErr := false
Loop:
	for _, arg := range args {
		prefix := ""
		//only take flag name, ignore value after '='
		arg = strings.Split(arg, "=")[0]
		if arg == "--h" || arg == "-h" || arg == "--help" || arg == "-help" {
			continue Loop
		}
		if strings.HasPrefix(arg, "--") {
			prefix = "--"
		} else if strings.HasPrefix(arg, "-") {
			prefix = "-"
		}
		arg = strings.TrimLeft(arg, prefix)
		//skip verification for negative integers, e.g. -i -10
		if lastPassed {
			lastPassed = false
			if _, err := strconv.ParseInt(arg, 10, 32); err == nil {
				continue Loop
			}
		}
		if prefix != "" {
			for _, flag := range flags {
				for _, f := range strings.Split(flag, ", ") {
					flag = strings.TrimSpace(f)
					if flag == arg {
						lastPassed = true
						continue Loop
					}
				}
			}
			if badFlag == "" {
				badFlag = fmt.Sprintf("\"%s%s\"", prefix, arg)
			} else {
				multipleFlagErr = true
				badFlag = badFlag + fmt.Sprintf(", \"%s%s\"", prefix, arg)
			}
		}
	}
	if multipleFlagErr && badFlag != "" {
		badFlag = fmt.Sprintf("%s %s", "Unknown flags:", badFlag)
	} else if badFlag != "" {
		badFlag = fmt.Sprintf("%s %s", "Unknown flag", badFlag)
	}
	return badFlag
}

func RequestHelp(args []string) bool {
	for _, v := range args {
		if v == "-h" || v == "--help" {
			return true
		}
	}
	return false
}

func CallCoreCommand(args []string, cliApp *cli.App) {
	err := cliApp.Run(args)
	if err != nil {
		os.Exit(1)
	}
}

func GetCommandFlags(app *cli.App, command string) []string {
	cmd, err := GetByCmdName(app, command)
	if err != nil {
		return []string{}
	}
	var flags []string
	for _, flag := range cmd.Flags {
		switch t := flag.(type) {
		default:
		case cli.StringSliceFlag:
			flags = append(flags, t.Name)
		case cli.IntFlag:
			flags = append(flags, t.Name)
		case cli.StringFlag:
			flags = append(flags, t.Name)
		case cli.BoolFlag:
			flags = append(flags, t.Name)
		case cli.DurationFlag:
			flags = append(flags, t.Name)
		}
	}
	return flags
}

func GetByCmdName(app *cli.App, cmdName string) (cmd *cli.Command, err error) {
	cmd = app.Command(cmdName)
	if cmd == nil {
		for _, c := range app.Commands {
			if c.ShortName == cmdName {
				return &c, nil
			}
		}
		err = errors.New("Command not found")
	}
	return
}
