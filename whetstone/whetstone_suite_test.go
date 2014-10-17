package whetstone_test

import (
	"testing"

	Bbs "github.com/cloudfoundry-incubator/runtime-schema/bbs"
	"github.com/cloudfoundry/gunk/timeprovider"
	"github.com/cloudfoundry/storeadapter/etcdstoreadapter"
	"github.com/cloudfoundry/storeadapter/workerpool"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager/lagertest"
)

var bbs *Bbs.BBS

const StackName = "lucid64"

func TestWhetstone(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Whetstone Suite")
}

var _ = BeforeEach(func() {
	adapter := etcdstoreadapter.NewETCDStoreAdapter([]string{"http://10.244.16.2:4001"}, workerpool.NewWorkerPool(20))

	err := adapter.Connect()
	Expect(err).ToNot(HaveOccurred())

	bbs = Bbs.NewBBS(adapter, timeprovider.NewTimeProvider(), lagertest.NewTestLogger("test"))
})
