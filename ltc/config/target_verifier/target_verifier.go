package target_verifier

import "github.com/cloudfoundry-incubator/receptor"

//go:generate counterfeiter -o fake_target_verifier/fake_target_verifier.go . TargetVerifier
type TargetVerifier interface {
	VerifyTarget(name string) (up bool, auth bool, err error)
}

func New(receptorClientFactory func(target string) receptor.Client) TargetVerifier {
	return &targetVerifier{receptorClientFactory}
}

type targetVerifier struct {
	receptorClientFactory func(target string) receptor.Client
}

func (t *targetVerifier) VerifyTarget(target string) (up, auth bool, err error) {
	receptorClient := t.receptorClientFactory(target)
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
