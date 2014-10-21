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

	"github.com/gorilla/websocket"
	"github.com/onsi/gomega/gbytes"

	"code.google.com/p/gogoprotobuf/proto"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
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
			appName = fmt.Sprintf("whetstone-%s", factories.GenerateGuid())
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

			outBuf := gbytes.NewBuffer()
			go streamAppLogsIntoGbytes(processGuid, outBuf)

			Eventually(outBuf, 2).Should(gbytes.Say("Diego Edge Docker App. Says Hello"))
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

func streamAppLogsIntoGbytes(logGuid string, outBuf *gbytes.Buffer) {
	ws, _, err := websocket.DefaultDialer.Dial(
		fmt.Sprintf("ws://%s/tail/?app=%s", loggregatorAddress, logGuid),
		http.Header{},
	)
	Expect(err).To(BeNil())

	for {
		_, data, err := ws.ReadMessage()
		if err != nil {
			return
		}

		receivedMessage := &logmessage.LogMessage{}
		err = proto.Unmarshal(data, receivedMessage)
		Expect(err).To(BeNil())

		outBuf.Write(receivedMessage.GetMessage())
	}

}

func desireLongRunningProcess(processGuid, appName, route string) error {
	return bbs.DesireLRP(models.DesiredLRP{
		Domain:      "whetstone",
		ProcessGuid: processGuid,
		Instances:   1,
		Stack:       "lucid64",
		RootFSPath:  "docker:///dajulia3/diego-edge-docker-app",
		Routes:      []string{route},
		MemoryMB:    128,
		DiskMB:      1024,
		Ports: []models.PortMapping{
			{ContainerPort: 8080},
		},
		Log: models.LogConfig{
			Guid:       processGuid,
			SourceName: "APP",
		},
		Actions: []models.ExecutorAction{
			models.Parallel(
				models.ExecutorAction{
					models.RunAction{
						Path: "/dockerapp",
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
