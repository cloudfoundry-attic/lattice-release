package buildpack_app_lifecycle

import "strings"

const (
	DetectFailMsg     = "None of the buildpacks detected a compatible application"
	CompileFailMsg    = "Failed to compile droplet"
	ReleaseFailMsg    = "Failed to build droplet release"
	DETECT_FAIL_CODE  = 222
	COMPILE_FAIL_CODE = 223
	RELEASE_FAIL_CODE = 224
)

func ExitCodeFromError(err error) int {
	errMsg := err.Error()
	switch {
	case strings.Contains(errMsg, DetectFailMsg):
		return DETECT_FAIL_CODE
	case strings.Contains(errMsg, CompileFailMsg):
		return COMPILE_FAIL_CODE
	case strings.Contains(errMsg, ReleaseFailMsg):
		return RELEASE_FAIL_CODE
	default:
		return 1
	}
}

type StagingResult struct {
	BuildpackKey         string            `json:"buildpack_key,omitempty"`
	DetectedBuildpack    string            `json:"detected_buildpack"`
	ExecutionMetadata    string            `json:"execution_metadata"`
	DetectedStartCommand map[string]string `json:"detected_start_command"`
}
