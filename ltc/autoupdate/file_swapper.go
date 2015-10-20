package autoupdate

import (
	"fmt"
	"io/ioutil"
	"os"
)

type AppFileSwapper struct{}

func (*AppFileSwapper) GetTempFile() (*os.File, error) {
	return ioutil.TempFile("", "")
}

func (*AppFileSwapper) SwapTempFile(destPath, srcPath string) error {
	destInfo, err := os.Stat(destPath)
	if err != nil {
		return fmt.Errorf("failed to stat dest file: %s", err.Error())
	}

	err = os.Rename(srcPath, destPath)
	if err != nil {
		return fmt.Errorf("failed to rename file: %s", err.Error())
	}

	err = os.Chmod(destPath, destInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to change file permissions: %s", err.Error())
	}

	return nil
}
