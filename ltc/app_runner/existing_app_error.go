package app_runner

import "fmt"

type existingAppError string

func newExistingAppError(appName string) existingAppError {
	return existingAppError(appName)
}

func (appName existingAppError) Error() string {
	return fmt.Sprintf("%s is already running", string(appName))
}
