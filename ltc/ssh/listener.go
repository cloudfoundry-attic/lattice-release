package ssh

import (
	"io"
	"net"
)

type ChannelListener struct{}

func (*ChannelListener) Listen(network, laddr string) (<-chan io.ReadWriteCloser, <-chan error) {
	connChan := make(chan io.ReadWriteCloser)
	errChan := make(chan error, 1)
	listener, err := net.Listen(network, laddr)
	if err != nil {
		errChan <- err
		return connChan, errChan
	}
	go func() {
		defer listener.Close()
		for {
			conn, err := listener.Accept()
			if err != nil {
				errChan <- err
				return
			}
			connChan <- conn
		}
	}()
	return connChan, errChan
}
