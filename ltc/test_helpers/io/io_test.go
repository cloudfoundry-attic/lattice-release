package io_test

import (
	. "github.com/cloudfoundry-incubator/lattice/ltc/test_helpers/io"
	"os"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("io helpers", func() {
	It("will never overflow the pipe", func() {
		str := strings.Repeat("z", 75000)
		output := CaptureOutput(func() {
			os.Stdout.Write([]byte(str))
		})

		Expect(output).To(Equal([]string{str}))
	})
})
