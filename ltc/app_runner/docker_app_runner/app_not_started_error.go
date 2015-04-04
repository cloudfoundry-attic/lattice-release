package docker_app_runner

import "fmt"

type appNotStartedError string

func newAppNotStartedError(appName string) appNotStartedError {
	return appNotStartedError(appName)
}

func (appName appNotStartedError) Error() string {
	return fmt.Sprintf("%s is not started. Please start an app first", string(appName))
}
