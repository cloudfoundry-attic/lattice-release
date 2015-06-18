package main

import (
	"flag"
	"os"

	"github.com/cloudfoundry-incubator/buildpack_app_lifecycle"
	"github.com/cloudfoundry-incubator/buildpack_app_lifecycle/buildpackrunner"
)

func main() {
	config := buildpack_app_lifecycle.NewLifecycleBuilderConfig([]string{}, false, false)

	if err := config.Parse(os.Args[1:len(os.Args)]); err != nil {
		println(err.Error())
		os.Exit(1)
	}

	if err := config.Validate(); err != nil {
		println(err.Error())
		usage()
	}

	runner := buildpackrunner.New(&config)

	err := runner.Run()
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}

	os.Exit(0)
}

func usage() {
	flag.PrintDefaults()
	os.Exit(1)
}
