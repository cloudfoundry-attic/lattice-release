package target_verifier

import "github.com/cloudfoundry-incubator/receptor"

type TargetVerifier interface {
	ValidateAuthorization(name string) (bool, error)
}

func New(receptorClientFactory func(target string) receptor.Client) TargetVerifier {
	return &targetVerifier{receptorClientFactory}
}

type targetVerifier struct {
	receptorClientFactory func(target string) receptor.Client
}

func (t *targetVerifier) ValidateAuthorization(target string) (bool, error) {
	receptorClient := t.receptorClientFactory(target)

	_, err := receptorClient.DesiredLRPs()
	if err != nil {
		if err.Error() == receptor.Unauthorized {
			return false, nil
		}
		return false, err
	}

	return true, nil
}
