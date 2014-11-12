package whetstone_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/cloudfoundry-incubator/runtime-schema/models/factories"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/gorilla/websocket"
	"github.com/onsi/gomega/gbytes"

	"code.google.com/p/gogoprotobuf/proto"
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
)

const (
	repUrlRelativeToExecutor string = "http://127.0.0.1:20515"
	spyDownloadUrl           string = "http://file_server.service.dc1.consul:8080/v1/static/docker-circus/docker-circus.tgz"
)

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
			err := receptorClient.DeleteDesiredLRP(processGuid)
			Expect(err).To(BeNil())

			Eventually(errorCheckForRoute(route), timeout, 1).Should(HaveOccurred())
		})

		It("eventually runs on an executor", func() {
			err := desireLongRunningProcess(processGuid, route, 1)
			Expect(err).To(BeNil())

			Eventually(errorCheckForRoute(route), timeout, 1).ShouldNot(HaveOccurred())

			outBuf := gbytes.NewBuffer()
			go streamAppLogsIntoGbytes(processGuid, outBuf)
			Eventually(outBuf, 4).Should(gbytes.Say("Diego Edge Docker App. Says Hello"))

			err = desireLongRunningProcess(processGuid, route, 3)
			Expect(err).To(BeNil())

			instanceCountChan := make(chan int, numCpu)
			go countInstances(route, instanceCountChan)

			Eventually(instanceCountChan, timeout).Should(Receive(Equal(3)))
		})
	})

})

func errorCheckForRoute(route string) func() error {
	return func() error {
		resp, err := makeGetRequestToRoute(route)
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
	defer GinkgoRecover()

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

func countInstances(route string, instanceCountChan chan<- int) {
	defer GinkgoRecover()
	instanceIndexRoute := fmt.Sprintf("%s/instance-index", route)
	instancesSeen := make(map[int]bool)
	instanceIndexChan := make(chan int, numCpu)

	for i := 0; i < numCpu; i++ {
		go pollForInstanceIndices(instanceIndexRoute, instanceIndexChan)
	}

	for {
		instanceIndex := <-instanceIndexChan
		instancesSeen[instanceIndex] = true
		instanceCountChan <- len(instancesSeen)
	}
}

func pollForInstanceIndices(route string, instanceIndexChan chan<- int) {
	defer GinkgoRecover()
	for {
		resp, err := makeGetRequestToRoute(route)
		Expect(err).To(BeNil())

		responseBody, err := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		Expect(err).To(BeNil())

		instanceIndex, err := strconv.Atoi(string(responseBody))
		if err != nil {
			continue
		}
		instanceIndexChan <- instanceIndex
	}
}

func makeGetRequestToRoute(route string) (*http.Response, error) {
	routeWithScheme := fmt.Sprintf("http://%s", route)
	resp, err := http.DefaultClient.Get(routeWithScheme)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func desireLongRunningProcess(processGuid, route string, instanceCount int) error {
	return receptorClient.CreateDesiredLRP(receptor.DesiredLRPCreateRequest{
		ProcessGuid: processGuid,
		Domain:      "whetstone",
		RootFSPath:  "docker:///diegoedge/diego-edge-docker-app",
		Instances:   instanceCount,
		Stack:       "lucid64",
		Routes:      []string{route},
		MemoryMB:    128,
		DiskMB:      1024,
		Ports: []receptor.PortMapping{
			{ContainerPort: 8080},
		},
		Log: receptor.LogConfig{
			Guid:       processGuid,
			SourceName: "APP",
		},
		Actions: []models.ExecutorAction{
			{
				Action: models.DownloadAction{
					From:     spyDownloadUrl,
					To:       "/tmp",
					CacheKey: "",
				},
			},
			models.Parallel(
				models.ExecutorAction{
					models.RunAction{
						Path: "/dockerapp",
					},
				},
				models.ExecutorAction{
					models.MonitorAction{
						Action: models.ExecutorAction{
							models.RunAction{
								Path: "/tmp/spy",
								Args: []string{"-addr", ":8080"},
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
