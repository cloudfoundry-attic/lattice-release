package command_factory

import (
	"io"

	"github.com/codegangsta/cli"
	"github.com/pivotal-cf-experimental/diego-edge-cli/config"
)

type commandFactory struct {
	config *config.Config
	output io.Writer
}

func NewConfigCommandFactory(config *config.Config, output io.Writer) *commandFactory {
	return &commandFactory{config, output}
}

func (c *commandFactory) MakeSetApiCommand() cli.Command {
	var startCommand = cli.Command{
		Name:        "api",
		ShortName:   "a",
		Description: "Set a target api location",
		Usage:       "diego-edge-cli api API_PATH",
		Action:      c.setApi,
		Flags:       []cli.Flag{},
	}

	return startCommand
}

func (c *commandFactory) setApi(context *cli.Context) {
	api := context.Args().First()

	if api == "" {
		c.output.Write([]byte("Incorrect Usage\n"))
		return
	}

	err := c.config.SetApi(api)
	if err != nil {
		c.output.Write([]byte(err.Error()))
		return
	}

	c.output.Write([]byte("Api Location Set\n"))
}
