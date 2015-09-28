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
	var (
		buffer      *gbytes.Buffer
		exitFunc    func(int)
		signalChan  chan os.Signal
		exitHandler exit_handler.ExitHandler
	)

	BeforeEach(func() {
		buffer = gbytes.NewBuffer()

		exitFunc = func(code int) {
			buffer.Write([]byte(fmt.Sprintf("Exit-Code=%d", code)))
		}

		signalChan = make(chan os.Signal)
		exitHandler = exit_handler.New(signalChan, exitFunc)
		go exitHandler.Run()

		exitHandler.OnExit(func() {
			buffer.Write([]byte("handler1 "))
		})

		exitHandler.OnExit(func() {
			buffer.Write([]byte("handler2 "))
		})
	})

	Context("Signals", func() {
		It("Executes exit handlers on os.Interupts", func() {
			signalChan <- os.Interrupt
			Eventually(buffer).Should(gbytes.Say("handler1 handler2 Exit-Code=160"))
		})

		It("Executes exit handlers on SIGTERM", func() {
			signalChan <- syscall.SIGTERM
			Eventually(buffer).Should(gbytes.Say("handler1 handler2 Exit-Code=160"))
		})
	})

	Describe("Exit", func() {
		It("triggers a system exit after calling all the exit funcs ", func() {
			exitHandler.Exit(222)
			Eventually(buffer).Should(gbytes.Say("handler1 handler2 Exit-Code=222"))
		})
	})
})
