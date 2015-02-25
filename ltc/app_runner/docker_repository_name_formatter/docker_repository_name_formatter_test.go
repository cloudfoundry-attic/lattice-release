package docker_repository_name_formatter_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/lattice/ltc/app_runner/docker_repository_name_formatter"
)

var _ = Describe("FormatForReceptor", func() {

	Context("With a well-formed docker repo name", func() {

		Context("with a fully qualified docker repo name", func() {
			It("Formats it as a url that the receptor can use as a rootfs", func() {
				formattedName, err := docker_repository_name_formatter.FormatForReceptor("jimbo/my-docker-app")
				Expect(err).NotTo(HaveOccurred())
				Expect(formattedName).To(Equal("docker:///jimbo/my-docker-app#latest"))
			})

			Context("when a tag is specified", func() {

				It("Converts it to a tagged url that receptor can use as a rootfs", func() {
					formattedName, err := docker_repository_name_formatter.FormatForReceptor("jimbo/my-docker-app:test")
					Expect(err).NotTo(HaveOccurred())
					Expect(formattedName).To(Equal("docker:///jimbo/my-docker-app#test"))
				})
			})

		})

		Context("With a shortened official docker repo name", func() {
			It("Formats it as a url that the receptor can use as a rootfs", func() {
				formattedName, err := docker_repository_name_formatter.FormatForReceptor("ubuntu")
				Expect(err).NotTo(HaveOccurred())
				Expect(formattedName).To(Equal("docker:///library/ubuntu#latest"))
			})

			It("Converts it to a tagged url that receptor can use as a rootfs", func() {
				formattedName, err := docker_repository_name_formatter.FormatForReceptor("ubuntu:test")
				Expect(err).NotTo(HaveOccurred())
				Expect(formattedName).To(Equal("docker:///library/ubuntu#test"))
			})
		})

	})

	Context("With a malformed docker repo name", func() {
		It("returns an error", func() {
			_, err := docker_repository_name_formatter.FormatForReceptor("¥¥¥¥¥suchabadname¥¥¥¥¥")
			Expect(err).To(HaveOccurred())
		})
	})

})
