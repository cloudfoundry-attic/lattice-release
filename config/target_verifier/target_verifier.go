package target_verifier

import "github.com/cloudfoundry-incubator/receptor"

type TargetVerifier interface {
	RequiresAuth(name string) bool
}

func New(receptorClientFactory func(target string) receptor.Client) TargetVerifier {
	return &targetVerifier{receptorClientFactory}
}

type targetVerifier struct {
	receptorClientFactory func(target string) receptor.Client
}

func (t *targetVerifier) RequiresAuth(target string) bool {
	receptorClient := t.receptorClientFactory(target)

	_, err := receptorClient.DesiredLRPs()
	if err != nil {
		return true
	}

	return false
}
