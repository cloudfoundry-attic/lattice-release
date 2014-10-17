package whetstone_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/cloudfoundry-incubator/runtime-schema/models/factories"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const repUrlRelativeToExecutor string = "http://127.0.0.1:20515"

var _ = Describe("Diego Edge", func() {
	Context("when desiring a docker-based LRP", func() {

		var (
			processGuid string
			appName     string
			route       string
		)

		BeforeEach(func() {
			processGuid = factories.GenerateGuid()
			appName = factories.GenerateGuid()
			route = fmt.Sprintf("%s.%s", appName, domain)
		})

		AfterEach(func() {
			err := bbs.RemoveDesiredLRPByProcessGuid(processGuid)
			Expect(err).To(BeNil())

			Eventually(errorCheckForRoute(route), 20, 1).Should(HaveOccurred())
		})

		It("eventually runs on an executor", func() {
			err := desireLongRunningProcess(processGuid, appName, route)
			Expect(err).To(BeNil())

			Eventually(errorCheckForRoute(route), 20, 1).ShouldNot(HaveOccurred())
		})
	})

})

func errorCheckForRoute(route string) func() error {
	routeWithScheme := fmt.Sprintf("http://%s", route)
	return func() error {
		resp, err := http.DefaultClient.Get(routeWithScheme)
		if err != nil {
			return err
		}

		io.Copy(ioutil.Discard, resp.Body)
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return fmt.Errorf("Status code %d should be 200", resp.StatusCode)
		}

		return nil
	}
}

func desireLongRunningProcess(processGuid, appName, route string) error {
	return bbs.DesireLRP(models.DesiredLRP{
		Domain:      "whetstone",
		ProcessGuid: processGuid,
		Instances:   1,
		Stack:       "lucid64",
		RootFSPath:  "docker:///onsi/grace-busybox",
		Routes:      []string{route},
		MemoryMB:    128,
		DiskMB:      1024,
		Ports: []models.PortMapping{
			{ContainerPort: 8080},
		},
		Actions: []models.ExecutorAction{
			models.Parallel(
				models.ExecutorAction{
					models.RunAction{
						Path: "/grace",
						Env: []models.EnvironmentVariable{
							{Name: "VCAP_APPLICATION", Value: `{"instance_index":0}`},
							{Name: "PORT", Value: "8080"},
						},
					},
				},
				models.ExecutorAction{
					models.MonitorAction{
						Action: models.ExecutorAction{
							models.RunAction{ //The spy. Is this container healthy? running on 8080?
								Path: "echo",
								Args: []string{"http://127.0.0.1:8080"},
							},
						},
						HealthyThreshold:   1,
						UnhealthyThreshold: 1,
						HealthyHook: models.HealthRequest{ //Teel the rep where to call back to on exit 0 of spy
							Method: "PUT",
							URL: fmt.Sprintf(
								"%s/lrp_running/%s/PLACEHOLDER_INSTANCE_INDEX/PLACEHOLDER_INSTANCE_GUID",
								repUrlRelativeToExecutor,
								processGuid,
							),
						},
					},
				},
			),
		},
	})
}
