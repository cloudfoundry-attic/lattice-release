package emitter

import (
	"net"

	"github.com/cloudfoundry/sonde-go/control"
	"github.com/gogo/protobuf/proto"
)

type udpEmitter struct {
	udpAddr *net.UDPAddr
	udpConn net.PacketConn
}

func NewUdpEmitter(remoteAddr string) (*udpEmitter, error) {
	addr, err := net.ResolveUDPAddr("udp4", remoteAddr)
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenPacket("udp4", "")
	if err != nil {
		return nil, err
	}

	emitter := &udpEmitter{udpAddr: addr, udpConn: conn}
	return emitter, nil
}

func (e *udpEmitter) Emit(data []byte) error {
	_, err := e.udpConn.WriteTo(data, e.udpAddr)
	return err
}

func (e *udpEmitter) Close() {
	e.udpConn.Close()
}

func (e *udpEmitter) Address() net.Addr {
	return e.udpConn.LocalAddr()
}

func (e *udpEmitter) ListenForHeartbeatRequest(responder func(*control.ControlMessage)) error {
	buf := make([]byte, 1024)
	for {
		n, _, err := e.udpConn.ReadFrom(buf)
		if err != nil {
			return err
		}

		controlMessage := &control.ControlMessage{}
		err = proto.Unmarshal(buf[:n], controlMessage)
		if err != nil {
			return err
		}

		responder(controlMessage)
	}
}
