package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/cloudfoundry-incubator/buildpack_app_lifecycle/protocol"
	"github.com/cloudfoundry-incubator/candiedyaml"
)

const launcher = `
cd "$1"

if [ -d .profile.d ]; then
  for env_file in .profile.d/*; do
    source $env_file
  done
fi

shift

exec bash -c "$@"
`

func main() {
	if len(os.Args) < 4 {
		exitWithUsage()
	}

	dir := os.Args[1]
	startCommand := os.Args[2]
	metadata := os.Args[3]

	absDir, err := filepath.Abs(dir)
	if err == nil {
		dir = absDir
	}
	os.Setenv("HOME", dir)

	tmpDir, err := filepath.Abs(filepath.Join(dir, "..", "tmp"))
	if err == nil {
		os.Setenv("TMPDIR", tmpDir)
	}

	vcapAppEnv := map[string]interface{}{}
	err = json.Unmarshal([]byte(os.Getenv("VCAP_APPLICATION")), &vcapAppEnv)
	if err == nil {
		vcapAppEnv["host"] = "0.0.0.0"

		vcapAppEnv["instance_id"] = os.Getenv("INSTANCE_GUID")

		port, err := strconv.Atoi(os.Getenv("PORT"))
		if err == nil {
			vcapAppEnv["port"] = port
		}

		index, err := strconv.Atoi(os.Getenv("INSTANCE_INDEX"))
		if err == nil {
			vcapAppEnv["instance_index"] = index
		}

		mungedAppEnv, err := json.Marshal(vcapAppEnv)
		if err == nil {
			os.Setenv("VCAP_APPLICATION", string(mungedAppEnv))
		}
	}

	var command string
	if startCommand != "" {
		command = startCommand
	} else if metadata != "" {
		var executionMetadata protocol.ExecutionMetadata
		err := json.Unmarshal([]byte(metadata), &executionMetadata)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid metadata - %s", err)
			os.Exit(1)
		} else {
			command = executionMetadata.StartCommand
		}
	} else {
		command, err = startCommandFromStagingInfo("staging_info.yml")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid staging info - %s", err)
			os.Exit(1)
		}
	}

	if command == "" {
		exitWithUsage()
	}

	syscall.Exec("/bin/bash", []string{
		"bash",
		"-c",
		launcher,
		os.Args[0],
		dir,
		command,
	}, os.Environ())
}

func exitWithUsage() {
	fmt.Fprintf(os.Stderr, "Usage: %s <app directory> <start command> <metadata>", os.Args[0])
	os.Exit(1)
}

type stagingInfo struct {
	StartCommand string `yaml:"start_command"`
}

func startCommandFromStagingInfo(stagingInfoPath string) (string, error) {
	stagingInfoFile, err := os.Open(stagingInfoPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	defer stagingInfoFile.Close()

	info := stagingInfo{}

	err = candiedyaml.NewDecoder(stagingInfoFile).Decode(&info)
	if err != nil {
		return "", errors.New("invalid YAML")
	}

	return info.StartCommand, nil
}
