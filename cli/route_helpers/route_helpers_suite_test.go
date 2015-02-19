package route_helpers_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestRouteHelpers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "RouteHelpers Suite")
}
