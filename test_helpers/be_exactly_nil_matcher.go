package test_helpers

import (
	"fmt"

	"github.com/onsi/gomega/types"
)

type exactlyNilMatcher struct{}

func BeExactlyNil() types.GomegaMatcher {
	return &exactlyNilMatcher{}
}

func (*exactlyNilMatcher) Match(actual interface{}) (success bool, err error) {
	if actual == nil {
		return true, nil
	} else {
		return false, nil
	}
}

func (*exactlyNilMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected %v to be exactly nil (not just a nil pointer or some other nil-like thing)", actual)
}

func (*exactlyNilMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected %v not to be exactly nil, but it really was nil (not just a nil pointer or some other nil-like thing)!", actual)
}
