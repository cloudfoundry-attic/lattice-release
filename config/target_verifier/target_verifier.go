package target_verifier

import "github.com/cloudfoundry-incubator/receptor"

type TargetVerifier interface {
	ValidateReceptor(name string) bool
}

func New(receptorClientFactory func(target string) receptor.Client) TargetVerifier {
	return &targetVerifier{receptorClientFactory}
}

type targetVerifier struct {
	receptorClientFactory func(target string) receptor.Client
}

func (t *targetVerifier) ValidateReceptor(target string) bool {
	receptorClient := t.receptorClientFactory(target)

	_, err := receptorClient.DesiredLRPs()
	if err != nil {
		return false
	}

	return true
}
