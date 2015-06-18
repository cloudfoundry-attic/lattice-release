package buildpackrunner

// Used to generate YAML file read by the DEA
type DeaStagingInfo struct {
	DetectedBuildpack string `yaml:"detected_buildpack"`
	StartCommand      string `yaml:"start_command"`
}
