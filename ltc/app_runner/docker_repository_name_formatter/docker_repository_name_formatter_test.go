package docker_repository_name_formatter_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/lattice/ltc/app_runner/docker_repository_name_formatter"
)

var _ = Describe("DockerRepositoryNameFormatter", func() {

	Describe("FormatForReceptor", func() {

		Context("with a well-formed docker repo name", func() {

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

			Context("with a shortened official docker repo name", func() {
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

			Context("when the docker image reference uses docker.io as the index", func() {
				It("returns the index name and remote name", func() {
					formattedName, err := docker_repository_name_formatter.FormatForReceptor("docker.io/app-name")
					Expect(err).ToNot(HaveOccurred())
					Expect(formattedName).To(Equal("docker://docker.io/library/app-name#latest"))
				})
			})

		})

		Context("with a non-standard docker registry name", func() {

			It("Converts it to a tagged url that receptor can use as a rootfs", func() {
				formattedName, err := docker_repository_name_formatter.FormatForReceptor("docker.gocd.cf-app.com:5000/my-app")
				Expect(err).NotTo(HaveOccurred())
				Expect(formattedName).To(Equal("docker://docker.gocd.cf-app.com:5000/my-app#latest"))
			})

			Context("when a tag is specified", func() {
				It("Converts it to a tagged url that receptor can use as a rootfs", func() {
					formattedName, err := docker_repository_name_formatter.FormatForReceptor("docker.gocd.cf-app.com:5000/my-app:test")
					Expect(err).NotTo(HaveOccurred())
					Expect(formattedName).To(Equal("docker://docker.gocd.cf-app.com:5000/my-app#test"))
				})
			})
		})

		Context("when the repository name fails docker validation", func() {
			Context("when the docker image reference contains the scheme", func() {
				It("returns an error", func() {
					_, err := docker_repository_name_formatter.FormatForReceptor("docker:///library/ubuntu")
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("docker URI [docker:///library/ubuntu] should not contain scheme"))
				})
			})

			It("returns an error for an invalid repository name", func() {
				_, err := docker_repository_name_formatter.FormatForReceptor("¥¥¥¥¥suchabadname¥¥¥¥¥")

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Invalid repository name (¥¥¥¥¥suchabadname¥¥¥¥¥), only [a-z0-9-_.] are allowed"))
			})

			It("returns an error for an invalid namespace name", func() {
				dockerImageReference := "jim/my-docker-app"
				_, err := docker_repository_name_formatter.FormatForReceptor(dockerImageReference)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Invalid namespace name (jim). Cannot be fewer than 4 or more than 30 characters."))
			})
		})
	})

	Describe("ParseRepoNameAndTagFromImageReference", func() {

		Context("with a well-formed docker repo name", func() {

			Context("with a shortened official docker repo name", func() {
				It("parses the repo and tag from a docker image reference", func() {
					dockerImageReference := "my-docker-app:test"
					indexName, remoteName, tag, err := docker_repository_name_formatter.ParseRepoNameAndTagFromImageReference(dockerImageReference)

					Expect(err).ToNot(HaveOccurred())
					Expect(indexName).To(BeEmpty())
					Expect(remoteName).To(Equal("library/my-docker-app"))
					Expect(tag).To(Equal("test"))
				})

				It("parses the repo and defaults tag to latest for docker image reference without a tag", func() {
					dockerImageReference := "my-docker-app"
					indexName, remoteName, tag, err := docker_repository_name_formatter.ParseRepoNameAndTagFromImageReference(dockerImageReference)

					Expect(err).ToNot(HaveOccurred())
					Expect(indexName).To(BeEmpty())
					Expect(remoteName).To(Equal("library/my-docker-app"))
					Expect(tag).To(Equal("latest"))
				})
			})

			Context("with a fully qualified docker repo name", func() {
				It("parses the repo and tag from a docker image reference", func() {
					dockerImageReference := "jimbo/my-docker-app:test"
					indexName, remoteName, tag, err := docker_repository_name_formatter.ParseRepoNameAndTagFromImageReference(dockerImageReference)

					Expect(err).ToNot(HaveOccurred())
					Expect(indexName).To(BeEmpty())
					Expect(remoteName).To(Equal("jimbo/my-docker-app"))
					Expect(tag).To(Equal("test"))
				})

				It("parses the repo and defaults tag to latest for docker image reference without a tag", func() {
					dockerImageReference := "jimbo/my-docker-app"
					indexName, remoteName, tag, err := docker_repository_name_formatter.ParseRepoNameAndTagFromImageReference(dockerImageReference)

					Expect(err).ToNot(HaveOccurred())
					Expect(indexName).To(BeEmpty())
					Expect(remoteName).To(Equal("jimbo/my-docker-app"))
					Expect(tag).To(Equal("latest"))
				})
			})

			Context("when the docker image reference uses docker.io as the index", func() {
				It("returns the index name, remote name, and tag", func() {
					dockerImageReference := "docker.io/app-name"
					indexName, remoteName, tag, err := docker_repository_name_formatter.ParseRepoNameAndTagFromImageReference(dockerImageReference)

					Expect(err).ToNot(HaveOccurred())
					Expect(indexName).To(Equal("docker.io"))
					Expect(remoteName).To(Equal("library/app-name"))
					Expect(tag).To(Equal("latest"))
				})
			})
		})

		Context("with a non-standard docker registry name", func() {
			It("Converts it to a tagged url that receptor can use as a rootfs", func() {
				indexName, remoteName, tag, err := docker_repository_name_formatter.ParseRepoNameAndTagFromImageReference("docker.gocd.cf-app.com:5000/my-app")

				Expect(err).NotTo(HaveOccurred())
				Expect(indexName).To(Equal("docker.gocd.cf-app.com:5000"))
				Expect(remoteName).To(Equal("my-app"))
				Expect(tag).To(Equal("latest"))
			})

			Context("when a tag is specified", func() {
				It("Converts it to a tagged url that receptor can use as a rootfs", func() {
					indexName, remoteName, tag, err := docker_repository_name_formatter.ParseRepoNameAndTagFromImageReference("docker.gocd.cf-app.com:5000/my-app:test")

					Expect(err).NotTo(HaveOccurred())
					Expect(indexName).To(Equal("docker.gocd.cf-app.com:5000"))
					Expect(remoteName).To(Equal("my-app"))
					Expect(tag).To(Equal("test"))
				})
			})
		})

		Context("when the repository name fails docker validation", func() {
			It("returns an error for an invalid repository name", func() {
				_, _, _, err := docker_repository_name_formatter.ParseRepoNameAndTagFromImageReference("¥¥¥¥¥suchabadname¥¥¥¥¥")

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Invalid repository name (¥¥¥¥¥suchabadname¥¥¥¥¥), only [a-z0-9-_.] are allowed"))
			})

			It("returns an error for an invalid namespace name", func() {
				dockerImageReference := "jim/my-docker-app"
				_, _, _, err := docker_repository_name_formatter.ParseRepoNameAndTagFromImageReference(dockerImageReference)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Invalid namespace name (jim). Cannot be fewer than 4 or more than 30 characters."))
			})
		})

	})
})
