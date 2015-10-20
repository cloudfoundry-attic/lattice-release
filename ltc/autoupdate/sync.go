package autoupdate

import (
	"fmt"
	"io"
	"net/http"
	"os"

	config_package "github.com/cloudfoundry-incubator/lattice/ltc/config"
)

//go:generate counterfeiter -o mocks/fake_file_swapper.go . FileSwapper
type FileSwapper interface {
	GetTempFile() (*os.File, error)
	SwapTempFile(destPath, srcPath string) error
}

type Sync struct {
	fileSwapper FileSwapper
}

func NewSync(fileSwapper FileSwapper) *Sync {
	return &Sync{fileSwapper}
}

func (s *Sync) SyncLTC(ltcPath string, arch string, config *config_package.Config) error {
	response, err := http.DefaultClient.Get(fmt.Sprintf("%s/v1/sync/%s/ltc", config.Receptor(), arch))
	if err != nil {
		return fmt.Errorf("failed to connect to receptor: %s", err.Error())
	}
	if response.StatusCode != 200 {
		return fmt.Errorf("failed to download ltc: %s", response.Status)
	}

	tmpFile, err := s.fileSwapper.GetTempFile()
	if err != nil {
		return fmt.Errorf("failed to open temp file: %s", err.Error())
	}
	defer tmpFile.Close()

	_, err = io.Copy(tmpFile, response.Body)
	if err != nil {
		return fmt.Errorf("failed to write to temp ltc: %s", err.Error())
	}

	err = s.fileSwapper.SwapTempFile(ltcPath, tmpFile.Name())
	if err != nil {
		return fmt.Errorf("failed to swap ltc: %s", err.Error())
	}

	return nil
}
