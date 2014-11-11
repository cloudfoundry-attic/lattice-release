package whetstone_test

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"testing"

	Bbs "github.com/cloudfoundry-incubator/runtime-schema/bbs"
	"github.com/cloudfoundry/gunk/timeprovider"
	"github.com/cloudfoundry/gunk/workpool"
	"github.com/cloudfoundry/storeadapter/etcdstoreadapter"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager/lagertest"
)

var (
	bbs                *Bbs.BBS
	etcdAddress        string
	domain             string
	loggregatorAddress string
	numCpu             int
	timeout            int
)

const StackName = "lucid64"

func init() {
	numCpu = runtime.NumCPU()
	runtime.GOMAXPROCS(numCpu)

	flag.StringVar(&etcdAddress, "etcdAddress", "", "Address of the etcd cluster - REQUIRED")
	flag.StringVar(&domain, "domain", "", "Domain to use for deployed apps - REQUIRED")
	flag.StringVar(&loggregatorAddress, "loggregatorAddress", "", "Address of the loggregator traffic controller - REQUIRED")
	flag.IntVar(&timeout, "timeout", 30, "How long whetstone will wait for docker apps to start")
}

func TestWhetstone(t *testing.T) {
	if etcdAddress == "" || domain == "" || loggregatorAddress == "" {
		fmt.Fprintf(os.Stderr, "To run this test suite, you must set the required flags.\nUsage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}
	RegisterFailHandler(Fail)
	RunSpecs(t, "Whetstone Suite")
}

var _ = BeforeEach(func() {
	etcdUrl := fmt.Sprintf("http://%s", etcdAddress)
	adapter := etcdstoreadapter.NewETCDStoreAdapter([]string{etcdUrl}, workpool.NewWorkPool(20))

	err := adapter.Connect()
	Expect(err).ToNot(HaveOccurred())

	bbs = Bbs.NewBBS(adapter, timeprovider.NewTimeProvider(), lagertest.NewTestLogger("test"))
})
