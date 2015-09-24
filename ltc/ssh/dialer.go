package ssh

import (
	"fmt"
	"io"

	config_package "github.com/cloudfoundry-incubator/lattice/ltc/config"
	"github.com/cloudfoundry-incubator/lattice/ltc/ssh/sshapi"
)

type AppDialer struct{}

func (*AppDialer) Dial(appName string, instanceIndex int, config *config_package.Config) (Client, error) {
	diegoSSHUser := fmt.Sprintf("diego:%s/%d", appName, instanceIndex)
	address := fmt.Sprintf("%s:2222", config.Target())

	client, err := sshapi.New(diegoSSHUser, config.Username(), config.Password(), address)
	if err != nil {
		return nil, err
	}

	return client, nil
}

//go:generate counterfeiter -o mocks/fake_client.go . Client
type Client interface {
	Open(width, height int, desirePTY bool) (*sshapi.Session, error)
	Forward(localConn io.ReadWriteCloser, remoteAddress string) error
}
