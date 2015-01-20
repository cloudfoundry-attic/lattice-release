package docker_repository_name_formatter_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf-experimental/lattice-cli/app_runner/docker_repository_name_formatter"
)

var _ = Describe("FormatForReceptor", func() {

	Context("With a well-formed docker repo name", func() {
		It("Formats a fully qualified docker repo name as a url that the receptor can use as a rootfs", func() {
			formattedName, err := docker_repository_name_formatter.FormatForReceptor("jimbo/my-docker-app")
			Expect(err).NotTo(HaveOccurred())
			Expect(formattedName).To(Equal("docker:///jimbo/my-docker-app"))
		})

		It("Formats a shortened official docker repo name as a url that the receptor can use as a rootfs", func() {
			formattedName, err := docker_repository_name_formatter.FormatForReceptor("ubuntu")
			Expect(err).NotTo(HaveOccurred())
			Expect(formattedName).To(Equal("docker:///library/ubuntu"))
		})
	})

	Context("With a malformed docker repo name", func() {
		It("returns an error", func() {
			_, err := docker_repository_name_formatter.FormatForReceptor("¥¥¥¥¥suchabadname¥¥¥¥¥")
			Expect(err).To(HaveOccurred())
		})
	})

})
