package ssh_test

import (
	"io/ioutil"
	"net"

	"github.com/cloudfoundry-incubator/lattice/ltc/ssh"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Listener", func() {
	var listener *ssh.ChannelListener

	BeforeEach(func() {
		listener = &ssh.ChannelListener{}
	})

	Describe("#Listen", func() {
		It("should accept connections", func() {
			address := freeAddress()
			acceptChan, _ := listener.Listen("tcp", address)

			writeConn, err := net.Dial("tcp", address)
			Expect(err).NotTo(HaveOccurred())

			readConn := <-acceptChan

			writeConn.Write([]byte("abcd"))
			writeConn.Close()
			Expect(ioutil.ReadAll(readConn)).To(Equal([]byte("abcd")))
		})

		Context("when we fail to listen", func() {
			It("should return an error on the returned error channel", func() {
				_, errChan := listener.Listen("bad network", "bad address")
				Expect(<-errChan).To(MatchError("listen bad network: unknown network bad network"))
			})
		})
	})
})

func freeAddress() string {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		Fail(err.Error())
	}
	defer listener.Close()
	return listener.Addr().String()
}
