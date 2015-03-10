package exit_handler_test

import (
	"fmt"
	"os"
	"syscall"

    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler"
)

var _ = Describe("ExitHandler", func() {
	var buffer *gbytes.Buffer

	BeforeEach(func() {
		buffer = gbytes.NewBuffer()
	})

	It("Executes exit handlers on os.Interupts", func() {
		exitFunc := func(code int) {
			buffer.Write([]byte(fmt.Sprintf("Exit-Code=%d", code)))
		}

		signalChan := make(chan os.Signal)
		exitHandler := exit_handler.New(signalChan, exitFunc)
		go exitHandler.Run()

		exitHandler.OnExit(func() {
			buffer.Write([]byte("handler1"))
		})

		exitHandler.OnExit(func() {
			buffer.Write([]byte("handler2"))
		})

		signalChan <- syscall.SIGHUP

		Consistently(buffer).ShouldNot(gbytes.Say("handler"))

		signalChan <- os.Interrupt

		Eventually(buffer).Should(gbytes.Say("handler1"))
		Eventually(buffer).Should(gbytes.Say("handler2"))
		Eventually(buffer).Should(gbytes.Say("Exit-Code=130"))
	})

	Describe("Exit", func() {
		It("triggers a system exit after calling all the exit funcs ", func() {
			exitFunc := func(code int) {
				buffer.Write([]byte(fmt.Sprintf("Exit-Code=%d", code)))
			}

			signalChan := make(chan os.Signal)
			exitHandler := exit_handler.New(signalChan, exitFunc)
			go exitHandler.Run()

			exitHandler.OnExit(func() {
				buffer.Write([]byte("handler1"))
			})

			exitHandler.OnExit(func() {
				buffer.Write([]byte("handler2"))
			})

			exitHandler.Exit(222)

			Eventually(buffer).Should(gbytes.Say("handler1"))
			Eventually(buffer).Should(gbytes.Say("handler2"))
			Eventually(buffer).Should(gbytes.Say("Exit-Code=222"))
		})
	})
})
