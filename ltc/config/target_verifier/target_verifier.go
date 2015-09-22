package target_verifier

import (
	"github.com/cloudfoundry-incubator/lattice/ltc/receptor_client"
	"github.com/cloudfoundry-incubator/receptor"
)

//go:generate counterfeiter -o fake_target_verifier/fake_target_verifier.go . TargetVerifier
type TargetVerifier interface {
	VerifyTarget(name string) (up bool, auth bool, err error)
}

func New(receptorClientCreator receptor_client.Creator) TargetVerifier {
	return &targetVerifier{receptorClientCreator}
}

type targetVerifier struct {
	receptorClientCreator receptor_client.Creator
}

func (t *targetVerifier) VerifyTarget(target string) (up, auth bool, err error) {
	receptorClient := t.receptorClientCreator.CreateReceptorClient(target)
	_, err = receptorClient.DesiredLRPs()

	if err != nil {
		receptorErr, ok := err.(receptor.Error)

		if !ok {
			return false, false, err
		}

		if receptorErr.Type == receptor.Unauthorized {
			return true, false, nil
		} else {
			// TODO: poor interface for return values: "true, false, err"
			return true, false, err
		}
	}

	return true, true, nil
}
