package terminal_test

import (
	"io"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"github.com/cloudfoundry-incubator/lattice/ltc/terminal"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal/password_reader"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal/password_reader/fake_password_reader"
	"github.com/cloudfoundry-incubator/lattice/ltc/test_helpers"
)

var _ = Describe("UI", func() {

	var (
		stdinReader        *io.PipeReader
		stdinWriter        *io.PipeWriter
		outputBuffer       *gbytes.Buffer
		fakePasswordReader *fake_password_reader.FakePasswordReader
		terminalUI         terminal.UI
	)

	BeforeEach(func() {
		stdinReader, stdinWriter = io.Pipe()
		outputBuffer = gbytes.NewBuffer()
		fakePasswordReader = &fake_password_reader.FakePasswordReader{}
		terminalUI = terminal.NewUI(stdinReader, outputBuffer, fakePasswordReader)
	})

	Describe("Instantiation", func() {
		It("instantiates a terminal", func() {
			Expect(terminalUI).ToNot(BeNil())

			_, readWriterOk := terminalUI.(io.ReadWriter)
			Expect(readWriterOk).To(BeTrue())

			_, passwordReaderOk := terminalUI.(password_reader.PasswordReader)
			Expect(passwordReaderOk).To(BeTrue())
		})
	})

	Describe("Output methods", func() {
		Describe("Say", func() {
			It("says the message to the terminal", func() {
				terminalUI.Say("Cloudy with a chance of meatballs")
				Expect(outputBuffer).To(test_helpers.Say("Cloudy with a chance of meatballs"))
			})
		})

		Describe("SayLine", func() {
			It("says the message to the terminal with a newline", func() {
				terminalUI.SayLine("Strange Clouds")
				Expect(outputBuffer).To(test_helpers.Say("Strange Clouds\n"))
			})
		})

		Describe("SayIncorrectUsage", func() {
			Context("when no message is passed", func() {
				It("outputs incorrect usage", func() {
					terminalUI.SayIncorrectUsage("")
					Expect(outputBuffer).To(test_helpers.SayIncorrectUsage())
				})
			})
			Context("when a message is passed", func() {
				It("outputs incorrect usage with the message", func() {
					terminalUI.SayIncorrectUsage("You did that thing wrong")
					Expect(outputBuffer).To(test_helpers.Say("Incorrect Usage: You did that thing wrong"))
				})
			})
		})

		Describe("SayNewLine", func() {
			It("says a newline", func() {
				terminalUI.SayNewLine()
				Expect(outputBuffer).To(test_helpers.SayNewLine())
			})
		})
	})

	Describe("Input Methods", func() {
		Describe("Prompt", func() {
			It("Prompts the user for input", func() {

				answerChan := make(chan string)
				go func() {
					defer GinkgoRecover()

					answerChan <- terminalUI.Prompt("Nickname: ")
					close(answerChan)
				}()

				Eventually(outputBuffer).Should(test_helpers.Say("Nickname: "))
				stdinWriter.Write([]byte("RockStar\n"))

				Expect(<-answerChan).To(Equal("RockStar"))
				Eventually(answerChan).Should(BeClosed())
			})
		})

		Describe("PasswordReader PromptForPassword", func() {
			It("Calls to PasswordReader, which contains untested content", func() {
				fakePasswordReader.PromptForPassword("Password: ")

				Expect(fakePasswordReader.PromptForPasswordCallCount()).To(Equal(1))
				Expect(fakePasswordReader.PromptForPasswordArgsForCall(0)).To(Equal("Password: "))
			})
		})
	})
})
