package whetstone_test

import (
	"flag"
	"fmt"
	"os"
	"testing"

	Bbs "github.com/cloudfoundry-incubator/runtime-schema/bbs"
	"github.com/cloudfoundry/gunk/timeprovider"
	"github.com/cloudfoundry/storeadapter/etcdstoreadapter"
	"github.com/cloudfoundry/storeadapter/workerpool"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager/lagertest"
)

var (
	bbs         *Bbs.BBS
	etcdAddress string
	domain      string
)

const StackName = "lucid64"

func init() {
	flag.StringVar(&etcdAddress, "etcdAddress", "", "Address of the etcd cluster - REQUIRED")
	flag.StringVar(&domain, "domain", "", "Domain to use for deployed apps - REQUIRED")
}

func TestWhetstone(t *testing.T) {
	if etcdAddress == "" || domain == "" {
		fmt.Println(etcdAddress)
		fmt.Println(domain)
		fmt.Fprintf(os.Stderr, "To run this test suite, you must set the required flags.\nUsage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}
	RegisterFailHandler(Fail)
	RunSpecs(t, "Whetstone Suite")
}

var _ = BeforeEach(func() {
	adapter := etcdstoreadapter.NewETCDStoreAdapter([]string{etcdAddress}, workerpool.NewWorkerPool(20))

	err := adapter.Connect()
	Expect(err).ToNot(HaveOccurred())

	bbs = Bbs.NewBBS(adapter, timeprovider.NewTimeProvider(), lagertest.NewTestLogger("test"))
})
