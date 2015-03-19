package terminal_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler/fake_exit_handler"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal"
)

var _ = Describe("PasswordReader", func() {

	It("instantiates a password reader", func() {

		exitHandler := &fake_exit_handler.FakeExitHandler{}
		passwordReader := terminal.NewPasswordReader(exitHandler)

		Expect(passwordReader).ToNot(BeNil())

	})

})
