package cli_app_factory

import (
	"fmt"
	"io"
	"reflect"
	"strings"
	"text/tabwriter"
	"text/template"
	"unicode/utf8"

	"github.com/codegangsta/cli"
)

type groupedCommands struct {
	Name             string
	CommandSubGroups [][]cmdPresenter
}

func (c groupedCommands) SubTitle(name string) string {
	return (name + ":")
}

type cmdPresenter struct {
	Name        string
	Description string
}

func presentCmdName(cmd cli.Command) (name string) {
	name = strings.Join(cmd.Names(), ", ")
	return
}

type appPresenter struct {
	cli.App
	Commands []groupedCommands
}

func newAppPresenter(app *cli.App) (presenter appPresenter) {
	maxNameLen := 0
	for _, cmd := range app.Commands {
		name := presentCmdName(cmd)
		if utf8.RuneCountInString(name) > maxNameLen {
			maxNameLen = len(name)
		}
	}

	presentCommand := func(commandName string) (presenter cmdPresenter) {
		cmd := app.Command(commandName)
		presenter.Name = presentCmdName(*cmd)

		padding := strings.Repeat(" ", maxNameLen-utf8.RuneCountInString(presenter.Name))
		presenter.Name = presenter.Name + padding
		presenter.Description = cmd.Usage

		return
	}

	presenter.Name = app.Name
	presenter.Flags = app.Flags
	presenter.Usage = app.Usage
	presenter.Version = app.Version
	presenter.Authors = app.Authors
	presenter.Commands = []groupedCommands{
		{
			Name: "TARGET LATTICE",
			CommandSubGroups: [][]cmdPresenter{
				{
					presentCommand("target"),
				},
			},
		}, {
			Name: "MANAGE DOCKER APPS",
			CommandSubGroups: [][]cmdPresenter{
				{
					presentCommand("create"),
				},
			},
		}, {
			Name: "MANAGE DROPLETS",
			CommandSubGroups: [][]cmdPresenter{
				{
					presentCommand("build-droplet"),
					presentCommand("export-droplet"),
					presentCommand("import-droplet"),
					presentCommand("launch-droplet"),
					presentCommand("list-droplets"),
					presentCommand("remove-droplet"),
				},
			},
		}, {
			Name: "MANAGE ALL APPS",
			CommandSubGroups: [][]cmdPresenter{
				{
					presentCommand("remove"),
					presentCommand("scale"),
					presentCommand("update-routes"),
				},
			},
		}, {
			Name: "TASKS",
			CommandSubGroups: [][]cmdPresenter{
				{
					presentCommand("cancel-task"),
					presentCommand("delete-task"),
					presentCommand("task"),
				},
			},
		}, {
			Name: "STREAM LOGS",
			CommandSubGroups: [][]cmdPresenter{
				{
					presentCommand("logs"),
				},
			},
		}, {
			Name: "SEE WHATS RUNNING",
			CommandSubGroups: [][]cmdPresenter{
				{
					presentCommand("cells"),
					presentCommand("list"),
					presentCommand("status"),
					presentCommand("visualize"),
				},
			},
		}, {
			Name: "ADVANCED",
			CommandSubGroups: [][]cmdPresenter{
				{
					presentCommand("submit-lrp"),
					presentCommand("submit-task"),
				},
			},
		}, {
			Name: "HELP AND DEBUG",
			CommandSubGroups: [][]cmdPresenter{
				{
					presentCommand("debug-logs"),
					presentCommand("help"),
					presentCommand("test"),
				},
			},
		},
	}

	return
}

func ShowHelp(w io.Writer, helpTemplate string, thingToPrint interface{}) {
	translatedTemplatedHelp := strings.Replace(helpTemplate, "{{", "[[", -1)
	translatedTemplatedHelp = strings.Replace(translatedTemplatedHelp, "[[", "{{", -1)

	switch thing := thingToPrint.(type) {
	case *cli.App:
		showAppHelp(w, translatedTemplatedHelp, thing)
	case cli.Command:
		commandPrintHelp(w, translatedTemplatedHelp, thing)
	default:
		panic(fmt.Errorf("Help printer has received something that is neither app nor command! The beast (%s) looks like this: %s", reflect.TypeOf(thing), thing))
	}
}

func showAppHelp(w io.Writer, helpTemplate string, appToPrint *cli.App) {
	presenter := newAppPresenter(appToPrint)
	tabWriter := tabwriter.NewWriter(w, 0, 8, 1, '\t', 0)
	t := template.Must(template.New("help").Parse(helpTemplate))
	err := t.Execute(tabWriter, presenter)
	if err != nil {
		panic(err)
	}
	tabWriter.Flush()
}

func commandPrintHelp(w io.Writer, templ string, data cli.Command) {
	funcMap := template.FuncMap{
		"join": strings.Join,
	}

	tabWriter := tabwriter.NewWriter(w, 0, 8, 1, '\t', 0)
	t := template.Must(template.New("help").Funcs(funcMap).Parse(templ))
	err := t.Execute(tabWriter, data)
	if err != nil {
		panic(err)
	}
	tabWriter.Flush()
}
