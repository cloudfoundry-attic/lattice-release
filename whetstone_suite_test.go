package whetstone_test

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"testing"

	"github.com/cloudfoundry-incubator/receptor"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	receptorClient     receptor.Client
	domain             string
	loggregatorAddress string
	receptorAddress    string
	receptorUsername    string
	receptorPassword    string
	numCpu             int
	timeout            int
)

const StackName = "lucid64"

func init() {
	numCpu = runtime.NumCPU()
	runtime.GOMAXPROCS(numCpu)

	flag.StringVar(&domain, "domain", "", "Domain to use for deployed apps - REQUIRED")
	flag.StringVar(&loggregatorAddress, "loggregatorAddress", "", "Address of the loggregator traffic controller - REQUIRED")
	flag.StringVar(&receptorAddress, "receptorAddress", "", "Address of the diego receptor - REQUIRED")
	flag.StringVar(&receptorUsername, "receptorUsername", "", "Username for the receptor")
	flag.StringVar(&receptorPassword, "receptorPassword", "", "Password for the receptor")
	flag.IntVar(&timeout, "timeout", 30, "How long whetstone will wait for docker apps to start")
}

func TestWhetstone(t *testing.T) {
	if receptorAddress == "" || domain == "" || loggregatorAddress == "" {
		fmt.Fprintf(os.Stderr, "To run this test suite, you must set the required flags.\nUsage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}
	RegisterFailHandler(Fail)
	RunSpecs(t, "Whetstone Suite")
}

var _ = BeforeEach(func() {
	receptorClient = receptor.NewClient(receptorAddress, receptorUsername, receptorPassword)
})
